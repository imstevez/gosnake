package gosnake

import (
	"encoding/json"
)

const (
	MSGCMDPing   = "PING"
	MSGCMDPong   = "PONG"
	MSGCMDMov    = "MOV"
	MSGCMDRender = "RENDER"
)

type Message struct {
	CMD  string `json:"cmd"`
	Data []byte `json:"data"`
}

type DirMsg struct {
	Dir Direction `json:"dir"`
}

func encodeMessage(msg *Message) []byte {
	data, _ := json.Marshal(msg)
	return data
}

func decodeMessage(data []byte) *Message {
	var msg Message
	json.Unmarshal(data, &msg)
	return &msg
}

func encodePingMsg() []byte {
	return encodeMessage(&Message{
		CMD: MSGCMDPing,
	})
}

func encodePongMsg() []byte {
	return encodeMessage(&Message{
		CMD: MSGCMDPong,
	})
}

func encodeMovMsg(dir Direction) []byte {
	dirData, _ := json.Marshal(
		[]Direction{dir},
	)
	return encodeMessage(&Message{
		CMD:  MSGCMDMov,
		Data: dirData,
	})
}

func decodeDirData(dirData []byte) Direction {
	var dirs []Direction
	json.Unmarshal(dirData, &dirs)
	return dirs[0]
}

func encodeRenderMsg(result string) []byte {
	resultData, _ := json.Marshal(
		[]string{result},
	)
	return encodeMessage(&Message{
		CMD:  MSGCMDRender,
		Data: resultData,
	})
}

func decodeRenderData(data []byte) string {
	var results []string
	json.Unmarshal(data, &results)
	return results[0]
}
