package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Go-static-server/internal/config"
	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type ErrorTemplateData struct {
	ErrorText string
	config.BaseTemplateData
}

var Renders PagesRenders

type PagesRenders struct {
	Index    *[]byte
	Login    *[]byte
	User     *[]byte
	Error404 *[]byte
}

func RenderTemplate(baseTemplates, additionalTemplates []string, templateData any) (*[]byte, error) {
	ts, err := template.ParseFiles(append(baseTemplates, additionalTemplates...)...)
	if err != nil {
		return nil, err
	}
	buff := &bytes.Buffer{}
	err = ts.Execute(buff, templateData)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(buff)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func RenderTemplates() error {
	baseTemplates := []string{fmt.Sprintf("%s/%s", config.Conf.Paths.Templates, "base.html"), fmt.Sprintf("%s/%s", config.Conf.Paths.Templates, "header.html")}
	pagesRenders := PagesRenders{}

	var err error
	pagesRenders.Index, err = RenderTemplate(baseTemplates, []string{fmt.Sprintf("%s/%s", config.Conf.Paths.Templates, "index.html"), fmt.Sprintf("%s/%s", config.Conf.Paths.Templates, "footer.html")}, config.Conf.BaseTemplateData)
	if err != nil {
		return err
	}
	pagesRenders.Login, err = RenderTemplate(baseTemplates, []string{fmt.Sprintf("%s/%s", config.Conf.Paths.Templates, "login.html")}, config.Conf.BaseTemplateData)
	if err != nil {
		return err
	}
	pagesRenders.User, err = RenderTemplate(baseTemplates, []string{fmt.Sprintf("%s/%s", config.Conf.Paths.Templates, "user.html")}, config.Conf.BaseTemplateData)
	if err != nil {
		return err
	}
	pagesRenders.Error404, err = RenderTemplate(baseTemplates, []string{fmt.Sprintf("%s/%s", config.Conf.Paths.Templates, "error.html")}, ErrorTemplateData{"Такой страницы не существует", config.Conf.BaseTemplateData})
	if err != nil {
		return err
	}
	Renders = pagesRenders
	return nil
}

func WriteRenderedPage(data *[]byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(*data)
	}
}

func NewStaticServer(Addr string) *http.Server {
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(WriteRenderedPage(Renders.Error404))

	router.HandleFunc("/", WriteRenderedPage(Renders.Index))
	router.HandleFunc("/users/{username}", WriteRenderedPage(Renders.User))
	router.HandleFunc("/login", WriteRenderedPage(Renders.Login))

	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/robots.txt", robotsHandler)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.Conf.Paths.CdnFiles))))
	return &http.Server{
		Addr:    Addr,
		Handler: tollbooth.LimitHandler(tollbooth.NewLimiter(50, nil), LoggingMiddleware(router)),
	}
}

func ServeServer(ctx context.Context, server *http.Server) error {
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := server.ListenAndServe()
		slog.Error("Go static server error", slog.String("error", err.Error()))
		return err
	})
	<-gCtx.Done()
	return server.Shutdown(ctx)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(config.Conf.Paths.Favicon)
	if err != nil {
		http.Error(w, "favicon not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "favicon not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	http.ServeContent(w, r, config.Conf.Paths.Favicon, fileInfo.ModTime(), file)
}

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(config.Conf.Paths.Robots)
	if err != nil {
		http.Error(w, "robots.txt path not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "robots.txt not found", http.StatusNotFound)
		return
	}

	http.ServeContent(w, r, config.Conf.Paths.Robots, fileInfo.ModTime(), file)
}
