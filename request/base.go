package request

import (
	"sync"

	"github.com/spf13/afero"

	"github.com/nhatthm/grpcmock/service"
)

type baseRequest struct {
	locker      sync.Locker
	serviceDesc *service.Method

	// Amount of times this request has been executed.
	totalCalls int

	// The number of times to return the return arguments when setting
	// expectations. 0 means to always return the value.
	repeatability RepeatedTime

	fs afero.Fs
}

func (r *baseRequest) lock() {
	r.locker.Lock()
}

func (r *baseRequest) unlock() {
	r.locker.Unlock()
}

func (r *baseRequest) withFs(fs afero.Fs) {
	r.lock()
	defer r.unlock()

	r.fs = fs
}

func (r *baseRequest) service() service.Method {
	return *r.serviceDesc
}

func (r *baseRequest) getRepeatability() RepeatedTime {
	return r.repeatability
}

func (r *baseRequest) setRepeatability(i RepeatedTime) {
	r.repeatability = i
}

func (r *baseRequest) numCalls() int {
	return r.totalCalls
}

func (r *baseRequest) countCall() {
	r.totalCalls++
}
