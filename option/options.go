package option

import "time"

const (
	defaultAdminAddr = "http://localhost:8080/xxl-job-admin/"
	defaultAppName   = "go-executor"
	defaultPort      = 8081
	defaultTimeout   = 5 * time.Second
	defaultBeatTime  = 20 * time.Second
)

type Option func(*ClientOptions)

type ClientOptions struct {
	//xxl admin 地址
	AdminAddr []string

	//token
	AccessToken string

	//执行期名
	AppName string

	//执行器端口
	Port int

	//开启http协议
	EnableHttp bool

	//请求admin超时时间
	Timeout time.Duration

	//执行器续约时间（超过30秒不续约admin会移除执行器，请设置到30秒以内）
	BeatTime time.Duration
}

func NewClientOptions(opts ...Option) ClientOptions {
	options := ClientOptions{
		AdminAddr:   []string{defaultAdminAddr},
		AccessToken: "",
		AppName:     defaultAppName,
		Port:        defaultPort,
		Timeout:     defaultTimeout,
		BeatTime:    defaultBeatTime,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

//xxl admin address
func WithAdminAddress(addrs ...string) Option {
	return func(o *ClientOptions) {
		o.AdminAddr = addrs
	}
}

//xxl admin accessToke
func WithAccessToken(token string) Option {
	return func(o *ClientOptions) {
		o.AccessToken = token
	}
}

//app name
func WithAppName(appName string) Option {
	return func(o *ClientOptions) {
		o.AppName = appName
	}
}

//xxl client port
func WithClientPort(port int) Option {
	return func(o *ClientOptions) {
		o.Port = port
	}
}

//xxl admin request timeout
func WithAdminTimeout(timeout time.Duration) Option {
	return func(o *ClientOptions) {
		o.Timeout = timeout
	}
}

//xxl admin renew time
func WithBeatTime(beatTime time.Duration) Option {
	return func(o *ClientOptions) {
		o.BeatTime = beatTime
	}
}

func WithEnableHttp(enable bool) Option {
	return func(o *ClientOptions) {
		o.EnableHttp = enable
	}
}
