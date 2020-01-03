package xxl

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type Job struct {
	JobGroup              string `json:"jobGroup"`
	JobCorn               string `json:"jobCorn"`
	JobDesc               string `json:"jobDesc"`
	Author                string `json:"author"`
	ExecutorRouteStrategy string `json:"executorRouteStrategy"`
	ExecutorBlockStrategy string `json:"executorBlockStrategy"`
	GlueType              string `json:"glueType"`
	GlueSource            string `json:"glueSource"`
	AlarmEmail            string `json:"alarmEmail"`
}

type JobGroup struct {
	AppName     string `json:"appName"`
	Title       string `json:"title"`
	AddressType int32  `json:"addressType"`
	AddressList string `json:"addressList"`
}

func ApiCallback(address, accessToken string, callbackParam []*HandleCallbackParam, timeout time.Duration) (respMap map[string]interface{}, err error) {
	bytesData, err := json.Marshal(callbackParam)
	if err != nil {
		return respMap, err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", address+"/api/callback", reader)
	if err != nil {
		return respMap, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("XXL-RPC-ACCESS-TOKEN", accessToken)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(request)
	if err != nil {
		return respMap, err
	} else {
		defer resp.Body.Close()
	}
	respMap, err = parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return respMap, nil
}

func RegisterJobExecutor(address, accessToken string, param *RegistryParam, timeout time.Duration) (respMap map[string]interface{}, err error) {
	bytesData, err := json.Marshal(param)
	if err != nil {
		return respMap, err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", address+"/api/registry", reader)
	if err != nil {
		return respMap, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("XXL-RPC-ACCESS-TOKEN", accessToken)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(request)
	if err != nil {
		return respMap, err
	} else {
		defer resp.Body.Close()
	}
	respMap, err = parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return respMap, nil
}

func RemoveJobExecutor(address, accessToken string, param *RegistryParam, timeout time.Duration) (respMap map[string]interface{}, err error) {
	bytesData, err := json.Marshal(param)
	if err != nil {
		return respMap, err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", address+"/api/registryRemove", reader)
	if err != nil {
		return respMap, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("XXL-RPC-ACCESS-TOKEN", accessToken)
	client := http.Client{Timeout: timeout}
	resp, err := client.Do(request)
	if err != nil {
		return respMap, err
	} else {
		defer resp.Body.Close()
	}
	respMap, err = parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return respMap, nil
}

func parseResponse(response *http.Response) (map[string]interface{}, error) {
	var result map[string]interface{}
	body, err := ioutil.ReadAll(response.Body)
	if err == nil {
		err = json.Unmarshal(body, &result)
	}

	return result, err
}
