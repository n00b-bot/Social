package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func respond(w http.ResponseWriter, v interface{}, statusCode int) {
	b, err := json.Marshal(v)
	if err != nil {
		respondError(w, err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func respondError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func writeSSE(w io.Writer, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "cound foramt %v\n", err)
		return
	}
	fmt.Fprintf(w, "data:%s\n\n", b)
}
