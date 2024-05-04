package createfiles

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateUniqueFilename() string {
	currentTime := time.Now().Format("20060102_150405")
	randomNumber := rand.Intn(1000000)
	uniqueFilename := fmt.Sprintf("%s_%06d", currentTime, randomNumber)
	return uniqueFilename
}
