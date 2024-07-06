package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/internal/entities"
	"github.com/KalashnikovProjects/RamGenerator/internal/ram_image_generator"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
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

// CreateRam Переходит на websocket
func (h *Handlers) CreateRam(w http.ResponseWriter, r *http.Request) {
	conf := config.New()
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
		http.Error(w, fmt.Sprintf("you can't create ram for another user"), http.StatusForbidden)
		return
	}
	log.Println(user.LastRamGenerated+conf.UsersConfig.TimeBetweenGenerations*60, int(time.Now().Unix()))
	if user.LastRamGenerated+conf.UsersConfig.TimeBetweenGenerations*60 > int(time.Now().Unix()) {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Now().Unix())-user.LastRamGenerated+conf.UsersConfig.TimeBetweenGenerations*60))
		http.Error(w, fmt.Sprintf("you need to wait %d hours before generating a new ram", conf.UsersConfig.TimeBetweenGenerations), http.StatusTooManyRequests)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed upgrade to websocket", http.StatusBadRequest)
		log.Print(err)
		return
	}
	h.websocketCreateRam(ctx, ws, user)
}

func (h *Handlers) websocketCreateRam(ctx context.Context, ws *websocket.Conn, user entities.User) {
	defer ws.Close()
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))

	ram := entities.Ram{UserId: user.Id}

	messageType, wsMessage, err := ws.ReadMessage()
	if err != nil {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("read message error")})
		log.Println(err)
		return
	}
	if messageType != websocket.TextMessage {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("invalid input")})
		return
	}

	userPrompt := string(wsMessage)
	ws.WriteMessage(websocket.TextMessage, []byte("generating prompt"))

	var prompt string
	if user.LastRamGenerated == 0 {
		prompt, err = ram_image_generator.GenerateStartPrompt(ctx, h.gRPCClient, userPrompt)
	} else {
		rams, err := database.GetRamsByUserIdContext(ctx, h.db, user.Id)
		if err != nil {
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("unexpected db error")})
			return
		}
		var descriptions []string
		for _, userRam := range rams {
			descriptions = append(descriptions, userRam.Description)
		}
		prompt, err = ram_image_generator.GenerateHybridPrompt(ctx, h.gRPCClient, userPrompt, descriptions)

	}
	if err != nil {
		if errors.Is(err, ram_image_generator.CensorshipError) {
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("user prompt or rams descriptions contains illegal content")})
			return
		}
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("prompt generating error")})
		return
	}
	ws.WriteMessage(websocket.TextMessage, []byte("generating image"))

	imageBase64, err := ram_image_generator.GenerateRamImage(ctx, h.gRPCClient, prompt)
	if err != nil {
		if errors.Is(err, ram_image_generator.ImageGenerationTimeout) {
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("image generation timeout")})
			return
		}
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("image generating error")})
		return
	}

	imageUrl, err := ram_image_generator.UploadImage(imageBase64)
	if err != nil {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("image uploading error")})
		return
	}
	ram.ImageUrl = imageUrl
	imageDescription, err := ram_image_generator.GenerateDescription(ctx, h.gRPCClient, imageUrl)
	if err != nil {
		if errors.Is(err, ram_image_generator.CensorshipError) {
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("user prompt or rams descriptions contains illegal content")})
			return
		}
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("image description generating error")})
		return
	}
	ram.Description = imageDescription
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("unexpected db error")})
		return
	}

	id, err := database.CreateRamContext(ctx, tx, ram)
	if err != nil {
		tx.Rollback()
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("unexpected db error")})
		return
	}
	err = database.UpdateUserContext(ctx, tx, user.Id, entities.User{LastRamGenerated: int(time.Now().Unix())})
	if err != nil {
		tx.Rollback()
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("unexpected db error")})
		return
	}
	tx.Commit()
	ws.WriteJSON(map[string]string{"id": fmt.Sprintf("%d", id), "image_url": imageUrl, "image_description": imageDescription})
	return
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

// PatchRam также выполняет функции PutRam
func (h *Handlers) PatchRam(w http.ResponseWriter, r *http.Request) {
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
	err = database.UpdateRamContext(ctx, h.db, id, ram)
	if err != nil {
		http.Error(w, fmt.Sprintf("unexpected db error"), http.StatusInternalServerError)
		return
	}
}

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
