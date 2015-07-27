package main

import (
    "fmt"
    "net/http"
    "time"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, time.Now().UTC())
}

func main() {
    http.HandleFunc("/time", handler)
    http.ListenAndServe(":8080", nil)
}