package main

import (
	"log"
	"net/http"
	"time"
)

type HandlerFunc = func(http.ResponseWriter, *http.Request)

func Logging(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println("START:", r.Method, r.URL.Path)
		next(w, r)
		log.Println("END:", r.Method, r.URL.Path, time.Since(start))
	}
}
