package main

import (
	"os"
	"path/filepath"
	"net/http"
	"live-voter-server/log"
)

const VERSION string = "0.3.0"

type ApiResponse struct {
	Error   string   `json:"error"`
	Message string   `json:"message"`
	Data    struct{} `json:data`
}

func handleRequests() {
	http.HandleFunc("/", matchAll)
	http.HandleFunc("/check", check)
	http.HandleFunc("/new-vote", newVote)
	http.HandleFunc("/vote-data/", voteData)
	http.HandleFunc("/upload-image/", uploadImage)
	http.HandleFunc("/image/", image)
	http.HandleFunc("/host-vote", hostVote)
	log.Error(http.ListenAndServe(":8080", nil))
}

func create_data_dir() {
	var err error
	var dir_name string
	err = os.Mkdir("data", 0600)
	if err != nil {
		log.Warning(err)
	}
	for _, dir_name = range []string{"vote_data", "active_votes"} {
		err = os.Mkdir(filepath.Join("data", dir_name), 0600)
		if err != nil {
			log.Warning(err)
		}
	}
}

func main() {
	log.Init("logs", "live-voter-server")
	log.Info("Live Voter server START v", VERSION)
	create_data_dir()
	handleRequests()
	log.Info("Live Voter server STOP")
}
