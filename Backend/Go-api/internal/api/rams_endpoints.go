package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (h *Handlers) GetRams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	user, err := database.GetUserByUsernameContext(ctx, h.db, params["username"])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no users with username = %s", user.Username), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	rams, err := database.GetRamsByUsernameContext(ctx, h.db, params["username"])
	if err != nil {
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(rams)
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

func (h *Handlers) GetRam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("ram id must be integer"), http.StatusBadRequest)
		return
	}

	user, err := database.GetUserByUsernameContext(ctx, h.db, params["username"])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no users with username = %s", user.Username), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	ram, err := database.GetRamContext(ctx, h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no rams with id = %d", id), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}

	if ram.UserId != user.Id {
		http.Error(w, fmt.Sprintf("user no rams with id = %d", id), http.StatusNotFound)
		return
	}

	res, err := json.Marshal(ram)
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

// PutPatchRam (НЕИСПОЛЬЗУЕТСЯ) также выполняет функции PutRam
func (h *Handlers) PutPatchRam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	user, err := database.GetUserContext(ctx, h.db, ctx.Value("userId").(int))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("can't recognize your permissions, please relogin"), http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if user.Username != params["username"] {
		http.Error(w, fmt.Sprintf("you can't edit another user ram"), http.StatusForbidden)
		return
	}

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("ram id must be integer"), http.StatusBadRequest)
		return
	}

	dbRam, err := database.GetRamContext(ctx, h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no rams with id = %d", id), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if user.Id != dbRam.UserId {
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
	if ram.Id != id {
		http.Error(w, "ram id cant be edited", http.StatusBadRequest)
		return
	}
	ram.Id = id
	ram.Taps = 0
	err = database.UpdateRamContext(ctx, h.db, id, ram)
	if err != nil {
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
}

// DeleteRam НЕ ИСПОЛЬЗУЕТСЯ
func (h *Handlers) DeleteRam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	user, err := database.GetUserContext(ctx, h.db, ctx.Value("userId").(int))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("can't recognize your permissions, please relogin"), http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if user.Username != params["username"] {
		http.Error(w, fmt.Sprintf("you can't delete another user's ram"), http.StatusForbidden)
		return
	}

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("ram id must be integer"), http.StatusBadRequest)
		return
	}
	dbRam, err := database.GetRamContext(ctx, h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no rams with id = %d", id), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
	if user.Id != dbRam.UserId {
		http.Error(w, fmt.Sprintf("it's not your ram"), http.StatusForbidden)
		return
	}

	err = database.DeleteRamContext(ctx, h.db, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
}
