package admin

import (
	"bytes"
	"encoding/json"
	"github.com/feixiaobo/go-xxl-job-client/v2/transport"
	"io/ioutil"
	"net/http"
	"time"
)

func ApiCallback(address string, accessToken map[string]string, callbackParam []*transport.HandleCallbackParam, timeout time.Duration) (respMap map[string]interface{}, err error) {
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
	if len(accessToken) > 0 {
		for k, v := range accessToken {
			request.Header.Set(k, v)
			//request.Header.Set("XXL-RPC-ACCESS-TOKEN", accessToken)
		}
	}
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

func RegisterJobExecutor(address string, accessToken map[string]string, param *transport.RegistryParam, timeout time.Duration) (respMap map[string]interface{}, err error) {
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
	if len(accessToken) > 0 {
		for k, v := range accessToken {
			request.Header.Set(k, v)
		}
	}
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

func RemoveJobExecutor(address string, accessToken map[string]string, param *transport.RegistryParam, timeout time.Duration) (respMap map[string]interface{}, err error) {
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
	if len(accessToken) > 0 {
		for k, v := range accessToken {
			request.Header.Set(k, v)
		}
	}
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
