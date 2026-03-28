package valueobject

import "fmt"

type InventoryState int

const (
	ToReportEmissions InventoryState = iota
	ToProvideEvidence
	ForAuditing
	Audited
	ForReview
)

var stateNames = []string{
	"to_report_emissions",
	"to_provide_evidence",
	"for_auditing",
	"audited",
	"for_review",
}

func (s InventoryState) String() string {
	if s < ToReportEmissions || s > ForReview {
		return ""
	}
	return stateNames[s]
}

func NewInventoryState(value string) (InventoryState, error) {
	for i, name := range stateNames {
		if name == value {
			return InventoryState(i), nil
		}
	}
	return 0, fmt.Errorf("invalid inventory state: %s", value)
}

func (s InventoryState) IsValid() bool {
	return s >= ToReportEmissions && s <= ForReview
}

func (s InventoryState) CanTransitionTo(target InventoryState) bool {
	transitions := map[InventoryState][]InventoryState{
		ToReportEmissions: {ToProvideEvidence},
		ToProvideEvidence: {ForAuditing, ForReview},
		ForAuditing:       {Audited},
		ForReview:         {ForAuditing},
		Audited:           {},
	}

	allowed, ok := transitions[s]
	if !ok {
		return false
	}

	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

func (s InventoryState) IsTerminal() bool {
	return s == Audited
}

func (s InventoryState) IsEditable() bool {
	return s == ToReportEmissions || s == ToProvideEvidence || s == ForReview
}