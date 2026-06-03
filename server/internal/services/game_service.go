package services

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alexsharav/layout-phone-backend/internal/models"
	"github.com/alexsharav/layout-phone-backend/internal/store"
	"github.com/alexsharav/layout-phone-backend/internal/utils"
	ws "github.com/alexsharav/layout-phone-backend/internal/websocket"

	"github.com/google/uuid"
)

type GameService struct {
	store            *store.MemoryStore
	hub              *ws.Hub
	defaultTopicTime int
	defaultRoundTime int
}

func NewGameService(store *store.MemoryStore, hub *ws.Hub, defaultTopicTime int, defaultRoundTime int) *GameService {
	return &GameService{
		store:            store,
		hub:              hub,
		defaultTopicTime: defaultTopicTime,
		defaultRoundTime: defaultRoundTime,
	}
}

func (s *GameService) CreateRoom(nickname string) (*models.Room, *models.Player, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return nil, nil, errors.New("Напиши имя, чтобы создать комнату.")
	}

	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	code := utils.RoomCode(5)
	for s.store.RoomsByCode[code] != nil {
		code = utils.RoomCode(5)
	}

	roomID := uuid.NewString()
	player := &models.Player{
		ID:       uuid.NewString(),
		RoomID:   roomID,
		Nickname: nickname,
		IsHost:   true,
	}

	room := &models.Room{
		ID:     roomID,
		Code:   code,
		HostID: player.ID,
		Status: models.RoomLobby,
		Settings: models.GameSettings{
			MaxPlayers:       4,
			TopicWritingTime: s.defaultTopicTime,
			CodingRoundTime:  s.defaultRoundTime,
		},
		Players:     []*models.Player{player},
		Assignments: make(map[string]models.Assignment),
	}

	s.store.RoomsByCode[code] = room
	return cloneRoom(room), clonePlayer(player), nil
}

func (s *GameService) JoinRoom(code string, nickname string) (*models.Room, *models.Player, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return nil, nil, errors.New("Напиши имя, чтобы зайти в комнату.")
	}

	s.store.Mu.Lock()
	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		s.store.Mu.Unlock()
		return nil, nil, errors.New("Такой комнаты нет. Проверь код или попроси новую ссылку.")
	}
	if room.Status != models.RoomLobby {
		s.store.Mu.Unlock()
		return nil, nil, errors.New("Игра уже началась, в это лобби больше не зайти.")
	}
	if len(room.Players) >= room.Settings.MaxPlayers {
		s.store.Mu.Unlock()
		return nil, nil, errors.New("В комнате уже полный состав.")
	}

	player := &models.Player{
		ID:       uuid.NewString(),
		RoomID:   room.ID,
		Nickname: nickname,
		IsHost:   false,
	}
	room.Players = append(room.Players, player)
	snapshot := cloneRoom(room)
	playerSnapshot := clonePlayer(player)
	s.store.Mu.Unlock()

	s.hub.Broadcast(room.Code, ws.Event{Type: "player_joined", Payload: playerSnapshot})
	s.broadcastRoom(room.Code, snapshot)

	return snapshot, playerSnapshot, nil
}

func (s *GameService) GetRoom(code string) (*models.Room, error) {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		return nil, errors.New("Комната не найдена.")
	}
	return cloneRoom(room), nil
}

func (s *GameService) UpdateSettings(code string, playerID string, settings models.GameSettings) (*models.Room, error) {
	s.store.Mu.Lock()
	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		s.store.Mu.Unlock()
		return nil, errors.New("Комната не найдена.")
	}
	if room.HostID != playerID {
		s.store.Mu.Unlock()
		return nil, errors.New("Настройки может менять только хост.")
	}
	if room.Status != models.RoomLobby {
		s.store.Mu.Unlock()
		return nil, errors.New("После старта настройки уже не трогаем.")
	}

	room.Settings = normalizeSettings(settings)
	if room.Settings.MaxPlayers < len(room.Players) {
		room.Settings.MaxPlayers = len(room.Players)
	}
	snapshot := cloneRoom(room)
	s.store.Mu.Unlock()

	s.hub.Broadcast(room.Code, ws.Event{Type: "settings_updated", Payload: roomSettingsPayload(snapshot)})
	s.broadcastRoom(room.Code, snapshot)
	return snapshot, nil
}

