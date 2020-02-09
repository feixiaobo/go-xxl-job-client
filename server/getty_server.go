package server

import (
	"github.com/dubbogo/getty"
	"github.com/dubbogo/getty/demo/util"
	"github.com/dubbogo/gost/sync"
	"github.com/feixiaobo/go-xxl-job-client/admin"
	"github.com/feixiaobo/go-xxl-job-client/handler"
	"strconv"
)

var (
	taskPool *gxsync.TaskPool
)

func StartServer() {
	jobLen := len(handler.JobMap)
	taskPool = gxsync.NewTaskPool(
		gxsync.WithTaskPoolTaskQueueLength(jobLen*5),
		gxsync.WithTaskPoolTaskQueueNumber(jobLen+4),
		gxsync.WithTaskPoolTaskPoolSize(jobLen*16),
	)

	port := ":" + strconv.Itoa(xxl.XxlAdmin.Port)
	options := []getty.ServerOption{getty.WithLocalAddress(port)}
	server := getty.NewTCPServer(
		options...,
	)
	server.RunEventLoop(newServerSession)
	util.WaitCloseSignals(server)
}

func newServerSession(session getty.Session) (err error) {
	err = handler.InitialSession(session)
	if err != nil {
		return
	}
	session.SetTaskPool(taskPool)
	return
}
