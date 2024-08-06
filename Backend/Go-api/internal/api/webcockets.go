package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/ram_image_generator"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func validateClickData(clicks int, lastClicks time.Time) bool {
	now := time.Now()
	timeSub := now.Sub(lastClicks)

	if timeSub < time.Second {
		return false
	}

	if timeSub > time.Minute {
		timeSub = time.Minute
	}

	if clicks < 0 || clicks > 100 {
		return false
	}

	if clicks > int(40*timeSub.Seconds()) {
		return false
	}

	return true
}

func PingOrCancelContext(ctx context.Context, ws *websocket.Conn, cancel func()) {
	ticker := time.NewTicker(time.Second * time.Duration(config.Conf.Websocket.PingPeriod))

	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(time.Second * time.Duration(config.Conf.Websocket.PongWait)))
		return nil
	})

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				cancel()
			}
		}
	}
}

func (h *Handlers) websocketNeedClicks(ctx context.Context, ws *websocket.Conn, amount int) error {
	var clicked int
	lastClicks := time.Now().Add(-1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			ws.WriteJSON(map[string]string{"error": "context canceled"})
			return ctx.Err()
		default:
			messageType, wsMessage, err := ws.ReadMessage()
			if err != nil {
				ws.WriteJSON(map[string]string{"error": "read message error"})
				continue
			}
			if messageType != websocket.TextMessage {
				ws.WriteJSON(map[string]string{"error": "invalid message type"})
				continue
			}
			messageClicks, err := strconv.Atoi(string(wsMessage))
			if !validateClickData(messageClicks, lastClicks) {
				ws.WriteMessage(websocket.TextMessage, []byte("invalid clicks"))
				continue
			}
			lastClicks = time.Now()

			clicked += messageClicks
			if clicked >= amount {
				return nil
			}
		}
	}
}

func (h *Handlers) WebsocketClicker(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, fmt.Sprintf("you can't tap another user ram"), http.StatusForbidden)
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

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed upgrade to websocket", http.StatusBadRequest)
		log.Print(err)
		return
	}
	h.upgradedWebsocketClicker(ctx, ws, dbRam.Id)
}

func (h *Handlers) upgradedWebsocketClicker(ctx context.Context, ws *websocket.Conn, ramId int) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	PingOrCancelContext(ctx, ws, cancel)

	var clicked int
	defer func() {
		if clicked == 0 {
			return
		}
		database.AddTapsRamContext(context.Background(), h.db, ramId, clicked)
	}()
	lastClicks := time.Now().Add(-1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			ws.WriteJSON(map[string]string{"error": "context canceled"})
			return
		default:
			messageType, wsMessage, err := ws.ReadMessage()
			if err != nil {
				ws.WriteJSON(map[string]string{"error": "read message error"})
				log.Println(err)
				continue
			}
			if messageType != websocket.TextMessage {
				ws.WriteJSON(map[string]string{"error": "invalid message type"})
				continue
			}
			messageClicks, err := strconv.Atoi(string(wsMessage))
			if !validateClickData(messageClicks, lastClicks) {
				ws.WriteMessage(websocket.TextMessage, []byte("invalid clicks"))
				continue
			}
			lastClicks = time.Now()

			clicked += messageClicks
		}
	}
}

func (h *Handlers) WebsocketCreateRam(w http.ResponseWriter, r *http.Request) {
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
	if user.RamsGeneratedLastDay >= len(config.Conf.Clicks.DailyRamsPrices) {
		http.Error(w, fmt.Sprintf("you can generate only %d rams per day", len(config.Conf.Clicks.DailyRamsPrices)), http.StatusTooManyRequests)
		return
	}
	log.Println(user.DailyRamGenerationTime+config.Conf.Users.TimeBetweenGenerations*60, int(time.Now().Unix()))
	if user.DailyRamGenerationTime+config.Conf.Users.TimeBetweenGenerations*60 > int(time.Now().Unix()) {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Now().Unix())-user.DailyRamGenerationTime+config.Conf.Users.TimeBetweenGenerations*60))
		http.Error(w, fmt.Sprintf("you need to wait %d hours before generating a new ram", config.Conf.Users.TimeBetweenGenerations), http.StatusTooManyRequests)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed upgrade to websocket", http.StatusBadRequest)
		log.Print(err)
		return
	}
	h.upgradedWebsocketCreateRam(ctx, ws, user)
}

func (h *Handlers) upgradedWebsocketCreateRam(ctx context.Context, ws *websocket.Conn, user entities.User) {
	defer ws.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	PingOrCancelContext(ctx, ws, cancel)

	messageType, wsMessage, err := ws.ReadMessage()
	if err != nil {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("read message error")})
		log.Println(err)
		return
	}
	if messageType != websocket.TextMessage {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("invalid message type")})
		return
	}

	userPrompt := string(wsMessage)
	ws.WriteMessage(websocket.TextMessage, []byte("generating prompt"))

	var prompt string
	if user.DailyRamGenerationTime == 0 {
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

	imageBase64, err := ram_image_generator.GenerateRamImage(ctx, h.gRPCClient, prompt)
	if err != nil {
		if errors.Is(err, ram_image_generator.ImageGenerationTimeout) {
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("image generation timeout")})
			return
		}
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("image generating error")})
		return
	}

	var needClicks int
	if user.DailyRamGenerationTime == 0 {
		needClicks = config.Conf.Clicks.FirstRam
	} else {
		if user.RamsGeneratedLastDay >= len(config.Conf.Clicks.DailyRamsPrices) {
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("you can generate only %d rams per day", len(config.Conf.Clicks.DailyRamsPrices))})
			return
		}
		needClicks = config.Conf.Clicks.DailyRamsPrices[user.RamsGeneratedLastDay]

	}
	err = h.websocketNeedClicks(ctx, ws, needClicks)
	if err != nil {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("clicks timeout")})
		return
	}

	ws.WriteMessage(websocket.TextMessage, []byte("generating image"))

	imageUrl, err := ram_image_generator.UploadImage(imageBase64)
	if err != nil {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("image uploading error")})
		return
	}
	imageDescription, err := ram_image_generator.GenerateDescription(ctx, h.gRPCClient, imageUrl)
	if err != nil {
		if errors.Is(err, ram_image_generator.CensorshipError) {
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("user prompt or rams descriptions contains illegal content")})
			return
		}
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("image description generating error")})
		return
	}
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("unexpected db error")})
		return
	}
	ram := entities.Ram{UserId: user.Id, Description: imageDescription, ImageUrl: imageUrl}
	id, err := database.CreateRamContext(ctx, tx, ram)
	if err != nil {
		tx.Rollback()
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("unexpected db error")})
		return
	}
	err = database.UpdateUserContext(ctx, tx, user.Id, entities.User{DailyRamGenerationTime: int(time.Now().Unix()), RamsGeneratedLastDay: user.RamsGeneratedLastDay + 1})
	if err != nil {
		tx.Rollback()
		ws.WriteJSON(map[string]string{"error": fmt.Sprintf("unexpected db error")})
		return
	}
	tx.Commit()
	ws.WriteJSON(map[string]string{"id": fmt.Sprintf("%d", id), "image_url": imageUrl, "image_description": imageDescription})
	return
}
