package xxl

import (
	"context"
	"github.com/apache/dubbo-go-hessian2"
	"github.com/feixiaobo/go-xxl-job-client/admin"
	"github.com/feixiaobo/go-xxl-job-client/handler"
	"github.com/feixiaobo/go-xxl-job-client/logger"
	"github.com/feixiaobo/go-xxl-job-client/utils"
	"net/http"
	"reflect"
	"runtime"
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

func RequestHandler(buf []byte) (res []byte, err error) {
	out := hessian.NewDecoder(buf)
	r, err := out.Decode()

	if r == nil || err != nil {
		return nil, err
	}

	response := xxl.XxlRpcResponse{}
	returnt := xxl.ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	refMap := utils.RefletcStructToMap(r)
	if refMap == nil {
		returnt.Content = http.StatusInternalServerError
		returnt.Content = "reflect request body error"
	} else {
		reqId := refMap["RequestId"].(string)
		response.RequestId = reqId
		if "BEAT_PING_PONG" != reqId { //处理非心跳请求
			if xxl.XxlAdmin.AccessToken != "" && refMap["AccessToken"].(string) != xxl.XxlAdmin.AccessToken {
				returnt.Content = http.StatusInternalServerError
				returnt.Content = "access token error"
			} else {
				if refMap["MethodName"].(string) != "beat" {
					req := r.(*xxl.XxlRpcRequest)
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
							handler.CancelJob(req.Parameters[0].(int32))
						default:
							go pushJob(req.Parameters[0].(*xxl.TriggerParam))
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

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func pushJob(trigger *xxl.TriggerParam) {
	returnt := xxl.ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	fun, ok := handler.JobMap[trigger.ExecutorHandler]
	if ok {
		funName := getFunctionName(fun)
		jobParam := &handler.JobRunParam{
			LogId:       trigger.LogId,
			LogDateTime: trigger.LogDateTime,
			JobName:     trigger.ExecutorHandler,
			JobFuncName: funName,
			InputParam:  trigger.ExecutorParams,
		}
		if !handler.PushJob(trigger.JobId, jobParam) {
			returnt.Code = http.StatusInternalServerError
			returnt.Content = "push job queue failed"
		}
	} else {
		returnt.Code = http.StatusInternalServerError
		returnt.Content = "job handle not found"
	}

	callback := &xxl.HandleCallbackParam{
		LogId:         trigger.LogId,
		LogDateTim:    trigger.LogDateTime,
		ExecuteResult: returnt,
	}
	xxl.CallbackAdmin([]*xxl.HandleCallbackParam{callback})
}

func RegisterJob(name string, function handler.JobHandlerFunc) {
	handler.AddJob(name, function)
}
