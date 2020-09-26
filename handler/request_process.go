package handler

import (
	"context"
	"encoding/json"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/feixiaobo/go-xxl-job-client/v2/admin"
	"github.com/feixiaobo/go-xxl-job-client/v2/transport"
	"net/http"
	"sync"
)

type RequestProcess struct {
	sync.RWMutex

	adminServer *admin.XxlAdminServer

	JobHandler *JobHandler

	ReqHandler RequestHandler
}

func NewRequestProcess(adminServer *admin.XxlAdminServer, handler RequestHandler) *RequestProcess {
	requestHandler := &RequestProcess{
		adminServer: adminServer,
		ReqHandler:  handler,
	}
	jobHandler := &JobHandler{
		QueueMap:     make(map[int32]*JobQueue),
		CallbackFunc: requestHandler.jobRunCallback,
	}
	requestHandler.JobHandler = jobHandler
	return requestHandler
}

func (j *RequestProcess) RegisterJob(jobName string, function JobHandlerFunc) {
	j.JobHandler.RegisterJob(jobName, function)
}

func (j *RequestProcess) pushJob(trigger *transport.TriggerParam) {
	returnt := transport.ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	err := j.JobHandler.PutJobToQueue(trigger)
	if err != nil {
		returnt.Code = http.StatusInternalServerError
		returnt.Content = err.Error()
		callback := &transport.HandleCallbackParam{
			LogId:         trigger.LogId,
			LogDateTim:    trigger.LogDateTime,
			ExecuteResult: returnt,
		}
		j.adminServer.CallbackAdmin([]*transport.HandleCallbackParam{callback})
	}
}

func (r *RequestProcess) jobRunCallback(trigger *JobRunParam, runErr error) {
	returnt := transport.ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	if runErr != nil {
		returnt.Code = http.StatusInternalServerError
		returnt.Content = runErr.Error()
	}

	callback := &transport.HandleCallbackParam{
		LogId:         trigger.LogId,
		LogDateTim:    trigger.LogDateTime,
		ExecuteResult: returnt,
	}
	r.adminServer.CallbackAdmin([]*transport.HandleCallbackParam{callback})
}

func (j *RequestProcess) RequestProcess(ctx context.Context, r interface{}) (res []byte, err error) {
	response := transport.XxlRpcResponse{}
	returnt := transport.ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	isOld := false
	reqId, accessToken, methodName, err := j.ReqHandler.ParseParam(ctx, r)
	if err != nil {
		returnt.Code = http.StatusInternalServerError
		returnt.Msg = err.Error()
	} else {
		isOld = reqId != ""
		isContinue := reqId == ""
		if reqId != "" && "BEAT_PING_PONG" != reqId { //老版本,处理非心跳请求
			response.RequestId = reqId
			isContinue = true
		}

		if isContinue {
			jt := j.adminServer.GetToken()
			if accessToken != jt {
				returnt.Code = http.StatusInternalServerError
				returnt.Msg = "access token error"
			} else {
				if methodName != "beat" {
					mn := j.ReqHandler.MethodName(ctx, r)
					switch mn {
					case "idleBeat":
						jobId, err := j.ReqHandler.IdleBeat(ctx, r)
						if err == nil {
							if j.JobHandler.HasRunning(jobId) {
								returnt.Content = http.StatusInternalServerError
								returnt.Content = "the server busy"
							}
						} else {
							returnt.Content = http.StatusInternalServerError
							returnt.Content = err.Error()
						}
					case "log":
						log, err := j.ReqHandler.Log(ctx, r)
						if err == nil {
							returnt.Content = log
						}
					case "kill":
						jobId, err := j.ReqHandler.Kill(ctx, r)
						if err == nil {
							j.JobHandler.cancelJob(jobId)
						}
					default:
						ta, err := j.ReqHandler.Run(ctx, r)
						if err == nil {
							go j.pushJob(ta)
						}
					}
				}
			}
		}
	}

	var bytes []byte
	if isOld {
		response.Result = returnt
		e := hessian.NewEncoder()
		err = e.Encode(response)
		if err != nil {
			return nil, err
		}
		bytes = e.Buffer()
	} else {
		bs, err := json.Marshal(&returnt)
		if err != nil {
			return nil, err
		}
		bytes = bs
	}
	return bytes, nil
}

func (j *RequestProcess) RemoveRegisterExecutor() {
	j.JobHandler.clearJob()
	j.adminServer.RemoveRegisterExecutor()
}

func (j *RequestProcess) RegisterExecutor() {
	j.adminServer.RegisterExecutor()
	go j.adminServer.AutoRegisterJobGroup()
}
