package model

import (
	"fmt"
	"strings"
)

type DiscSource string

func (s DiscSource) IsOptical() bool {
	return strings.HasPrefix(string(s), "disc:") || strings.HasPrefix(string(s), "dev:")
}

func DiscSourceFromIndex(n int) DiscSource {
	return DiscSource(fmt.Sprintf("disc:%d", n))
}
