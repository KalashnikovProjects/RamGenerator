package api

import (
	"context"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	pb "github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/proto_generated"
	"github.com/didip/tollbooth"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
)

type Handlers struct {
	db         database.SQLTXQueryExec
	gRPCClient pb.RamGeneratorClient
}

func NewRamGeneratorServer(ctx context.Context, Addr string, db database.SQLTXQueryExec, gRPCClient pb.RamGeneratorClient) *http.Server {
	handlers := Handlers{db: db, gRPCClient: gRPCClient}
	router := mux.NewRouter()

	router.Handle("/api/register", http.HandlerFunc(handlers.Register)).Methods("GET", "POST")
	router.Handle("/api/login", http.HandlerFunc(handlers.Login)).Methods("GET", "POST")
	router.Handle("/api/me", AuthorizationMiddleware(http.HandlerFunc(handlers.Me))).Methods("GET", "POST")

	router.Handle("/api/top-rams", http.HandlerFunc(handlers.GetTopRams)).Methods("GET")

	router.Handle("/api/users/{username}", http.HandlerFunc(handlers.GetUser)).Methods("GET")
	router.Handle("/api/users/{username}", AuthorizationMiddleware(http.HandlerFunc(handlers.PutPatchUser))).Methods("PUT", "PATCH")
	router.Handle("/api/users/{username}", AuthorizationMiddleware(http.HandlerFunc(handlers.DeleteUser))).Methods("DELETE")

	router.Handle("/api/users/{username}/ws/generate-ram", tollbooth.LimitHandler(tollbooth.NewLimiter(3, nil), http.HandlerFunc(handlers.WebsocketGenerateRam))).Methods("GET", "POST")

	router.Handle("/api/users/{username}/rams", http.HandlerFunc(handlers.GetRams)).Methods("GET")
	router.Handle("/api/users/{username}/rams/{id}", http.HandlerFunc(handlers.GetRam)).Methods("GET")

	router.Handle("/api/users/{username}/rams/{id}/ws/clicker", tollbooth.LimitHandler(tollbooth.NewLimiter(3, nil), http.HandlerFunc(handlers.WebsocketClicker))).Methods("GET", "POST")

	// Maybe made in future:
	// router.HandleFunc("/api/ws/trade", handlers.TradeWebsocket)

	slog.Info("API server created", slog.String("addr", Addr))
	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowCredentials(),
	)
	httpServer := &http.Server{
		Addr:    Addr,
		Handler: corsHandler(tollbooth.LimitHandler(tollbooth.NewLimiter(50, nil), LoggingMiddleware(router))),
	}

	return httpServer
}

func ServeServer(ctx context.Context, server *http.Server) error {
	slog.Info("Running API server", slog.String("addr", server.Addr))

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := server.ListenAndServe()
		slog.Error("Go api server error", slog.String("error", err.Error()))
		return err
	})
	<-gCtx.Done()
	return server.Shutdown(ctx)
}
