package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/auth"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"log/slog"
	"net/http"
	"strings"
)

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	user, err := database.GetUserByUsernameWithAvatarUrlContext(ctx, h.db, params["username"])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no users with username = %s", params["username"]), http.StatusNotFound)
			return
		}
		slog.Error("unexpected db error", slog.String("endpoint", "get user"), slog.String("function", "database.GetUserByUsernameContext"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = ""
	user.Password = ""
	res, err := json.Marshal(user)
	if err != nil {
		slog.Error("json marshal error", slog.String("endpoint", "get user"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("json marshal error"), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		slog.Error("response writing error", slog.String("endpoint", "get user"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("response writing error"), http.StatusInternalServerError)
		return
	}
}

// PutPatchUser также выполняет функции PutUser
// Рекомендуется использовать Patch
func (h *Handlers) PutPatchUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	var user entities.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, fmt.Sprintf("json decoding error"), http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodPut {
		if user.Username == "" || user.Password == "" || user.AvatarRamId == 0 || user.AvatarBox == nil {
			http.Error(w, "all fields must be filled for PUT request", http.StatusBadRequest)
			return
		}
	}

	dbUser, err := database.GetUserContext(ctx, h.db, ctx.Value("userId").(int))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("can't recognize your permissions, please relogin"), http.StatusUnauthorized)
			return
		}
		slog.Error("unexpected db error", slog.String("function", "database.GetUserContext"), slog.String("endpoint", fmt.Sprintf("%s user", strings.ToLower(r.Method))), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if dbUser.Username != params["username"] {
		http.Error(w, fmt.Sprintf("you can't edit another user"), http.StatusForbidden)
		return
	}

	if user.AvatarRamId != 0 {
		dbRam, err := database.GetRamContext(ctx, h.db, user.AvatarRamId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, fmt.Sprintf("no rams with id = %d", user.AvatarRamId), http.StatusNotFound)
				return
			}
			slog.Error("unexpected db error", slog.String("function", "database.GetRamContext"), slog.String("endpoint", fmt.Sprintf("%s user", strings.ToLower(r.Method))), slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
			return
		}
		if dbUser.Id != dbRam.UserId {
			http.Error(w, fmt.Sprintf("it's not your ram"), http.StatusForbidden)
			return
		}
	}

	if user.Username != "" && !ValidateUsername(user.Username) {
		http.Error(w, fmt.Sprintf("username must be 3-%d lenght and contain only English letters, numbers and _", config.Conf.Users.MaxUsernameLen), http.StatusBadRequest)
		return
	}
	if user.Password != "" {
		hashed, err := auth.GenerateHashedPassword(user.Password)
		if err != nil {
			slog.Error("hashing password error", slog.String("function", "auth.GenerateHashedPassword"), slog.String("endpoint", fmt.Sprintf("%s user", strings.ToLower(r.Method))), slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("hashing password error"), http.StatusInternalServerError)
			return
		}
		user.PasswordHash = hashed
		user.Password = ""
	}

	// Неизменяемые поля
	user.Id = 0
	user.DailyRamGenerationTime = 0
	user.RamsGeneratedLastDay = 0

	err = database.UpdateUserByUsernameContext(ctx, h.db, params["username"], user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no users with username = %s", params["username"]), http.StatusNotFound)
			return
		}

		var pgErr *pq.Error
		isPgErr := errors.As(err, &pgErr)
		if isPgErr && pgErr.Code == "23505" {
			http.Error(w, fmt.Sprintf("username %s is already taken", user.Username), http.StatusBadRequest)
			return
		}
		slog.Error("unexpected db error", slog.String("function", "database.UpdateUserByUsernameContext"), slog.String("endpoint", fmt.Sprintf("%s user", strings.ToLower(r.Method))), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	dbUser, err := database.GetUserContext(ctx, h.db, ctx.Value("userId").(int))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("can't recognize your permissions, please relogin"), http.StatusUnauthorized)
			return
		}
		slog.Error("unexpected db error", slog.String("endpoint", "delete user"), slog.String("function", "database.GetUserContext"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if dbUser.Username != params["username"] {
		http.Error(w, fmt.Sprintf("you can't delete another user"), http.StatusForbidden)
		return
	}

	err = database.DeleteUserByUsernameContext(ctx, h.db, params["username"])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no users with username = %s", params["username"]), http.StatusNotFound)
			return
		}
		slog.Error("unexpected db error", slog.String("endpoint", "delete user"), slog.String("function", "database.DeleteUserByUsernameContext"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
}
