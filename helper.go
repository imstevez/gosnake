package gosnake

func IfStr(condition bool, val1, val2 string) string {
	if condition {
		return val1
	}
	return val2
}

func IfInt(condition bool, val1, val2 int) int {
	if condition {
		return val1
	}
	return val2
}
