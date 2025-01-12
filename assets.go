package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetUrl(assetpath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetpath)
}

func getAssetPath(mediaType string) string {
	base := make([]byte, 32)

	_, err := rand.Read(base)

	if err != nil {
		panic("Failed to generate name")
	}

	id := base64.RawURLEncoding.EncodeToString(base)

	ext := mediatypeToExt(mediaType)
	return fmt.Sprintf("%s%s", id, &ext)
}

func mediatypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}
