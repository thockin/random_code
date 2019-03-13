package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/good", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("good")
		fmt.Fprintf(w, "GOOD")
	})
	http.HandleFunc("/redirect-to-good", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("redirect-to-good")
		http.Redirect(w, r, "/good", http.StatusFound)
	})
	http.HandleFunc("/redirect-to-bad", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("redirect-to-bad")
		http.Redirect(w, r, "/bad", http.StatusFound)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
