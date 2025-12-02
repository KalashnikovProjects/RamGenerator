package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	MaxFileSize int // in kilobytes. For no max size -1
	UploadToken string
	RootPath    string
	ImageDir    string // "./uploads"
}

var Conf *Config

func initConfig() error {
	rootPath, valid := os.LookupEnv("ROOT_PATH")
	if !valid {
		return fmt.Errorf("required env var IMAGE_CDN_TOKEN is empty")
	}
	_ = godotenv.Load(fmt.Sprintf("%s/.env", rootPath))
	token, valid := os.LookupEnv("IMAGE_CDN_API_KEY")
	if !valid {
		return fmt.Errorf("required env var IMAGE_CDN_TOKEN is empty")
	}
	Conf = &Config{
		MaxFileSize: 32 * 1024,
		UploadToken: token,
		RootPath:    rootPath,
		ImageDir:    filepath.Join(rootPath, "cdn", "uploads"),
	}
	imageDir, valid := os.LookupEnv("IMAGE_CDN_DIR")
	if valid {
		Conf.ImageDir = imageDir
	}
	return nil
}

func main() {
	if err := initConfig(); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(Conf.ImageDir, 0755); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/img/", imageHandler)

	fmt.Println("CDN running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkToken(r *http.Request) bool {
	auth := r.Header.Get("Authorization")

	if auth != "" {
		parts := strings.Fields(auth)

		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] == Conf.UploadToken {
			return true
		}
	}
	return false
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("uploading image")

	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkToken(r) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="upload"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "body empty or read error", http.StatusBadRequest)
		return
	}

	data, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		http.Error(w, "invalid base64", http.StatusBadRequest)
		return
	}

	ct := r.Header.Get("Content-Type")
	ext := ""
	switch ct {
	case "image/jpeg":
		ext = ".jpg"
	case "image/jpg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	default:
		ext = ".png"
		return
	}

	id := uuid.New().String()
	filename := id + ext
	path := fmt.Sprintf("%s/%s", Conf.ImageDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}

	url := "/img/" + filename
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(url))
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	path := filepath.Join(Conf.ImageDir, filename)

	fmt.Println("get image: ", path)
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	etag := genETag(filename, stat.ModTime())

	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, etag) || match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	if since := r.Header.Get("If-Modified-Since"); since != "" {
		if t, err := http.ParseTime(since); err == nil {
			if stat.ModTime().Before(t.Add(1 * time.Second)) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "cannot read", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	ct := mime.TypeByExtension(filepath.Ext(filename))
	if ct == "" {
		ct = "application/octet-stream"
	}

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", stat.ModTime().UTC().Format(http.TimeFormat))

	http.ServeContent(w, r, filename, stat.ModTime(), f)
}

func genETag(name string, t time.Time) string {
	h := sha1.Sum([]byte(name + t.String()))
	return `"` + hex.EncodeToString(h[:]) + `"`
}
