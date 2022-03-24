package main

import (
	"log"
	"time"
)

// Game Data
var gameData ServerGameData

func game() {
	gameData.GameStatus = true
	log.Println("Game Running")
	gameLoop()
	log.Println("Game Ended")
}

func startGame() {
	resetGameData()
	go game()
}

func gameLoop() {
	tick := time.Tick(time.Second / 60)

	for gameStatus() {
		select {
		case <-tick:
			gameInteraction()
			sendGameDataMessage()
			if !gameStatus() {
				break
			}
		}
	}
}

func increasePointsPlayer(player *Player) {
	(*player).Score += 1
	if (*player).Score == 10 {
		gameData.GameStatus = false
		gameData.PlayerWon = true
	}
}

func isBallCollidedPlayers(player *int) bool {
	var ball = gameData.Ball
	var player1 = gameData.Player1
	var player2 = gameData.Player2

	// Check colision with Player 1
	if ball.Position.X <= (player1.Bar.Position.X+player1.Bar.Size.Width) && ball.Position.X >= (player1.Bar.Position.X) {
		if ball.Position.Y <= player1.Bar.Position.Y+player1.Bar.Size.Height && ball.Position.Y+ball.Size.Height >= player1.Bar.Position.Y {
			*player = 1
			return true
		}
	}

	// Check colision with Player 2
	if ball.Position.X+ball.Size.Width >= player2.Bar.Position.X && ball.Position.X <= player2.Bar.Position.X+player2.Bar.Size.Width {
		if ball.Position.Y <= player2.Bar.Position.Y+player2.Bar.Size.Height && ball.Position.Y+ball.Size.Height >= player2.Bar.Position.Y {
			*player = 2
			return true
		}
	}

	return false
}

func handlerBallColisionWithXAxis() {
	// Detect colision with X Limits
	var ball = &gameData.Ball
	var board_limit = &gameData.Board.Size

	if ((*ball).Position.X) < 0 {
		increasePointsPlayer(&gameData.Player2)
		resetBall(1)
		return
	} else if (*ball).Position.X+(*ball).Size.Width > (*board_limit).Width {
		increasePointsPlayer(&gameData.Player1)
		resetBall(2)
		return
	}
}

func handlerBallColisionWithYAxis() {
	//Detect colision in the Y limits
	var ball = &gameData.Ball
	var board_limit = &gameData.Board.Size
	if (*ball).Position.Y < 0 || (*ball).Position.Y+(*ball).Size.Height > (*board_limit).Height {
		((*ball).Speed.Y) *= -1
	}
}

func handlerBallColisionWithPlayers() {
	// Detect colision with the Players Bar
	var player_nr int = 0
	if isBallCollidedPlayers(&player_nr) {
		handlerBallHasTouchedPlayer(player_nr)
		handlerBallSpeedColisionPlayers(player_nr)
	}
}

func setBallSpeedAfterColisionWithPlayer(player *Player) {
	if player == nil {
		return
	}

	ball := &gameData.Ball
	// Check if the Ball it's the Superior / Inferior / Middle Part of the Player Bar
	if (*ball).Position.Y+(*ball).Size.Height/2 < (*player).Bar.Position.Y+((*player).Bar.Size.Height/3) {
		if (*ball).Speed.Y == 0 {
			(*ball).Speed.Y = -(*ball).MaxSpeed.Y
		} else if (*ball).Speed.Y > 0 {
			(*ball).Speed.Y *= -1
		}
	} else if (*ball).Position.Y+(*ball).Size.Height/2 > (*player).Bar.Position.Y+(2*((*player).Bar.Size.Height/3)) {
		if (*ball).Speed.Y == 0 {
			(*ball).Speed.Y = (*ball).MaxSpeed.Y
		} else if (*ball).Speed.Y < 0 {
			(*ball).Speed.Y *= -1
		}
	} else {
		(*ball).Speed.Y = 0
	}

	(*ball).Speed.X *= -1
}

func handlerBallSpeedColisionPlayers(player_nr int) {
	var player *Player

	if player_nr == 1 {
		player = &gameData.Player1
	} else if player_nr == 2 {
		player = &gameData.Player2
	}

	setBallSpeedAfterColisionWithPlayer(player)
}

func handlerBallHasTouchedPlayer(player_nr int) {
	ball := &gameData.Ball
	if (*ball).HasTouchedPlayer {
		return
	}

	if player_nr == 1 {
		(*ball).Speed.X = -(*ball).MaxSpeed.X
	} else if player_nr == 2 {
		(*ball).Speed.X = (*ball).MaxSpeed.X
	}

	(*ball).HasTouchedPlayer = true
}

func ballMovement() {
	ball := &gameData.Ball
	((*ball).Position.X) += (*ball).Speed.X
	((*ball).Position.Y) += (*ball).Speed.Y

	handlerBallColisionWithXAxis()
	handlerBallColisionWithYAxis()
	handlerBallColisionWithPlayers()
}

func resetGameData() {
	gameData.GameStatus = false
	gameData.PlayerWon = false
	gameData.PlayLobby = 0
	resetPositions()
}

