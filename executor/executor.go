package executor

import (
	"fmt"
	"net"
)

type Executor struct {
	Protocol    string
	AppName     string
	Port        int
	gettyClient *GettyClient
}

func NewExecutor(protocol, appName string, port int) *Executor {
	return &Executor{
		Protocol: protocol,
		AppName:  appName,
		Port:     port,
	}
}
func (e *Executor) SetClient(gettyClient *GettyClient) {
	e.gettyClient = gettyClient
}

func (e *Executor) GetRegisterAddr() string {
	if e.Protocol == "" {
		return fmt.Sprintf("%s:%d", getIp(), e.Port)
	} else {
		return fmt.Sprintf("%s://%s:%d/", e.Protocol, getIp(), e.Port)
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

func getIp() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}
	panic("Unable to determine local IP address (non loopback). Exiting.")
}

func (e *Executor) Run(taskSize int) {
	if e.gettyClient == nil {
		panic("executor client has not been set")
	}
	e.gettyClient.Run(e.Port, taskSize)
}
