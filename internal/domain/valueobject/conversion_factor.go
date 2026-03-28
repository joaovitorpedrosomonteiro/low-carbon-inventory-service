package valueobject

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type ConversionFactor struct {
	GasType GasType
	Value   decimal.Decimal
}

func NewConversionFactor(gasType GasType, value decimal.Decimal) (ConversionFactor, error) {
	if !gasType.IsValid() {
		return ConversionFactor{}, fmt.Errorf("invalid gas type")
	}
	if value.IsNegative() {
		return ConversionFactor{}, fmt.Errorf("conversion factor cannot be negative")
	}
	if value.IsZero() {
		return ConversionFactor{}, fmt.Errorf("conversion factor cannot be zero")
	}

	return ConversionFactor{
		GasType: gasType,
		Value:   value,
	}, nil
}

func (cf ConversionFactor) String() string {
	return fmt.Sprintf("%s: %s", cf.GasType.Name, cf.Value.String())
}

func (cf ConversionFactor) IsValid() bool {
	return cf.GasType.IsValid() && !cf.Value.IsZero() && !cf.Value.IsNegative()
}