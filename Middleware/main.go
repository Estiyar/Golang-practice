package main

import (
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/hello", Logging(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World"))
		},
	))

	studentHandlers := NewStudentHandlers()

	http.HandleFunc("/student", Logging(studentHandlers.CreateStudent))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
