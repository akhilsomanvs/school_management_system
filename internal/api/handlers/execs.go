package handlers

import "net/http"

func ExecsHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Hello root Route")
	w.Write([]byte("Hello Execs route"))
}
