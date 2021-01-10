package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/womat/debug"
)

func initWebService() (err error) {
	for pattern, f := range map[string]func(http.ResponseWriter, *http.Request){
		"version":     httpGetVersion,
		"currentdata": httpReadCurrentData,
	} {
		if set, ok := Config.Webserver.Webservices[pattern]; ok && set {
			http.HandleFunc("/"+pattern, f)
		}
	}

	port := ":" + strconv.Itoa(Config.Webserver.Port)
	go http.ListenAndServe(port, nil)
	return
}

// httpGetVersion prints the SW Version
func httpGetVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(VERSION)); err != nil {
		debug.ErrorLog.Println(err)
		return
	}
}

// httpReadCurrentData supplies the data of all meters
func httpReadCurrentData(w http.ResponseWriter, r *http.Request) {
	var j []byte
	var err error

	func() {
		Measurements.Lock()
		defer Measurements.Unlock()
		if j, err = json.MarshalIndent(&Measurements, "", "  "); err != nil {
			debug.ErrorLog.Println(err)
			return
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(j); err != nil {
		debug.ErrorLog.Println(err)
		return
	}
}
