package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/dkarella/mylinks/mylinks"
)

const fileName = "links.csv"

var links mylinks.T
var lock sync.RWMutex

func main() {
	if err := links.Load(fileName); err != nil {
		log.Fatal(fmt.Errorf("failed to load links: %s", err))
	}
	defer links.Close()

	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/setlink", setLinkHandler)
	http.HandleFunc("/404", notFoundHandler)
	http.HandleFunc("/", redirectHandler)

	port := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	lock.RLock()
	defer lock.RUnlock()

	redirect := "/404"
	if l, ok := links.Get(r.URL.Path[1:]); ok {
		redirect = l
	}

	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

func setLinkHandler(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()

	q := r.URL.Query()
	key := q.Get("key")
	value := q.Get("value")

	if key == "" || value == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("key and value query parameters are required"))
		return
	}

	if err := links.Set(key, value); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("unexpected error occurred: %s", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
