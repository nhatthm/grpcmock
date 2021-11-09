package planner

import (
	"github.com/nhatthm/grpcmock/request"
)

func newUnaryRequestWithTimes(i int) *request.UnaryRequest {
	r := &request.UnaryRequest{}

	request.SetRepeatability(r, i)

	return r
}
