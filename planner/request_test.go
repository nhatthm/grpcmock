package planner

import (
	"go.nhat.io/grpcmock/request"
)

func newUnaryRequestWithTimes(i request.RepeatedTime) *request.UnaryRequest {
	r := &request.UnaryRequest{}

	request.SetRepeatability(r, i)

	return r
}
