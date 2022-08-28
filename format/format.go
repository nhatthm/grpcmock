package format

import (
	"fmt"
	"io"
	"sort"

	"go.nhat.io/matcher/v2"

	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/must"
	xreflect "go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/value"
)

const indent = "    "

// ExpectedRequest formats an expected request.
func ExpectedRequest(w io.Writer, svc service.Method, header xmatcher.HeaderMatcher, payload matcher.Matcher) {
	ExpectedRequestTimes(w, svc, header, payload, 0, 0)
}

// ExpectedRequestTimes formats an expected request with total and remaining calls.
func ExpectedRequestTimes(w io.Writer, svc service.Method, header xmatcher.HeaderMatcher, payload matcher.Matcher, totalCalls, remainingCalls int) {
	expectedHeader := map[string]interface{}(nil)

	if header != nil {
		expectedHeader = make(map[string]interface{}, len(header))

		for k, v := range header {
			expectedHeader[k] = v
		}
	}

	formatRequestTimes(w, svc, expectedHeader, payload, totalCalls, remainingCalls)
}

// Request formats a request.
func Request(w io.Writer, svc service.Method, header map[string]string, payload interface{}) {
	interfaceHeader := map[string]interface{}(nil)
	if header != nil {
		interfaceHeader = make(map[string]interface{}, len(header))

		for k, v := range header {
			interfaceHeader[k] = v
		}
	}

	formatRequestTimes(w, svc, interfaceHeader, payload, 0, 0)
}

func formatRequestTimes(w io.Writer, svc service.Method, header map[string]interface{}, payload interface{}, totalCalls, remainingCalls int) {
	_, _ = fmt.Fprintf(w, "%s %s", svc.MethodType, svc.FullName())

	if remainingCalls > 0 && (totalCalls != 0 || remainingCalls != 1) {
		_, _ = fmt.Fprintf(w, " (called: %d time(s), remaining: %d time(s))", totalCalls, remainingCalls)
	}

	_, _ = fmt.Fprintln(w)

	if len(header) > 0 {
		_, _ = fmt.Fprintf(w, "%swith header:\n", indent)

		keys := make([]string, len(header))
		i := 0

		for k := range header {
			keys[i] = k
			i++
		}

		sort.Strings(keys)

		for _, key := range keys {
			_, _ = fmt.Fprintf(w, "%s%s%s: %s\n", indent, indent, key, formatValueInline(header[key]))
		}
	}

	if !xreflect.IsNil(payload) {
		bodyStr := formatValue(payload)

		if bodyStr != "" {
			_, _ = fmt.Fprintf(w, "%swith payload%s\n", indent, formatType(payload))
			_, _ = fmt.Fprintf(w, "%s%s%s\n", indent, indent, bodyStr)
		}
	}
}

func formatValueInline(v interface{}) string {
	if v == nil {
		return "<nil>"
	}

	switch m := v.(type) {
	case matcher.ExactMatcher,
		xmatcher.FnMatcher,
		[]byte,
		string:
		return formatValue(v)

	case matcher.Callback:
		return formatValueInline(m.Matcher())

	case *xmatcher.PayloadMatcher:
		if m == nil {
			return ""
		}

		return formatValueInline(m.Matcher())

	case matcher.Matcher:
		return fmt.Sprintf("%T(%q)", v, m.Expected())

	default:
		if xreflect.IsNil(v) {
			return ""
		}

		panic("unknown value type")
	}
}

func formatType(v interface{}) string {
	if xreflect.IsNil(v) {
		return ""
	}

	switch m := v.(type) {
	case matcher.ExactMatcher,
		xmatcher.FnMatcher,
		[]byte,
		string:
		return ""

	case matcher.Callback:
		return formatType(m.Matcher())

	case *xmatcher.PayloadMatcher:
		return formatType(m.Matcher())

	default:
		return fmt.Sprintf(" using %T", v)
	}
}

// nolint: cyclop
func formatValue(v interface{}) string {
	if v == nil {
		return "<nil>"
	}

	switch m := v.(type) {
	case []byte:
		return string(m)

	case string:
		return m

	case matcher.Callback:
		return formatValue(m.Matcher())

	case xmatcher.FnMatcher:
		if e := m.Expected(); e != "" {
			return e
		}

		return "matches custom expectation"

	case *xmatcher.PayloadMatcher:
		if m == nil {
			return ""
		}

		return formatValue(m.Matcher())

	case matcher.Matcher:
		return m.Expected()

	default:
		if xreflect.IsNil(v) {
			return ""
		}

		data, err := value.Marshal(v)
		must.NotFail(err)

		return data
	}
}
