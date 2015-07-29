package main

import (
    "fmt"
    "net/http"
    "time"
    "log"
)

func main() {
    http.HandleFunc("/time", func (w http.ResponseWriter, r *http.Request) {
	    if _, err := fmt.Fprint(w, time.Now().UTC()); err != nil {
	        log.Fatal(err)
	    }
	} )
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}