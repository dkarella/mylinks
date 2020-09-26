package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const fileName = "links.csv"

type MyLinks struct {
	file  *os.File
	links map[string]string
}

var links MyLinks

func main() {
	if err := links.Load(); err != nil {
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
	redirect := "/404"
	if l, ok := links.links[r.URL.Path[1:]]; ok {
		redirect = l
	}

	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

func setLinkHandler(w http.ResponseWriter, r *http.Request) {
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

func (l *MyLinks) Load() error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			file.Close()
		}
	}()

	l.file = file
	l.links = make(map[string]string)

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ",")

		if len(tokens) != 2 {
			return fmt.Errorf("invalid line: %s", line)
		}

		l.links[tokens[0]] = tokens[1]
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (l *MyLinks) Set(key, value string) error {
	if _, err := l.file.WriteString(fmt.Sprintf("\n%s,%s", key, value)); err != nil {
		return err
	}
	l.links[key] = value
	return nil
}

func (l *MyLinks) Close() {
	l.file.Close()
}
