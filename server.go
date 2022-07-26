/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net"
    "sync"
    "time"

    encoder "github.com/vmihailenco/msgpack/v5"
)

type HandlerFunc =  func(*Context) error

type Service struct {
    handlers    map[string]HandlerFunc
    ctx         context.Context
    cancel      context.CancelFunc
    wg          *sync.WaitGroup
    preMw       []HandlerFunc
    postMw      []HandlerFunc
    keepalive   bool
    kaTime      time.Duration
    kaMtx       sync.Mutex
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

func (svc *Service) Handler(method string, handler HandlerFunc) {
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
        go svc.handleConn(conn, svc.wg)
    }
    return err
}

func notFound(context *Context) error {
    execErr := errors.New("method not found")
    err := context.SendError(execErr)
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

func (svc *Service) handleConn(conn *net.TCPConn, wg *sync.WaitGroup) {
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
    context := CreateContext(conn)

    remoteAddr := conn.RemoteAddr().String()
    remoteHost, _, _ := net.SplitHostPort(remoteAddr)
    context.remoteHost = remoteHost

    context.binReader = conn
    context.binWriter = io.Discard

    exitFunc := func() {
            conn.Close()
            wg.Done()
            if err != nil {
                logError("conn handler err:", err)
            }
    }
    defer exitFunc()

    recovFunc := func () {
        panicMsg := recover()
        if panicMsg != nil {
            logError("handler panic message:", panicMsg)
        }
    }
    defer recovFunc()

    err = context.ReadRequest()
    if err != nil {
        err = Err(err)
        return
    }

    err = context.BindMethod()
    if err != nil {
        err = Err(err)
        return
    }
    for _, mw := range svc.preMw {
        err = mw(context)
        if err != nil {
            err = Err(err)
            return
        }
    }
    err = svc.Route(context)
    if err != nil {
        err = Err(err)
        return
    }
    for _, mw := range svc.postMw {
        err = mw(context)
        if err != nil {
            err = Err(err)
            return
        }
    }
    return
}

func (svc *Service) Route(context *Context) error {
    handler, ok := svc.handlers[context.reqRPC.Method]
    if ok {
        return Err(handler(context))
    }
    return Err(notFound(context))
}

func (context *Context) ReadRequest() error {
    var err error

    context.reqPacket.header, err = ReadBytes(context.sockReader, headerSize)
    if err != nil {
        return Err(err)
    }
    context.reqHeader, err = UnpackHeader(context.reqPacket.header)
    if err != nil {
        return Err(err)
    }

    rpcSize := context.reqHeader.rpcSize
    context.reqPacket.rcpPayload, err = ReadBytes(context.sockReader, rpcSize)
    if err != nil {
        return Err(err)
    }
    return Err(err)
}

func (context *Context) BinWriter() io.Writer {
    return context.sockWriter
}

func (context *Context) BinReader() io.Reader {
    return context.sockReader
}

func (context *Context) BinSize() int64 {
    return context.reqHeader.binSize
}

func (context *Context) ReadBin(writer io.Writer) error {
    var err error
    _, err = CopyBytes(context.sockReader, writer, context.reqHeader.binSize)
    return Err(err)
}


func (context *Context) BindMethod() error {
    var err error
    err = encoder.Unmarshal(context.reqPacket.rcpPayload, context.reqRPC)
    return Err(err)
}

func (context *Context) BindParams(params any) error {
    var err error
    context.reqRPC.Params = params
    err = encoder.Unmarshal(context.reqPacket.rcpPayload, context.reqRPC)
    if err != nil {
        return Err(err)
    }
    return Err(err)
}

func (context *Context) SendResult(result any, binSize int64) error {
    var err error
    context.resRPC.Result = result

    context.resPacket.rcpPayload, err = context.resRPC.Pack()
    if err != nil {
        return Err(err)
    }
    context.resHeader.rpcSize = int64(len(context.resPacket.rcpPayload))
    context.resHeader.binSize = binSize

    context.resPacket.header, err = context.resHeader.Pack()
    if err != nil {
        return Err(err)
    }
    _, err = context.sockWriter.Write(context.resPacket.header)
    if err != nil {
        return Err(err)
    }
    _, err = context.sockWriter.Write(context.resPacket.rcpPayload)
    if err != nil {
        return Err(err)
    }
    return Err(err)
}


func (context *Context) SendError(execErr error) error {
    var err error

    context.resRPC.Error = execErr.Error()
    context.resRPC.Result = NewEmpty()

    context.resPacket.rcpPayload, err = context.resRPC.Pack()
    if err != nil {
        return Err(err)
    }
    context.resHeader.rpcSize = int64(len(context.resPacket.rcpPayload))
    context.resPacket.header, err = context.resHeader.Pack()
    if err != nil {
        return Err(err)
    }
    _, err = context.sockWriter.Write(context.resPacket.header)
    if err != nil {
        return Err(err)
    }
    _, err = context.sockWriter.Write(context.resPacket.rcpPayload)
    if err != nil {
        return Err(err)
    }
    return Err(err)
}
