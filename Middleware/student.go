package main

import "net/http"

type StudentHandlers struct{}

func NewStudentHandlers() *StudentHandlers {
	return &StudentHandlers{}
}

func (h *StudentHandlers) CreateStudent(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	age := r.URL.Query().Get("age")

	if name == "" || age == "" {
		w.Write([]byte("Give name and age"))
		return
	}

	w.Write([]byte("Student: " + name + ", age: " + age))
}