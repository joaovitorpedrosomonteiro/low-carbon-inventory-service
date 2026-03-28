package inventory

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/joaovitorpedrosomonteiro/low-carbon-inventory-service/internal/domain/valueobject"
	"github.com/shopspring/decimal"
)

type Emission struct {
	ID            uuid.UUID             `json:"id"`
	Name          string                `json:"name"`
	GasType       valueobject.GasType   `json:"gas_type"`
	Formula       valueobject.Formula   `json:"formula"`
	Evidences     []*Link               `json:"evidences"`
	Category      *Category             `json:"category"`
	ReliabilityJobID *string             `json:"reliability_job_id"`
}

func NewEmission(id uuid.UUID, name string, gasType valueobject.GasType, formula valueobject.Formula, category *Category) (*Emission, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if !gasType.IsValid() {
		return nil, fmt.Errorf("invalid gas type")
	}
	if category == nil {
		return nil, fmt.Errorf("category cannot be nil")
	}

	return &Emission{
		ID:            id,
		Name:          name,
		GasType:       gasType,
		Formula:       formula,
		Evidences:     make([]*Link, 0),
		Category:      category,
		ReliabilityJobID: nil,
	}, nil
}

func (e *Emission) AddEvidence(link *Link) error {
	if link == nil {
		return fmt.Errorf("link cannot be nil")
	}
	if !link.IsValid() {
		return fmt.Errorf("invalid link")
	}

	e.Evidences = append(e.Evidences, link)
	return nil
}

func (e Emission) IsComplete() bool {
	return e.Formula.IsCalculable()
}

func (e Emission) TotalEmissionTons() (decimal.Decimal, error) {
	if !e.IsComplete() {
		return decimal.Zero, fmt.Errorf("formula not calculable")
	}

	result, err := e.Formula.Calculate()
	if err != nil {
		return decimal.Zero, err
	}

	return result, nil
}

func (e Emission) TotalCO2Equivalent(gwp decimal.Decimal) (decimal.Decimal, error) {
	if !e.IsComplete() {
		return decimal.Zero, fmt.Errorf("formula not calculable")
	}

	total, err := e.TotalEmissionTons()
	if err != nil {
		return decimal.Zero, err
	}

	return total.Mul(gwp), nil
}

func (e *Emission) SetVariable(name string, value decimal.Decimal) {
	e.Formula.SetVariable(name, value)
}