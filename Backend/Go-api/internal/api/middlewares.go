package api

import (
	"context"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/auth"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// AuthorizationMiddleware check Authorization: Bearer token
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userId, err := auth.LoadUserIdFromToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userId", userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		slog.Info("Request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Duration("duration", duration),
		)
	})
}
