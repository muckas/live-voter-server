package main

import (
	"os"
	"flag"
	"path/filepath"
	"net/http"
	"live-voter-server/log"
)

const VERSION string = "0.4.0"

var SERVER_ADDRESS string
var URL_PATH string
var DATA_DIR string
var VOTE_LIFETIME int
var CLIENT_LIFETIME int
var MAX_VOTES int

func handleRequests() {
	http.HandleFunc(URL_PATH, matchAll)
	http.HandleFunc(URL_PATH + "check", check)
	http.HandleFunc(URL_PATH + "new-vote", newVote)
	http.HandleFunc(URL_PATH + "vote-data/", voteData)
	http.HandleFunc(URL_PATH + "upload-image/", uploadImage)
	http.HandleFunc(URL_PATH + "image/", image)
	http.HandleFunc(URL_PATH + "host-vote", hostVote)
	http.HandleFunc(URL_PATH + "get-active-vote/", getActiveVote)
	http.HandleFunc(URL_PATH + "keep-active-vote/", keepActiveVote)
	http.HandleFunc(URL_PATH + "update-active-vote/", updateActiveVote)
	http.HandleFunc(URL_PATH + "join-vote/", joinVote)
	http.HandleFunc(URL_PATH + "send-vote/", sendVote)
	log.Info("Server listening on ", SERVER_ADDRESS)
	log.Error(http.ListenAndServe(SERVER_ADDRESS, nil))
}

func create_data_dir() {
	var err error
	var dir_name string
	err = os.Mkdir(DATA_DIR, 0600)
	if err != nil {
		log.Debug(err)
	}
	for _, dir_name = range []string{"vote_data", "active_votes"} {
		err = os.Mkdir(filepath.Join(DATA_DIR, dir_name), 0600)
		if err != nil {
			log.Debug(err)
		}
	}
}

func parse_flags() {
	flag.StringVar(&SERVER_ADDRESS, "address", ":8080", "server port")
	flag.StringVar(&URL_PATH, "urlpath", "/", "url path")
	flag.StringVar(&DATA_DIR, "datadir", "data", "data directory")
	flag.IntVar(&VOTE_LIFETIME, "votelifetime", 10, "vote lifetime in minutes")
	flag.IntVar(&CLIENT_LIFETIME, "clientlifetime", 60, "client lifetime in seconds")
	flag.IntVar(&MAX_VOTES, "maxvotes", 100, "max active votes allowed")
	flag.Parse()
}

func main() {
	parse_flags()
	log.Init("logs", "live-voter-server")
	log.Info("Live Voter server START v", VERSION)
	log.Info("Server address: ", SERVER_ADDRESS)
	log.Info("URL path: ", URL_PATH)
	log.Info("Data directory: ", DATA_DIR)
	log.Info("Vote lifetime: ", VOTE_LIFETIME, "m")
	log.Info("Client lifetime: ", CLIENT_LIFETIME, "s")
	log.Info("Max votes allowed: ", MAX_VOTES)
	create_data_dir()
	handleRequests()
	log.Info("Live Voter server STOP")
}
