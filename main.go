package main

import (
	"net/http"
	"encoding/json"
	"live-voter-server/log"
)

type ApiResponse struct {
	Error string `json:"error"`
	Message string `json:"message"`
	Data map[string]string `json:data`
}

func check(w http.ResponseWriter, r *http.Request) {
	var response ApiResponse
	log.Debug("endpoint hit: /check")
	response = ApiResponse{
		Error: "OK",
		Message: "OK",
		Data: map[string]string {},
	}
	json.NewEncoder(w).Encode(response)
}

func handleRequests() {
	http.HandleFunc("/check", check)
	log.Error(http.ListenAndServe(":8080", nil).Error())
}

func main() {
	log.Init("logs", "live-voter-server")
	log.Info("Live Voter server START")
	handleRequests()
	log.Info("Live Voter server STOP")
}
