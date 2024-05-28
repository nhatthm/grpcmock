package matcher

import (
	"context"
	"fmt"

	"go.nhat.io/matcher/v2"
	"google.golang.org/grpc/metadata"
)

// HeaderMatcher matches the header values.
type HeaderMatcher map[string]matcher.Matcher

// Match matches the header in context.
func (m HeaderMatcher) Match(ctx context.Context) error {
	if len(m) == 0 {
		return nil
	}

	md := incomingHeader(ctx)

	for h, m := range m {
		value := getHeader(md, h)

		matched, err := m.Match(value)
		if err != nil {
			return fmt.Errorf("could not match header: %w", err)
		}

		if !matched {
			return fmt.Errorf("header %q with value %q expected, %q received", h, m.Expected(), value) //nolint: goerr113
		}
	}

	return nil
}

func getHeader(md metadata.MD, k string) string {
	values := md.Get(k)

	if len(values) > 0 {
		return values[0]
	}

	return ""
}

func incomingHeader(ctx context.Context) metadata.MD {
	md, _ := metadata.FromIncomingContext(ctx)
	if md == nil {
		md = metadata.New(nil)
	}

	return md
}