func (s *GameService) StartGame(code string, hostID string) (*models.Room, error) {
	s.store.Mu.Lock()
	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		s.store.Mu.Unlock()
		return nil, errors.New("Комната не найдена.")
	}
	if room.HostID != hostID {
		s.store.Mu.Unlock()
		return nil, errors.New("Стартовать игру может только хост.")
	}
	if room.Status != models.RoomLobby {
		s.store.Mu.Unlock()
		return nil, errors.New("Игра уже стартовала.")
	}
	if len(room.Players) < 2 {
		s.store.Mu.Unlock()
		return nil, errors.New("Нужны хотя бы два игрока.")
	}

	room.Status = models.RoomWritingTopics
	room.CurrentRound = 0
	room.Topics = []*models.Topic{}
	room.Submissions = []*models.Submission{}
	room.ActiveRound = nil
	room.Assignments = make(map[string]models.Assignment)
	snapshot := cloneRoom(room)
	duration := room.Settings.TopicWritingTime
	s.store.Mu.Unlock()

	s.hub.Broadcast(room.Code, ws.Event{Type: "game_started", Payload: map[string]string{"status": string(models.RoomWritingTopics)}})
	s.broadcastRoom(room.Code, snapshot)
	go s.runTopicTimer(room.Code, duration)

	return snapshot, nil
}

func (s *GameService) SubmitTopic(code string, playerID string, text string) (*models.Room, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("Тема пустая. Напиши хотя бы пару слов.")
	}

	s.store.Mu.Lock()
	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		s.store.Mu.Unlock()
		return nil, errors.New("Комната не найдена.")
	}
	if room.Status != models.RoomWritingTopics {
		s.store.Mu.Unlock()
		return nil, errors.New("Время придумывать темы уже закончилось.")
	}
	if findPlayer(room, playerID) == nil {
		s.store.Mu.Unlock()
		return nil, errors.New("Не нашел тебя среди игроков этой комнаты.")
	}

	existing := topicByAuthor(room, playerID)
	if existing != nil {
		existing.Text = text
	} else {
		room.Topics = append(room.Topics, &models.Topic{
			ID:       uuid.NewString(),
			RoomID:   room.ID,
			AuthorID: playerID,
			Text:     text,
		})
	}

	allSubmitted := len(room.Topics) >= len(room.Players)
	snapshot := cloneRoom(room)
	s.store.Mu.Unlock()

	s.hub.Broadcast(snapshot.Code, ws.Event{Type: "topic_submitted", Payload: map[string]string{"playerId": playerID}})
	s.broadcastRoom(snapshot.Code, snapshot)
	if allSubmitted {
		go s.FinishTopicWritingPhase(snapshot.Code)
	}

	return snapshot, nil
}

func (s *GameService) FinishTopicWritingPhase(code string) {
	s.store.Mu.Lock()
	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil || room.Status != models.RoomWritingTopics {
		s.store.Mu.Unlock()
		return
	}

	for _, player := range room.Players {
		if topicByAuthor(room, player.ID) == nil {
			room.Topics = append(room.Topics, &models.Topic{
				ID:       uuid.NewString(),
				RoomID:   room.ID,
				AuthorID: player.ID,
				Text:     fmt.Sprintf("Свободная тема от %s", player.Nickname),
			})
		}
	}
	s.store.Mu.Unlock()

	s.hub.Broadcast(strings.ToUpper(code), ws.Event{Type: "topics_completed", Payload: map[string]bool{"ready": true}})
	s.StartCodingRound(code, 1)
}

