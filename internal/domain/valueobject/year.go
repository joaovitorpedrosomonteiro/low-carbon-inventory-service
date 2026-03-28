package valueobject

import (
	"fmt"
	"time"
)

type Year int

func NewYear(value int) (Year, error) {
	currentYear := time.Now().Year()
	maxYear := currentYear + 1

	if value < 2000 || value > maxYear {
		return 0, fmt.Errorf("year must be between 2000 and %d, got %d", maxYear, value)
	}
	return Year(value), nil
}

func CurrentYear() Year {
	return Year(time.Now().Year())
}

func (y Year) Int() int {
	return int(y)
}

func (y Year) IsValid() bool {
	currentYear := time.Now().Year()
	return int(y) >= 2000 && int(y) <= currentYear+1
}