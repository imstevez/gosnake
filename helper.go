package gosnake

func TreeStr(condi bool, val1, val2 string) string {
	if condi {
		return val1
	}
	return val2
}
