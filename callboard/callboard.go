//Package callboard implements simple callboard realisation.
package callboard

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"mime"
	"net/http"
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

//Callboard implements http.Handler interface. It creates callboard
//available at the given URL.
type Callboard struct {
	Adverts []Advert
	html    string
	sync.Mutex
	*template.Template
}

//NewCallboard returns initialized Callboard.
func NewCallboard(html_temp string) *Callboard {
	return &Callboard{Adverts: make([]Advert, 0), Template: template.Must(template.ParseFiles(html_temp)), html: html_temp}
}

//ServeHTTP reads Request and writes reply data to ResponseWriter.
//It is required by http.Handler interface.
func (c *Callboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := c.viewHandler(w, r); err != nil {
		log.Println(err.Error)
		http.Error(w, err.Error.Error(), err.Code)
	}
}

//appError used to return both error and http error code.
type appError struct {
	Error error
	Code  int
}

func (c *Callboard) viewHandler(w http.ResponseWriter, r *http.Request) *appError {
	//Authorization check
	username, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="callboard"`)
		return &appError{errors.New("401 Unauthorized"), http.StatusUnauthorized}
	}
	//Checks if JSON required
	JSON, err := ifMIMETypePreferred(r, "application/json")
	if err != nil {
		return &appError{err, http.StatusInternalServerError}
	}
	if JSON {
		log.Printf("Sending JSON to %s", username)
		b, err := json.Marshal(c.Adverts)
		if err != nil {
			return &appError{err, http.StatusInternalServerError}
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(b); err != nil {
			return &appError{err, http.StatusInternalServerError}
		}
		return nil
	}
	//Checks if adverts adding or deleting requested
	if title, button := r.FormValue("title"), r.FormValue("button"); button != "" {
		c.Lock()
		switch button {
		case "Save":
			log.Printf("Adding advert '%v' as %s", title, username)
			c.Adverts = append(c.Adverts, Advert{title, r.FormValue("body"), username, time.Now().Format(time.RFC850)})
		case "DeleteAll":
			log.Printf("Deleting all c.Adverts as %s", username)
			c.Adverts = make([]Advert, 0)
		}
		c.Unlock()
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
		return nil
	}
	//Shows callboard
	log.Println("Executing templates")
	w.Header().Set("Content-Type", "text/html")
	if err := c.ExecuteTemplate(w, c.html, c.Adverts); err != nil {
		return &appError{err, http.StatusInternalServerError}
	}
	return nil
}

//ifMIMETypePreferred checks if given MIME type is one of the most preferred
//by given http.Request.
func ifMIMETypePreferred(r *http.Request, mtype string) (bool, error) {
	accept := make([]string, 0)
	//While Header has map[string][]string type, for some reason Header["accept"]
	//sometimes contains one string with all MIME types accepted. So it is to be
	//parsed.
	for _, val := range r.Header["Accept"] {
		accept = append(accept, strings.Split(val, ",")...)
	}
	//Accept may also contain "qvalue" indicating a relative preference for given
	//media-range. It needs to be parsed too.
	mediatypes := make(map[string]map[string]string)
	for _, val := range accept {
		name, parameters, err := mime.ParseMediaType(val)
		if err != nil {
			return false, err
		}
		//Qvalue default value is 1
		if _, present := parameters["q"]; !present {
			parameters["q"] = "1"
		}
		mediatypes[name] = parameters
	}
	//If given MIME type is not present, it is not preferred.
	if mediatypes[mtype]["q"] == "" {
		return false, nil
	}
	//Looking for max qvalue
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
	//Checking if given MIME type qvalue equal to max qvalue
	num, err := strconv.ParseFloat(mediatypes[mtype]["q"], 32)
	if err != nil {
		return false, err
	}
	if num == max {
		return true, nil
	}
	return false, nil
}
