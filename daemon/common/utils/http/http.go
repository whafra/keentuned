package http

import (
	"keentune/daemon/common/log"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Protocol  ...
const Protocol = "http"

// StatusOk ...
const StatusOk = 200

type requester struct {
	client *http.Client
	url    string
	body   []byte
	method string
}

func newRequester(method string, uri string, data interface{}) (*requester, error) {
	url := Protocol + "://" + uri
	if data == nil && method == "GET" {
		return &requester{
			url:    url,
			body:   nil,
			method: method,
		}, nil
	}

	var client = &http.Client{}
	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("[newRequester] error in Unexpected type %T\n", data)
		return nil, err
	}

	log.Debugf(log.ParamTune, "[%v] request info [%+v]", url, string(bytesData))
	return &requester{
		client: client,
		url:    url,
		body:   bytesData,
		method: method,
	}, nil
}

func (r *requester) execute() (*http.Response, error) {
	if r.method == "GET" && r.body == nil {
		response, err := http.Get(r.url)
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	request, err := http.NewRequest(r.method, r.url, bytes.NewReader(r.body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// RemoteCall remote call the url with the specified method and request data
func RemoteCall(method, url string, data interface{}) ([]byte, error) {
	requester, err := newRequester(method, url, data)
	if err != nil {
		return nil, err
	}

	response, err := requester.execute()
	if err != nil {
		return nil, err
	}	

	if response.Body == nil {
		return nil, fmt.Errorf("[%v] response is nil", method+url)
	}

	defer response.Body.Close()

	bytesData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != StatusOk {
		return bytesData, fmt.Errorf("request [%v] status is:%v", Protocol+"://"+url, response.StatusCode)
	}

	log.Debugf(log.ParamTune, "[%v] response info [%+v]", Protocol+"://"+url, string(bytesData))
	return bytesData, nil
}

// ResponseSuccess get success response after request url
func ResponseSuccess(method, url string, request interface{}) error {
	resp, err := RemoteCall(method, url, request)
	if err != nil {
		return err
	}

	var response struct {
		Success bool        `json:"suc"`
		Message interface{} `json:"msg"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("response suc is false, msg is [%v].", response.Message)
	}

	return nil
}
