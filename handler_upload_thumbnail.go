package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	const maxMemory int64 = 10 << 20

	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse form file", err)
		return
	}

	defer file.Close()

	mediaType := header.Header.Get("Content-Type")

	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content-type for thumbnail", nil)
		return
	}

	assetPath := getAssetPath(mediaType)
	assetdiskPath := cfg.getAssetDiskPath(assetPath)

	dst, err := os.Create(assetdiskPath)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create file on server", err)
		return
	}

	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt copy contents", err)
	}

	video, err := cfg.db.GetVideo(videoID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldnt find video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not the owner of the vid", nil)
		return
	}

	url := cfg.getAssetUrl(assetPath)
	video.ThumbnailURL = &url

	err = cfg.db.UpdateVideo(video)

	if err != nil {
		delete(videoThumbnails, videoID)

		respondWithError(w, http.StatusBadRequest, "couldnt update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
