package xxl

import (
	"context"
	"github.com/apache/dubbo-go-hessian2"
	"github.com/feixiaobo/go-xxl-job-client/logger"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type JobHandlerFunc func(ctx context.Context) error

var JobMap map[string]JobHandlerFunc

type XxlRpcRequest struct {
	RequestId        string
	CreateMillisTime int64
	AccessToken      string
	ClassName        string
	MethodName       string
	ParameterTypes   []hessian.Object
	Parameters       []hessian.Object
	Version          string
}

func (XxlRpcRequest) JavaClassName() string {
	return "com.xxl.rpc.remoting.net.params.XxlRpcRequest"
}

type TriggerParam struct {
	JobId                 int32
	ExecutorHandler       string
	ExecutorParams        string
	ExecutorBlockStrategy string
	ExecutorTimeout       int32
	LogId                 int64
	LogDateTime           int64
	GlueType              string
	GlueSource            string
	GlueUpdatetime        int64
	BroadcastIndex        int32
	BroadcastTotal        int32
}

func (TriggerParam) JavaClassName() string {
	return "com.xxl.job.core.biz.model.TriggerParam"
}

type Beat struct {
	RequestId        string
	CreateMillisTime int64
	AccessToken      string
	ClassName        string
	MethodName       string
	ParameterTypes   []hessian.Object
	Parameters       []hessian.Object
	Version          string
}

func (Beat) JavaClassName() string {
	return "com.xxl.rpc.remoting.net.params.Beat"
}

type XxlRpcResponse struct {
	RequestId string
	ErrorMsg  string
	Result    hessian.Object
}

func (XxlRpcResponse) JavaClassName() string {
	return "com.xxl.rpc.remoting.net.params.XxlRpcResponse"
}

type ReturnT struct {
	Code    int32       `json:"code"`
	Msg     string      `json:"msg"`
	Content interface{} `json:"content"`
}

func (ReturnT) JavaClassName() string {
	return "com.xxl.job.core.biz.model.ReturnT"
}

type HandleCallbackParam struct {
	LogId         int64   `json:"logId"`
	LogDateTim    int64   `json:"logDateTim"`
	ExecuteResult ReturnT `json:"executeResult"`
}

func (HandleCallbackParam) JavaClassName() string {
	return "com.xxl.job.core.biz.model.HandleCallbackParam"
}

type RegistryParam struct {
	RegistryGroup string `json:"registryGroup"`
	RegistryKey   string `json:"registryKey"`
	RegistryValue string `json:"registryValue"`
}

func (RegistryParam) JavaClassName() string {
	return "com.xxl.job.core.biz.model.RegistryParam"
}

func InitExecutor(addresses []string, accessToken, appName string, port int) {
	hessian.RegisterPOJO(&XxlRpcRequest{})
	hessian.RegisterPOJO(&TriggerParam{})
	hessian.RegisterPOJO(&Beat{})
	hessian.RegisterPOJO(&XxlRpcResponse{})
	hessian.RegisterPOJO(&ReturnT{})
	hessian.RegisterPOJO(&HandleCallbackParam{})
	hessian.RegisterPOJO(&logger.LogResult{})
	hessian.RegisterPOJO(&RegistryParam{})
	RegisterExecutor(addresses, accessToken, appName, port, 30*time.Second)
	logrus.RegisterExitHandler(RemoveRegisterExecutor)
	go AutoRegisterJobGroup()
}

func RegisterJob(jobName string, function JobHandlerFunc) {
	if JobMap == nil {
		JobMap = make(map[string]JobHandlerFunc)
	}
	JobMap[jobName] = function
}

func GetParam(ctx context.Context, key string) string {
	jobMap := ctx.Value("jobParam")
	if jobMap != nil {
		inputParam, ok := jobMap.(map[string]map[string]interface{})["inputParam"]
		if ok {
			val, vok := inputParam[key]
			if vok {
				return val.(string)
			}
		}
	}
	return ""
}

func RequestHandler(buf []byte) (res []byte, err error) {
	out := hessian.NewDecoder(buf)
	r, err := out.Decode()

	if r == nil || err != nil {
		return nil, err
	}
	req := r.(*XxlRpcRequest)
	response := XxlRpcResponse{
		RequestId: req.RequestId,
	}
	returnt := ReturnT{
		Code:    http.StatusOK,
		Content: "success",
	}

	if len(req.Parameters) == 0 {
		response.ErrorMsg = "job parameters is empty"
		returnt.Code = http.StatusInternalServerError
		returnt.Content = "job parameters is empty"
	} else {
		if req.MethodName != "log" {
			trigger := req.Parameters[0].(*TriggerParam)
			ctx := context.Background()
			jobParam := make(map[string]map[string]interface{})

			if trigger.ExecutorParams != "" {
				params := strings.Split(trigger.ExecutorParams, ",")
				if len(params) > 0 {
					inputParam := make(map[string]interface{})
					for _, param := range params {
						if param != "" {
							jobP := strings.Split(param, "=")
							if len(jobP) > 0 {
								inputParam[jobP[0]] = jobP[1]
							}
						}
					}
					jobParam["inputParam"] = inputParam
				}
			}

			fun, ok := JobMap[trigger.ExecutorHandler]
			if ok {
				funName := getFunctionName(fun)
				logParam := make(map[string]interface{})
				logParam["logId"] = trigger.LogId
				logParam["jobId"] = trigger.JobId
				logParam["jobName"] = trigger.ExecutorHandler
				logParam["jobFunc"] = funName
				jobParam["logParam"] = logParam

				valueCtx := context.WithValue(ctx, "jobParam", jobParam)
				logger.Info(valueCtx, "job begin start!")
				err := fun(valueCtx)
				if err != nil {
					logger.Info(valueCtx, "job run failed! msg:", err.Error())
				} else {
					logger.Info(valueCtx, "job run success!")
				}

				callback := &HandleCallbackParam{
					LogId:         trigger.LogId,
					LogDateTim:    trigger.LogDateTime,
					ExecuteResult: returnt,
				}
				go CallbackAdmin([]*HandleCallbackParam{callback})
			}
		} else {
			fromLine := req.Parameters[2].(int32)
			line, content := logger.ReadLog(req.Parameters[0].(int64), req.Parameters[1].(int64), fromLine)
			log := logger.LogResult{
				FromLineNum: fromLine,
				ToLineNum:   line,
				LogContent:  content,
				IsEnd:       true,
			}
			returnt.Content = log
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
