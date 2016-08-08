package helpers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
)

func Put(
	url string,
	headers http.Header,
	check_ssl_cert bool,
	payload interface{},
) (
	response *http.Response,
	err error,
) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !check_ssl_cert,
		},
	}
	client := &http.Client{Transport: tr}

	json_raw, err := json.Marshal(
		payload,
	)
	if err != nil {
		return
	}
	request, _ := http.NewRequest(
		"PUT",
		url,
		bytes.NewBuffer(json_raw),
	)
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}

func Post(
	url string,
	headers http.Header,
	check_ssl_cert bool,
	payload interface{},
) (
	response *http.Response,
	err error,
) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !check_ssl_cert,
		},
	}
	client := &http.Client{Transport: tr}

	json_raw, err := json.Marshal(
		payload,
	)
	if err != nil {
		return
	}
	request, _ := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(json_raw),
	)
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}

func Patch(
	url string,
	headers http.Header,
	check_ssl_cert bool,
	payload interface{},
) (
	response *http.Response,
	err error,
) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !check_ssl_cert,
		},
	}
	client := &http.Client{Transport: tr}

	json_raw, err := json.Marshal(
		payload,
	)
	if err != nil {
		return
	}
	request, _ := http.NewRequest(
		"PATCH",
		url,
		bytes.NewBuffer(json_raw),
	)
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}

func Delete(url string, headers http.Header, check_ssl_cert bool) (
	response *http.Response,
	err error,
) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !check_ssl_cert,
		},
	}
	client := &http.Client{Transport: tr}

	request, _ := http.NewRequest(
		"DELETE", url, nil,
	)
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}

func Get(url string, headers http.Header, check_ssl_cert bool) (
	response *http.Response,
	err error,
) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !check_ssl_cert,
		},
	}
	client := &http.Client{Transport: tr}

	request, _ := http.NewRequest(
		"GET", url, nil,
	)
	for k, list_v := range headers {
		for _, v := range list_v {
			request.Header.Set(k, v)
		}
	}
	response, err = client.Do(request)
	return
}
