package api

import "net/http"

type APIHandler struct {
	mux *http.ServeMux
}

func CreateAPIHandler() *APIHandler {
	return &APIHandler{
		mux: &http.ServeMux{},
	}
}

func (a *APIHandler) AddEndpoint(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	a.mux.HandleFunc(pattern, handler)
}

func (a *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
