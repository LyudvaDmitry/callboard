package main

import (
	"net/http"
	"github.com/LyudvaDmitry/test_repository/callboard"
	"log"
)

func main() {
	http.Handle("/board1/", callboard.NewCallboard())
	cb := callboard.NewCallboard()
	http.Handle("/board2/", cb)
	cb.Adverts = append(cb.Adverts, callboard.Advert{"MyTitle", "Greetings from testcallboard.go", "User", "Time"})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
