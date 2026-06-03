package utils

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RoomCode(size int) string {
	var builder strings.Builder
	for i := 0; i < size; i++ {
		builder.WriteByte(alphabet[rand.Intn(len(alphabet))])
	}
	return builder.String()
}
