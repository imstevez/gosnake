package gosnake

import (
	"bytes"
	"encoding/gob"
)

type GameSceneData struct {
	PlayerID   string
	Options    GameRoomOptions
	Players    PlayerList
	FoodPos    Position
	FoodSymbol string
}

func (data *GameSceneData) EncodeForPlayer(playerID string) []byte {
	data.PlayerID = playerID
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	encoder.Encode(data)
	return buf.Bytes()
}
