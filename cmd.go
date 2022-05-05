package gosnake

import (
	"gosnake/base"
	"gosnake/keys"
)

type PlayerCMD uint8

const (
	CMDPing PlayerCMD = iota + 1
	CMDJoin
	CMDPause
	CMDReplay
	CMDQuit
	CMDMoveLeft
	CMDMoveRight
	CMDMoveUp
	CMDMoveDown
)

func EncodePlayerData(cmd PlayerCMD) []byte {
	return []byte{byte(cmd)}
}

func DecodePlayerData(data []byte) PlayerCMD {
	return PlayerCMD(data[0])
}

type GameCMD uint16

const (
	CMDPong GameCMD = iota + 1
	CMDUpdate
)

func EncodeGameData(cmd GameCMD, content []byte) (data []byte) {
	data = append(content, byte(cmd))
	return
}

func DecodeGameData(data []byte) (cmd PlayerCMD, content []byte) {
	cmd = PlayerCMD(data[len(data)-1])
	content = data[:len(data)-1]
	return
}

func GetKeyCodeCMD(code keys.Code) PlayerCMD {
	return keyCodeToPlayerCMD[code]
}

var keyCodeToPlayerCMD = map[keys.Code]PlayerCMD{
	keys.CodePause:  CMDPause,
	keys.CodeReplay: CMDReplay,
	keys.CodeQuit:   CMDQuit,
	keys.CodeUp:     CMDMoveUp,
	keys.CodeRight:  CMDMoveRight,
	keys.CodeDown:   CMDMoveDown,
	keys.CodeLeft:   CMDMoveLeft,
	keys.CodeUp2:    CMDMoveUp,
	keys.CodeRight2: CMDMoveRight,
	keys.CodeDown2:  CMDMoveDown,
	keys.CodeLeft2:  CMDMoveLeft,
}

func GetMoveCMDDir(cmd PlayerCMD) base.Direction2D {
	return playerMoveCMDToDir[cmd]
}

var playerMoveCMDToDir = map[PlayerCMD]base.Direction2D{
	CMDMoveUp:    base.Dir2DUp,
	CMDMoveRight: base.Dir2DRight,
	CMDMoveDown:  base.Dir2DDown,
	CMDMoveLeft:  base.Dir2DLeft,
}
