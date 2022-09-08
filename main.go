package main

import (
	"os"
	"net/http"
	"encoding/json"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"live-voter-server/log"
)

const VERSION string = "0.3.0"

type ApiResponse struct {
	Error string `json:"error"`
	Message string `json:"message"`
	Data map[string]string `json:data`
}

func matchAll(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, r.URL)
	http.NotFound(w, r)
}

func check(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: VERSION,
		Data: map[string]string{},
	}
	json.NewEncoder(w).Encode(response)
}

func newVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, r.URL)
	var vote_id = uuid.New().String()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vote_data, _ := io.ReadAll(r.Body)
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: vote_id,
		Data: map[string]string{},
	}
	err := os.Mkdir(filepath.Join("data", vote_id), 0600)
	if err != nil {
		log.Error(err)
	}
	f, err := os.OpenFile(filepath.Join("data", "vote_data", vote_id, "vote_data.json"), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()
	_, err = f.Write(vote_data)
	if err != nil {
		log.Error(err)
	}
	log.Debug("Created new vote:", vote_id)
	json.NewEncoder(w).Encode(response)
}

func voteData(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-1]
	http.ServeFile(w, r, filepath.Join("data", "vote_data", vote_id, "vote_data.json"))
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, r.URL)
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-2]
	var image_index string = url_fields[len(url_fields)-1]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	form_file, _, err := r.FormFile("fileupload")
	if err != nil {
		log.Error("form_file", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer form_file.Close()
	upload_file, err := os.OpenFile(filepath.Join("data", "vote_data", vote_id, image_index+".png"), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error("upload_file", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer upload_file.Close()
	io.Copy(upload_file, form_file)
}

func image(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, r.URL)
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-2]
	var image_index string = url_fields[len(url_fields)-1]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeFile(w, r, filepath.Join("data", "vote_data", vote_id, image_index+".png"))
}

func generateNewVoteCode() (string, error) {
	var code_length int = 2
	bytes := make([]byte, code_length)
  if _, err := rand.Read(bytes); err != nil {
    return "", err
  }
  return hex.EncodeToString(bytes), nil
}

func startNewVote() (string, error) {
	var max_votes int = 65000
	var votes_dir string = filepath.Join("data", "active_votes")
	var code string
	var err error
	var dir_entry []fs.DirEntry
	dir_entry, err = os.ReadDir(votes_dir)
	if len(dir_entry) >= max_votes {
		return "", errors.New("max votes exceeded")
	}
	for {
		code, err = generateNewVoteCode()
		_, err = os.Stat(filepath.Join(votes_dir, code))
		if os.IsNotExist(err) {
			break
		}
	}
	err = os.Mkdir(filepath.Join(votes_dir, code), 0600)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return code, nil
}

func hostVote(w http.ResponseWriter, r *http.Request) {
	var response ApiResponse
	log.Debug(r.RemoteAddr, r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vote_code, err := startNewVote()
	if err == nil {
		response = ApiResponse{
			Error: "OK",
			Message: vote_code,
			Data: map[string]string{},
		}
	} else {
		response = ApiResponse{
			Error: "ERROR",
			Message: err.Error(),
			Data: map[string]string{},
		}
	}
	json.NewEncoder(w).Encode(response)
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
	err := os.Mkdir("data", 0600)
	if err != nil {
		log.Warning(err)
	}
	for _, dir_name := range []string{"vote_data", "active_votes"} {
		err := os.Mkdir(filepath.Join("data", dir_name), 0600)
		if err != nil {
			log.Warning(err)
		}
	}
}

func main() {
	log.Init("logs", "live-voter-server")
	log.Info("Live Voter server START", "v" + VERSION)
	create_data_dir()
	handleRequests()
	log.Info("Live Voter server STOP")
}
