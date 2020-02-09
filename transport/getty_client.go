package transport

import (
	"github.com/dubbogo/getty"
	"log"
	"sync"
)

type GettyRPCClient struct {
	lock     sync.RWMutex
	sessions []getty.Session
}

func (c *GettyRPCClient) AddSession(session getty.Session) {
	if session == nil {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	if c.sessions == nil {
		c.sessions = make([]getty.Session, 0, 16)
	}
	c.sessions = append(c.sessions, session)
	c.lock.Unlock()
}

func (c *GettyRPCClient) RemoveSession(session getty.Session) {
	if session == nil {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	if c.sessions == nil {
		return
	}

	for i, s := range c.sessions {
		if s == session {
			c.sessions = append(c.sessions[:i], c.sessions[i+1:]...)
			log.Print("delete session{%s}, its index{%d}", session.Stat(), i)
			break
		}
	}
	log.Print("after remove session{%s}, left session number:%d", session.Stat(), len(c.sessions))
}
