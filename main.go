package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	Routes := mux.NewRouter()
	Routes.HandleFunc("/home", home)

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))

	http.Handle("/home", Routes)
	http.ListenAndServe(":"+port, nil)
	log.Println("Listening now...")
}

func home(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")
	log.Println("called something")

}
