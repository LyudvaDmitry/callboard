package main

import (
    "fmt"
    "net/http"
    "time"
    "log"
)

func handler(w http.ResponseWriter, r *http.Request) {
    _, err := fmt.Fprint(w, time.Now().UTC())
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    http.HandleFunc("/time", handler)
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal(err)
    }
}