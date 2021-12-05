package planner

import (
	"fmt"

	"github.com/nhatthm/grpcmock/request"
)

func recovered(v interface{}) string {
	switch v := v.(type) {
	case error:
		return v.Error()

	case string:
		return v
	}

	return fmt.Sprintf("%+v", v)
}

func trackRepeatable(r request.Request) bool {
	t := request.Repeatability(r)

	if t == 0 {
		return true
	}

	if t > 0 {
		request.SetRepeatability(r, t-1)
	}

	return request.Repeatability(r) > 0
}

func removeExpectations(expectations []request.Request, index int) []request.Request {
	if index == 0 {
		return expectations[1:]
	}

	max := len(expectations) - 1

	if index == max {
		return expectations[:max]
	}

	remains := make([]request.Request, 0, max)

	remains = append(remains, expectations[:index]...)
	remains = append(remains, expectations[index+1:]...)

	return remains
}
