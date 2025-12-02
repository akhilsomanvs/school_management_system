package router

import (
	"net/http"
	"simpleapi/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/teachers", handlers.TeachersHandler)
	mux.HandleFunc("/students/", handlers.StudentHandler)

	mux.HandleFunc("/execs/", handlers.ExecsHandler)
	return mux
}
