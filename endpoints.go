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

func matchAll(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	http.NotFound(w, r)
}

func check(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var response ApiResponse = ApiResponse{
		Error:   "OK",
		Message: VERSION,
		Data:    struct{}{},
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
	var response ApiResponse = ApiResponse{
		Error:   "OK",
		Message: vote_id,
		Data:    struct{}{},
	}
	err = os.Mkdir(filepath.Join("data", vote_id), 0600)
	if err != nil {
		log.Error(err)
	}
	f, err = os.OpenFile(filepath.Join("data", "vote_data", vote_id, "vote_data.json"), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()
	_, err = f.Write(vote_data)
	if err != nil {
		log.Error(err)
	}
	log.Debug("Created new vote: ", vote_id)
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
	var err error
	var form_file multipart.File
	var upload_file *os.File
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-2]
	var image_index string = url_fields[len(url_fields)-1]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err = r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	form_file, _, err = r.FormFile("fileupload")
	if err != nil {
		log.Error("form_file: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer form_file.Close()
	upload_file, err = os.OpenFile(filepath.Join("data", "vote_data", vote_id, image_index+".png"), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error("upload_file ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer upload_file.Close()
	io.Copy(upload_file, form_file)
}

func image(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-2]
	var image_index string = url_fields[len(url_fields)-1]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeFile(w, r, filepath.Join("data", "vote_data", vote_id, image_index+".png"))
}

func hostVote(w http.ResponseWriter, r *http.Request) {
	var err error
	var vote_code string
	var response ApiResponse
	log.Debug(r.RemoteAddr, " ", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vote_code, err = startNewVote()
	if err == nil {
		response = ApiResponse{
			Error:   "OK",
			Message: vote_code,
			Data:    struct{}{},
		}
	} else {
		response = ApiResponse{
			Error:   "ERROR",
			Message: err.Error(),
			Data:    struct{}{},
		}
	}
	json.NewEncoder(w).Encode(response)
}
