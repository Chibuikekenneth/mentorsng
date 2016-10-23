package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/adesokanayo/mentorsng/views"
	"github.com/gorilla/mux"
)

var homeTemplate *template.Template
var contactTemplate *template.Template

var homeView views.View
var contactView *views.View

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	//homeView := views.NewView("bootstrap", "views/home.gohtml")
	//contactView := views.NewView("bootstrap", "views/contact.gohtml")

	var err error
	homeTemplate, err = template.ParseFiles(
		"views/home.gohtml",
		"views/layouts/footer.gohtml")
	if err != nil {
		panic(err)
	}

	contactTemplate, err = template.ParseFiles(
		"views/contact.gohtml",
		"views/layouts/footer.gohtml")
	if err != nil {
		panic(err)
	}

	Routes := mux.NewRouter()
	Routes.HandleFunc("/home", home)

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))

	http.Handle("/s", Routes)
	http.ListenAndServe(":"+port, nil)
	log.Println("Listening...")
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	homeTemplate.ExecuteTemplate(w, homeView.Layout, nil)

}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	contactTemplate.ExecuteTemplate(w, contactView.Layout, nil)
}
