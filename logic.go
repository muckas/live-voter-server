package main

import (
	"errors"
	"time"
	"os"
	"io/fs"
	"path/filepath"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"live-voter-server/log"
)

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
	var file_info fs.FileInfo
	var file_time time.Time
	var now time.Time = time.Now()
	for index < len(dir_content) {
		file_info, err = dir_content[index].Info()
		file_time = file_info.ModTime()
		if file_time.Add(vote_lifetime).Before(now) { // if modification time is more then vote_lifetime ago
			err = os.RemoveAll(filepath.Join(votes_dir, file_info.Name()))
			if err != nil {
				log.Error(err)
				return 0, err
			}
			log.Debug("Removed inactive vote: ", file_info.Name())
			dir_content = append(dir_content[:index], dir_content[index+1:]...)
		} else {
			index++
		}
	}
	log.Debug("Active votes num: ", len(dir_content))
	return len(dir_content), nil
}

func startNewVote(host_id string, vote_name string) (string, error) {
	if host_id == "" {
		return "", errors.New("Invalid host id")
	}
	var max_votes int = 100
	var vote_lifetime time.Duration = 10 * time.Minute
	var votes_dir string = filepath.Join("data", "active_votes")
	var code string
	var err error
	var num_active_votes int
	num_active_votes, err = voteCleanup(votes_dir, vote_lifetime)
	if err != nil {
		log.Error(err)
		return "", errors.New("Error creating a vote")
	}
	if num_active_votes >= max_votes {
		log.Info("Max active votes exceeded (max ", max_votes, ")")
		return "", errors.New("Max active votes exceeded")
	}
	for {
		code, err = generateNewVoteCode()
		_, err = os.Stat(filepath.Join(votes_dir, code + ".json"))
		if os.IsNotExist(err) {
			break
		}
	}
	var vote_info ActiveVoteInfo = ActiveVoteInfo {
		HostID: host_id,
		Clients: []ActiveVoteClient{},
		VoteData: ActiveVoteData {
			State: Intro,
			ClientCount: 0,
			VoteName: vote_name,
			VoteItems: map[int]VoteItem{},
		},
	}
	var vote_byte_data []byte
	vote_byte_data, err = json.Marshal(vote_info)
	if err != nil {
		log.Error(err)
		return "", errors.New("Error creating a vote")
	}
	var file *os.File
	file, err = os.Create(filepath.Join(votes_dir, code + ".json"))
	if err != nil {
		log.Error(err)
		return "", errors.New("Error creating a vote")
	}
	defer file.Close()
	file.Write(vote_byte_data)
	log.Debug("Created active vote: ", code)
	return code, nil
}
