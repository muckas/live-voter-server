package main

type ApiResponse struct {
	Error   string   `json:"error"`
	Message string   `json:"message"`
	Data    struct{} `json:data`
}
