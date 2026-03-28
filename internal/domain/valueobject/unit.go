package valueobject

import "fmt"

type Unit struct {
	Numerator   string
	Denominator string
}

var validUnits = map[string]bool{
	"kg":    true,
	"t":     true,
	"L":     true,
	"m3":    true,
	"kWh":   true,
	"MWh":   true,
	"g":     true,
	"mL":    true,
	" GJ":   true,
	" RPM":  true,
}

var validGWPGases = map[string]bool{
	"CO2":  true,
	"CH4":  true,
	"N2O":  true,
	"SF6":  true,
	"NF3":  true,
}

func NewUnit(numerator, denominator string) (Unit, error) {
	num := numerator
	den := denominator

	if num == "" {
		return Unit{}, fmt.Errorf("numerator cannot be empty")
	}

	if !validUnits[num] {
		return Unit{}, fmt.Errorf("invalid numerator unit: %s", num)
	}

	if den != "" && !validUnits[den] {
		return Unit{}, fmt.Errorf("invalid denominator unit: %s", den)
	}

	u := Unit{
		Numerator:   num,
		Denominator: den,
	}
	return u, nil
}

func (u Unit) String() string {
	if u.Denominator == "" {
		return u.Numerator
	}
	return fmt.Sprintf("%s/%s", u.Numerator, u.Denominator)
}

func (u Unit) IsValid() bool {
	if u.Numerator == "" {
		return false
	}
	if !validUnits[u.Numerator] {
		return false
	}
	if u.Denominator != "" && !validUnits[u.Denominator] {
		return false
	}
	return true
}

func (u Unit) IsGHGUnit() bool {
	return validGWPGases[u.Numerator]
}