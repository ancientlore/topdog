package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func midTier(resp http.ResponseWriter, req *http.Request) {
	result, err := queryDownstreamService(*backendURL+"/backend", req)
	if err != nil {
		log.Print("Cannot query backend service: ", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	result.MidtierVersion = *version
	b, err := json.Marshal(result)
	if err != nil {
		log.Print("Cannot marshal JSON: ", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-type", "application/json")
	resp.Write(b)
}
