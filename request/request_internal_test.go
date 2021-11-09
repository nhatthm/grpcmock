package request

import "sync"

func emptyBaseRequest() baseRequest {
	return baseRequest{locker: &sync.Mutex{}}
}
