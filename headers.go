package main

import "net/http"

var headersToCopy = []string{
	"x-request-id",
	"x-b3-traceid",
	"x-b3-spanid",
	"x-b3-parentspanid",
	"x-b3-sampled",
	"x-b3-flags",
	"x-ot-span-context",
}

func copyHeaders(toReq *http.Request, fromReq *http.Request) {
	// Copy headers needed for Istio
	for _, h := range headersToCopy {
		val := fromReq.Header.Get(h)
		if val != "" {
			toReq.Header.Set(h, val)
		}
	}
}
