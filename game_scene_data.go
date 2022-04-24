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

type PlayerList []*Player

func (pl PlayerList) Len() int {
	return len(pl)
}

func (pl PlayerList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

func (pl PlayerList) Less(i, j int) bool {
	if pl[i].Score != pl[j].Score {
		return pl[i].Score > pl[j].Score
	}
	return pl[i].CreatedAt.Before(
		pl[j].CreatedAt,
	)
}
