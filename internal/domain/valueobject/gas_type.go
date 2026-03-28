package valueobject

import "fmt"

type GasType struct {
	Formula ChemicalFormula
	Name    string
}

var predefinedGasTypes = map[string]GasType{
	"CO2":  {Formula: ChemicalFormula{value: "CO2"}, Name: "Dióxido de Carbono"},
	"CH4":  {Formula: ChemicalFormula{value: "CH4"}, Name: "Metano"},
	"N2O":  {Formula: ChemicalFormula{value: "N2O"}, Name: "Óxido Nitroso"},
	"SF6":  {Formula: ChemicalFormula{value: "SF6"}, Name: "Hexafluoreto de Enxofre"},
	"NF3":  {Formula: ChemicalFormula{value: "NF3"}, Name: "Trifluoreto de Nitrogênio"},
	"F-gas": {Formula: ChemicalFormula{value: "F-gas"}, Name: "Gás Fluorado"},
}

func NewGasType(formula ChemicalFormula, name string) (GasType, error) {
	if !formula.IsValid() {
		return GasType{}, fmt.Errorf("invalid formula")
	}
	if name == "" {
		return GasType{}, fmt.Errorf("name cannot be empty")
	}

	return GasType{
		Formula: formula,
		Name:    name,
	}, nil
}

func NewGasTypeFromFormula(formula string) (GasType, error) {
	cf, err := NewChemicalFormula(formula)
	if err != nil {
		return GasType{}, err
	}

	if gt, ok := predefinedGasTypes[cf.String()]; ok {
		return gt, nil
	}

	return GasType{
		Formula: cf,
		Name:    cf.String(),
	}, nil
}

func (gt GasType) String() string {
	return gt.Formula.String()
}

func (gt GasType) IsValid() bool {
	return gt.Formula.IsValid() && gt.Name != ""
}