package website

import (
	"net/http"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	indexHTML, err := Templates.ReadFile("templates/index.html")
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Write([]byte(indexHTML))
}
