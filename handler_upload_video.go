package main

import (
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	closer := http.MaxBytesReader(w, r.Body, 1<<30)

	videoIDString := r.PathValue("videoID")
	vidoID, err := uuid.Parse(videoIDString)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt parse video id", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldnt get token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldnt validate jwt", err)
		return
	}

	videoData, err := cfg.db.GetVideo(vidoID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "no vid in db", err)
		return
	}

	if userID != videoData.UserID {
		respondWithError(w, http.StatusUnauthorized, "not owner of vid", err)
		return
	}

	file, header, err := r.FormFile("video")

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to parse", err)
		return
	}
	defer file.Close()

}
