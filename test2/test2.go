package main

import (
	"flag"
	"github.com/LyudvaDmitry/test_repository/callboard"
	"log"
	"net/http"
)

var port = flag.String("port", ":8080", "Number of port to use. Example: ':8080'")

func main() {
	flag.Parse()
	http.Handle("/", callboard.NewCallboard("view_bootstrap.html"))
	log.Fatal(http.ListenAndServe(*port, nil))
	//log.Fatal(http.ListenAndServeTLS(*port, "cert.pem", "key.pem", nil))
}
