package handler

import (
	"github.com/dubbogo/getty"
	"github.com/feixiaobo/go-xxl-job-client/transport"
	"log"
	"net/http"
	"time"
)

const WritePkg_Timeout = 5 * time.Second

var gettyClient transport.GettyRPCClient

type MessageHandler struct {
	SessionOnOpen func(session getty.Session)
}

func (h *MessageHandler) OnOpen(session getty.Session) error {
	log.Print("OnOpen session{%s} open", session.Stat())
	if h.SessionOnOpen != nil {
		h.SessionOnOpen(session)
		gettyClient.AddSession(session)
	}
	return nil
}

func (h *MessageHandler) OnError(session getty.Session, err error) {
	log.Print("OnError session{%s} got error{%v}, will be closed.", session.Stat(), err)
}

func (h *MessageHandler) OnClose(session getty.Session) {
	log.Print("OnClose session{%s} is closing......", session.Stat())
	gettyClient.RemoveSession(session)
}

func (h *MessageHandler) OnMessage(session getty.Session, pkg interface{}) {
	s, ok := pkg.([]interface{})
	if !ok {
		log.Print("illegal package{%#v}", pkg)
		return
	}

	for _, v := range s {
		if v != nil {
			res, err := RequestHandler(v)
			reply(session, res, err)
		}
	}
}

func (h *MessageHandler) OnCron(session getty.Session) {
	active := session.GetActive()
	if 20e9 < time.Since(active).Nanoseconds() {
		log.Print("OnCorn session{%s} timeout{%s}", session.Stat(), time.Since(active).String())
		session.Close()
		gettyClient.RemoveSession(session)
	}
}

func reply(session getty.Session, resBy []byte, err error) {
	pkg := transport.NewHttpResponsePkg(http.StatusOK, resBy)
	if err != nil || resBy == nil {
		pkg = transport.NewHttpResponsePkg(http.StatusInternalServerError, resBy)
	}

	if err := session.WritePkg(pkg, WritePkg_Timeout); err != nil {
		log.Print("WritePkg error: %#v, %#v", pkg, err)
	}
}
