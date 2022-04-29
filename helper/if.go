package helper

func IfUint(condition bool, v1, v2 uint) uint {
	if condition {
		return v1
	}
	return v2
}

func IfInt(condition bool, v1, v2 int) int {
	if condition {
		return v1
	}
	return v2
}

func IfStr(condition bool, v1, v2 string) string {
	if condition {
		return v1
	}
	return v2
}
