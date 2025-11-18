package internal

import (
	"net/http"
)

func NewRouter(submitHandler http.Handler, reportHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/submit", submitHandler)
	mux.Handle("/report", reportHandler)
	return mux
}
