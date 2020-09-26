package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
)

const fileName = "links.csv"

var links map[string]string = make(map[string]string)

func main() {
	load()

	http.HandleFunc("/health", healthCheckHandler)
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
	if l, ok := links[r.URL.Path[1:]]; ok {
		redirect = l
	}

	http.Redirect(w, r, redirect, 302)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func load() {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ",")

		if len(tokens) != 2 {
			log.Fatalf("failed to load line: %s", line)
		}

		links[tokens[0]] = tokens[1]
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
