package main

type PlayersGameData struct {
	Key  string `json:"key"`
	Play bool   `json:"play"`
}
type Dimensions struct {
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
}

type Object struct {
	Position     Vector2    `json:"pos"`
	Size         Dimensions `json:"size"`
	Speed        Vector2
	DefaultSpeed Vector2
	MaxSpeed     Vector2
}

type Board struct {
	Size Dimensions `json:"size"`
	Bar  Dimensions `json:"bar"`
}

type Player struct {
	Bar     Object `json:"bar"`
	Score   int    `json:"score"`
	LastKey string `json:"lastKey"`
}

type Ball struct {
	Object
	HasTouchedPlayer bool `json:"hasTouchedPlayer"`
}

type ServerGameData struct {
	GameStatus bool   `json:"gameStatus"`
	PlayerWon  bool   `json:"hasPlayerWon"`
	PlayLobby  int    // Players in Lobby waiting for game to start, after click on Play Again
	Board      Board  `json:"board"`
	Ball       Ball   `json:"ball"`
	Player1    Player `json:"player1"`
	Player2    Player `json:"player2"`
}

// Vectors
type Vector2 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}
