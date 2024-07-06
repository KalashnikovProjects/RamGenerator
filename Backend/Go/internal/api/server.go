package api

import (
	"context"
	"github.com/KalashnikovProjects/RamGenerator/internal/database"
	pb "github.com/KalashnikovProjects/RamGenerator/proto_generated"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
)

type Handlers struct {
	db         database.SQLTXQueryExec
	gRPCClient pb.RamGeneratorClient
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewRamGeneratorServer(ctx context.Context, Addr string, db database.SQLTXQueryExec, gRPCClient pb.RamGeneratorClient) *http.Server {
	handlers := Handlers{db: db, gRPCClient: gRPCClient}
	router := mux.NewRouter()

	router.Handle("/api/register", http.HandlerFunc(handlers.Register)).Methods("POST")
	router.Handle("/api/login", http.HandlerFunc(handlers.Login)).Methods("POST")

	// Необходим токен
	router.Handle("/api/users/{username}", http.HandlerFunc(handlers.GetUser)).Methods("GET")
	router.Handle("/api/users/{username}", AuthenticationMiddleware(http.HandlerFunc(handlers.PatchUser))).Methods("PUT", "PATCH")
	router.Handle("/api/users/{username}", AuthenticationMiddleware(http.HandlerFunc(handlers.DeleteUser))).Methods("DELETE")

	router.Handle("/api/users/{username}/ws/create-ram", AuthenticationMiddleware(http.HandlerFunc(handlers.CreateRam))).Methods("GET", "POST")

	router.Handle("/api/users/{username}/rams", http.HandlerFunc(handlers.GetRams)).Methods("GET")
	router.Handle("/api/users/{username}/rams/{id}", http.HandlerFunc(handlers.GetRam)).Methods("GET")
	router.Handle("/api/users/{username}/rams/{id}", AuthenticationMiddleware(http.HandlerFunc(handlers.DeleteRam))).Methods("DELETE")

	// Барана нельзя редактировать после изменения (только передавать другому пользователю через трейды)
	// router.Handle("/api/users/{username}/rams/{id}", AuthenticationMiddleware(http.HandlerFunc(handlers.PatchRam))).Methods("PUT", "PATCH")

	//TODO:
	// router.HandleFunc("/api/trade", handlers.TradeWebsocket)

	log.Printf("API running on %s\n", Addr)

	// Пока-что выключу
	//corsHandler := handlers.CORS(
	//	handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
	//	handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	//	handlers.AllowedOrigins([]string{"*"}),
	//	handlers.AllowCredentials(),
	//)
	httpServer := &http.Server{
		Addr:    Addr,
		Handler: router,
	}

	return httpServer
}

func ServeServer(ctx context.Context, server *http.Server) error {
	log.Println("Running api server...")
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := server.ListenAndServe()
		log.Println(err)
		return err
	})
	<-gCtx.Done()
	return server.Shutdown(ctx)
}
