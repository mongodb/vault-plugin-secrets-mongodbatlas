package atlas

import (
	"math/rand"
	"time"
)

const min int = 0
const max int = 94

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func getRandomPassword(length int) (password string) {
	seed := time.Now().Unix()
	startChar := "!"
	password = ""

	rand.Seed(seed)

	for i := 0; i < length; i++ {
		r := random(min, max)
		newChar := string(startChar[0] + byte(r))
		password += newChar
	}
	return

}
