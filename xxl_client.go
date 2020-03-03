package xxl

import (
	"context"
	"github.com/apache/dubbo-go-hessian2"
	"github.com/feixiaobo/go-xxl-job-client/admin"
	"github.com/feixiaobo/go-xxl-job-client/handler"
	"github.com/feixiaobo/go-xxl-job-client/logger"
	"github.com/feixiaobo/go-xxl-job-client/server"
	"time"
)

func InitExecutor(addresses []string, accessToken, appName string, port int) {
	hessian.RegisterPOJO(&xxl.XxlRpcRequest{})
	hessian.RegisterPOJO(&xxl.TriggerParam{})
	hessian.RegisterPOJO(&xxl.Beat{})
	hessian.RegisterPOJO(&xxl.XxlRpcResponse{})
	hessian.RegisterPOJO(&xxl.ReturnT{})
	hessian.RegisterPOJO(&xxl.HandleCallbackParam{})
	hessian.RegisterPOJO(&logger.LogResult{})
	hessian.RegisterPOJO(&xxl.RegistryParam{})
	xxl.RegisterExecutor(addresses, accessToken, appName, port, 5*time.Second)
	go logger.InitLogPath()
	go xxl.AutoRegisterJobGroup()
}

func ExitApplication() {
	xxl.RemoveRegisterExecutor()
	handler.ClearJob()
}

func GetParam(ctx context.Context, key string) (val string, has bool) {
	jobMap := ctx.Value("jobParam")
	if jobMap != nil {
		inputParam, ok := jobMap.(map[string]map[string]interface{})["inputParam"]
		if ok {
			val, vok := inputParam[key]
			if vok {
				return val.(string), true
			}
		}
	}
	return "", false
}

func RunServer() {
	if xxl.XxlAdmin.Port == 0 {
		panic("executor must be init before run server")
	}
	server.StartServer()
}

func RegisterJob(name string, function handler.JobHandlerFunc) {
	handler.AddJob(name, function)
}
