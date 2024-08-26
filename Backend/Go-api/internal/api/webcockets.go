package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/auth"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/database"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/entities"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/ram_image_generator"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"time"
)

var (
	errorGenerateRamLimitExceed = errors.New("error generating ram: daily rams limit exceed")
	errorGenerateRamWaitTime    = errors.New("error generating ram: too many requests")
)

type wsError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type wsErrorRateLimit struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
	Next  int    `json:"next"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ValidateClickData(clicks int, lastClicks time.Time) bool {
	now := time.Now()
	timeSub := now.Sub(lastClicks)

	if timeSub < time.Second {
		return false
	}

	if timeSub > time.Minute {
		timeSub = time.Minute
	}

	if clicks < 0 || clicks > 200 {
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

func wsFirstMessageAuthorization(ws *websocket.Conn) (int, error) {
	ws.SetReadDeadline(time.Now().Add(time.Second * 2))

	messageType, wsMessage, err := ws.ReadMessage()
	if err != nil {
		ws.WriteJSON(wsError{"read message error", 400})
		return 0, err
	}
	if messageType != websocket.TextMessage {
		ws.WriteJSON(wsError{"invalid message type", 400})
		return 0, err
	}

	token := string(wsMessage)
	userId, err := auth.LoadUserIdFromToken(token)
	if err != nil {
		ws.WriteJSON(wsError{"unauthorized, first message must be token", 401})
		return 0, err
	}
	ws.SetReadDeadline(time.Time{})

	return userId, err
}

func checkWsUserAccess(ctx context.Context, db database.SQLTXQueryExec, ws *websocket.Conn, userId int, params map[string]string) (entities.User, error) {
	user, err := database.GetUserContext(ctx, db, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ws.WriteJSON(wsError{"can't recognize your permissions, please relogin", 401})
			return entities.User{}, err
		}
		ws.WriteJSON(wsError{"unexpected db error", 500})
		return entities.User{}, err
	}
	if user.Username != params["username"] {
		ws.WriteJSON(wsError{"you can't tap another user ram", 403})
		return entities.User{}, err
	}
	return user, nil
}

func checkWsRamAccess(ctx context.Context, db database.SQLTXQueryExec, ws *websocket.Conn, user entities.User, params map[string]string) (entities.Ram, error) {
	ramId, err := strconv.Atoi(params["id"])
	if err != nil {
		ws.WriteJSON(wsError{"ram id must be integer", 400})
		return entities.Ram{}, err
	}

	dbRam, err := database.GetRamContext(ctx, db, ramId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ws.WriteJSON(wsError{"ram not found", 404})
			return entities.Ram{}, err
		}
		ws.WriteJSON(wsError{"unexpected db error", 500})
		return entities.Ram{}, err
	}
	if user.Id != dbRam.UserId {
		ws.WriteJSON(wsError{"it's not your ram", 403})
		return entities.Ram{}, err
	}
	return dbRam, nil
}

func checkWsCanGenerateRam(ws *websocket.Conn, user entities.User) error {
	ramsGeneratedLastDay := user.CalculateRamsGeneratedLastDay(config.Conf.Generation.TimeBetweenDaily)
	if ramsGeneratedLastDay == 0 {
		return nil
	}
	if ramsGeneratedLastDay >= len(config.Conf.Clicks.DailyRams) {
		targetTime := user.DailyRamGenerationTime + config.Conf.Generation.TimeBetweenDaily*60
		ws.WriteJSON(wsErrorRateLimit{fmt.Sprintf("you can generate only %d rams per day, you can generate next in %d (unix)", len(config.Conf.Clicks.DailyRams), targetTime), 429, targetTime})
		return errorGenerateRamLimitExceed
	}
	targetTime := user.DailyRamGenerationTime
	for _, t := range config.Conf.Generation.TimeBetweenDailyAnother[:ramsGeneratedLastDay] {
		targetTime += t * 60
	}
	if targetTime > int(time.Now().Unix()) {
		ws.WriteJSON(wsErrorRateLimit{fmt.Sprintf("you can generate next ram in %d (unix)", targetTime), 429, targetTime})
		return errorGenerateRamWaitTime
	}
	return nil
}

func (h *Handlers) websocketNeedClicks(ctx context.Context, ws *websocket.Conn, amount int) error {
	var clicked int
	lastClicks := time.Now().Add(-1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			ws.WriteJSON(wsError{"context canceled", 499})
			return ctx.Err()
		default:
			messageType, wsMessage, err := ws.ReadMessage()
			if err != nil {
				ws.WriteJSON(wsError{"read message error", 400})
				return err
			}
			if messageType != websocket.TextMessage {
				ws.WriteJSON(wsError{"invalid message type", 400})
				continue
			}
			messageClicks, err := strconv.Atoi(string(wsMessage))
			if !ValidateClickData(messageClicks, lastClicks) {
				ws.WriteJSON(wsError{"invalid clicks", 400})
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

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed upgrade to websocket", http.StatusBadRequest)
		return
	}
	h.upgradedWebsocketClicker(ctx, ws, params)
}

func (h *Handlers) upgradedWebsocketClicker(ctx context.Context, ws *websocket.Conn, params map[string]string) {
	//TODO: defer Control Message

	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()
	userId, err := wsFirstMessageAuthorization(ws)
	if err != nil {
		return
	}
	user, err := checkWsUserAccess(ctx, h.db, ws, userId, params)
	if err != nil {
		return
	}
	ram, err := checkWsRamAccess(ctx, h.db, ws, user, params)
	if err != nil {
		return
	}

	if err = database.UpdateUserClickersBlockedIfCan(ctx, h.db, userId, int(time.Now().Unix())+7200, int(time.Now().Unix())); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ws.WriteJSON(wsError{"cant tap or create 2 rams parallel", 409})
			return
		}
		ws.WriteJSON(wsError{"unexpected db error", 500})
		return
	}
	defer database.UpdateUserClickersBlockedToZero(context.WithoutCancel(ctx), h.db, userId)

	ctx = context.WithValue(ctx, "userId", userId)

	go PingOrCancelContext(ctx, ws, cancel)

	var clicked int
	defer func() {
		if clicked == 0 {
			return
		}
		database.AddTapsRamContext(context.WithoutCancel(ctx), h.db, ram.Id, clicked)
	}()
	lastClicks := time.Now().Add(-1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			ws.WriteJSON(wsError{"context canceled", 499})
			return
		default:
			messageType, wsMessage, err := ws.ReadMessage()
			if err != nil {
				ws.WriteJSON(wsError{"read message error", 400})
				return
			}
			if messageType != websocket.TextMessage {
				ws.WriteJSON(wsError{"invalid message type", 400})
				continue
			}

			messageClicks, err := strconv.Atoi(string(wsMessage))
			if !ValidateClickData(messageClicks, lastClicks) {
				ws.WriteJSON(wsError{"invalid clicks", 400})
				continue
			}
			lastClicks = time.Now()

			clicked += messageClicks
		}
	}
}

func (h *Handlers) WebsocketGenerateRam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed upgrade to websocket", http.StatusBadRequest)
		return
	}
	h.upgradedWebsocketGenerateRam(ctx, ws, params)
}

func (h *Handlers) upgradedWebsocketGenerateRam(ctx context.Context, ws *websocket.Conn, params map[string]string) {
	//TODO: defer Control Message

	defer ws.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	userId, err := wsFirstMessageAuthorization(ws)
	if err != nil {
		return
	}
	user, err := checkWsUserAccess(ctx, h.db, ws, userId, params)
	if err != nil {
		return
	}
	err = checkWsCanGenerateRam(ws, user)
	if err != nil {
		return
	}

	if err = database.UpdateUserClickersBlockedIfCan(ctx, h.db, userId, int(time.Now().Unix())+7200, int(time.Now().Unix())); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ws.WriteJSON(wsError{"cant tap or create 2 rams parallel", 409})
			return
		}
		ws.WriteJSON(wsError{"unexpected db error", 500})
		return
	}
	defer database.UpdateUserClickersBlockedToZero(context.WithoutCancel(ctx), h.db, userId)

	ctx = context.WithValue(ctx, "userId", userId)

	go PingOrCancelContext(ctx, ws, cancel)

	if user.DailyRamGenerationTime == 0 {
		ws.WriteJSON(map[string]any{"status": "need first ram prompt"})
	} else {
		ws.WriteJSON(map[string]any{"status": "need ram prompt"})
	}

	messageType, wsMessage, err := ws.ReadMessage()
	if err != nil {
		ws.WriteJSON(wsError{"read message error", 400})
		return
	}
	if messageType != websocket.TextMessage {
		ws.WriteJSON(wsError{"invalid message type", 400})
		return
	}

	userPrompt := string(wsMessage)
	if len(userPrompt) > config.Conf.Generation.MaxPromptLen {
		ws.WriteJSON(wsError{fmt.Sprintf("user prompt too long (max %d symbols)", config.Conf.Generation.MaxPromptLen), 400})
		return
	}
	aiGeneratedRam := make(chan entities.Ram)
	go func() {
		var prompt string
		if user.DailyRamGenerationTime == 0 {
			prompt, err = ram_image_generator.GenerateStartPrompt(ctx, h.gRPCClient, userPrompt)
		} else {
			rams, err := database.GetRamsByUserIdContext(ctx, h.db, user.Id)
			if err != nil {
				ws.WriteJSON(wsError{"unexpected db error", 500})
				ws.Close()
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
				ws.WriteJSON(wsError{"user prompt or rams descriptions contains illegal content", 400})
				ws.Close()
				return
			}
			if errors.Is(err, ram_image_generator.TooLongPromptError) {
				ws.WriteJSON(wsError{fmt.Sprintf("user prompt too long (max %d symbols)", config.Conf.Generation.MaxPromptLen), 400})
				ws.Close()
				return
			}
			ws.WriteJSON(wsError{"prompt generating error", 500})
			ws.Close()
			return
		}

		imageBase64, err := ram_image_generator.GenerateRamImage(ctx, h.gRPCClient, prompt)
		if err != nil {
			if errors.Is(err, ram_image_generator.ImageGenerationTimeout) {
				ws.WriteJSON(wsError{"image generation timeout", 500})
				ws.Close()
				return
			}
			ws.WriteJSON(wsError{"prompt generating error", 500})
			ws.Close()
			return
		}
		imageUrl, err := ram_image_generator.UploadImage(imageBase64)
		if err != nil {
			ws.WriteJSON(wsError{"image uploading error", 500})
			ws.Close()
			return
		}
		imageDescription, err := ram_image_generator.GenerateDescription(ctx, h.gRPCClient, imageUrl)
		if err != nil {
			if errors.Is(err, ram_image_generator.CensorshipError) {
				ws.WriteJSON(wsError{"user prompt or rams descriptions contains illegal content", 400})
				ws.Close()
				return
			}
			ws.WriteJSON(wsError{"image description generating error", 500})
			ws.Close()
			return
		}
		ws.WriteJSON(map[string]string{"status": "image generated"})
		aiGeneratedRam <- entities.Ram{UserId: user.Id, Description: imageDescription, ImageUrl: imageUrl}
	}()

	var needClicks int
	ramsGeneratedLastDay := user.CalculateRamsGeneratedLastDay(config.Conf.Generation.TimeBetweenDaily)
	if user.DailyRamGenerationTime == 0 {
		needClicks = config.Conf.Clicks.FirstRam
	} else {
		if ramsGeneratedLastDay >= len(config.Conf.Clicks.DailyRams) {
			ws.WriteJSON(wsError{fmt.Sprintf("you can generate only %d rams per day", len(config.Conf.Clicks.DailyRams)), 429})
			return
		}
		needClicks = config.Conf.Clicks.DailyRams[ramsGeneratedLastDay]
	}
	ws.WriteJSON(map[string]any{"status": "need clicks", "clicks": needClicks})
	err = h.websocketNeedClicks(ctx, ws, needClicks)
	if err != nil {
		return
	}

	ram := <-aiGeneratedRam
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		ws.WriteJSON(wsError{"unexpected db error", 500})
		return
	}
	_, err = database.CreateRamContext(ctx, tx, ram)
	if err != nil {
		tx.Rollback()
		ws.WriteJSON(wsError{"unexpected db error", 500})
		return
	}

	err = database.UpdateUserContext(ctx, tx, user.Id, entities.User{DailyRamGenerationTime: int(time.Now().Unix()), RamsGeneratedLastDay: ramsGeneratedLastDay + 1})
	if err != nil {
		tx.Rollback()
		ws.WriteJSON(wsError{"unexpected db error", 500})
		return
	}
	tx.Commit()
	ws.WriteJSON(ram)
	return
}
