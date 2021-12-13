package planner

import (
	"github.com/nhatthm/grpcmock/request"
)

func newUnaryRequestWithTimes(i request.RepeatedTime) *request.UnaryRequest {
	r := &request.UnaryRequest{}

	request.SetRepeatability(r, i)

	return r
}
