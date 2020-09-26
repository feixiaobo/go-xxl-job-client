package rpc

import (
	"context"
	"errors"
	"github.com/feixiaobo/go-xxl-job-client/v2/logger"
	"github.com/feixiaobo/go-xxl-job-client/v2/transport"
	"github.com/feixiaobo/go-xxl-job-client/v2/utils"
)

type RpcRequestHandler struct {
}

func (h *RpcRequestHandler) MethodName(ctx context.Context, r interface{}) string {
	req := r.(*transport.XxlRpcRequest)
	return req.MethodName
}

func (h *RpcRequestHandler) ParseParam(ctx context.Context, r interface{}) (reqId, accessToken, methodName string, err error) {
	refMap := utils.ReflectStructToMap(r)
	if refMap == nil {
		return "", "", "", errors.New("reflect request body error")
	}
	return refMap["RequestId"].(string), refMap["AccessToken"].(string), refMap["MethodName"].(string), nil
}

func (h *RpcRequestHandler) Beat(ctx context.Context, r interface{}) (err error) {
	return nil
}

func (h *RpcRequestHandler) IdleBeat(ctx context.Context, r interface{}) (jobId int32, err error) {
	req := r.(*transport.XxlRpcRequest)
	if len(req.Parameters) == 0 {
		return 0, errors.New("job parameters is empty")
	}
	return req.Parameters[0].(int32), err
}

func (h *RpcRequestHandler) Run(ctx context.Context, r interface{}) (triggerParam *transport.TriggerParam, err error) {
	req := r.(*transport.XxlRpcRequest)
	if len(req.Parameters) == 0 {
		return nil, errors.New("job parameters is empty")
	}
	return req.Parameters[0].(*transport.TriggerParam), err
}

func (h *RpcRequestHandler) Kill(ctx context.Context, r interface{}) (jobId int32, err error) {
	req := r.(*transport.XxlRpcRequest)
	if len(req.Parameters) == 0 {
		return 0, errors.New("job parameters is empty")
	}
	return req.Parameters[0].(int32), err
}

func (h *RpcRequestHandler) Log(ctx context.Context, r interface{}) (log *logger.LogResult, err error) {
	req := r.(*transport.XxlRpcRequest)
	if len(req.Parameters) == 0 {
		return nil, errors.New("job parameters is empty")
	}
	fromLine := req.Parameters[2].(int32)
	line, content := logger.ReadLog(req.Parameters[0].(int64), req.Parameters[1].(int64), fromLine)
	log = &logger.LogResult{
		FromLineNum: fromLine,
		ToLineNum:   line,
		LogContent:  content,
		IsEnd:       true,
	}
	return log, err
}
