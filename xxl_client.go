package xxl

import (
	"context"
	"github.com/apache/dubbo-go-hessian2"
	"github.com/dubbogo/getty"
	"github.com/feixiaobo/go-xxl-job-client/v2/admin"
	"github.com/feixiaobo/go-xxl-job-client/v2/constants"
	executor2 "github.com/feixiaobo/go-xxl-job-client/v2/executor"
	"github.com/feixiaobo/go-xxl-job-client/v2/handler"
	"github.com/feixiaobo/go-xxl-job-client/v2/handler/http"
	"github.com/feixiaobo/go-xxl-job-client/v2/handler/rpc"
	"github.com/feixiaobo/go-xxl-job-client/v2/logger"
	"github.com/feixiaobo/go-xxl-job-client/v2/option"
	"github.com/feixiaobo/go-xxl-job-client/v2/transport"
)

type XxlClient struct {
	executor       *executor2.Executor
	requestHandler *handler.RequestProcess
}

func NewXxlClient(opts ...option.Option) *XxlClient {
	clientOps := option.NewClientOptions(opts...)

	executor := executor2.NewExecutor(
		"",
		clientOps.AppName,
		clientOps.Port,
	)

	adminServer := admin.NewAdminServer(
		clientOps.AdminAddr,
		clientOps.Timeout,
		clientOps.BeatTime,
		executor,
	)

	var requestHandler *handler.RequestProcess
	var gettyClient *executor2.GettyClient
	if clientOps.EnableHttp {
		adminServer.AccessToken = map[string]string{
			"XXL-JOB-ACCESS-TOKEN": clientOps.AccessToken,
		}
		requestHandler = handler.NewRequestProcess(adminServer, &http.HttpRequestHandler{})
		executor.Protocol = constants.HttpProtocol
		gettyClient = &executor2.GettyClient{
			PkgHandler:    http.NewHttpPackageHandler(),
			EventListener: http.NewHttpMessageHandler(&transport.GettyRPCClient{}, requestHandler.RequestProcess),
		}
	} else {
		//register java POJO
		hessian.RegisterPOJO(&transport.XxlRpcRequest{})
		hessian.RegisterPOJO(&transport.TriggerParam{})
		hessian.RegisterPOJO(&transport.Beat{})
		hessian.RegisterPOJO(&transport.XxlRpcResponse{})
		hessian.RegisterPOJO(&transport.ReturnT{})
		hessian.RegisterPOJO(&transport.HandleCallbackParam{})
		hessian.RegisterPOJO(&logger.LogResult{})
		hessian.RegisterPOJO(&transport.RegistryParam{})

		adminServer.AccessToken = map[string]string{
			"XXL-RPC-ACCESS-TOKEN": clientOps.AccessToken,
		}
		requestHandler = handler.NewRequestProcess(adminServer, &rpc.RpcRequestHandler{})
		gettyClient = &executor2.GettyClient{
			PkgHandler:    rpc.NewPackageHandler(),
			EventListener: rpc.NewRpcMessageHandler(&transport.GettyRPCClient{}, requestHandler.RequestProcess),
		}
	}
	executor.SetClient(gettyClient)

	return &XxlClient{
		requestHandler: requestHandler,
		executor:       executor,
	}
}

func (c *XxlClient) ExitApplication() {
	c.requestHandler.RemoveRegisterExecutor()
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

func GetSharding(ctx context.Context) (shardingIdx, shardingTotal int32) {
	jobMap := ctx.Value("jobParam")
	if jobMap != nil {
		shardingParam, ok := jobMap.(map[string]map[string]interface{})["sharding"]
		if ok {
			idx, vok := shardingParam["shardingIdx"]
			if vok {
				shardingIdx = idx.(int32)
			}
			total, ok := shardingParam["shardingTotal"]
			if ok {
				shardingTotal = total.(int32)
			}
		}
	}
	return shardingIdx, shardingTotal
}

func (c *XxlClient) Run() {
	c.requestHandler.RegisterExecutor()
	go logger.InitLogPath()
	c.executor.Run(c.requestHandler.JobHandler.BeanJobLength())
}

func (c *XxlClient) RegisterJob(jobName string, function handler.JobHandlerFunc) {
	c.requestHandler.RegisterJob(jobName, function)
}

func (c *XxlClient) SetLogger(logger getty.Logger) {
	getty.SetLogger(logger)
}
