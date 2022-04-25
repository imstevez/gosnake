package gosnake

import "gosnake/keys"

type CMD string

const (
	CMDPing     CMD = "PING"
	CMDPong     CMD = "PONG"
	CMDPause    CMD = "PAUSE"
	CMDReplay   CMD = "REPLAY"
	CMDQuit     CMD = "QUIT"
	CMDMovLeft  CMD = "MOVE_LEFT"
	CMDMovRight CMD = "MOVE_RIGHT"
	CMDMovUp    CMD = "MOVE_UP"
	CMDMovDown  CMD = "MOVE_DOWN"
)

var keyCodeToCMD = map[keys.Code]CMD{
	keys.CodePause:  CMDPause,
	keys.CodeReplay: CMDReplay,
	keys.CodeQuit:   CMDQuit,
	keys.CodeUp:     CMDMovUp,
	keys.CodeDown:   CMDMovDown,
	keys.CodeLeft:   CMDMovLeft,
	keys.CodeRight:  CMDMovRight,
	keys.CodeUp2:    CMDMovUp,
	keys.CodeDown2:  CMDMovDown,
	keys.CodeLeft2:  CMDMovLeft,
	keys.CodeRight2: CMDMovRight,
}

func GetKeyCodeCMD(keycode keys.Code) CMD {
	return keyCodeToCMD[keycode]
}

var cmdToDir = map[CMD]Direction{
	CMDMovUp:    DirUp,
	CMDMovDown:  DirDown,
	CMDMovLeft:  DirLeft,
	CMDMovRight: DirRight,
}

func GetCMDDir(cmd CMD) (dir Direction, ok bool) {
	dir, ok = cmdToDir[cmd]
	return
}
