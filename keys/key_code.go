package keys

type Code byte

const (
	CodeQuit     Code = 'q'
	CodePause    Code = 'p'
	CodeReplay   Code = 'r'
	CodeUp       Code = 38
	CodeRight    Code = 39
	CodeDown     Code = 40
	CodeLeft     Code = 37
	CodeMacDir   Code = 91
	CodeWaitExit Code = 0
)

var macDirCodeMap = map[Code]Code{
	65: CodeUp,
	66: CodeDown,
	67: CodeRight,
	68: CodeLeft,
}
