package main

import (
	"fmt"
	"net/http"
	"live-voter-server/log"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage, yo!")
	log.Debug("Request: /")
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	log.Error(http.ListenAndServe(":8080", nil).Error())
}

func main() {
	log.Init("logs", "live-voter-server")
	log.Info("Live Voter server START")
	handleRequests()
}
