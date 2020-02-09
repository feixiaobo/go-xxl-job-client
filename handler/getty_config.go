package handler

import (
	"fmt"
	"github.com/dubbogo/getty"
	"net"
	"time"
)

var (
	pkgHandler    = &PackageHandler{}
	eventListener = &MessageHandler{}
	CronPeriod    = 20e9
)

func InitialSession(session getty.Session) (err error) {
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
	if err = tcpConn.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		return err
	}
	if err = tcpConn.SetReadBuffer(262144); err != nil {
		return err
	}
	if err = tcpConn.SetWriteBuffer(65536); err != nil { //考虑查看日志时候返回数据可能会多，会不会太小？
		return err
	}

	session.SetName("tcp")
	session.SetMaxMsgLen(102400)
	session.SetWQLen(512)
	session.SetReadTimeout(time.Second)
	session.SetWriteTimeout(5 * time.Second)
	session.SetCronPeriod(int(CronPeriod / 1e6))
	session.SetWaitTime(time.Second)
	session.SetPkgHandler(pkgHandler)
	session.SetEventListener(eventListener)
	return err
}