func (s *GameService) StartCodingRound(code string, roundNumber int) {
	s.store.Mu.Lock()
	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		s.store.Mu.Unlock()
		return
	}
	if room.Status != models.RoomWritingTopics && room.Status != models.RoomCodingRound {
		s.store.Mu.Unlock()
		return
	}
	if roundNumber > len(room.Players) {
		room.Status = models.RoomResults
		room.ActiveRound = nil
		snapshot := cloneRoom(room)
		s.store.Mu.Unlock()
		s.hub.Broadcast(snapshot.Code, ws.Event{Type: "game_finished", Payload: map[string]string{"resultsUrl": "/results/" + snapshot.Code}})
		s.broadcastRoom(snapshot.Code, snapshot)
		return
	}

	now := time.Now()
	room.Status = models.RoomCodingRound
	room.CurrentRound = roundNumber
	room.ActiveRound = &models.Round{
		RoomID:    room.ID,
		Number:    roundNumber,
		StartedAt: now.Unix(),
		EndsAt:    now.Add(time.Duration(room.Settings.CodingRoundTime) * time.Second).Unix(),
	}
	room.Assignments = buildAssignments(room, roundNumber)

	snapshot := cloneRoom(room)
	assignments := make([]models.Assignment, 0, len(room.Assignments))
	for _, assignment := range room.Assignments {
		assignments = append(assignments, assignment)
	}
	duration := room.Settings.CodingRoundTime
	s.store.Mu.Unlock()

	s.broadcastRoom(snapshot.Code, snapshot)
	for _, assignment := range assignments {
		s.hub.SendToPlayer(snapshot.Code, assignment.PlayerID, ws.Event{Type: "round_started", Payload: map[string]interface{}{
			"roundNumber": roundNumber,
			"assignment":  assignment,
		}})
	}
	go s.runRoundTimer(snapshot.Code, roundNumber, duration)
}

func (s *GameService) GetAssignment(code string, playerID string) (models.Assignment, error) {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		return models.Assignment{}, errors.New("Комната не найдена.")
	}
	assignment, ok := room.Assignments[playerID]
	if !ok {
		return models.Assignment{}, errors.New("Задание еще не выдали. Подожди пару секунд.")
	}
	return assignment, nil
}

func (s *GameService) SubmitCode(code string, request models.Submission) (*models.Room, error) {
	s.store.Mu.Lock()
	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		s.store.Mu.Unlock()
		return nil, errors.New("Комната не найдена.")
	}
	if room.Status != models.RoomCodingRound {
		s.store.Mu.Unlock()
		return nil, errors.New("Сейчас не идет раунд верстки.")
	}
	if request.RoundNumber != room.CurrentRound {
		s.store.Mu.Unlock()
		return nil, errors.New("Этот раунд уже не актуален.")
	}
	if findPlayer(room, request.PlayerID) == nil {
		s.store.Mu.Unlock()
		return nil, errors.New("Не нашел тебя среди игроков этой комнаты.")
	}
	if submissionByPlayerRound(room, request.PlayerID, request.RoundNumber) != nil {
		s.store.Mu.Unlock()
		return nil, errors.New("Этот раунд от тебя уже записан.")
	}

	assignment := room.Assignments[request.PlayerID]
	if assignment.PlayerID == "" {
		s.store.Mu.Unlock()
		return nil, errors.New("Задание для тебя не нашлось. Обнови страницу.")
	}
	if request.TopicID == "" {
		request.TopicID = assignment.TopicID
	}
	if request.PreviousSubmissionID == "" {
		request.PreviousSubmissionID = assignment.PreviousSubmissionID
	}

	submission := &models.Submission{
		ID:                   uuid.NewString(),
		RoomID:               room.ID,
		TopicID:              request.TopicID,
		RoundNumber:          request.RoundNumber,
		PlayerID:             request.PlayerID,
		PreviousSubmissionID: request.PreviousSubmissionID,
		HTML:                 request.HTML,
		CSS:                  request.CSS,
		JS:                   request.JS,
	}
	room.Submissions = append(room.Submissions, submission)
	allSubmitted := submissionsCountForRound(room, room.CurrentRound) >= len(room.Players)
	snapshot := cloneRoom(room)
	s.store.Mu.Unlock()

	s.hub.Broadcast(snapshot.Code, ws.Event{Type: "submission_received", Payload: map[string]interface{}{
		"playerId": request.PlayerID,
		"round":    request.RoundNumber,
	}})
	s.broadcastRoom(snapshot.Code, snapshot)
	if allSubmitted {
		go s.FinishCodingRound(snapshot.Code, request.RoundNumber)
	}

	return snapshot, nil
}

