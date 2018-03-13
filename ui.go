package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

var (
	once sync.Once
	tpl  *template.Template
)

func ui(resp http.ResponseWriter, req *http.Request) {
	once.Do(func() {
		tpl = template.Must(template.ParseGlob(filepath.Join(*staticPath, "*.html")))
		log.Print("Loaded templates")
	})
	d := make(map[string]interface{})
	d["Dogs"] = dogs
	d["Midtier"] = *midtierURL
	d["Backend"] = *backendURL
	d["ServicePort"] = *port
	d["Version"] = *version
	tpl.ExecuteTemplate(resp, "index.html", d)
}

func jsonQuery(resp http.ResponseWriter, req *http.Request) {
	result, err := queryDownstreamService(*midtierURL+"/midtier", req)
	if err != nil {
		log.Print("Cannot query midtier service: ", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	result.UIVersion = *version
	b, err := json.Marshal(result)
	if err != nil {
		log.Print("Cannot marshal JSON: ", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-type", "application/json")
	resp.Write(b)
}
