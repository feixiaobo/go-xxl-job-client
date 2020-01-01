package xxl

import (
	"fmt"
	"github.com/feixiaobo/go-xxl-job-client/utils"
	"github.com/gin-gonic/gin"
	"net"
	"time"
)

var XxlAdmin XxlAdminInfo

type JobFunc func()

func (f JobFunc) RunJob(c *gin.Context) { f() }

type XxlAdminInfo struct {
	AccessToken string
	Port        int
	Timeout     time.Duration
	Addresses   map[string]*utils.SafeMap
	Registry    *RegistryParam
}

func RegisterExecutor(addresses []string, accessToken, appName string, port int, timeout time.Duration) {
	if len(addresses) == 0 {
		panic("xxl admin address is null")
	}
	if appName == "" {
		panic("appName is executor, it can't be null")
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

	addressMap := make(map[string]*utils.SafeMap)
	for _, address := range addresses {
		validMap := utils.NewSafeMap(len(addresses))
		if !hasValid {
			resMap, err := RegisterJobExecutor(address, XxlAdmin.AccessToken, param, XxlAdmin.Timeout)
			if err == nil && resMap["code"].(float64) == 200 {
				validMap.WriteMap("valid", 1)
				hasValid = true
			} else {
				validMap.WriteMap("valid", -1)
			}
			validMap.WriteMap("requestTime", time.Now().Unix())
		} else {
			validMap.WriteMap("valid", 0)
			validMap.WriteMap("requestTime", time.Now().Unix())
		}
		addressMap[address] = validMap
	}

	if !hasValid {
		panic("register executor failed, please check xxl admin address")
	}
	XxlAdmin.Addresses = addressMap

}

func AutoRegisterJobGroup() {
	XxlAdmin.Registry.RegistryValue = fmt.Sprintf("%s:%d", getLocalIP(), XxlAdmin.Port)
	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			requestAdminApi(registerExe, XxlAdmin.Registry)
			t = time.NewTicker(10 * time.Second)
		}
	}
}

func RemoveRegisterExecutor() {
	requestAdminApi(removerRegister, XxlAdmin.Registry)
}

func CallbackAdmin(callbackParam []*HandleCallbackParam) {
	requestAdminApi(apiCallback, callbackParam)
}

//使用有效地址请求，没有有效地址遍历调用
func requestAdminApi(op func(string, interface{}) bool, param interface{}) {
	reqTime := time.Now().Unix()
	for k, v := range XxlAdmin.Addresses {
		if v.ReadMap("valid") == 0 || v.ReadMap("valid") == 1 {
			if op(k, param) {
				return
			} else {
				setAddressValid(k, -1)
			}
		} else if reqTime-v.ReadMap("requestTime").(int64) > 10*1000*1000 {
			if op(k, param) {
				return
			} else {
				setAddressValid(k, -1)
			}
		}
	}

	for k, _ := range XxlAdmin.Addresses {
		if op(k, param) {
			return
		} else {
			setAddressValid(k, -1)
		}
	}
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
	validMap, ok := XxlAdmin.Addresses[address]
	if ok {
		validMap.WriteMap("valid", flag)
		validMap.WriteMap("requestTime", time.Now().Unix())
	}
}
