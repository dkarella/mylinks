package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/dkarella/mylinks/mylinks"
)

const fileName = "links.csv"

var setlinkTemplate = template.Must(template.New("setlink").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MyLinks</title>
</head>
<body>
	<div>{{.NotFoundMessage}}</div>
    <form action="/setlink" method="post">
		<input type="text" id="key" name="key" placeholder="key" value="{{.NotFoundKey}}"/>
		<br/>
		<br/>
		<input type="text" id="value" name="value" placeholder="url"/><br><br>
		<br/>
		<br/>
		<input type="submit" value="Submit"/>
    </form>
</body>
</html>
`))

type setLinkValues struct {
	NotFoundMessage string
	NotFoundKey     string
}

var links mylinks.T

func main() {
	if err := links.Load(fileName); err != nil {
		log.Fatal(fmt.Errorf("failed to load links: %s", err))
	}
	defer links.Close()

	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/setlink", setLinkHandler)
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
	links.RLock()
	defer links.RUnlock()

	key := r.URL.Path[1:]
	redirect := fmt.Sprintf("/setlink?not_found=%s", key)
	if l, ok := links.Get(key); ok {
		redirect = l
	}

	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

func setLinkHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		setLinkPostHandler(w, r)
	case http.MethodGet:
		setLinkGetHandler(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("invalid method"))
	}
}

func setLinkGetHandler(w http.ResponseWriter, r *http.Request) {
	var notFoundMessage string
	notFoundKey := r.URL.Query().Get("not_found")
	if notFoundKey != "" {
		notFoundMessage = "That key doesn't exist yet, add it now!"
	}

	w.WriteHeader(http.StatusOK)
	if err := setlinkTemplate.Execute(w, setLinkValues{
		NotFoundKey:     notFoundKey,
		NotFoundMessage: notFoundMessage,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}

func setLinkPostHandler(w http.ResponseWriter, r *http.Request) {
	links.Lock()
	defer links.Unlock()

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	key := r.PostForm.Get("key")
	value := r.PostForm.Get("value")
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

	http.Redirect(w, r, value, http.StatusTemporaryRedirect)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
