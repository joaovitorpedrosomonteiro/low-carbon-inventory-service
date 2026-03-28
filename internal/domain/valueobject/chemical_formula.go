package valueobject

import (
	"fmt"
	"regexp"
	"strings"
)

type ChemicalFormula struct {
	value string
}

var validFormulaRegex = regexp.MustCompile(`^[A-Z][a-z]?([0-9]*[A-Z][a-z]?)*$`)

var knownFormulas = map[string]bool{
	"CO2":  true,
	"CH4":  true,
	"N2O":  true,
	"SF6":  true,
	"NF3":  true,
	"NOV":  true,
	"F-gas": true,
}

func NewChemicalFormula(value string) (ChemicalFormula, error) {
	cleaned := strings.TrimSpace(value)
	if cleaned == "" {
		return ChemicalFormula{}, fmt.Errorf("chemical formula cannot be empty")
	}

	if !validFormulaRegex.MatchString(cleaned) {
		return ChemicalFormula{}, fmt.Errorf("invalid chemical formula format: %s", cleaned)
	}

	return ChemicalFormula{value: cleaned}, nil
}

func (c ChemicalFormula) String() string {
	return c.value
}

func (c ChemicalFormula) IsKnown() bool {
	return knownFormulas[c.value]
}

func (c ChemicalFormula) IsValid() bool {
	return c.value != "" && validFormulaRegex.MatchString(c.value)
}