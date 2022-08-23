package main

import (
	"os"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"live-voter-server/log"
)

type ApiResponse struct {
	Error string `json:"error"`
	Message string `json:"message"`
	Data map[string]string `json:data`
}

func matchAll(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
	log.Debug(r.RemoteAddr, r.URL, "404")
}

func check(w http.ResponseWriter, r *http.Request) {
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: "OK",
		Data: map[string]string {},
	}
	json.NewEncoder(w).Encode(response)
	log.Debug(r.RemoteAddr, r.URL, "Response:", response)
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	var image_path string = filepath.Join("data", "test.png")
	buf, err := ioutil.ReadFile(image_path)
	if err != nil {
		log.Error(err)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(buf)
	log.Debug(r.RemoteAddr, r.URL, "image/png")
}

func handleRequests() {
	http.HandleFunc("/", matchAll)
	http.HandleFunc("/check", check)
	http.HandleFunc("/image", serveImage)
	log.Error(http.ListenAndServe(":8080", nil))
}

func main() {
	log.Init("logs", "live-voter-server")
	log.Info("Live Voter server START ")
	err := os.Mkdir("data", 0600)
	if err != nil {
		log.Warning(err)
	}
	handleRequests()
	log.Info("Live Voter server STOP")
}
