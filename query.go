package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	transport = &http.Transport{DisableKeepAlives: false, MaxIdleConnsPerHost: 10, DisableCompression: false, ResponseHeaderTimeout: time.Second * 5}
	client    = &http.Client{Transport: transport, Timeout: time.Second * 10}
)

func queryDownstreamService(url string, originalRequest *http.Request) (*backEndResponse, error) {
	// create request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Close = false

	// copy headers for Istio and correlation id
	copyHeaders(request, originalRequest)

	// issue request
	response, err := client.Do(request)
	if err != nil {
		log.Print("HTTP request error on "+url+": ", err)
		return nil, err
	}

	var data []byte
	data, err = ioutil.ReadAll(response.Body)
	response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		err = errors.New(string(data))
		log.Printf("HTTP error %d on %s: %s", response.StatusCode, url, err)
		return nil, err
	}

	var result backEndResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Print("Unable to parse JSON from "+url+": ", err)
		return nil, err
	}

	return &result, nil
}
