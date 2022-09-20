package main

type ApiResponse struct {
	Error   string   `json:"error"`
	Message string   `json:"message"`
	Data    any `json:data`
}

type ApiHostVoteRequest struct {
	VoteName string `json:"vote_name"`
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
	VoteName string `json:"vote_name"`
	PageName string `json:"page_name"`
	VoteItems map[int]VoteItem `json:"vote_items"`
}
