package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
)

type backEndResponse struct {
	TopDog         string `json:"topDog"`
	BackendVersion int    `json:"backendVersion,omitempty"`
	MidtierVersion int    `json:"midtierVersion,omitempty"`
	UIVersion      int    `json:"uiVersion,omitempty"`
}

var v1dogs = append(dogs, "mike", "mike", "mike", "mike")

func voteV1() (string, error) {
	v := rand.Int31n(int32(len(v1dogs)))
	return v1dogs[v], nil
}

var v2dogs = dogs

func voteV2() (string, error) {
	v := rand.Int31n(int32(len(v2dogs)))
	ev := rand.Int31n(int32(4))
	if ev == 1 {
		return "", errors.New("Oops")
	}
	return v2dogs[v], nil
}

var v3dogs = append(dogs, "amit", "amit", "mike", "dan", "dan", "dan", "dan", "reuben", "prashanth")

func voteV3() (string, error) {
	v := rand.Int31n(int32(len(v3dogs)))
	return v3dogs[v], nil
}

func getVoteFunc() func() (string, error) {
	switch *version {
	case 1:
		return voteV1
	case 2:
		return voteV2
	case 3:
		return voteV3
	}
	return voteV1
}

func backEnd(resp http.ResponseWriter, req *http.Request) {
	voteFunc := getVoteFunc()
	dog, err := voteFunc()
	if err != nil {
		log.Print("Vote failure: ", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	r := backEndResponse{
		TopDog:         dog,
		BackendVersion: *version,
	}
	b, err := json.Marshal(&r)
	if err != nil {
		log.Print("Write failure: ", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Header().Set("Content-type", "application/json")
	_, err = resp.Write(b)
	if err != nil {
		log.Print("Write failure: ", err)
	}
}
