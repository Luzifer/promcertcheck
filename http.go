package main

//go:generate go-bindata display.html

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/flosch/pongo2"
)

func htmlHandler(res http.ResponseWriter, r *http.Request) {
	tplsrc, _ := Asset("display.html")

	template, err := pongo2.FromString(string(tplsrc))
	if err != nil {
		log.Fatal(err)
	}

	template.ExecuteWriter(pongo2.Context{
		"results":                probeMonitors,
		"certificateOK":          certificateOK,
		"certificateExpiresSoon": certificateExpiresSoon,
		"version":                version,
	}, res)
}

func httpStatusHandler(res http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusOK
	for _, mon := range probeMonitors {
		if mon.Status != certificateOK {
			httpStatus = http.StatusInternalServerError
		}
	}

	res.WriteHeader(httpStatus)
}

func jsonHandler(res http.ResponseWriter, r *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(probeMonitors)
}