func (s *GameService) FinishCodingRound(code string, roundNumber int) {
	s.store.Mu.Lock()
	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil || room.Status != models.RoomCodingRound || room.CurrentRound != roundNumber {
		s.store.Mu.Unlock()
		return
	}

	for _, player := range room.Players {
		if submissionByPlayerRound(room, player.ID, roundNumber) != nil {
			continue
		}
		assignment := room.Assignments[player.ID]
		room.Submissions = append(room.Submissions, &models.Submission{
			ID:                   uuid.NewString(),
			RoomID:               room.ID,
			TopicID:              assignment.TopicID,
			RoundNumber:          roundNumber,
			PlayerID:             player.ID,
			PreviousSubmissionID: assignment.PreviousSubmissionID,
			HTML:                 fallbackHTML(player.Nickname, assignment),
			CSS:                  fallbackCSS(),
			JS:                   "",
		})
	}

	lastRound := roundNumber >= len(room.Players)
	if lastRound {
		room.Status = models.RoomResults
		room.ActiveRound = nil
		snapshot := cloneRoom(room)
		s.store.Mu.Unlock()
		s.hub.Broadcast(snapshot.Code, ws.Event{Type: "game_finished", Payload: map[string]string{"resultsUrl": "/results/" + snapshot.Code}})
		s.broadcastRoom(snapshot.Code, snapshot)
		return
	}

	s.store.Mu.Unlock()
	s.StartCodingRound(code, roundNumber+1)
}

func (s *GameService) GetResults(code string) ([]models.TopicResult, error) {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil {
		return nil, errors.New("Комната не найдена.")
	}

	results := make([]models.TopicResult, 0, len(room.Topics))
	for _, topic := range room.Topics {
		author := findPlayer(room, topic.AuthorID)
		if author == nil {
			continue
		}

		submissions := make([]models.SubmissionResult, 0)
		for _, submission := range room.Submissions {
			if submission.TopicID != topic.ID {
				continue
			}
			player := findPlayer(room, submission.PlayerID)
			if player == nil {
				continue
			}
			submissions = append(submissions, models.SubmissionResult{
				ID:                   submission.ID,
				RoundNumber:          submission.RoundNumber,
				Player:               *clonePlayer(player),
				PreviousSubmissionID: submission.PreviousSubmissionID,
				HTML:                 submission.HTML,
				CSS:                  submission.CSS,
				JS:                   submission.JS,
			})
		}
		sort.Slice(submissions, func(i, j int) bool {
			return submissions[i].RoundNumber < submissions[j].RoundNumber
		})

		results = append(results, models.TopicResult{
			Topic:       *cloneTopic(topic),
			Author:      *clonePlayer(author),
			Submissions: submissions,
		})
	}

	return results, nil
}

func (s *GameService) runTopicTimer(code string, duration int) {
	for left := duration; left >= 0; left-- {
		if !s.isRoomStatus(code, models.RoomWritingTopics, 0) {
			return
		}
		s.hub.Broadcast(strings.ToUpper(code), ws.Event{Type: "timer_tick", Payload: map[string]int{"secondsLeft": left}})
		if left == 0 {
			s.FinishTopicWritingPhase(code)
			return
		}
		time.Sleep(time.Second)
	}
}

func (s *GameService) runRoundTimer(code string, roundNumber int, duration int) {
	for left := duration; left >= 0; left-- {
		if !s.isRoomStatus(code, models.RoomCodingRound, roundNumber) {
			return
		}
		s.hub.Broadcast(strings.ToUpper(code), ws.Event{Type: "timer_tick", Payload: map[string]int{"secondsLeft": left}})
		if left == 0 {
			s.FinishCodingRound(code, roundNumber)
			return
		}
		time.Sleep(time.Second)
	}
}

func (s *GameService) isRoomStatus(code string, status models.RoomStatus, roundNumber int) bool {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	room := s.store.RoomsByCode[strings.ToUpper(code)]
	if room == nil || room.Status != status {
		return false
	}
	return roundNumber == 0 || room.CurrentRound == roundNumber
}

func (s *GameService) broadcastRoom(code string, room *models.Room) {
	s.hub.Broadcast(code, ws.Event{Type: "room_updated", Payload: room})
}

