package valueobject

import "fmt"

type Month int

const (
	January Month = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

var monthNames = []string{
	"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
	"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
}

func (m Month) String() string {
	if m < January || m > December {
		return ""
	}
	return monthNames[m-1]
}

func (m Month) Int() int {
	return int(m)
}

func NewMonth(value int) (Month, error) {
	if value < 1 || value > 12 {
		return 0, fmt.Errorf("month must be between 1 and 12, got %d", value)
	}
	return Month(value), nil
}

func NewMonthFromString(value string) (Month, error) {
	for i, name := range monthNames {
		if name == value {
			return Month(i + 1), nil
		}
	}
	return 0, fmt.Errorf("invalid month name: %s", value)
}

func (m Month) IsValid() bool {
	return m >= January && m <= December
}