package handlers

import "net/http"

func RootHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Hello root Route")
	w.Write([]byte("Hello root route"))
}
