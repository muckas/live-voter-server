package main

import (
	"os"
	"io"
	"path/filepath"
	"net/http"
	"mime/multipart"
	"encoding/json"
	"strings"
	"live-voter-server/log"
	"github.com/google/uuid"
)

func isError(err error, w http.ResponseWriter, error_message string) bool {
	var response ApiResponse
	if err != nil {
		if error_message == "" {
			error_message = err.Error()
		}
		log.Error(err)
		response = ApiResponse{
			Error: "ERROR",
			Message: error_message,
			Data: struct{}{},
		}
		json.NewEncoder(w).Encode(response)
		return true
	}
	return false
}

func matchAll(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.NotFound(w, r)
}

func check(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: VERSION,
		Data: struct{}{},
	}
	json.NewEncoder(w).Encode(response)
}

func newVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var err error
	var vote_data []byte
	var f *os.File
	var vote_id = uuid.New().String()
	vote_data, _ = io.ReadAll(r.Body)
	err = os.Mkdir(filepath.Join("data", vote_id), 0600)
	if isError(err, w, "Error on saving vote") {
		return
	}
	f, err = os.OpenFile(filepath.Join("data", "vote_data", vote_id, "vote_data.json"), os.O_WRONLY|os.O_CREATE, 0600)
	if isError(err, w, "Error on saving vote") {
		return
	}
	defer f.Close()
	_, err = f.Write(vote_data)
	if isError(err, w, "Error on saving vote") {
		return
	}
	log.Debug("Created new vote: ", vote_id)
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: vote_id,
		Data: struct{}{},
	}
	json.NewEncoder(w).Encode(response)
}

func voteData(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-1]
	http.ServeFile(w, r, filepath.Join("data", "vote_data", vote_id, "vote_data.json"))
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var err error
	var form_file multipart.File
	var upload_file *os.File
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-2]
	var image_index string = url_fields[len(url_fields)-1]
	err = r.ParseMultipartForm(5 * 1024 * 1024)
	if isError(err, w, "Error parsing request form") {
		return
	}
	form_file, _, err = r.FormFile("fileupload")
	if isError(err, w, "FileForm 'fileupload' not found") {
		return
	}
	defer form_file.Close()
	upload_file, err = os.OpenFile(filepath.Join("data", "vote_data", vote_id, image_index+".png"), os.O_WRONLY|os.O_CREATE, 0666)
	if isError(err, w, "Error saving file") {
		return
	}
	defer upload_file.Close()
	io.Copy(upload_file, form_file)
}

func image(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-2]
	var image_index string = url_fields[len(url_fields)-1]
	http.ServeFile(w, r, filepath.Join("data", "vote_data", vote_id, image_index+".png"))
}

func hostVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var err error
	var vote_code string
	var vote_byte_data []byte
	var host_vote_request ApiHostVoteRequest
	var response ApiResponse
	vote_byte_data, err = io.ReadAll(r.Body)
	if isError(err, w, "Error reading request") {
		return
	}
	err = json.Unmarshal(vote_byte_data, &host_vote_request)
	if isError(err, w, "Invalid request") {
		return
	}
	vote_code, err = startNewVote(host_vote_request.VoteName)
	if isError(err, w, "") {
		return
	}
	response = ApiResponse{
		Error: "OK",
		Message: vote_code,
		Data: struct{}{},
	}
	json.NewEncoder(w).Encode(response)
}
