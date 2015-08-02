package main

import (
	"net/http"
	"html/template"
	"log"
	"sync"
	"encoding/base64"
	"strings"
	"flag"
	"encoding/json"
)

type Advert struct {
	Title string
	Body  string
	Username string
}

var (
	templates = template.Must(template.ParseFiles("view.html"))
	adv_list = make([]Advert, 0)
	mutex sync.Mutex
	fs = http.FileServer(http.Dir(""))
	port = flag.String("port", ":8080", "Number of port to use. Example: ':8080'")
)

//func (*Request) BasicAuth
//Нашел ее случайно и слишком поздно
func GetUsername(r *http.Request) string {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		return ""
	}
	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return ""
	}
	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return ""
	}
	return pair[0]
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	username := GetUsername(r)
	if username == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="callboard"`)
		w.WriteHeader(401)
		if _, err := w.Write([]byte("401 Unauthorized\n")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	if (r.Header.Get("Content-Type") == "application/json") {
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
	http.Handle("/static/", fs)
	http.HandleFunc("/view", ViewHandler)
	//log.Fatal(http.ListenAndServe(*port, nil))
	log.Fatal(http.ListenAndServeTLS(*port, "cert.pem", "key.pem", nil))
}