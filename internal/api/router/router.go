package router

import (
	"net/http"
	"simpleapi/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)

	//Teachers Handler
	mux.HandleFunc("GET /teachers/", handlers.GetTeachersHandler)
	mux.HandleFunc("POST /teachers/", handlers.AddTeacherHandler)
	mux.HandleFunc("PATCH /teachers/", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers/", handlers.DeleteTeachersHandler)

	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}", handlers.GetTeachersHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeachersHandler)

	mux.HandleFunc("/students/", handlers.StudentHandler)

	mux.HandleFunc("/execs/", handlers.ExecsHandler)
	return mux
}
