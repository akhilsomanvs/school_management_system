package handlers

import "net/http"

func StudentHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Hello root Route")
	w.Write([]byte("Hello students route"))
}
