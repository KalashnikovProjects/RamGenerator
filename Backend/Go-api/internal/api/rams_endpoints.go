package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handlers) GetTopRams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rams, err := database.GetTopRams(ctx, h.db, config.Conf.Another.TopRamsCount)
	if err != nil {
		slog.Error("unexpected db error", slog.String("function", "database.GetTopRams"), slog.String("endpoint", "get top rams"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(rams)
	if err != nil {
		slog.Error("json marshal error", slog.String("endpoint", "get top rams"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("json marshal error"), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		slog.Error("response writing error", slog.String("endpoint", "get top rams"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("response writing error"), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) GetRams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	user, err := database.GetUserByUsernameContext(ctx, h.db, params["username"])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no users with username = %s", user.Username), http.StatusNotFound)
			return
		}
		slog.Error("unexpected db error", slog.String("function", "database.GetUserByUsernameContext"), slog.String("endpoint", "get rams"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	rams, err := database.GetRamsByUsernameContext(ctx, h.db, params["username"])
	if err != nil {
		slog.Error("unexpected db error", slog.String("function", "database.GetRamsByUsernameContext"), slog.String("endpoint", "get rams"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(rams)
	if err != nil {
		slog.Error("json marshal error", slog.String("endpoint", "get rams"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("json marshal error"), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		slog.Error("response writing error", slog.String("endpoint", "get rams"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("response writing error"), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) GetRam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("ram id must be integer"), http.StatusBadRequest)
		return
	}

	ram, err := database.GetRamWithUsernameContext(ctx, h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no rams with id = %d", id), http.StatusNotFound)
			return
		}
		slog.Error("unexpected db error", slog.String("function", "database.GetRamContext"), slog.String("endpoint", "get ram"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	if ram.User.Username != params["username"] {
		http.Error(w, fmt.Sprintf("this is %s ram", params["username"]), http.StatusNotFound)
		return
	}

	res, err := json.Marshal(ram)
	if err != nil {
		slog.Error("json marshal error", slog.String("endpoint", "get ram"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("json marshal error"), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		slog.Error("response writing error", slog.String("endpoint", "get ram"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("response writing error"), http.StatusInternalServerError)
		return
	}
}

// PutPatchRam (Unused) also running for Put method, recommended Patch
func (h *Handlers) PutPatchRam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("ram id must be integer"), http.StatusBadRequest)
		return
	}

	dbRam, err := database.GetRamWithUsernameContext(ctx, h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no rams with id = %d", id), http.StatusNotFound)
			return
		}
		slog.Error("unexpected db error", slog.String("function", "database.GetRamContext"), slog.String("endpoint", fmt.Sprintf("%s ram", strings.ToLower(r.Method))), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if ctx.Value("userId").(int) != dbRam.UserId || dbRam.User.Username != params["username"] {
		http.Error(w, fmt.Sprintf("it's not your ram"), http.StatusForbidden)
		return
	}

	var ram entities.Ram
	err = json.NewDecoder(r.Body).Decode(&ram)
	if err != nil {
		http.Error(w, fmt.Sprintf("json decoding error"), http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodPut {
		if ram.Description == "" || ram.ImageUrl == "" {
			http.Error(w, "all fields must be filled for PUT request", http.StatusBadRequest)
			return
		}
	}
	ram.Id = id
	ram.Taps = 0
	ram.UserId = 0
	err = database.UpdateRamContext(ctx, h.db, id, ram)
	if err != nil {
		slog.Error("unexpected db error", slog.String("function", "database.UpdateRamContext"), slog.String("endpoint", fmt.Sprintf("%s ram", strings.ToLower(r.Method))), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
}

// DeleteRam (Unused)
func (h *Handlers) DeleteRam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("ram id must be integer"), http.StatusBadRequest)
		return
	}
	dbRam, err := database.GetRamWithUsernameContext(ctx, h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no rams with id = %d", id), http.StatusNotFound)
			return
		}
		slog.Error("unexpected db error", slog.String("endpoint", "delete ram"), slog.String("function", "database.GetRamContext"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if ctx.Value("userId").(int) != dbRam.UserId || dbRam.User.Username != params["username"] {
		http.Error(w, fmt.Sprintf("it's not your ram"), http.StatusForbidden)
		return
	}

	err = database.DeleteRamContext(ctx, h.db, id)
	if err != nil {
		slog.Error("unexpected db error", slog.String("endpoint", "delete ram"), slog.String("function", "database.DeleteRamContext"), slog.String("error", err.Error()))
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
}
