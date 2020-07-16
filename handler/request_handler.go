package handler

import (
	"github.com/apache/dubbo-go-hessian2"
	"github.com/feixiaobo/go-xxl-job-client/v2/admin"
	"github.com/feixiaobo/go-xxl-job-client/v2/logger"
	"github.com/feixiaobo/go-xxl-job-client/v2/option"
	"github.com/feixiaobo/go-xxl-job-client/v2/transport"
	"github.com/feixiaobo/go-xxl-job-client/v2/utils"
	"net/http"
	"sync"
)

type RequestHandler struct {
	sync.RWMutex

	AdminServer *admin.XxlAdminServer

	JobHandler *JobHandler
}

func NewRequestHandler(options option.ClientOptions) *RequestHandler {
	requestHandler := &RequestHandler{
		AdminServer: admin.NewAdminServer(options.AdminAddr, options.AccessToken,
			options.Timeout, options.BeatTime),
	}
	jobHandler := &JobHandler{
		QueueMap:     make(map[int32]*JobQueue),
		CallbackFunc: requestHandler.jobRunCallback,
	}
	requestHandler.JobHandler = jobHandler
	return requestHandler
}

func (j *RequestHandler) RegisterJob(jobName string, function JobHandlerFunc) {
	j.JobHandler.RegisterJob(jobName, function)
}

func (j *RequestHandler) pushJob(trigger *transport.TriggerParam) {
	returnt := transport.ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	err := j.JobHandler.PutJobToQueue(trigger)
	if err != nil {
		returnt.Code = http.StatusInternalServerError
		returnt.Content = err.Error()
	}

	callback := &transport.HandleCallbackParam{
		LogId:         trigger.LogId,
		LogDateTim:    trigger.LogDateTime,
		ExecuteResult: returnt,
	}
	j.AdminServer.CallbackAdmin([]*transport.HandleCallbackParam{callback})
}

func (r *RequestHandler) jobRunCallback(trigger *JobRunParam, runErr error) {
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
	r.AdminServer.CallbackAdmin([]*transport.HandleCallbackParam{callback})
}

func (j *RequestHandler) RequestHandler(r interface{}) (res []byte, err error) {
	response := transport.XxlRpcResponse{}
	returnt := transport.ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	refMap := utils.ReflectStructToMap(r)
	if refMap == nil {
		returnt.Content = http.StatusInternalServerError
		returnt.Content = "reflect request body error"
	} else {
		reqId := refMap["RequestId"].(string)
		response.RequestId = reqId
		if "BEAT_PING_PONG" != reqId { //处理非心跳请求
			if j.AdminServer.AccessToken != "" &&
				refMap["AccessToken"].(string) != j.AdminServer.AccessToken {
				returnt.Content = http.StatusInternalServerError
				returnt.Content = "access token error"
			} else {
				if refMap["MethodName"].(string) != "beat" {
					req := r.(*transport.XxlRpcRequest)
					response.RequestId = req.RequestId
					if len(req.Parameters) == 0 {
						response.ErrorMsg = "job parameters is empty"
						returnt.Code = http.StatusInternalServerError
						returnt.Content = "job parameters is empty"
					} else {
						switch req.MethodName {
						case "log":
							fromLine := req.Parameters[2].(int32)
							line, content := logger.ReadLog(req.Parameters[0].(int64), req.Parameters[1].(int64), fromLine)
							log := logger.LogResult{
								FromLineNum: fromLine,
								ToLineNum:   line,
								LogContent:  content,
								IsEnd:       true,
							}
							returnt.Content = log
						case "kill":
							j.JobHandler.cancelJob(req.Parameters[0].(int32))
						default:
							go j.pushJob(req.Parameters[0].(*transport.TriggerParam))
						}
					}
				}
			}
		}
	}

	response.Result = returnt
	e := hessian.NewEncoder()
	err = e.Encode(response)
	if err != nil {
		return nil, err
	}

	bytes := e.Buffer()
	return bytes, nil
}

func (j *RequestHandler) RemoveRegisterExecutor() {
	j.JobHandler.clearJob()
	j.AdminServer.RemoveRegisterExecutor()
}

func (j *RequestHandler) RegisterExecutor(appName string, port int) {
	j.AdminServer.RegisterExecutor(appName, port)
	go j.AdminServer.AutoRegisterJobGroup(port)
}
