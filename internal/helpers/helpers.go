package helpers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CalculateLuhn return the check number
func CalculateLuhn(number int64) int64 {
	checkNumber := checksum(number)

	if checkNumber == 0 {
		return 0
	}
	return 10 - checkNumber
}

// ValidLuhn Valid check number is valid or not based on Luhn algorithm
func ValidLuhn(number int64) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int64) int64 {
	var luhn int64

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}

func RespondError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"error": message})
}

func RespondInternalServerError(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
