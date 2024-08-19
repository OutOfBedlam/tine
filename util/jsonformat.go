package util

import "fmt"

type JsonFloat struct {
	Value   float64
	Decimal int
}

func (l JsonFloat) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%.*f", l.Decimal, l.Value)
	return []byte(s), nil
}
