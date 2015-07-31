package main

import (
	"net/http"
	"html/template"
	"log"
	"sync"
)

type Advert struct {
	Title string
	Body  string
}

var templates = template.Must(template.ParseFiles("view.html"))
var adv_list []Advert
var mutex = sync.Mutex{}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	title, button := r.FormValue("title"), r.FormValue("button")
	if button == "Save" {
		mutex.Lock()
		log.Printf("Adding advert '%v'", title)
		adv_list = append(adv_list, Advert{title, r.FormValue("body")})
		mutex.Unlock()
	} else if button == "DeleteAll" {
		mutex.Lock()
		log.Println("Deleting all adverts")
		adv_list = make([]Advert, 0)
		mutex.Unlock()
	}
	log.Println("Executing templates")
	if err := templates.ExecuteTemplate(w, "view.html", adv_list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	adv_list = make([]Advert, 0)
	http.HandleFunc("/view", ViewHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}