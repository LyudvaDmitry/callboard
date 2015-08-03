package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type Advert struct {
	Title    string
	Body     string
	Username string
}

var (
	templates = template.Must(template.ParseFiles("view.html"))
	adv_list  = make([]Advert, 0)
	mutex     sync.Mutex
	port      = flag.String("port", ":8080", "Number of port to use. Example: ':8080'")
)

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	username, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="callboard"`)
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Header.Get("Accept") == "application/json" {
		log.Printf("Sending JSON to %s", username)
		b, err := json.Marshal(adv_list)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(b); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	title, button := r.FormValue("title"), r.FormValue("button")
	if button != "" {
		mutex.Lock()
		switch button {
		case "Save":
			log.Printf("Adding advert '%v' as %s", title, username)
			adv_list = append(adv_list, Advert{title, r.FormValue("body"), username})
		case "DeleteAll":
			log.Printf("Deleting all adverts as %s", username)
			adv_list = make([]Advert, 0)
		}
		mutex.Unlock()
		http.Redirect(w, r, "/view", http.StatusFound)
		return
	}
	log.Println("Executing templates")
	w.Header().Set("Content-Type", "text/html")
	if err := templates.ExecuteTemplate(w, "view.html", adv_list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	flag.Parse()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/view", ViewHandler)
	//log.Fatal(http.ListenAndServe(*port, nil))
	log.Fatal(http.ListenAndServeTLS(*port, "cert.pem", "key.pem", nil))
}
