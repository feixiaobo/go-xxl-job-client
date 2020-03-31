package admin

import (
	"fmt"
	"github.com/feixiaobo/go-xxl-job-client/transport"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type XxlAdminServer struct {
	AccessToken string
	Timeout     time.Duration
	Addresses   sync.Map
	Registry    *transport.RegistryParam
	BeatTime    time.Duration
}

const (
	renewTimeWaring = 30 * time.Second
	requestTime     = 10
)

type Address struct {
	Valid       int
	RequestTime int64
}

func NewAdminServer(addresses []string, accessToken string, timeout, beatTime time.Duration) *XxlAdminServer {
	if len(addresses) == 0 {
		panic("xxl admin address is null")
	}

	if beatTime >= renewTimeWaring {
		log.Print("Waring! Risk of your executor can't be renewed")
	}

	s := &XxlAdminServer{
		AccessToken: accessToken,
		Timeout:     timeout,
		BeatTime:    beatTime,
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

func (s *XxlAdminServer) RegisterExecutor(appName string, port int) {
	if appName == "" {
		panic("appName is executor name, it can't be null")
	}

	param := &transport.RegistryParam{
		RegistryGroup: "EXECUTOR",
		RegistryKey:   appName,
		RegistryValue: fmt.Sprintf("%s:%d", getLocalIP(), port),
	}
	s.Registry = param

	hasValid := s.requestAdminApi(s.registerExe, s.Registry)
	if !hasValid {
		panic("register executor failed, please check xxl admin address")
	}
}

func (s *XxlAdminServer) AutoRegisterJobGroup(port int) {
	s.Registry.RegistryValue = fmt.Sprintf("%s:%d", getLocalIP(), port)
	t := time.NewTicker(s.BeatTime)
	for {
		select {
		case <-t.C:
			res := s.requestAdminApi(s.registerExe, s.Registry)
			log.Print("register job executor beat")
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

func getLocalIP() string {
	ip := getIPFromInterface("eth0")
	if ip == "" {
		ip = getIPFromInterface("en0")
	}
	if ip == "" {
		panic("Unable to determine local IP address (non loopback). Exiting.")
	}
	return ip
}

func getIPFromInterface(interfaceName string) string {
	itf, _ := net.InterfaceByName(interfaceName)
	item, _ := itf.Addrs()
	var ip net.IP
	for _, addr := range item {
		switch v := addr.(type) {
		case *net.IPNet:
			if !v.IP.IsLoopback() {
				if v.IP.To4() != nil {
					ip = v.IP
				}
			}
		}
	}
	if ip != nil {
		return ip.String()
	} else {
		return ""
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
