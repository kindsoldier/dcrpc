/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
        "context"
        "crypto/tls"
        "errors"
        "fmt"
        "io"
        "net"
        "sync"
        "time"

        encoder "encoding/json"
)

type HandlerFunc = func(*Content) error

type Service struct {
        handlers  map[string]HandlerFunc
        ctx       context.Context
        cancel    context.CancelFunc
        wg        *sync.WaitGroup
        preMw     []HandlerFunc
        postMw    []HandlerFunc
        keepalive bool
        kaTime    time.Duration
        kaMtx     sync.Mutex
}

func NewService() *Service {
        rdrpc := &Service{}
        rdrpc.handlers = make(map[string]HandlerFunc)
        ctx, cancel := context.WithCancel(context.Background())
        rdrpc.ctx = ctx
        rdrpc.cancel = cancel
        var wg sync.WaitGroup
        rdrpc.wg = &wg
        rdrpc.preMw = make([]HandlerFunc, 0)
        rdrpc.postMw = make([]HandlerFunc, 0)

        return rdrpc
}

func (svc *Service) PreMiddleware(mw HandlerFunc) {
        svc.preMw = append(svc.preMw, mw)
}

func (svc *Service) PostMiddleware(mw HandlerFunc) {
        svc.postMw = append(svc.postMw, mw)
}

func (svc *Service) Handle(method string, handler HandlerFunc) {
        svc.handlers[method] = handler
}

func (svc *Service) SetKeepAlive(flag bool) {
        svc.kaMtx.Lock()
        defer svc.kaMtx.Unlock()
        svc.keepalive = true
}

func (svc *Service) SetKeepAlivePeriod(interval time.Duration) {
        svc.kaMtx.Lock()
        defer svc.kaMtx.Unlock()
        svc.kaTime = interval
}

func (svc *Service) Listen(address string) error {
        var err error
        logInfo("server listen:", address)

        addr, err := net.ResolveTCPAddr("tcp", address)
        if err != nil {
                err = fmt.Errorf("unable to resolve adddress: %s", err)
                return err
        }
        listener, err := net.ListenTCP("tcp", addr)
        if err != nil {
                err = fmt.Errorf("unable to start listener: %s", err)
                return err
        }

        for {
                conn, err := listener.AcceptTCP()
                if err != nil {
                        logError("conn accept err:", err)
                }
                select {
                case <-svc.ctx.Done():
                        return err
                default:
                }
                svc.wg.Add(1)
                go svc.handleTCPConn(conn, svc.wg)
        }
        return err
}

func (svc *Service) ListenTLS(address string, tlsConfig *tls.Config) error {
        var err error
        logInfo("server listen:", address)

        listener, err := tls.Listen("tcp", address, tlsConfig)
        if err != nil {
                err = fmt.Errorf("unable to start listener: %s", err)
                return err
        }

        for {
                conn, err := listener.Accept()
                if err != nil {
                        logError("conn accept err:", err)
                }
                select {
                case <-svc.ctx.Done():
                        return err
                default:
                }
                svc.wg.Add(1)
                go svc.handleConn(conn, svc.wg)
        }
        return err
}

func notFound(content *Content) error {
        execErr := errors.New("method not found")
        err := content.SendError(execErr)
        return err
}

func (svc *Service) Stop() error {
        var err error
        // Disable new connection
        logInfo("cancel rpc accept loop")
        svc.cancel()
        // Wait handlers
        logInfo("wait rpc handlers")
        svc.wg.Wait()
        return err
}

func (svc *Service) handleTCPConn(conn *net.TCPConn, wg *sync.WaitGroup) {
        var err error
        if svc.keepalive {
                err = conn.SetKeepAlive(true)
                if err != nil {
                        err = fmt.Errorf("unable to set keepalive: %s", err)
                        return
                }
                if svc.kaTime > 0 {
                        err = conn.SetKeepAlivePeriod(svc.kaTime)
                        if err != nil {
                                err = fmt.Errorf("unable to set keepalive period: %s", err)
                                return
                        }
                }
        }
        svc.handleConn(conn, wg)
}

