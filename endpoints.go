package main

import (
	"time"
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
	if err != nil {
		if error_message == "" {
			error_message = err.Error()
		}
		log.Error("Error in ", log.TraceCaller(3), ": ", err)
		var response ApiResponse = ApiResponse{
			Error: "ERROR",
			Message: error_message,
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return true
	}
	return false
}

func set_headers(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func matchAll(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	http.NotFound(w, r)
}

func check(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: VERSION,
		Data: nil,
	}
	json.NewEncoder(w).Encode(response)
}

func newVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var err error
	var vote_id = uuid.New().String()
	var vote_data []byte
	vote_data, _ = io.ReadAll(r.Body)
	err = os.Mkdir(filepath.Join("data", "vote_data", vote_id), 0600)
	if isError(err, w, "Error saving vote") {
		return
	}
	var file *os.File
	file, err = os.OpenFile(filepath.Join("data", "vote_data", vote_id, "vote_data.json"), os.O_WRONLY|os.O_CREATE, 0600)
	if isError(err, w, "Error saving vote") {
		return
	}
	defer file.Close()
	_, err = file.Write(vote_data)
	if isError(err, w, "Error saving vote") {
		return
	}
	log.Debug("Created new vote: ", vote_id)
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: vote_id,
		Data: nil,
	}
	json.NewEncoder(w).Encode(response)
}

func voteData(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-1]
	http.ServeFile(w, r, filepath.Join("data", "vote_data", vote_id, "vote_data.json"))
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var err error
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-2]
	var image_index string = url_fields[len(url_fields)-1]
	err = r.ParseMultipartForm(5 * 1024 * 1024)
	if isError(err, w, "Error parsing request form") {
		return
	}
	var form_file multipart.File
	form_file, _, err = r.FormFile("fileupload")
	if isError(err, w, "FileForm 'fileupload' not found") {
		return
	}
	defer form_file.Close()
	var upload_file *os.File
	upload_file, err = os.OpenFile(filepath.Join("data", "vote_data", vote_id, image_index+".png"), os.O_WRONLY|os.O_CREATE, 0666)
	if isError(err, w, "Error saving file") {
		return
	}
	defer upload_file.Close()
	io.Copy(upload_file, form_file)
}

func image(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_id string = url_fields[len(url_fields)-2]
	var image_index string = url_fields[len(url_fields)-1]
	http.ServeFile(w, r, filepath.Join("data", "vote_data", vote_id, image_index+".png"))
}

func hostVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var err error
	var vote_byte_data []byte
	vote_byte_data, err = io.ReadAll(r.Body)
	if isError(err, w, "Error reading request") {
		return
	}
	var vote_code string
	var host_vote_request ApiHostVoteRequest
	err = json.Unmarshal(vote_byte_data, &host_vote_request)
	if isError(err, w, "Invalid request") {
		return
	}
	vote_code, err = startNewVote(host_vote_request.HostID, host_vote_request.VoteName)
	if isError(err, w, "") {
		return
	}
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: vote_code,
		Data: nil,
	}
	json.NewEncoder(w).Encode(response)
}

func getActiveVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var err error
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_code string = url_fields[len(url_fields)-1]
	var active_vote_bytes []byte
	active_vote_bytes, err = os.ReadFile(filepath.Join("data", "active_votes", vote_code + ".json"))
	if isError(err, w, "Invalid vote code") {
		return
	}
	var active_vote_info ActiveVoteInfo
	err = json.Unmarshal(active_vote_bytes, &active_vote_info)
	if isError(err, w, "Corrupted vote data, unable to proceed") {
		return
	}
	var get_vote_byte_data []byte
	get_vote_byte_data, err = io.ReadAll(r.Body)
	if isError(err, w, "Error reading request") {
		return
	}
	var client_request ClientRequest
	err = json.Unmarshal(get_vote_byte_data, &client_request)
	if isError(err, w, "Invalid request") {
		return
	}
	var now time.Time = time.Now()
	active_vote_info.Clients[client_request.ID] = now
	var file *os.File
	file, err = os.Create(filepath.Join("data", "active_votes", vote_code + ".json"))
	if isError(err, w, "Inable to change active vote") {
		return
	}
	defer file.Close()
	active_vote_bytes, err = json.Marshal(active_vote_info)
	if isError(err, w, "Error writing vote data") {
		return
	}
	_, err = file.Write(active_vote_bytes)
	if isError(err, w, "Error writing vote data") {
		return
	}
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: vote_code,
		Data: active_vote_info.VoteData,
	}
	json.NewEncoder(w).Encode(response)
}

func keepActiveVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var err error
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_code string = url_fields[len(url_fields)-1]
	var now time.Time = time.Now().Local()
	_, err = os.Stat(filepath.Join("data", "active_votes", vote_code + ".json"))
	if os.IsNotExist(err) {
		log.Warning(err)
		var response ApiResponse = ApiResponse{
			Error: "ERROR",
			Message: "Invalid vote code",
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	err = os.Chtimes(filepath.Join("data", "active_votes", vote_code + ".json"), now, now)
	if isError(err, w, "Unable to keep vote alive") {
		return
	}
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: "OK",
		Data: nil,
	}
	json.NewEncoder(w).Encode(response)
}

func updateActiveVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var err error
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_code string = url_fields[len(url_fields)-1]
	var update_vote_byte_data []byte
	update_vote_byte_data, err = io.ReadAll(r.Body)
	if isError(err, w, "Error reading request") {
		return
	}
	var update_vote_request UpdateActiveVoteRequest
	err = json.Unmarshal(update_vote_byte_data, &update_vote_request)
	if isError(err, w, "Invalid vote data") {
		return
	}
	var active_vote_bytes []byte
	active_vote_bytes, err = os.ReadFile(filepath.Join("data", "active_votes", vote_code + ".json"))
	if isError(err, w, "Invalid vote code") {
		return
	}
	var active_vote_info ActiveVoteInfo
	err = json.Unmarshal(active_vote_bytes, &active_vote_info)
	if isError(err, w, "Unable to read active vote") {
		return
	}
	if active_vote_info.HostID != update_vote_request.HostID {
		var response ApiResponse = ApiResponse{
			Error: "ERROR",
			Message: "Invalid host id",
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	active_vote_info.VoteData = update_vote_request.VoteData
	var file *os.File
	file, err = os.Create(filepath.Join("data", "active_votes", vote_code + ".json"))
	if isError(err, w, "Inable to change active vote") {
		return
	}
	defer file.Close()
	active_vote_bytes, err = json.Marshal(active_vote_info)
	if isError(err, w, "Error writing vote data") {
		return
	}
	_, err = file.Write(active_vote_bytes)
	if isError(err, w, "Error writing vote data") {
		return
	}
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: "OK",
		Data: nil,
	}
	json.NewEncoder(w).Encode(response)
}

func joinVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var err error
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_code string = url_fields[len(url_fields)-1]
	var active_vote_bytes []byte
	active_vote_bytes, err = os.ReadFile(filepath.Join("data", "active_votes", vote_code + ".json"))
	if isError(err, w, "Invalid vote code") {
		return
	}
	var active_vote_info ActiveVoteInfo
	err = json.Unmarshal(active_vote_bytes, &active_vote_info)
	if isError(err, w, "Corrupted vote data, unable to proceed") {
		return
	}
	var join_vote_byte_data []byte
	join_vote_byte_data, err = io.ReadAll(r.Body)
	if isError(err, w, "Error reading request") {
		return
	}
	var client_request ClientRequest
	err = json.Unmarshal(join_vote_byte_data, &client_request)
	if isError(err, w, "Invalid request") {
		return
	}
	var now time.Time = time.Now()
	active_vote_info.Clients[client_request.ID] = now
	active_vote_info.VoteData.ClientCount += 1
	var file *os.File
	file, err = os.Create(filepath.Join("data", "active_votes", vote_code + ".json"))
	if isError(err, w, "Inable to change active vote") {
		return
	}
	defer file.Close()
	active_vote_bytes, err = json.Marshal(active_vote_info)
	if isError(err, w, "Error writing vote data") {
		return
	}
	_, err = file.Write(active_vote_bytes)
	if isError(err, w, "Error writing vote data") {
		return
	}
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: "OK",
		Data: active_vote_info.VoteData,
	}
	json.NewEncoder(w).Encode(response)
}

func sendVote(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " ", r.URL)
	set_headers(w)
	var err error
	var url_fields []string = strings.Split(r.URL.Path, "/")
	var vote_code string = url_fields[len(url_fields)-1]
	var active_vote_bytes []byte
	active_vote_bytes, err = os.ReadFile(filepath.Join("data", "active_votes", vote_code + ".json"))
	if isError(err, w, "Invalid vote code") {
		return
	}
	var active_vote_info ActiveVoteInfo
	err = json.Unmarshal(active_vote_bytes, &active_vote_info)
	if isError(err, w, "Corrupted vote data, unable to proceed") {
		return
	}
	var send_vote_byte_data []byte
	send_vote_byte_data, err = io.ReadAll(r.Body)
	if isError(err, w, "Error reading request") {
		return
	}
	var send_vote_request SendVoteRequest
	err = json.Unmarshal(send_vote_byte_data, &send_vote_request)
	if isError(err, w, "Invalid request") {
		return
	}
	var vote_items map[int]VoteItem = active_vote_info.VoteData.VoteItems
	if len(vote_items) >= send_vote_request.ItemID {
		var vote_item = vote_items[send_vote_request.ItemID]
		vote_item.Votes += 1
		active_vote_info.VoteData.VoteItems[send_vote_request.ItemID] = vote_item
	} else {
		var response ApiResponse = ApiResponse{
			Error: "ERROR",
			Message: "Vote Item not found",
			Data: nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	var file *os.File
	file, err = os.Create(filepath.Join("data", "active_votes", vote_code + ".json"))
	if isError(err, w, "Inable to change active vote") {
		return
	}
	defer file.Close()
	active_vote_bytes, err = json.Marshal(active_vote_info)
	if isError(err, w, "Error writing vote data") {
		return
	}
	_, err = file.Write(active_vote_bytes)
	if isError(err, w, "Error writing vote data") {
		return
	}
	var response ApiResponse = ApiResponse{
		Error: "OK",
		Message: "OK",
		Data: nil,
	}
	json.NewEncoder(w).Encode(response)
}
