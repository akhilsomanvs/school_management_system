package utils

import "net/http"

type MiddlewareFunc func(http.Handler) http.Handler

func ApplyMiddleWares(mux http.Handler, middlewares []MiddlewareFunc) http.Handler {
	var middleWare http.Handler = mux
	for _, mwFunc := range middlewares {
		middleWare = mwFunc(middleWare)
	}
	return middleWare
}
