package value_test

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/test/grpctest"
	"go.nhat.io/grpcmock/value"
)

func TestString_Success(t *testing.T) {
	t.Parallel()

	const expected = `id:42`

	testCases := []struct {
		scenario string
		input    interface{}
		expected string
	}{
		{
			scenario: "string",
			input:    expected,
		},
		{
			scenario: "[]byte",
			input:    []byte(expected),
		},
		{
			scenario: "fmt.Stringer",
			input:    &grpctest.Item{Id: 42},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := value.String(tc.input)

			assert.Equal(t, expected, actual)
		})
	}
}

func TestString_Panic(t *testing.T) {
	t.Parallel()

	assert.PanicsWithError(t, `unsupported data type`, func() {
		value.String(42)
	})
}

func TestMarshal(t *testing.T) {
	t.Parallel()

	const payload = `[{"id":42}]`

	testCases := []struct {
		scenario       string
		in             interface{}
		expectedResult string
		expectedError  string
	}{
		{
			scenario:       "[]byte",
			in:             []byte(payload),
			expectedResult: payload,
		},
		{
			scenario:       "string",
			in:             payload,
			expectedResult: payload,
		},
		{
			scenario:      "chan",
			in:            make(chan struct{}, 1),
			expectedError: `json: unsupported type: chan struct {}`,
		},
		{
			scenario:       "object",
			in:             streamer.NewBidirectionalStreamer(nil, nil, nil),
			expectedResult: "",
		},
		{
			scenario:       "object",
			in:             []*grpctest.Item{{Id: 42}},
			expectedResult: payload,
		},
		{
			scenario: "client stream error",
			in: streamer.NewClientStreamer(xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Return(errors.New("recv error"))
			})(t), reflect.TypeOf(&grpctest.Item{}), reflect.TypeOf(&grpctest.CreateItemsResponse{})),
			expectedError: `recv error`,
		},
		{
			scenario: "client stream success",
			in: streamer.NewClientStreamer(xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Once().
					Run(func(args mock.Arguments) {
						msg := args.Get(0).(*grpctest.Item) // nolint: errcheck
						msg.Id = 42
					}).
					Return(nil)

				s.On("RecvMsg", &grpctest.Item{}).
					Return(io.EOF)
			})(t), reflect.TypeOf(&grpctest.Item{}), reflect.TypeOf(&grpctest.CreateItemsResponse{})),
			expectedResult: payload,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result, err := value.Marshal(tc.in)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
