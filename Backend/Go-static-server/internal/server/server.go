package server

import (
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Go-static-server/internal/config"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

type TemplateData struct {
	Title         string
	ApiUrl        string
	DefaultAvatar string
}

type ErrorTemplateData struct {
	ErrorText string
	TemplateData
}

func Run() {
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(ErrorPage404)

	router.HandleFunc("/", Index)
	router.HandleFunc("/users/{username}", User)

	router.HandleFunc("/login", Login)

	router.HandleFunc("/favicon.ico", faviconHandler)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.Conf.CdnFilesPath))))

	fmt.Printf("Static server started at http://localhost:%d\n", config.Conf.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Conf.Port), router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles(fmt.Sprintf("%s/%s", config.Conf.TemplatesPath, "base.html"), fmt.Sprintf("%s/%s", config.Conf.TemplatesPath, "index.html"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, TemplateData{Title: fmt.Sprintf("Ram Generator"), ApiUrl: config.Conf.ApiUrl, DefaultAvatar: config.Conf.DefaultAvatar})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}

func User(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles(fmt.Sprintf("%s/%s", config.Conf.TemplatesPath, "base.html"), fmt.Sprintf("%s/%s", config.Conf.TemplatesPath, "user.html"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	err = ts.Execute(w, TemplateData{Title: fmt.Sprintf("%s - Ram Generator", vars["username"]), ApiUrl: config.Conf.ApiUrl, DefaultAvatar: config.Conf.DefaultAvatar})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles(fmt.Sprintf("%s/%s", config.Conf.TemplatesPath, "base.html"), fmt.Sprintf("%s/%s", config.Conf.TemplatesPath, "login.html"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, TemplateData{Title: "Ram Generator", ApiUrl: config.Conf.ApiUrl, DefaultAvatar: config.Conf.DefaultAvatar})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}

func ErrorPage404(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles(fmt.Sprintf("%s/%s", config.Conf.TemplatesPath, "base.html"), fmt.Sprintf("%s/%s", config.Conf.TemplatesPath, "error.html"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, ErrorTemplateData{ErrorText: "Такой страницы не существует", TemplateData: TemplateData{Title: fmt.Sprintf("404 - Ram Generator"), ApiUrl: config.Conf.ApiUrl, DefaultAvatar: config.Conf.DefaultAvatar}})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения файла: %v", err), http.StatusInternalServerError)
		return
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(config.Conf.FaviconPath)
	if err != nil {
		http.Error(w, "FaviconPath not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "FaviconPath not found", http.StatusNotFound)
		return
	}

	// Устанавливаем заголовки HTTP
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	http.ServeContent(w, r, config.Conf.FaviconPath, fileInfo.ModTime(), file)
}
