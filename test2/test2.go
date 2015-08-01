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
var adv_list = make([]Advert, 0)
var mutex sync.Mutex
var fs = http.FileServer(http.Dir(""))

func ViewHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
	title, button := r.FormValue("title"), r.FormValue("button")
	if button != "" {
		mutex.Lock()
		switch button {
		case "Save":
			log.Printf("Adding advert '%v'", title)
			adv_list = append(adv_list, Advert{title, r.FormValue("body")})
		case "DeleteAll":
			log.Println("Deleting all adverts")
			adv_list = make([]Advert, 0)
		}
		mutex.Unlock()
		http.Redirect(w, r, "/view", http.StatusFound)
		return
	}
	log.Println("Executing templates")
	if err := templates.ExecuteTemplate(w, "view.html", adv_list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/view", ViewHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}