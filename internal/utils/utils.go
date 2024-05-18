package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func Contains(s []string, v ...string) bool {
	for _, e := range s {
		for _, f := range v {
			if strings.EqualFold(e, f) {
				return true
			}
		}
	}
	return false
}

func ValidateOrderNumber(number string) error {
	if len(number) == 0 {
		return fmt.Errorf("invalid order number: %s", number)
	}
	sum := int64(0)
	digits := make([]int64, len(number))
	for i, c := range number {
		if c < '0' || c > '9' {
			return fmt.Errorf("invalid order number: %s, should contain only 0-9", number)
		}
		digits[i] = int64(c - '0')
	}

	parity := len(digits) % 2

	for i := 0; i < len(digits); i++ {
		digit := digits[i]
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	if sum%10 != 0 {
		return fmt.Errorf("invalid order number: %s", number)
	}
	return nil
}

func ParseOrderNumber(orderNumber string) (int64, error) {
	return strconv.ParseInt(orderNumber, 10, 64)
}

func FormatOrderNumber(orderNumber int64) string {
	return strconv.FormatInt(orderNumber, 10)
}

func Float64Compare(a, b float64) int64 {
	comparePrecision := math.Pow(10, -6)
	diff := a - b

	switch {
	case diff > comparePrecision:
		return 1
	case diff < -comparePrecision:
		return -1
	default:
		return 0
	}
}
