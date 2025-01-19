package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	http.MaxBytesReader(w, r.Body, 1<<30)

}
