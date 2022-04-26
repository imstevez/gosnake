package gosnake

import (
	"bytes"
	"encoding/gob"
)

type SceneData struct {
	PlayerID     string
	BorderWidth  int
	BorderHeight int
	PlayerSnake  *CompressLayer
	Snakes       *CompressLayer
	Food         *CompressLayer
	PlayerStats  PlayerStats
}

func (scd *SceneData) Encode() []byte {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(scd)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func DecodeSceneData(data []byte) *SceneData {
	scd := &SceneData{}
	buf := bytes.NewBuffer(data)
	err := gob.NewDecoder(buf).Decode(scd)
	if err != nil {
		panic(err)
	}
	return scd
}
