package main

import (
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)

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

	file, handler, err := r.FormFile("video")

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to parse", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(handler.Header.Get("Content-Type"))

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt parse media type", err)
		return
	}

	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "not video type", nil)
		return
	}

	tmpfile, err := os.CreateTemp("", "tubely-upload.mp4")

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt create temp file", err)
		return
	}

	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := io.Copy(tmpfile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not copy the file", err)
		return
	}

	_, err = tmpfile.Seek(0, io.SeekStart)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt get to the beginning of the file", err)
		return
	}
	key := getAssetPath(mediaType)
	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.s3Bucket),
		Key:         aws.String(key),
		Body:        tmpfile,
		ContentType: aws.String(mediaType),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error uploading file to s3", err)
		return
	}

	url := cfg.getObjectUrl(key)

	videoData.VideoURL = &url
	err = cfg.db.UpdateVideo(videoData)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoData)

}
