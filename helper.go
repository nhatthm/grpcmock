package grpcmock

import (
	"strings"

	"google.golang.org/grpc"
)

func methodName(v string) string {
	return "/" + strings.TrimLeft(v, "/")
}

func serviceSorter(services []*grpc.ServiceDesc) ([]*grpc.ServiceDesc, func(i, j int) bool) {
	return services, func(i, j int) bool {
		return services[i].ServiceName < services[j].ServiceName
	}
}

func closeNothing() error {
	return nil
}
