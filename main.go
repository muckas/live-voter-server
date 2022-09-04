package main

import (
	"os"
	"net/http"
	"encoding/json"
	"io"
	"path/filepath"
	"live-voter-server/log"
	"github.com/google/uuid"
)

const VERSION string = "0.2.0"

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: VERSION,
		Data: map[string]string{},
	}
	json.NewEncoder(w).Encode(response)
	log.Debug(r.RemoteAddr, r.URL, "Response:", response)
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var image_path string = filepath.Join("data", "image.png")
	buf, err := os.ReadFile(image_path)
	if err != nil {
		log.Error(err)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(buf)
	log.Debug(r.RemoteAddr, r.URL, "image/png")
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		log.Error(err)
	}
	form_file, _, err := r.FormFile("fileupload")
	if err != nil {
		log.Error("form_file", err)
	}
	defer form_file.Close()
	upload_file, err := os.OpenFile(filepath.Join("data", "image.png"), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error("upload_file", err)
	}
	defer upload_file.Close()
	io.Copy(upload_file, form_file)
	log.Debug(r.RemoteAddr, r.URL, "data/image.png")
}

func newVote(w http.ResponseWriter, r *http.Request) {
	var vote_id = uuid.New().String()
	vote_data, _ := io.ReadAll(r.Body)
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: vote_id,
		Data: map[string]string{},
	}
	err := os.Mkdir(filepath.Join("data", vote_id), 0600)
	if err != nil {
		log.Warning(err)
	}
	f, err := os.OpenFile(filepath.Join("data", vote_id, "vote_data.json"), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Warning(err)
	}
	defer f.Close()
	_, err = f.Write(vote_data)
	if err != nil {
		log.Warning(err)
	}
	json.NewEncoder(w).Encode(response)
	log.Debug(r.RemoteAddr, r.URL, "Data:", string(vote_data), "Response:", response)
}

func handleRequests() {
	http.HandleFunc("/", matchAll)
	http.HandleFunc("/check", check)
	http.HandleFunc("/image", serveImage)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/new-vote", newVote)
	log.Error(http.ListenAndServe(":8080", nil))
}

func main() {
	log.Init("logs", "live-voter-server")
	log.Info("Live Voter server START", "v" + VERSION)
	err := os.Mkdir("data", 0600)
	if err != nil {
		log.Warning(err)
	}
	handleRequests()
	log.Info("Live Voter server STOP")
}
