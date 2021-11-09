package matcher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestIncomingHeader(t *testing.T) {
	t.Parallel()

	data := map[string]string{"locale": "en-US"}

	testCases := []struct {
		scenario string
		ctx      context.Context
		expected metadata.MD
	}{
		{
			scenario: "no md",
			ctx:      context.Background(),
			expected: metadata.New(nil),
		},
		{
			scenario: "has md",
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.New(data)),
			expected: metadata.New(data),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := incomingHeader(tc.ctx)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestHeader_Get(t *testing.T) {
	t.Parallel()

	data := map[string]string{"locale": "en-US"}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(data))

	h := incomingHeader(ctx)

	actual := getHeader(h, "locale")
	expected := "en-US"

	assert.Equal(t, expected, actual)

	assert.Empty(t, getHeader(h, "random"))
}
