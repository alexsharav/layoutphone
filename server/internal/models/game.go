package models

type RoomStatus string

const (
	RoomLobby         RoomStatus = "lobby"
	RoomWritingTopics RoomStatus = "writing_topics"
	RoomCodingRound   RoomStatus = "coding_round"
	RoomResults       RoomStatus = "results"
	RoomFinished      RoomStatus = "finished"
)

type GameSettings struct {
	MaxPlayers       int `json:"maxPlayers"`
	TopicWritingTime int `json:"topicWritingTime"`
	CodingRoundTime  int `json:"codingRoundTime"`
}

type Player struct {
	ID       string `json:"id"`
	RoomID   string `json:"roomId"`
	Nickname string `json:"nickname"`
	IsHost   bool   `json:"isHost"`
}

type Topic struct {
	ID       string `json:"id"`
	RoomID   string `json:"roomId"`
	AuthorID string `json:"authorId"`
	Text     string `json:"text"`
}

type Submission struct {
	ID                   string `json:"id"`
	RoomID               string `json:"roomId"`
	TopicID              string `json:"topicId"`
	RoundNumber          int    `json:"roundNumber"`
	PlayerID             string `json:"playerId"`
	PreviousSubmissionID string `json:"previousSubmissionId"`
	HTML                 string `json:"html"`
	CSS                  string `json:"css"`
	JS                   string `json:"js"`
}

type Round struct {
	RoomID    string `json:"roomId"`
	Number    int    `json:"number"`
	StartedAt int64  `json:"startedAt"`
	EndsAt    int64  `json:"endsAt"`
}

type Assignment struct {
	PlayerID             string `json:"playerId"`
	RoundNumber          int    `json:"roundNumber"`
	TopicID              string `json:"topicId"`
	TopicText            string `json:"topicText,omitempty"`
	PreviousSubmissionID string `json:"previousSubmissionId,omitempty"`
	HTML                 string `json:"html,omitempty"`
	CSS                  string `json:"css,omitempty"`
	JS                   string `json:"js,omitempty"`
}

type Room struct {
	ID           string                `json:"id"`
	Code         string                `json:"code"`
	HostID       string                `json:"hostId"`
	Status       RoomStatus            `json:"status"`
	Settings     GameSettings          `json:"settings"`
	Players      []*Player             `json:"players"`
	Topics       []*Topic              `json:"topics"`
	Submissions  []*Submission         `json:"submissions"`
	CurrentRound int                   `json:"currentRound"`
	ActiveRound  *Round                `json:"activeRound,omitempty"`
	Assignments  map[string]Assignment `json:"-"`
}

type TopicResult struct {
	Topic       Topic              `json:"topic"`
	Author      Player            `json:"author"`
	Submissions []SubmissionResult `json:"submissions"`
}

type SubmissionResult struct {
	ID                   string `json:"id"`
	RoundNumber          int    `json:"roundNumber"`
	Player               Player `json:"player"`
	PreviousSubmissionID string `json:"previousSubmissionId"`
	HTML                 string `json:"html"`
	CSS                  string `json:"css"`
	JS                   string `json:"js"`
}
