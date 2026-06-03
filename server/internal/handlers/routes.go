package handlers

import (
	"net/http"
	"strings"

	"github.com/alexsharav/layout-phone-backend/internal/models"
	"github.com/alexsharav/layout-phone-backend/internal/services"
	"github.com/alexsharav/layout-phone-backend/internal/utils"
	internalws "github.com/alexsharav/layout-phone-backend/internal/websocket"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

type Handler struct {
	gameService *services.GameService
	hub         *internalws.Hub
}

func RegisterRoutes(router *gin.Engine, gameService *services.GameService, hub *internalws.Hub) {
	handler := &Handler{gameService: gameService, hub: hub}

	api := router.Group("/api")
	{
		api.POST("/rooms", handler.createRoom)
		api.GET("/rooms/:code", handler.getRoom)
		api.POST("/rooms/:code/join", handler.joinRoom)
		api.PATCH("/rooms/:code/settings", handler.updateSettings)
		api.POST("/rooms/:code/start", handler.startGame)
		api.POST("/rooms/:code/topics", handler.submitTopic)
		api.GET("/rooms/:code/assignment", handler.getAssignment)
		api.POST("/rooms/:code/submissions", handler.submitCode)
		api.GET("/rooms/:code/results", handler.getResults)
	}

	router.GET("/ws/rooms/:code", handler.websocketRoom)
}

type nicknameRequest struct {
	Nickname string `json:"nickname"`
}

type playerRequest struct {
	PlayerID string `json:"playerId"`
}

type settingsRequest struct {
	PlayerID         string `json:"playerId"`
	MaxPlayers       int    `json:"maxPlayers"`
	TopicWritingTime int    `json:"topicWritingTime"`
	CodingRoundTime  int    `json:"codingRoundTime"`
}

type topicRequest struct {
	PlayerID string `json:"playerId"`
	Text     string `json:"text"`
}

type submissionRequest struct {
	PlayerID             string `json:"playerId"`
	RoundNumber          int    `json:"roundNumber"`
	TopicID              string `json:"topicId"`
	PreviousSubmissionID string `json:"previousSubmissionId"`
	HTML                 string `json:"html"`
	CSS                  string `json:"css"`
	JS                   string `json:"js"`
}

func (h *Handler) createRoom(c *gin.Context) {
	var request nicknameRequest
	if !bindJSON(c, &request) {
		return
	}

	room, player, err := h.gameService.CreateRoom(request.Nickname)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.JSON(c, http.StatusCreated, gin.H{"room": room, "player": player})
}

func (h *Handler) getRoom(c *gin.Context) {
	room, err := h.gameService.GetRoom(c.Param("code"))
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}
	utils.JSON(c, http.StatusOK, gin.H{"room": room})
}

func (h *Handler) joinRoom(c *gin.Context) {
	var request nicknameRequest
	if !bindJSON(c, &request) {
		return
	}

	room, player, err := h.gameService.JoinRoom(c.Param("code"), request.Nickname)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.JSON(c, http.StatusOK, gin.H{"room": room, "player": player})
}

func (h *Handler) updateSettings(c *gin.Context) {
	var request settingsRequest
	if !bindJSON(c, &request) {
		return
	}

	room, err := h.gameService.UpdateSettings(c.Param("code"), request.PlayerID, models.GameSettings{
		MaxPlayers:       request.MaxPlayers,
		TopicWritingTime: request.TopicWritingTime,
		CodingRoundTime:  request.CodingRoundTime,
	})
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.JSON(c, http.StatusOK, gin.H{"room": room})
}

func (h *Handler) startGame(c *gin.Context) {
	var request playerRequest
	if !bindJSON(c, &request) {
		return
	}

	room, err := h.gameService.StartGame(c.Param("code"), request.PlayerID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.JSON(c, http.StatusOK, gin.H{"room": room})
}

func (h *Handler) submitTopic(c *gin.Context) {
	var request topicRequest
	if !bindJSON(c, &request) {
		return
	}

	room, err := h.gameService.SubmitTopic(c.Param("code"), request.PlayerID, request.Text)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.JSON(c, http.StatusOK, gin.H{"room": room})
}

func (h *Handler) getAssignment(c *gin.Context) {
	playerID := c.Query("playerId")
	if playerID == "" {
		utils.BadRequest(c, "Не вижу игрока. Перезайди в комнату.")
		return
	}

	assignment, err := h.gameService.GetAssignment(c.Param("code"), playerID)
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}
	utils.JSON(c, http.StatusOK, gin.H{"assignment": assignment})
}

func (h *Handler) submitCode(c *gin.Context) {
	var request submissionRequest
	if !bindJSON(c, &request) {
		return
	}

	room, err := h.gameService.SubmitCode(c.Param("code"), models.Submission{
		PlayerID:             request.PlayerID,
		RoundNumber:          request.RoundNumber,
		TopicID:              request.TopicID,
		PreviousSubmissionID: request.PreviousSubmissionID,
		HTML:                 request.HTML,
		CSS:                  request.CSS,
		JS:                   request.JS,
	})
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.JSON(c, http.StatusOK, gin.H{"room": room})
}

func (h *Handler) getResults(c *gin.Context) {
	results, err := h.gameService.GetResults(c.Param("code"))
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}
	utils.JSON(c, http.StatusOK, gin.H{"results": results})
}

func (h *Handler) websocketRoom(c *gin.Context) {
	playerID := c.Query("playerId")
	if playerID == "" {
		utils.BadRequest(c, "Не вижу игрока. Перезайди в комнату.")
		return
	}

	upgrader := gorilla.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	code := strings.ToUpper(c.Param("code"))
	client := internalws.NewClient(code, playerID, conn, h.hub)
	h.hub.Register(client)
	client.ReadLoop()
}

func bindJSON(c *gin.Context, request interface{}) bool {
	if err := c.ShouldBindJSON(request); err != nil {
		utils.BadRequest(c, "Не понял запрос. Проверь поля и попробуй еще раз.")
		return false
	}
	return true
}
