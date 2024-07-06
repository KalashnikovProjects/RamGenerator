package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/internal/auth"
	"github.com/KalashnikovProjects/RamGenerator/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/internal/entities"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"log"
	"net/http"
)

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	user, err := database.GetUserByUsernameContext(ctx, h.db, params["username"])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no users with username = %s", params["username"]), http.StatusNotFound)
			return
		}
		log.Println(err)
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = ""
	user.Password = ""
	res, err := json.Marshal(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("json marshal error"), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("response writing error"), http.StatusInternalServerError)
		return
	}
}

// PatchUser также выполняет функции PutUser
// Рекомендуется использовать Patch
func (h *Handlers) PatchUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	conf := config.New()

	dbUser, err := database.GetUserContext(ctx, h.db, ctx.Value("userId").(int))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("can't recognize your permissions, please relogin"), http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if dbUser.Username != params["username"] {
		http.Error(w, fmt.Sprintf("you can't edit another user"), http.StatusForbidden)
		return
	}

	var user entities.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, fmt.Sprintf("json decoding error"), http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodPut {
		if user.Username == "" || user.Password == "" || user.AvatarUrl == "" || user.AvatarBox == nil {
			http.Error(w, "all fields must be filled for PUT request", http.StatusBadRequest)
			return
		}
	}
	if user.Username != "" && !ValidateUsername(user.Username) {
		http.Error(w, fmt.Sprintf("username must be 3-%d lenght and contain only English letters, numbers and _", conf.UsersConfig.MaxUsernameLen), http.StatusBadRequest)
		return
	}
	if user.Password != "" {
		hashed, err := auth.GenerateHashedPassword(user.Password)
		if err != nil {
			http.Error(w, fmt.Sprintf("hashing password error"), http.StatusInternalServerError)
			return
		}
		user.PasswordHash = hashed
		user.Password = ""
	}

	// Неизменяемое поле
	user.LastRamGenerated = 0

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
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
}