func resetPositions() {
	// Board
	resetBoard()

	// Ball
	resetBall(0)

	// Players
	resetPlayers()
}

func resetBoard() {
	gameData.Board.Size.Width = 640
	gameData.Board.Size.Height = 480
	gameData.Board.Bar.Width = 10
	gameData.Board.Bar.Height = 100
}

func resetBall(num int) {
	ball := &gameData.Ball
	(*ball).Position = Vector2{X: gameData.Board.Size.Width / 2, Y: gameData.Board.Size.Height / 2}
	(*ball).DefaultSpeed = Vector2{X: 2, Y: 2}
	(*ball).MaxSpeed = Vector2{X: 6, Y: 6}
	(*ball).Size.Height = 15
	(*ball).Size.Width = 15
	(*ball).HasTouchedPlayer = false
	(*ball).Speed = (*ball).DefaultSpeed

	// Ball goes to Player 1
	if num == 1 {
		(*ball).Speed.X *= -1
	}
}

func resetPlayers() {
	// Player 1
	resetPlayer(&gameData.Player1)
	gameData.Player1.Bar.Position.X = (gameData.Player1.Bar.Size.Width / 2)
	gameData.Player1.Bar.Position.Y = (gameData.Board.Size.Height / 2) - (gameData.Player1.Bar.Size.Height / 2)

	// Player 2
	resetPlayer(&gameData.Player2)
	gameData.Player2.Bar.Position.X = gameData.Board.Size.Width - (1.5 * gameData.Player2.Bar.Size.Width)
	gameData.Player2.Bar.Position.Y = (gameData.Board.Size.Height / 2) - (gameData.Player2.Bar.Size.Height / 2)

}

func resetPlayer(player *Player) {
	(*player).Bar.Size.Width = gameData.Board.Bar.Width
	(*player).Bar.Size.Height = gameData.Board.Bar.Height
	(*player).Bar.Speed.Y = 20
	(*player).Bar.Speed.X = 0
	(*player).Score = 0
	(*player).LastKey = ""
}

func gameInteraction() {
	ballMovement()
}

func gameStatus() (status bool) {
	return gameData.GameStatus
}

func setLastKeyPressedInPlayer(player *Player, key string) {
	(*player).LastKey = key
}

func handlerWhenKey_W_isPressed() {
	if gameData.Player1.Bar.Position.Y > 0 {
		player := &gameData.Player1
		var keyDescription string = "w"
		handlerLastKey(player, keyDescription)
		(*player).Bar.Position.Y -= (*player).Bar.Speed.Y
	} else {
		return
	}
}

func handlerWhenKey_S_isPressed() {
	if gameData.Player1.Bar.Position.Y+gameData.Player1.Bar.Size.Height < gameData.Board.Size.Height {
		var keyDescription string = "s"
		player := &gameData.Player1
		handlerLastKey(player, keyDescription)
		(*player).Bar.Position.Y += (*player).Bar.Speed.Y
	} else {
		return
	}
}

func handlerWhenKey_ArrowUp_isPressed() {
	if gameData.Player2.Bar.Position.Y > 0 {
		var keyDescription string = "arrowUp"
		player := &gameData.Player2
		handlerLastKey(player, keyDescription)
		(*player).Bar.Position.Y -= (*player).Bar.Speed.Y
	} else {
		return
	}
}

func handlerWhenKey_ArrowDown_isPressed() {
	if gameData.Player2.Bar.Position.Y+gameData.Player2.Bar.Size.Height < gameData.Board.Size.Height {
		var keyDescription string = "arrowDown"
		player := &gameData.Player2
		handlerLastKey(player, keyDescription)
		(*player).Bar.Position.Y += (*player).Bar.Speed.Y
	} else {
		return
	}
}

func handlerPlayerKeyPress(playerData PlayersGameData) {
	var key = playerData.Key
	switch key {
	case "w":
		handlerWhenKey_W_isPressed()
	case "s":
		handlerWhenKey_S_isPressed()
	case "ArrowUp":
		handlerWhenKey_ArrowUp_isPressed()
	case "ArrowDown":
		handlerWhenKey_ArrowDown_isPressed()
	default:
		log.Println("The key sent can't be handled by the server side. Accepted Keys are -> 'w', 's', 'ArrowDown', 'ArrowUp' ")
		return
	}
}

func handlerLastKey(player *Player, keyDescription string) {
	const maxSpeed float32 = 50
	if (*player).LastKey == keyDescription {
		if (*player).Bar.Speed.Y < maxSpeed {
			(*player).Bar.Speed.Y += 5
		}
	} else {
		setLastKeyPressedInPlayer(player, keyDescription)
	}
}

func handlerPlayLobbyVariableStatus() {
	if gameData.PlayLobby == 2 {
		startGame()
	}
}

func handlerPlayerMessage(playerData PlayersGameData) {
	if gameData.GameStatus {
		handlerPlayerKeyPress(playerData)
	} else {
		if playerData.Play {
			gameData.PlayLobby += 1
			handlerPlayLobbyVariableStatus()
		}
	}
}
