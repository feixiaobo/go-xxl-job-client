package executor

import (
	"fmt"
	"github.com/dubbogo/getty"
	"github.com/dubbogo/getty/demo/util"
	"github.com/dubbogo/gost/sync"
	"net"
	"strconv"
	"time"
)

const (
	cronPeriod      = 20e9 / 1e6
	queueLen        = 4
	queueNum        = 4
	queuePool       = 16
	maxMsgLen       = 102400
	wqLen           = 512
	keepAliveTime   = 3 * time.Minute
	writeTimeout    = 5 * time.Second
	ReadBufferSize  = 262144
	writeBufferSize = 65536
)

type GettyClient struct {
	PkgHandler getty.ReadWriter

	EventListener getty.EventListener
}

func newGettyClient(pkgHandler getty.ReadWriter, eventListener getty.EventListener) *GettyClient {
	return &GettyClient{
		PkgHandler:    pkgHandler,
		EventListener: eventListener,
	}
}

func (c *GettyClient) Run(port, taskSize int) {
	portStr := ":" + strconv.Itoa(port)
	server := getty.NewTCPServer(
		getty.WithLocalAddress(portStr),
	)

	server.RunEventLoop(func(session getty.Session) error {
		taskPool := gxsync.NewTaskPool(
			gxsync.WithTaskPoolTaskQueueLength(taskSize*queueLen),
			gxsync.WithTaskPoolTaskQueueNumber(taskSize+queueNum),
			gxsync.WithTaskPoolTaskPoolSize(taskSize*queuePool),
		)

		err := c.initialSession(session)
		if err != nil {
			return err
		}
		session.SetTaskPool(taskPool)
		return err
	})
	util.WaitCloseSignals(server)
}

func (c *GettyClient) initialSession(session getty.Session) (err error) {
	tcpConn, ok := session.Conn().(*net.TCPConn)
	if !ok {
		panic(fmt.Sprintf("newSession: %s, session.conn{%#v} is not tcp connection", session.Stat(), session.Conn()))
	}

	if err = tcpConn.SetNoDelay(true); err != nil {
		return err
	}
	if err = tcpConn.SetKeepAlive(true); err != nil {
		return err
	}
	if err = tcpConn.SetKeepAlivePeriod(keepAliveTime); err != nil {
		return err
	}
	if err = tcpConn.SetReadBuffer(ReadBufferSize); err != nil {
		return err
	}
	if err = tcpConn.SetWriteBuffer(writeBufferSize); err != nil { //考虑查看日志时候返回数据可能会多，会不会太小？
		return err
	}

	session.SetName("tcp")
	session.SetMaxMsgLen(maxMsgLen)
	session.SetWQLen(wqLen)
	session.SetReadTimeout(time.Second)
	session.SetWriteTimeout(writeTimeout)
	session.SetCronPeriod(int(cronPeriod))
	session.SetWaitTime(time.Second)
	session.SetPkgHandler(c.PkgHandler)
	session.SetEventListener(c.EventListener)
	return err
}
