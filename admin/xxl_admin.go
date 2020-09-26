package admin

import (
	"github.com/feixiaobo/go-xxl-job-client/v2/executor"
	"github.com/feixiaobo/go-xxl-job-client/v2/transport"
	"log"
	"net/http"
	"sync"
	"time"
)

type XxlAdminServer struct {
	AccessToken map[string]string
	Timeout     time.Duration
	Addresses   sync.Map
	Registry    *transport.RegistryParam
	BeatTime    time.Duration
	executor    *executor.Executor
}

const (
	renewTimeWaring = 30 * time.Second
	requestTime     = 10
)

type Address struct {
	Valid       int
	RequestTime int64
}

func NewAdminServer(addresses []string, timeout, beatTime time.Duration, executor *executor.Executor) *XxlAdminServer {
	if len(addresses) == 0 {
		panic("xxl admin address is null")
	}

	if beatTime >= renewTimeWaring {
		log.Print("Waring! Risk of your executor can't be renewed")
	}

	s := &XxlAdminServer{
		Timeout:  timeout,
		BeatTime: beatTime,
		executor: executor,
	}

	addressMap := sync.Map{}
	for _, add := range addresses {
		address := &Address{
			Valid:       0,
			RequestTime: time.Now().Unix(),
		}
		addressMap.Store(add, address)
	}

	s.Addresses = addressMap
	return s
}

func (s *XxlAdminServer) RegisterExecutor() {
	if s.executor.AppName == "" {
		panic("appName is executor name, it can't be null")
	}

	param := &transport.RegistryParam{
		RegistryGroup: "EXECUTOR",
		RegistryKey:   s.executor.AppName,
		RegistryValue: s.executor.GetRegisterAddr(),
	}
	s.Registry = param

	hasValid := s.requestAdminApi(s.registerExe, s.Registry)
	if !hasValid {
		panic("register executor failed, please check xxl admin address or accessToken")
	}
}

func (s *XxlAdminServer) AutoRegisterJobGroup() {
	s.Registry.RegistryValue = s.executor.GetRegisterAddr()
	t := time.NewTicker(s.BeatTime)
	for {
		select {
		case <-t.C:
			res := s.requestAdminApi(s.registerExe, s.Registry)
			if !res {
				log.Print("register job executor failed")
			}
		}
	}
}

func (s *XxlAdminServer) RemoveRegisterExecutor() {
	log.Print("remove job executor register")
	s.requestAdminApi(s.removerRegister, s.Registry)
}

func (s *XxlAdminServer) CallbackAdmin(callbackParam []*transport.HandleCallbackParam) {
	res := s.requestAdminApi(s.apiCallback, callbackParam)
	if !res {
		log.Print("job callback failed")
	}
}

//使用有效地址请求，没有有效地址遍历调用
func (s *XxlAdminServer) requestAdminApi(op func(string, interface{}) bool, param interface{}) bool {
	reqTime := time.Now().Unix()
	reqSuccess := false
	s.Addresses.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(*Address)
		if v.Valid == 0 || v.Valid == 1 { //admin地址没有请求过或者有效时直接使用该地址
			reqSuccess = op(k, param)
			if reqSuccess {
				if v.Valid == 0 {
					s.setAddressValid(k, 1)
				}
				return false
			} else {
				s.setAddressValid(k, -1)
			}
		} else if reqTime-v.RequestTime > requestTime { //地址无效且上次请求时间少于10秒内暂时跳过
			reqSuccess = op(k, param)
			if reqSuccess {
				s.setAddressValid(k, 1)
				return false
			} else {
				s.setAddressValid(k, -1)
			}
		}
		return true
	})

	if !reqSuccess { //遍历所有有效admin地址仍然没有有效请求时
		s.Addresses.Range(func(key, value interface{}) bool {
			k := key.(string)
			v := value.(*Address)
			reqSuccess = op(k, param)
			if reqSuccess {
				if v.Valid == 0 || v.Valid == -1 {
					s.setAddressValid(k, 1)
				}
				return false
			} else {
				s.setAddressValid(k, -1)
			}
			return true
		})
	}
	return reqSuccess
}

func (s *XxlAdminServer) registerExe(address string, param interface{}) bool {
	resMap, err := RegisterJobExecutor(address, s.AccessToken, param.(*transport.RegistryParam), s.Timeout)
	if err == nil && resMap["code"].(float64) == http.StatusOK {
		return true
	} else {
		return false
	}
}

func (s *XxlAdminServer) removerRegister(address string, param interface{}) bool {
	resMap, err := RemoveJobExecutor(address, s.AccessToken, param.(*transport.RegistryParam), s.Timeout)
	if err == nil && resMap["code"].(float64) == http.StatusOK {
		return true
	} else {
		return false
	}
}

func (s *XxlAdminServer) apiCallback(address string, param interface{}) bool {
	resMap, err := ApiCallback(address, s.AccessToken, param.([]*transport.HandleCallbackParam), s.Timeout)
	if err == nil && resMap["code"].(float64) == http.StatusOK {
		return true
	} else {
		return false
	}
}

func (s *XxlAdminServer) setAddressValid(address string, flag int) {
	add, ok := s.Addresses.Load(address)
	if ok {
		address := add.(*Address)
		address.Valid = flag
		address.RequestTime = time.Now().Unix()
	}
}

func (s *XxlAdminServer) GetToken() string {
	if len(s.AccessToken) > 0 {
		for _, v := range s.AccessToken {
			return v
		}
	}
	return ""
}
