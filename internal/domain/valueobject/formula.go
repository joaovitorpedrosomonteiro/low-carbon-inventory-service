package valueobject

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type Formula struct {
	expression string
	variables  map[string]decimal.Decimal
}

var variableRegex = regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)`)

func NewFormula(expression string) (Formula, error) {
	cleaned := strings.TrimSpace(expression)
	if cleaned == "" {
		return Formula{}, fmt.Errorf("formula expression cannot be empty")
	}

	return Formula{
		expression: cleaned,
		variables:  make(map[string]decimal.Decimal),
	}, nil
}

func (f Formula) String() string {
	return f.expression
}

func (f Formula) GetVariables() []string {
	matches := variableRegex.FindAllStringSubmatch(f.expression, -1)
	vars := make([]string, 0)
	seen := make(map[string]bool)

	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			vars = append(vars, m[1])
			seen[m[1]] = true
		}
	}
	return vars
}

func (f *Formula) SetVariable(name string, value decimal.Decimal) {
	if f.variables == nil {
		f.variables = make(map[string]decimal.Decimal)
	}
	f.variables[name] = value
}

func (f Formula) IsCalculable() bool {
	vars := f.GetVariables()
	for _, v := range vars {
		if _, ok := f.variables[v]; !ok {
			return false
		}
	}
	return len(f.variables) > 0
}

func (f Formula) Calculate() (decimal.Decimal, error) {
	if !f.IsCalculable() {
		return decimal.Zero, fmt.Errorf("not all variables are set")
	}

	expr := f.expression
	for varName, value := range f.variables {
		placeholder := fmt.Sprintf("$%s", varName)
		expr = strings.ReplaceAll(expr, placeholder, value.String())
	}

	result, err := evaluateExpression(expr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to evaluate formula: %w", err)
	}

	return result, nil
}

func evaluateExpression(expr string) (decimal.Decimal, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	expr = strings.ToUpper(expr)

	numStr := ""
	hasDecimal := false
	numbers := []decimal.Decimal{}
	operators := []rune{}

	for i, r := range expr {
		if (r >= '0' && r <= '9') || r == '.' {
			if r == '.' && hasDecimal {
				return decimal.Zero, fmt.Errorf("invalid decimal number")
			}
			if r == '.' {
				hasDecimal = true
			}
			numStr += string(r)
			continue
		}

		if numStr != "" {
			if strings.Contains(numStr, ".") {
				f, err := strconv.ParseFloat(numStr, 64)
				if err != nil {
					return decimal.Zero, err
				}
				numbers = append(numbers, decimal.NewFromFloat(f))
			} else {
				f, err := strconv.ParseFloat(numStr, 64)
				if err != nil {
					return decimal.Zero, err
				}
				numbers = append(numbers, decimal.NewFromFloat(f))
			}
			numStr = ""
			hasDecimal = false
		}

		if r == '+' || r == '-' || r == '*' || r == '/' {
			if len(numbers) == 0 && len(operators) > 0 && r == '-' {
				numStr += string(r)
				continue
			}
			operators = append(operators, r)
		} else if r == '(' || r == ')' {
			continue
		} else if r == '^' {
			if len(numbers) >= 2 {
				pow := math.Pow(numbers[len(numbers)-2].InexactFloat64(), numbers[len(numbers)-1].InexactFloat64())
				numbers = numbers[:len(numbers)-2]
				numbers = append(numbers, decimal.NewFromFloat(pow))
			}
		} else if r >= 'A' && r <= 'Z' {
			if i < len(expr)-1 && expr[i+1] >= 'a' && expr[i+1] <= 'z' {
				continue
			}
		}
	}

	if numStr != "" {
		f, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return decimal.Zero, err
		}
		numbers = append(numbers, decimal.NewFromFloat(f))
	}

	result := decimal.Zero
	if len(numbers) > 0 {
		result = numbers[0]
	}

	for i, op := range operators {
		if i+1 >= len(numbers) {
			break
		}
		switch op {
		case '+':
			result = result.Add(numbers[i+1])
		case '-':
			result = result.Sub(numbers[i+1])
		case '*':
			result = result.Mul(numbers[i+1])
		case '/':
			if numbers[i+1].IsZero() {
				return decimal.Zero, fmt.Errorf("division by zero")
			}
			result = result.Div(numbers[i+1])
		}
	}

	return result, nil
}