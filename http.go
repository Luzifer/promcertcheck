package main

//go:generate go-bindata display.html

import (
	"log"
	"net/http"

	"github.com/flosch/pongo2"
)

func httpHandler(res http.ResponseWriter, r *http.Request) {
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
