//Package callboard implements simple callboard realisation.
package callboard

import (
	"encoding/json"
	"html/template"
	"log"
	"mime"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Advert contains all the information about advert.
type Advert struct {
	Title    string
	Body     string
	Username string
	Time     string
}

//Callboard implements http.Header interface. It creates callboard
//available at the given URL.
type Callboard struct {
	Adverts  []Advert
	mutex    sync.Mutex
	template *template.Template
}

//NewCallboard returns initialized Callboard.
func NewCallboard() Callboard {
	adverts := make([]Advert, 1)
	adverts[0] = Advert{"Sample title", "It's a sample Advert. Delete it using 'Delete all adverts' button.", "Callboard", time.Now().Format(time.RFC850)}
	return Callboard{Adverts: adverts, template: template.Must(template.ParseFiles(html))}
}

const html = "view_bootstrap.html"

func (c Callboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/view") {
		c.viewHandler(w, r)
	}
	matched, err := regexp.MatchString(`.*/static/[^/]*\.(css|jpg|png)`, r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if matched {
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
	}
}

func ifMIMETypePreferred(r *http.Request, mtype string) (bool, error) {
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

func (c Callboard) viewHandler(w http.ResponseWriter, r *http.Request) {
	username, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="callboard"`)
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
		return
	}
	JSON, err := ifMIMETypePreferred(r, "application/json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if JSON {
		log.Printf("Sending JSON to %s", username)
		b, err := json.Marshal(c.Adverts)
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
		c.mutex.Lock()
		switch button {
		case "Save":
			log.Printf("Adding advert '%v' as %s", title, username)
			c.Adverts = append(c.Adverts, Advert{title, r.FormValue("body"), username, time.Now().Format(time.RFC850)})
		case "DeleteAll":
			log.Printf("Deleting all c.Adverts as %s", username)
			c.Adverts = make([]Advert, 0)
		}
		c.mutex.Unlock()
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
		return
	}
	log.Println("Executing templates")
	w.Header().Set("Content-Type", "text/html")
	if err := c.template.ExecuteTemplate(w, html, c.Adverts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
