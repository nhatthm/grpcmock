package planner

import (
	"fmt"
)

func recovered(v any) string {
	switch v := v.(type) {
	case error:
		return v.Error()

	case string:
		return v
	}

	return fmt.Sprintf("%+v", v)
}

func trackRepeatable(r repeatableExpectation) bool {
	t := r.RemainTimes()

	if t == UnlimitedTimes {
		return true
	}

	return t > 1
}

func removeExpectations(expectations []Expectation, index int) []Expectation {
	if index == 0 {
		return expectations[1:]
	}

	max := len(expectations) - 1

	if index == max {
		return expectations[:max]
	}

	remains := make([]Expectation, 0, max)

	remains = append(remains, expectations[:index]...)
	remains = append(remains, expectations[index+1:]...)

	return remains
}
