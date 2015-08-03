package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"sync"
    "time"
	"strings"
	"mime"
	"strconv"
)

const html = "view_bootstrap.html"

type Advert struct {
	Title    string
	Body     string
	Username string
	Time 	 string
}

var (
	templates = template.Must(template.ParseFiles(html))
	adv_list  = make([]Advert, 0)
	mutex     sync.Mutex
	port      = flag.String("port", ":8080", "Number of port to use. Example: ':8080'")
)

//Я уже сожалею, что я это написал.
//Буду рад узнать, какая стандартная функция это делает.
func IfMIMETypePreferred(r *http.Request, mtype string) (bool, error) {
	log.Println(r.Header["Accept"])
	accept := make([]string, 0)
	for _, val := range r.Header["Accept"] {
		accept = append(accept, strings.Split(val, ",")...)
	}
	mediatypes := make(map[string]map[string]string)
	for _, val := range accept {
		name, parameters, err := mime.ParseMediaType(val)
		if err != nil {
			return false, err
		}
		if _, present := parameters["q"]; !present {
			parameters["q"] = "1"
		}
		mediatypes[name] = parameters
	}
	if mediatypes[mtype]["q"] == "" {
		return false, nil
	}
	max := 0.0
	for _, val := range mediatypes {
		num, err := strconv.ParseFloat(val["q"], 32)
		if err != nil {
			return false, err
		}
		if num > max {
			max = num
		}
	}
	num, err := strconv.ParseFloat(mediatypes[mtype]["q"], 32)
	if err != nil {
		return false, err
	}
	if num == max {
		return true, nil
	}
	return false, nil
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	username, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="callboard"`)
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
		return
	}
	JSON, err := IfMIMETypePreferred(r, "application/json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if JSON {
		log.Printf("Sending JSON to %s", username)
		b, err := json.Marshal(adv_list)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(b); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	title, button := r.FormValue("title"), r.FormValue("button")
	if button != "" {
		mutex.Lock()
		switch button {
		case "Save":
			log.Printf("Adding advert '%v' as %s", title, username)
			adv_list = append(adv_list, Advert{title, r.FormValue("body"), username, time.Now().Format(time.RFC850)})
		case "DeleteAll":
			log.Printf("Deleting all adverts as %s", username)
			adv_list = make([]Advert, 0)
		}
		mutex.Unlock()
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
		return
	}
	log.Println("Executing templates")
	w.Header().Set("Content-Type", "text/html")
	if err := templates.ExecuteTemplate(w, html, adv_list); err != nil {
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
