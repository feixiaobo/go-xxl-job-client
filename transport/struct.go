package transport

import "github.com/apache/dubbo-go-hessian2"

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
	return "com.xxl.rpc.remoting.net.params.Beat$1"
}

type XxlRpcResponse struct {
	RequestId string
	ErrorMsg  interface{}
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