func buildAssignments(room *models.Room, roundNumber int) map[string]models.Assignment {
	assignments := make(map[string]models.Assignment)
	topicsInPlayerOrder := make([]*models.Topic, 0, len(room.Players))
	for _, player := range room.Players {
		if topic := topicByAuthor(room, player.ID); topic != nil {
			topicsInPlayerOrder = append(topicsInPlayerOrder, topic)
		}
	}
	if len(topicsInPlayerOrder) == 0 {
		return assignments
	}

	for index, player := range room.Players {
		topic := topicsInPlayerOrder[(index+roundNumber)%len(topicsInPlayerOrder)]

		if roundNumber == 1 {
			assignments[player.ID] = models.Assignment{
				PlayerID:    player.ID,
				RoundNumber: roundNumber,
				TopicID:     topic.ID,
				TopicText:   topic.Text,
			}
			continue
		}

		submission := submissionByTopicRound(room, topic.ID, roundNumber-1)
		if submission == nil {
			continue
		}
		assignments[player.ID] = models.Assignment{
			PlayerID:             player.ID,
			RoundNumber:          roundNumber,
			TopicID:              submission.TopicID,
			PreviousSubmissionID: submission.ID,
			HTML:                 submission.HTML,
			CSS:                  submission.CSS,
			JS:                   submission.JS,
		}
	}

	return assignments
}

func normalizeSettings(settings models.GameSettings) models.GameSettings {
	if settings.MaxPlayers < 2 {
		settings.MaxPlayers = 2
	}
	if settings.MaxPlayers > 10 {
		settings.MaxPlayers = 10
	}
	if settings.TopicWritingTime < 10 {
		settings.TopicWritingTime = 10
	}
	if settings.CodingRoundTime < 30 {
		settings.CodingRoundTime = 30
	}
	return settings
}

func roomSettingsPayload(room *models.Room) map[string]int {
	return map[string]int{
		"maxPlayers":       room.Settings.MaxPlayers,
		"topicWritingTime": room.Settings.TopicWritingTime,
		"codingRoundTime":  room.Settings.CodingRoundTime,
	}
}

func findPlayer(room *models.Room, playerID string) *models.Player {
	for _, player := range room.Players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}

func topicByAuthor(room *models.Room, playerID string) *models.Topic {
	for _, topic := range room.Topics {
		if topic.AuthorID == playerID {
			return topic
		}
	}
	return nil
}

func submissionByPlayerRound(room *models.Room, playerID string, roundNumber int) *models.Submission {
	for _, submission := range room.Submissions {
		if submission.PlayerID == playerID && submission.RoundNumber == roundNumber {
			return submission
		}
	}
	return nil
}

func submissionsCountForRound(room *models.Room, roundNumber int) int {
	count := 0
	for _, submission := range room.Submissions {
		if submission.RoundNumber == roundNumber {
			count++
		}
	}
	return count
}

func submissionByTopicRound(room *models.Room, topicID string, roundNumber int) *models.Submission {
	for _, submission := range room.Submissions {
		if submission.TopicID == topicID && submission.RoundNumber == roundNumber {
			return submission
		}
	}
	return nil
}

func fallbackHTML(nickname string, assignment models.Assignment) string {
	if assignment.HTML != "" {
		return assignment.HTML
	}
	title := assignment.TopicText
	if title == "" {
		title = "Сайт, который пришлось спасать на бегу"
	}
	return fmt.Sprintf("<main class=\"page\"><h1>%s</h1><p>%s не успел сдать раунд, поэтому игра бережно оставила эту страницу жить дальше.</p><button>посмотреть</button></main>", title, nickname)
}

func fallbackCSS() string {
	return "body{margin:0;font-family:Inter,Arial,sans-serif;background:#f6f2ea;color:#202426}.page{min-height:100vh;display:grid;place-items:center;text-align:center;padding:48px}button{border:0;background:#2f6f5e;color:white;padding:12px 18px;border-radius:6px}"
}

func cloneRoom(room *models.Room) *models.Room {
	clone := *room
	clone.Players = make([]*models.Player, 0, len(room.Players))
	for _, player := range room.Players {
		clone.Players = append(clone.Players, clonePlayer(player))
	}
	clone.Topics = make([]*models.Topic, 0, len(room.Topics))
	for _, topic := range room.Topics {
		clone.Topics = append(clone.Topics, cloneTopic(topic))
	}
	clone.Submissions = make([]*models.Submission, 0, len(room.Submissions))
	for _, submission := range room.Submissions {
		item := *submission
		clone.Submissions = append(clone.Submissions, &item)
	}
	if room.ActiveRound != nil {
		activeRound := *room.ActiveRound
		clone.ActiveRound = &activeRound
	}
	clone.Assignments = nil
	return &clone
}

func clonePlayer(player *models.Player) *models.Player {
	clone := *player
	return &clone
}

func cloneTopic(topic *models.Topic) *models.Topic {
	clone := *topic
	return &clone
}
