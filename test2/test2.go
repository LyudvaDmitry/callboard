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

//html или javascript в объявлении?
//Добавить многопоточность

var templates = template.Must(template.ParseFiles("add.html", "adv.html"))
var adv_list []Advert

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting template execution")
	for _, adv := range adv_list {
		if err := templates.ExecuteTemplate(w, "adv.html", adv); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := templates.ExecuteTemplate(w, "add.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Successfully executed templates")
}

func AddHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Starting adding advert '%v'", r.FormValue("title"))
	adv_list = append(adv_list, Advert{r.FormValue("title"), r.FormValue("body")})
	http.Redirect(w, r, "/view", http.StatusFound)
	log.Printf("Added advert '%v'", r.FormValue("title"))
}

func DelHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting deleting all adverts")
	adv_list = make([]Advert, 0)
	http.Redirect(w, r, "/view", http.StatusFound)
	log.Println("Deleted all adverts")
}

func main() {
	http.HandleFunc("/view", ViewHandler)
	http.HandleFunc("/add", AddHandler)
	http.HandleFunc("/deleteall", DelHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}