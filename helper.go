package gosnake

func NoitherStr(condi bool, val1, val2 string) string {
	if condi {
		return val1
	}
	return val2
}

func NoitherInt(condi bool, val1, val2 int) int {
	if condi {
		return val1
	}
	return val2
}
