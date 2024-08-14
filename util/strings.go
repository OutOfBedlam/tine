package util

func CamelToSnake(s string) string {
	return CamelToSep(s, "_")
}

func CamelToKebab(s string) string {
	return CamelToSep(s, "-")
}

func CamelToSep(s string, sep string) string {
	ret := ""
	for i, c := range s {
		if 'A' <= c && c <= 'Z' {
			if i > 0 {
				ret += sep
			}
			ret += string(c + 32)
		} else {
			ret += string(c)
		}
	}
	return ret
}
