package xxl

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var XxlAdmin XxlAdminInfo

type Address struct {
	Valid       int
	RequestTime int64
}

type XxlAdminInfo struct {
	AccessToken string
	Port        int
	Timeout     time.Duration
	Addresses   sync.Map
	Registry    *RegistryParam
}

func RegisterExecutor(addresses []string, accessToken, appName string, port int, timeout time.Duration) {
	if len(addresses) == 0 {
		panic("xxl admin address is null")
	}
	if appName == "" {
		panic("appName is executor name, it can't be null")
	}

	hasValid := false
	XxlAdmin = XxlAdminInfo{
		Port:        port,
		AccessToken: accessToken,
		Timeout:     timeout,
	}

	param := &RegistryParam{
		RegistryGroup: "EXECUTOR",
		RegistryKey:   appName,
		RegistryValue: fmt.Sprintf("%s:%d", getLocalIP(), port),
	}
	XxlAdmin.Registry = param

	addressMap := sync.Map{}
	for _, add := range addresses {
		address := &Address{RequestTime: time.Now().Unix()}
		if !hasValid {
			resMap, err := RegisterJobExecutor(add, XxlAdmin.AccessToken, param, XxlAdmin.Timeout)
			if err == nil && resMap["code"].(float64) == 200 {
				address.Valid = 1
				hasValid = true
			} else {
				address.Valid = -1
			}
		} else {
			address.Valid = 0
		}
		addressMap.Store(add, address)
	}

	if !hasValid {
		panic("register executor failed, please check xxl admin address")
	}
	XxlAdmin.Addresses = addressMap

}

func AutoRegisterJobGroup() {
	XxlAdmin.Registry.RegistryValue = fmt.Sprintf("%s:%d", getLocalIP(), XxlAdmin.Port)
	t := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-t.C:
			res := requestAdminApi(registerExe, XxlAdmin.Registry)
			log.Print("register job executor beat")
			if !res {
				log.Print("register job executor failed")
			}
		}
	}
}

func RemoveRegisterExecutor() {
	log.Print("remove job executor register")
	requestAdminApi(removerRegister, XxlAdmin.Registry)
}

func CallbackAdmin(callbackParam []*HandleCallbackParam) {
	requestAdminApi(apiCallback, callbackParam)
}

//使用有效地址请求，没有有效地址遍历调用
func requestAdminApi(op func(string, interface{}) bool, param interface{}) bool {
	reqTime := time.Now().Unix()
	reqSuccess := false
	XxlAdmin.Addresses.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(*Address)
		if v.Valid == 0 || v.Valid == 1 { //admin地址没有请求过或者有效时直接使用该地址
			reqSuccess = op(k, param)
			if reqSuccess {
				if v.Valid == 0 {
					setAddressValid(k, 1)
				}
				return false
			} else {
				setAddressValid(k, -1)
			}
		} else if reqTime-v.RequestTime > 10 { //地址无效且上次请求时间少于10秒内暂时跳过
			reqSuccess = op(k, param)
			if reqSuccess {
				setAddressValid(k, 1)
				return false
			} else {
				setAddressValid(k, -1)
			}
		}
		return true
	})

	if !reqSuccess { //遍历所有有效admin地址仍然没有有效请求时
		XxlAdmin.Addresses.Range(func(key, value interface{}) bool {
			k := key.(string)
			v := value.(*Address)
			reqSuccess = op(k, param)
			if reqSuccess {
				if v.Valid == 0 || v.Valid == -1 {
					setAddressValid(k, 1)
				}
				return false
			} else {
				setAddressValid(k, -1)
			}
			return true
		})
	}
	return reqSuccess
}

func registerExe(address string, param interface{}) bool {
	resMap, err := RegisterJobExecutor(address, XxlAdmin.AccessToken, param.(*RegistryParam), XxlAdmin.Timeout)
	if err == nil && resMap["code"].(float64) == 200 {
		return true
	} else {
		return false
	}
}

func removerRegister(address string, param interface{}) bool {
	resMap, err := RemoveJobExecutor(address, XxlAdmin.AccessToken, param.(*RegistryParam), XxlAdmin.Timeout)
	if err == nil && resMap["code"].(float64) == 200 {
		return true
	} else {
		return false
	}
}

func apiCallback(address string, param interface{}) bool {
	resMap, err := ApiCallback(address, XxlAdmin.AccessToken, param.([]*HandleCallbackParam), XxlAdmin.Timeout)
	if err == nil && resMap["code"].(float64) == 200 {
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

func setAddressValid(address string, flag int) {
	add, ok := XxlAdmin.Addresses.Load(address)
	if ok {
		address := add.(*Address)
		address.Valid = flag
		address.RequestTime = time.Now().Unix()
	}
}
