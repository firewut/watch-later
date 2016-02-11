package helpers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Put(url string, headers http.Header, payload interface{}) (response *http.Response, err error) {
	client := &http.Client{}
	json_raw, err := json.Marshal(map[string]interface{}{
		"payload": payload,
	})
	if err != nil {
		return
	}
	request, _ := http.NewRequest("PUT", url, bytes.NewBuffer(json_raw))
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}

func Post(url string, headers http.Header, payload interface{}) (response *http.Response, err error) {
	client := &http.Client{}
	json_raw, err := json.Marshal(map[string]interface{}{
		"payload": payload,
	})
	if err != nil {
		return
	}
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(json_raw))
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}

func Patch(url string, headers http.Header, payload interface{}) (response *http.Response, err error) {
	client := &http.Client{}
	json_raw, err := json.Marshal(map[string]interface{}{
		"payload": payload,
	})
	if err != nil {
		return
	}
	request, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(json_raw))
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}

func Delete(url string, headers http.Header) (response *http.Response, err error) {
	client := &http.Client{}
	request, _ := http.NewRequest("DELETE", url, nil)
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}

func Get(url string, headers http.Header) (response *http.Response, err error) {
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}
