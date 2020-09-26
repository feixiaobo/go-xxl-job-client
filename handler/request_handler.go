package handler

import (
	"context"
	"github.com/feixiaobo/go-xxl-job-client/v2/logger"
	"github.com/feixiaobo/go-xxl-job-client/v2/transport"
)

type RequestHandler interface {
	MethodName(ctx context.Context, r interface{}) string
	ParseParam(ctx context.Context, r interface{}) (reqId, accessToken, methodName string, err error)
	Beat(ctx context.Context, r interface{}) error
	IdleBeat(ctx context.Context, r interface{}) (jobId int32, err error)
	Run(ctx context.Context, r interface{}) (triggerParam *transport.TriggerParam, err error)
	Kill(ctx context.Context, r interface{}) (jobId int32, err error)
	Log(ctx context.Context, r interface{}) (log *logger.LogResult, err error)
}
