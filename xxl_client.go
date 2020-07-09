package xxl

import (
	"context"
	"github.com/apache/dubbo-go-hessian2"
	"github.com/dubbogo/getty"
	"github.com/feixiaobo/go-xxl-job-client/handler"
	"github.com/feixiaobo/go-xxl-job-client/logger"
	"github.com/feixiaobo/go-xxl-job-client/option"
	"github.com/feixiaobo/go-xxl-job-client/transport"
)

type XxlClient struct {
	executor *executor

	gettyClient *GettyClient

	requestHandler *handler.RequestHandler
}

type executor struct {
	appName string
	port    int
}

func NewXxlClient(opts ...option.Option) *XxlClient {
	hessian.RegisterPOJO(&transport.XxlRpcRequest{})
	hessian.RegisterPOJO(&transport.TriggerParam{})
	hessian.RegisterPOJO(&transport.Beat{})
	hessian.RegisterPOJO(&transport.XxlRpcResponse{})
	hessian.RegisterPOJO(&transport.ReturnT{})
	hessian.RegisterPOJO(&transport.HandleCallbackParam{})
	hessian.RegisterPOJO(&logger.LogResult{})
	hessian.RegisterPOJO(&transport.RegistryParam{})

	clientOps := option.NewClientOptions(opts...)
	requestHandler := handler.NewRequestHandler(clientOps)
	gettyClient := &GettyClient{
		PkgHandler: handler.NewPackageHandler(),
		EventListener: &handler.MessageHandler{
			GettyClient: &transport.GettyRPCClient{},
			MsgHandle:   requestHandler.RequestHandler,
		},
	}

	return &XxlClient{
		requestHandler: requestHandler,
		executor: &executor{
			appName: clientOps.AppName,
			port:    clientOps.Port,
		},
		gettyClient: gettyClient,
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
	c.requestHandler.RegisterExecutor(c.executor.appName, c.executor.port)
	go logger.InitLogPath()
	c.gettyClient.Run(c.executor.port, c.requestHandler.JobHandler.BeanJobLength())
}

func (c *XxlClient) RegisterJob(jobName string, function handler.JobHandlerFunc) {
	c.requestHandler.JobHandler.RegisterJob(jobName, function)
}

func (c *XxlClient) SetLogger(logger getty.Logger) {
	getty.SetLogger(logger)
}
