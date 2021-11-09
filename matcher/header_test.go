package matcher_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/nhatthm/go-matcher"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	grpcMatcher "github.com/nhatthm/grpcmock/matcher"
)

func TestHeaderMatcher_Match(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		matcher       grpcMatcher.HeaderMatcher
		header        map[string]string
		expectedError string
	}{
		{
			scenario: "nil",
		},
		{
			scenario: "empty",
			matcher:  grpcMatcher.HeaderMatcher{},
		},
		{
			scenario: "match error",
			matcher: grpcMatcher.HeaderMatcher{
				"Authorization": grpcMatcher.Fn("", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				}),
			},
			expectedError: `could not match header: match error`,
		},
		{
			scenario: "mismatched",
			matcher: grpcMatcher.HeaderMatcher{
				"Authorization": matcher.Match("Bearer token"),
			},
			header: map[string]string{
				"Authorization": "Bearer foobar",
			},
			expectedError: `header "Authorization" with value "Bearer token" expected, "Bearer foobar" received`,
		},
		{
			scenario: "matched",
			matcher: grpcMatcher.HeaderMatcher{
				"Authorization": matcher.Match(regexp.MustCompile(`Bearer .*`)),
			},
			header: map[string]string{
				"Authorization": "Bearer foobar",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(tc.header))
			err := tc.matcher.Match(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