func (svc *Service) handleConn(conn net.Conn, wg *sync.WaitGroup) {
        var err error

        content := CreateContent(conn)

        remoteAddr := conn.RemoteAddr().String()
        remoteHost, _, _ := net.SplitHostPort(remoteAddr)
        content.remoteHost = remoteHost

        content.binReader = conn
        content.binWriter = io.Discard

        exitFunc := func() {
                conn.Close()
                wg.Done()
                if err != nil {
                        logError("conn handler err:", err)
                }
        }
        defer exitFunc()

        recovFunc := func() {
                panicMsg := recover()
                if panicMsg != nil {
                        logError("handler panic message:", panicMsg)
                }
        }
        defer recovFunc()

        err = content.ReadRequest()
        if err != nil {
                err = err
                return
        }

        err = content.BindMethod()
        if err != nil {
                err = err
                return
        }
        for _, mw := range svc.preMw {
                err = mw(content)
                if err != nil {
                        err = err
                        return
                }
        }
        err = svc.Route(content)
        if err != nil {
                err = err
                return
        }
        for _, mw := range svc.postMw {
                err = mw(content)
                if err != nil {
                        err = err
                        return
                }
        }
        return
}

func (svc *Service) Route(content *Content) error {
        handler, ok := svc.handlers[content.reqBlock.Method]
        if ok {
                return handler(content)
        }
        return notFound(content)
}

func (content *Content) ReadRequest() error {
        var err error

        content.reqPacket.header, err = ReadBytes(content.sockReader, headerSize)
        if err != nil {
                return err
        }
        content.reqHeader, err = UnpackHeader(content.reqPacket.header)
        if err != nil {
                return err
        }

        rpcSize := content.reqHeader.rpcSize
        content.reqPacket.rcpPayload, err = ReadBytes(content.sockReader, rpcSize)
        if err != nil {
                return err
        }
        return err
}

func (content *Content) BinWriter() io.Writer {
        return content.sockWriter
}

func (content *Content) BinReader() io.Reader {
        return content.sockReader
}

func (content *Content) BinSize() int64 {
        return content.reqHeader.binSize
}

func (content *Content) ReadBin(ctx context.Context, writer io.Writer) error {
        var err error
        _, err = CopyBytes(ctx, content.sockReader, writer, content.reqHeader.binSize)
        return err
}

func (content *Content) BindMethod() error {
        var err error
        err = encoder.Unmarshal(content.reqPacket.rcpPayload, content.reqBlock)
        return err
}

func (content *Content) BindParams(params any) error {
        var err error
        content.reqBlock.Params = params
        err = encoder.Unmarshal(content.reqPacket.rcpPayload, content.reqBlock)
        if err != nil {
                return err
        }
        return err
}

func (content *Content) SendResult(result any, binSize int64) error {
        var err error
        content.resBlock.Result = result

        content.resPacket.rcpPayload, err = content.resBlock.Pack()
        if err != nil {
                return err
        }
        content.resHeader.rpcSize = int64(len(content.resPacket.rcpPayload))
        content.resHeader.binSize = binSize

        content.resPacket.header, err = content.resHeader.Pack()
        if err != nil {
                return err
        }
        _, err = content.sockWriter.Write(content.resPacket.header)
        if err != nil {
                return err
        }
        _, err = content.sockWriter.Write(content.resPacket.rcpPayload)
        if err != nil {
                return err
        }
        return err
}

func (content *Content) SendError(execErr error) error {
        var err error

        content.resBlock.Error = execErr.Error()
        content.resBlock.Result = NewEmptyResult()

        content.resPacket.rcpPayload, err = content.resBlock.Pack()
        if err != nil {
                return err
        }
        content.resHeader.rpcSize = int64(len(content.resPacket.rcpPayload))
        content.resPacket.header, err = content.resHeader.Pack()
        if err != nil {
                return err
        }
        _, err = content.sockWriter.Write(content.resPacket.header)
        if err != nil {
                return err
        }
        _, err = content.sockWriter.Write(content.resPacket.rcpPayload)
        if err != nil {
                return err
        }
        return err
}
