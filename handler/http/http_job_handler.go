package http

import (
	"context"
	"encoding/json"
	"github.com/feixiaobo/go-xxl-job-client/v2/logger"
	"github.com/feixiaobo/go-xxl-job-client/v2/transport"
)

type HttpRequestHandler struct{}

func (h HttpRequestHandler) MethodName(ctx context.Context, r interface{}) string {
	req := r.(*transport.HttpRequestPkg)
	return req.MethodName
}

func (h HttpRequestHandler) ParseParam(ctx context.Context, r interface{}) (reqId, accessToken, methodName string, err error) {
	p := r.(*transport.HttpRequestPkg)
	headerMap := p.Header
	return headerMap["RequestId"], headerMap["XXL-JOB-ACCESS-TOKEN"], p.MethodName, nil
}

func (h HttpRequestHandler) Beat(ctx context.Context, r interface{}) error {
	return nil
}

func (h HttpRequestHandler) IdleBeat(ctx context.Context, r interface{}) (jobId int32, err error) {
	req := r.(*transport.HttpRequestPkg)
	job := &JobId{}
	err = json.Unmarshal(req.Body, job)
	if err != nil {
		return 0, err
	}
	return job.JobId, err
}

func (h HttpRequestHandler) Run(ctx context.Context, r interface{}) (triggerParam *transport.TriggerParam, err error) {
	req := r.(*transport.HttpRequestPkg)
	err = json.Unmarshal(req.Body, &triggerParam)
	return triggerParam, err
}

type JobId struct {
	JobId int32 `json:"jobId"`
}

func (h HttpRequestHandler) Kill(ctx context.Context, r interface{}) (jobId int32, err error) {
	req := r.(*transport.HttpRequestPkg)
	job := &JobId{}
	err = json.Unmarshal(req.Body, job)
	if err != nil {
		return 0, err
	}
	return job.JobId, err
}

func (h HttpRequestHandler) Log(ctx context.Context, r interface{}) (log *logger.LogResult, err error) {
	req := r.(*transport.HttpRequestPkg)
	lq := &transport.LogRequest{}
	err = json.Unmarshal(req.Body, lq)
	if err != nil {
		return nil, err
	}

	line, content := logger.ReadLog(lq.LogDateTim, lq.LogId, lq.FromLineNum)
	log = &logger.LogResult{
		FromLineNum: lq.FromLineNum,
		ToLineNum:   line,
		LogContent:  content,
		IsEnd:       true,
	}
	return log, err
}
