package main

import (
	"errors"
	"time"
	"os"
	"io"
	"io/fs"
	"path/filepath"
	"net/http"
	"mime/multipart"
	"encoding/json"
	"strings"
	"crypto/rand"
	"encoding/hex"
	"live-voter-server/log"
	"github.com/google/uuid"
)

const VERSION string = "0.3.0"

type ApiResponse struct {
	Error   string   `json:"error"`
	Message string   `json:"message"`
	Data    struct{} `json:data`
}

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

func generateNewVoteCode() (string, error) {
	var err error
	var bytes []byte
	var code_length int = 2
	bytes = make([]byte, code_length)
	if _, err = rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func voteCleanup(votes_dir string, vote_lifetime time.Duration) (int, error) {
	var err error
	var dir_content []fs.DirEntry
	dir_content, err = os.ReadDir(votes_dir)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	var index int = 0
	var dir_info fs.FileInfo
	var dir_time time.Time
	var now time.Time = time.Now()
	for index < len(dir_content) {
		dir_info, err = dir_content[index].Info()
		dir_time = dir_info.ModTime()
		if dir_time.Add(vote_lifetime).Before(now) { // if modification time is more then vote_lifetime ago
			err = os.RemoveAll(filepath.Join(votes_dir, dir_info.Name()))
			if err != nil {
				log.Warning(err)
			}
			log.Debug("Removed inactive vote: ", dir_info.Name())
			dir_content = append(dir_content[:index], dir_content[index+1:]...)
		} else {
			index++
		}
	}
	log.Debug("Active votes num: ", len(dir_content))
	return len(dir_content), nil
}

func startNewVote() (string, error) {
	var max_votes int = 100
	var vote_lifetime time.Duration = 10 * time.Minute
	var code string
	var err error
	var num_active_votes int
	var votes_dir string = filepath.Join("data", "active_votes")
	num_active_votes, err = voteCleanup(votes_dir, vote_lifetime)
	if num_active_votes >= max_votes {
		log.Info("Max votes exceeded (max ", max_votes, ")")
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
	log.Debug("Created active vote: ", code)
	return code, nil
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
