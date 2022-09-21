package main

import (
	"time"
)

type ApiResponse struct {
	Error   string   `json:"error"`
	Message string   `json:"message"`
	Data    any `json:"data"`
}

type ApiHostVoteRequest struct {
	HostID string `json:"host_id"`
	VoteName string `json:"vote_name"`
}

type UpdateActiveVoteRequest struct {
	HostID string `json:"host_id"`
	VoteData ActiveVoteData `json:"vote_data"`
}

type SendVoteRequest struct {
	ID string `json:"id"`
	ItemID int `json:"item_id"`
}

type ClientRequest struct {
	ID string `json:"id"`
}

type VoteState string
const (
	Intro VoteState = "intro"
	Presenting = "presenting"
	Voting = "voting"
	Outro = "outro"
)

type VoteItem struct {
	Name string `json:"item_name"`
	Votes int `json:"item_votes"`
}

type ActiveVoteData struct {
	State VoteState `json:"state"`
	ClientCount int `json:"client_count"`
	VoteName string `json:"vote_name"`
	PageName string `json:"page_name"`
	VoteItems map[int]VoteItem `json:"vote_items"`
}

type ActiveVoteInfo struct {
	HostID string `json:"host_id"`
	Clients map[string]time.Time `json:"clients"`
	VoteData ActiveVoteData `json:"vote_data"`
}
