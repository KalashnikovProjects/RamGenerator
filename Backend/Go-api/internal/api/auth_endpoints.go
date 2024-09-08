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
	"github.com/lib/pq"
	"log/slog"
	"net/http"
	"regexp"
)

var UsernameRegex, _ = regexp.Compile("^[a-zA-Z0-9_]{3,24}$")

func ValidateUsername(username string) bool {
	return UsernameRegex.Match([]byte(username))
}

type LoginUser struct {
	Username string
	Password string
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqUser LoginUser
	err := json.NewDecoder(r.Body).Decode(&reqUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("json decoding error"), http.StatusBadRequest)
		return
	}
	if reqUser.Username == "" {
		http.Error(w, fmt.Sprintf("required fields are not specified: username"), http.StatusBadRequest)
		return
	}
	if reqUser.Password == "" {
		http.Error(w, fmt.Sprintf("required fields are not specified: password"), http.StatusBadRequest)
		return
	}
	if !ValidateUsername(reqUser.Username) {
		http.Error(w, fmt.Sprintf("username must be 3-%d lenght and contain only English letters, numbers and _", config.Conf.Users.MaxUsernameLen), http.StatusBadRequest)
		return
	}
	hashed, err := auth.GenerateHashedPassword(reqUser.Password)
	if err != nil {
		slog.Error("hashing password error", slog.String("function", "auth.GenerateHashedPassword"), slog.String("endpoint", "register"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("hashing password error"), http.StatusInternalServerError)
		return
	}
	user := entities.User{
		Username:               reqUser.Username,
		PasswordHash:           hashed,
		DailyRamGenerationTime: 0,
		RamsGeneratedLastDay:   0,
		ClickersBlockedUntil:   0,
		AvatarRamId:            0,
		AvatarBox:              &config.Conf.Users.DefaultAvatarBox,
	}
	user.Id, err = database.CreateUserContext(ctx, h.db, user)
	if err != nil {
		var pgErr *pq.Error
		isPgErr := errors.As(err, &pgErr)
		if isPgErr && pgErr.Code == "23505" {
			http.Error(w, fmt.Sprintf("username %s is already taken", user.Username), http.StatusBadRequest)
			return
		}
		slog.Error("unexpected db error", slog.String("function", "database.CreateUserContext"), slog.String("endpoint", "register"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	// Записывает id вместо токена при регистрации
	//res, err := json.Marshal(entities.IdResponse{Id: id})
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("json marshal error"), http.StatusInternalServerError)
	//	return
	//}
	//_, err = w.Write(res)
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("response writing error"), http.StatusInternalServerError)
	//	return
	//}
	// return
	token, err := auth.GenerateToken(user.Id)
	if err != nil {
		slog.Error("token generation error", slog.String("function", "auth.GenerateToken"), slog.String("endpoint", "register"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("token generation error"), http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(token))
	if err != nil {
		slog.Error("token writing error", slog.String("endpoint", "register"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("token writing error"), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqUser LoginUser
	err := json.NewDecoder(r.Body).Decode(&reqUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("json decoding error"), http.StatusBadRequest)
		return
	}
	if reqUser.Username == "" {
		http.Error(w, fmt.Sprintf("required fields are not specified: username"), http.StatusBadRequest)
		return
	}
	if reqUser.Password == "" {
		http.Error(w, fmt.Sprintf("required fields are not specified: password"), http.StatusBadRequest)
		return
	}
	dbUser, err := database.GetUserByUsernameContext(ctx, h.db, reqUser.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no users with username = %s", reqUser.Username), http.StatusNotFound)
			return
		}
		slog.Error("unexpected db error", slog.String("endpoint", "login"), slog.String("function", "database.GetUserByUsernameContext"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	err = auth.ComparePasswordWithHashed(reqUser.Password, dbUser.PasswordHash)
	if err != nil {
		http.Error(w, fmt.Sprintf("wrong password"), http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(dbUser.Id)
	if err != nil {
		slog.Error("token generation error", slog.String("function", "auth.GenerateToken"), slog.String("endpoint", "login"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("token generation error"), http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(token))
	if err != nil {
		slog.Error("token writing error", slog.String("endpoint", "login"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("token writing error"), http.StatusInternalServerError)
		return
	}
}

// Me возвращает информацию о пользователе, владельце токена из Authorization
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbUser, err := database.GetUserContext(ctx, h.db, ctx.Value("userId").(int))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("can't recognize your permissions, please relogin"), http.StatusUnauthorized)
			return
		}
		slog.Error("unexpected db error", slog.String("endpoint", "me"), slog.String("function", "database.GetUserContext"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	dbUser.PasswordHash = ""
	dbUser.Password = ""
	res, err := json.Marshal(dbUser)
	if err != nil {
		slog.Error("json marshal error", slog.String("endpoint", "me"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("json marshal error"), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		slog.Error("response writing error", slog.String("endpoint", "me"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("response writing error"), http.StatusInternalServerError)
		return
	}
}
