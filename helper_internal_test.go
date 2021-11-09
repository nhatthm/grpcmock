package grpcmock

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestServiceSorter(t *testing.T) {
	t.Parallel()

	actual := []*grpc.ServiceDesc{
		{ServiceName: "ListItems"},
		{ServiceName: "CreateItems"},
		{ServiceName: "TransformItems"},
		{ServiceName: "GetItems"},
	}

	sort.Slice(serviceSorter(actual))

	expected := []*grpc.ServiceDesc{
		{ServiceName: "CreateItems"},
		{ServiceName: "GetItems"},
		{ServiceName: "ListItems"},
		{ServiceName: "TransformItems"},
	}

	assert.Equal(t, expected, actual)
}
