package main

import (
	"fmt"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage, yo!")
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequests()
}
