package main

import (
	"net/http"
	"html/template"
	"log"
)

type Advert struct {
	Title string
	Body  string
//	Body  []byte
}

var templates = template.Must(template.ParseFiles("view.html"))
var adv_list []Advert

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("button") == "Save" {
		log.Printf("Starting adding advert '%v'", r.FormValue("title"))
		adv_list = append(adv_list, Advert{r.FormValue("title"), r.FormValue("body")})
		log.Printf("Added advert '%v'", r.FormValue("title"))
	} else if r.FormValue("button") == "Delete all adverts" {
		log.Println("Starting deleting all adverts")
		adv_list = make([]Advert, 0)
		log.Println("Deleted all adverts")
	}
	log.Println("Starting template execution")
	if err := templates.ExecuteTemplate(w, "view.html", adv_list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Successfully executed templates")
}

func main() {
	adv_list = make([]Advert, 0)
	http.HandleFunc("/view", ViewHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}