# Mock a gRPC server

## Table of Contents

- [Usage](#usage)
    - [Create a new gRPC Server](#create-a-new-grpc-server)
    - [Testing with mocked gRPC Server](#testing-with-mocked-grpc-server)
    - [Register a service](#register-a-service)
        - [Register a Golang service](#register-a-golang-service)
            - [With `RegisterService(registerFunction)`](#with-registerserviceregisterfunction)
            - [With `RegisterServiceFromInstance(id string, instance interface{})`](#with-registerservicefrominstanceid-string-instance-interface)
        - [Register a random service](#register-a-random-service)
- [Match a value](#match-a-value)
    - [Exact](#exact)
    - [Regexp](#regexp)
    - [JSON](#json)
    - [Custom Matcher](#custom-matcher)
- [Mock a Unary Method](#mock-a-unary-method)
    - [Expect Header](#expect-header)
    - [Expect Payload](#expect-payload)
    - [Return](#return)
        - [Return an error](#return-an-error)
        - [Return a payload](#return-a-payload)
        - [Return with a custom handler](#return-with-a-custom-handler)
- [Mock a Client-Stream Method](#mock-a-client-stream-method)
    - [Expect Header](#expect-header-1)
    - [Expect Payload](#expect-payload-1)
    - [Return](#return-1)
        - [Return an error](#return-an-error-1)
        - [Return a payload](#return-a-payload-1)
        - [Return with a custom handler](#return-with-a-custom-handler-1)
- [Mock a Server-Stream Method](#mock-a-server-stream-method)
    - [Expect Header](#expect-header-2)
    - [Expect Payload](#expect-payload-2)
    - [Return](#return-2)
        - [Return an error](#return-an-error-2)
        - [Return a payload](#return-a-payload-2)
        - [Return with custom stream behaviors](#return-with-custom-stream-behaviors)
        - [Return with a custom handler](#return-with-a-custom-handler-2)
- [Mock a Bidirectional-Stream Method](#mock-a-bidirectional-stream-method)
    - [Expect Header](#expect-header)
    - [Return](#return-3)
        - [Return an error](#return-an-error-3)
        - [Return with a custom handler](#return-with-a-custom-handler-3)
- [Execution Plan](#execution-plan)
    - [First Match](#first-match)
- [Examples](#examples)

## Usage

### Create a new gRPC Server

Use the constructor `NewServer(opts ...ServerOption)` to create a new gRPC server, and you need to [register your service](#register-a-service) before mocking
it. After mocking the server, you need to start it, for example:

```go
package main

import (
	"context"
	"net"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s := grpcmock.NewServer(grpcmock.RegisterService(RegisterItemServiceServer))

	// Mock the server.
	s.ExpectUnary("grpctest.Service/GetItem")

	// Start the server.
	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	go func() {
		defer l.Close() // nolint: errcheck

		_ = s.Serve(l) // nolint: errcheck
	}()

	// Close the server on exit.
	defer s.Close(context.Background()) // nolint: errcheck

	// Call the server and assertions.
}
```

At the end, you could use `Server.ExpectationsWereMet() error` to check if all the expectations were met during the execution. It's always good to find out if
there are missing requests or the number of executions does not match your expectation.

Further reading:

- [Testing with mocked gRPC Server](#testing-with-mocked-grpc-server)
- [Register a service](#register-a-service)
- [Match a value](#match-a-value)
- [Mock a Unary Method](#mock-a-unary-method)
- [Mock a Client-Stream Method](#mock-a-client-stream-method)
- [Mock a Server-Stream Method](#mock-a-server-stream-method)
- [Mock a Bidirectional-Stream Method](#mock-a-bidirectional-stream-method)
- [Execution Plan](#execution-plan)

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Testing with mocked gRPC Server

After [creating a new gRPC Server](#create-a-new-grpc-server), you see it's a bit lengthy, and need to do several steps to ensure that the execution matches
your expectation. Furthermore, you may want to stop the test right away when the server receives an unexpected request. You can do that by specifying the test
that you're running with `Server.WithTest(t *testing.T)`. However, it still does not simplify your setup.

Therefore, for testing with a mocked gRPC server, you could use the `MockServer()` constructor, it does all the jobs for you except the start and stop a running
server. You need to do that yourself, or just use the `MockAndStartServer()` which starts a new server
with [`bufconn`](https://pkg.go.dev/google.golang.org/grpc/test/bufconn). This is the recommended way to test your application with a mocked gRPC server.

For example:

```go
package main

import (
	"context"
	"testing"

	"github.com/nhatthm/grpcmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockItemServiceServer(m ...grpcmock.ServerOption) grpcmock.ServerMockerWithContextDialer {
	opts := []grpcmock.ServerOption{grpcmock.RegisterService(RegisterItemServiceServer)}
	opts = append(opts, m...)

	return grpcmock.MockAndStartServer(opts...)
}

func TestServer(t *testing.T) {
	t.Parallel()

	const getItem = "grpctest.ItemService/GetItem"

	testCases := []struct {
		scenario   string
		mockServer grpcmock.ServerMockerWithContextDialer
		request    GetItemRequest
		expected   Item
	}{
		{
			scenario: "success",
			mockServer: mockItemServiceServer(func(s *grpcmock.Server) {
				s.ExpectUnary(getItem).
					WithPayload(&GetItemRequest{Id: 1}).
					Return(&Item{Id: 1, Name: "Item #1"})
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			_, dialer := tc.mockServer(t)

			// Use the dialer in your client, do the request and assertions.
			// For example:
			out := &Item{}
			err := grpcmock.InvokeUnary(context.Background(),
				getItem, &GetItemRequest{Id: 1}, out,
				grpcmock.WithInsecure(),
				grpcmock.WithContextDialer(dialer),
			)

			require.NoError(t, err)

			assert.Equal(t, "Item #1", out.Name)

			// Server is closed at the end, and the ExpectationsWereMet() is also called, automatically!
		})
	}
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Register a service

#### Register a Golang service

For example, there is an `ItemService`

```protobuf
service ItemService {
    rpc GetItem(GetItemRequest) returns (Item);
    rpc ListItems(ListItemsRequest) returns (stream Item);
    rpc CreateItems(stream Item) returns (CreateItemsResponse);
    rpc TransformItems(stream Item) returns (stream Item);
}
```

And a Golang implementation is generated with that `protobuf` definition, there are 2 ways to register this service to a gRPC server.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

##### With `RegisterService(registerFunction)`

In the generated code, you can find something like this:

```go
package grpctest

import "google.golang.org/grpc"

func RegisterItemServiceServer(s grpc.ServiceRegistrar, srv ItemServiceServer) {
	s.RegisterService(&ItemService_ServiceDesc, srv)
}
```

You can use this register function for the gRPC server, for example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			// Mock your server here.
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

##### With `RegisterServiceFromInstance(id string, instance interface{})`

In the generated code, you can find something like this:

```go
package grpctest

// ItemServiceServer is the server API for ItemService service.
// All implementations must embed UnimplementedItemServiceServer
// for forward compatibility
type ItemServiceServer interface {
	GetItem(context.Context, *GetItemRequest) (*Item, error)
	ListItems(*ListItemsRequest, ItemService_ListItemsServer) error
	CreateItems(ItemService_CreateItemsServer) error
	TransformItems(ItemService_TransformItemsServer) error
	mustEmbedUnimplementedItemServiceServer()
}
```

For registration, it's like this:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterServiceFromInstance("grpctest.ItemService", (*ItemServiceServer)(nil)),
		func(s *grpcmock.Server) {
			// Mock your server here.
		},
	)(t)

	// Your request and assertions.
}
```

All the service methods are discovered by using `reflect`. The service `id` is important because you need to put it in your mocks, for example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterServiceFromInstance("grpctest.ItemService", (*ItemServiceServer)(nil)),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.Service/GetItem")
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Register a random service

You can mock a gRPC server for a specific service without the generated Golang code, however it's quite lengthy. The method
is `RegisterServiceFromMethods(serviceMethods ...service.Method)`

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
	"github.com/nhatthm/grpcmock/service"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterServiceFromMethods(service.Method{
			// Provide a service definition with request and response type.
			ServiceName: "grpctest.ItemService",
			MethodName:  "GetItem",
			MethodType:  service.TypeUnary,
			Input:       &GetItemRequest{},
			Output:      &Item{},
		}),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.Service/GetItem")
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Match a value

`grpcmock` is using [`nhatthm/go-matcher`](https://github.com/nhatthm/go-matcher) for matching values and that makes `grpcmock` more powerful and convenient
than ever. When writing expectations for the header or the payload, you can use any kind of matchers for your needs.

For example, the `UnaryRequest.WithHeader(header string, value interface{})` means you expect a header that matches a value, you can put any of these into
the `value`

| Type | Explanation | Example |
| :---: | :--- | :--- |
| `string`<br/>`[]byte` | Match the exact string, case-sensitive | `.WithHeader("locale", "en-US")` |
| `*regexp.Regexp` | Match using `regexp.Regex.MatchString` | `.WithHeader("locale", regexp.MustCompile("^en-"))` |
| `matcher.RegexPattern` |  Match using `regexp.Regex.MatchString` | `.WithHeader("locale", matcher.RegexPattern("^en-"))` |

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Exact

`matcher.Exact` matches a value by using [`testify/assert.ObjectsAreEqual()`](https://github.com/stretchr/testify/assert).

| Matcher | Actual | Result |
| :---: | :---: | :---: |
| `matcher.Exact("en-US")` | `"en-US"` | `true` |
| `matcher.Exact("en-US")` | `"en-us"` | `false` |
| `matcher.Exact([]byte("en-US))` | `[]byte("en-US")` | `true` |
| `matcher.Exact([]byte("en-US))` | `"en-US"` | `false` |

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Regexp

`matcher.Regex` and `matcher.RegexPattern` matches a value by using [`Regexp.MatchString`](https://pkg.go.dev/regexp#Regexp.MatchString). `matcher.Regex`
expects a `*regexp.Regexp` while `matcher.RegexPattern` expects only a regexp pattern. However, in the end, they are the same because we create a
new `*regexp.Regexp` from the pattern using `regexp.MustCompile(pattern)`.

Notice, if the given value is not a `string` or `[]byte`, the matcher always fails.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## JSON

`matcher.JSON` matches a value by using [`swaggest/assertjson.FailNotEqual`](https://github.com/swaggest/assertjson). The matcher will marshal the input if it
is not a `string` or a `[]byte`, and then check against the expectation. For example, the expectation is ``matcher.JSON(`{"message": "hello"}`)``

These inputs match that expectation:

- `{"message":"hello"}` (notice there is no space after the `:` and it still matches)
- ``[]byte(`{"message":"hello"}`)``
- `map[string]string{"message": "hello"}`
- Or any objects that produce the same JSON object after calling `json.Marshal()`

You could also ignore some fields that you don't want to match. For example, the expectation is ``matcher.JSON(`{"name": "John Doe"}`)``. If you match it
with `{"name": "John Doe", "message": "hello"}`, that will fail because the `message` is unexpected. Therefore,
use ``matcher.JSON(`{"name": "John Doe", "message": "<ignore-diff>"}`)``

The `"<ignore-diff>"` can be used against any data types, not just the `string`. For example, `{"id": "<ignore-diff>"}` and `{"id": 42}` is a match.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Custom Matcher

You can use your own matcher as long as it implements the [`matcher.Matcher`](https://github.com/nhatthm/go-matcher/blob/master/matcher.go#L13-L17) interface.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Mock a Unary Method

### Expect Header

There are 2 methods for matching the headers:

`UnaryRequest.WithHeader(header string, value interface{})`

It checks whether a header matches the given `value`. The `value` could be `string`, `[]byte`, or a [`matcher.Matcher`](#match-a-value). If the `value` is
a `string` or a `[]byte`, the header is checked by using the [`matcher.Exact`](#exact).

For example:

```go
package main

import (
	"regexp"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.Service/GetItem").
				WithHeader("locale", regexp.MustCompile(`-US$`)).
				WithHeader("country", "US")
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

`UnaryRequest.WithHeaders(headers map[string]interface{})`

Similar to `WithHeader()`, this method checks for multiple headers. For example:

```go
package main

import (
	"regexp"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.Service/GetItem").
				WithHeaders(map[string]interface{}{
					"locale":  regexp.MustCompile(`-US$`),
					"country": "US",
				})
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Expect Payload

There are 2 methods for matching the request payload:

| Method | Explanation |
| :--- | :--- |
| `WithPayload(in interface{})` | Match the incoming payload with an expectation. See the table below for the supported types. |
| `WithPayloadf(format string, args ...interface{})` | An old school `fmt.Sprintf()` call will be made with `format` and `args`. The result will be passed to `WithPayload()` |

| `in` Type | Matcher | Explanation |
| :--- | :--- | :--- |
| `string`, `[]byte` | [`matcher.JSON`](#json) | Match the payload with a json string.
| `*regexp.Regexp` | [`matcher.Regex`](#regexp) | Match the payload using Regular Expressions. |
| [`matcher.Matcher`](#match-a-value) | The same matcher | Match the payload using the provided matcher. |
| `func(interface{}) (bool, error)` | The same matcher | Match the payload using a custom matcher. |
| Others | [`matcher.JSON`](#json) | `in` is marshaled to `string` and matched using `matcher.JSON`. |

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.Service/GetItem").
				WithPayload(`{"id": 41}`)

			s.ExpectUnary("grpctest.Service/GetItem").
				WithPayload(&GetItemRequest{Id: 41})

			s.ExpectUnary("grpctest.Service/GetItem").
				WithPayload(func(actual interface{}) (bool, error) {
					in, ok := actual.(*Item)
					if !ok {
						return false, nil
					}

					return in.Id == 42, nil
				})

			s.ExpectUnary("grpctest.Service/GetItem").
				WithPayload(matcher.RegexPattern(`"id":\d+`))
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Return

By default, if you don't specify anything, the mocked gRPC server will return a `codes.Unimplemented` error. You can return an error, a payload or write a
custom handler to feed the test scenario.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return an error

There are 4 methods, they are straightforward:

| Method | Explanation |
| :--- | :--- |
| `ReturnCode(code codes.Code)` | Change status code. If it is `codes.OK`, the error message is removed. |
| `ReturnErrorMessage(msg string)` | Change error message. Tf the current status code is `codes.OK`, it's changed to `codes.Internal` |
| `ReturnError(code codes.Code, msg string)` | Change status code and error message. If the code is `codes.OK`, the error message is removed. |
| `ReturnErrorf(code codes.Code, format string, args ...interface{})` | Same as `ReturnError` but with the support of `fmt.Sprintf() |

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/codes"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.Service/GetItem").
				ReturnError(codes.Internal, `server went away`)
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return a payload

There are 4 methods:

| Method | Explanation |
| :--- | :--- |
| `Return(v interface{})` | The response is a `string`, a `[]byte` or an object of the same type of the method. If it's a `string` or `[]byte`, the response will be unmarshalled to the object. |
| `Returnf(format string, args ...interface{})` | Same as `Return()`, but with support for formatting using `fmt.Sprintf()` |
| `ReturnFile(filePath string)` | The response is the content of given file, read by `io.ReadFile()` |
| `ReturnJSON(v interface{})` | The input is marshalled by `json.Marshal(v)` and then unmarshalled to an object of the same type of the method. |

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			// string,[]byte --json.Unmarshal()--> &Item{}
			s.ExpectUnary("grpctest.Service/GetItem").
				Return(`{"id": 41}`)

			// string --json.Unmarshal()--> &Item{}
			s.ExpectUnary("grpctest.Service/GetItem").
				Returnf(`{"id": %d}`, 41)

			s.ExpectUnary("grpctest.Service/GetItem").
				Return(&Item{Id: 41})

			// filePath --io.ReadFile()--> []byte --json.Unmarshal()--> &Item{}
			s.ExpectUnary("grpctest.Service/GetItem").
				ReturnFile("resources/fixtures/item41.json")

			// map[string]interface{} --json.Marshal()--> []byte --json.Unmarshal()--> &Item{}
			s.ExpectUnary("grpctest.Service/GetItem").
				ReturnJSON(map[string]interface{}{"id": 41}) // `{"id": 41}`
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return with a custom handler

You can write your own logic for handling the request, for example:

```go
package main

import (
	"context"
	"testing"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/metadata"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.Service/GetItem").
				Run(func(ctx context.Context, in interface{}) (interface{}, error) {
					var locale string

					if md, ok := metadata.FromIncomingContext(ctx); ok {
						if values := md.Get("locale"); len(values) > 0 {
							locale = values[0]
						}
					}

					req := in.(*GetItemRequest)

					return &Item{
						ID:     req.ID,
						Locale: locale,
						Name:   fmt.Sprintf("Item #%d", req.ID),
					}, nil
				})
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Mock a Client-Stream Method

### Expect Header

There are 2 methods for matching the headers:

`ClientStreamRequest.WithHeader(header string, value interface{})`

It checks whether a header matches the given `value`. The `value` could be `string`, `[]byte`, or a [`matcher.Matcher`](#match-a-value). If the `value` is
a `string` or a `[]byte`, the header is checked by using the [`matcher.Exact`](#exact).

For example:

```go
package main

import (
	"regexp"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream("grpctest.Service/CreateItems").
				WithHeader("locale", regexp.MustCompile(`-US$`)).
				WithHeader("country", "US")
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

`ClientStreamRequest.WithHeaders(headers map[string]interface{})`

Similar to `WithHeader()`, this method checks for multiple headers. For example:

```go
package main

import (
	"regexp"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream("grpctest.Service/CreateItems").
				WithHeaders(map[string]interface{}{
					"locale":  regexp.MustCompile(`-US$`),
					"country": "US",
				})
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Expect Payload

There are 2 methods for matching the request payload:

| Method | Explanation |
| :--- | :--- |
|`WithPayload(in interface{})` | Match the incoming payload with an expectation. See the table below for the supported types. |
|`WithPayloadf(format string, args ...interface{})` | An old school `fmt.Sprintf()` call will be made with `format` and `args`. The result will be passed to `WithPayload()` |

_* The incoming `payload` is tee from the stream until `io.EOF`._

| `in` Type | Matcher | Explanation |
| :--- | :--- | :--- |
| `string`, `[]byte` | [`matcher.JSON`](#json) | Match the payload with a json string.
| `*regexp.Regexp` | [`matcher.Regex`](#regexp) | Match the payload using Regular Expressions. |
| [`matcher.Matcher`](#match-a-value) | The same matcher | Match the payload using the provided matcher. |
| `func(in interface{}) (bool, error)` | The same matcher | Match the payload using a custom matcher. |
| Others | [`matcher.JSON`](#json) | `in` is marshaled to `string` and matched with the payload using `matcher.JSON`. |

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream("grpctest.Service/CreateItems").
				WithPayload(`[{"id": 41}]`)

			s.ExpectClientStream("grpctest.Service/CreateItems").
				WithPayload([]*Item{{Id: 41}})

			s.ExpectClientStream("grpctest.Service/CreateItems").
				WithPayload(func(in interface{}) (bool, error) {
					items, ok := in.([]*Item)
					if !ok {
						return false, nil
					}

					return len(items) == 1, nil
				})

			s.ExpectClientStream("grpctest.Service/CreateItems").
				WithPayload(matcher.RegexPattern(`"id":\d+`))
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Return

By default, if you don't specify anything, the mocked gRPC server will return a `codes.Unimplemented` error. You can return an error, a payload or write a
custom handler to feed the test scenario.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return an error

There are 4 methods, they are straightforward:

| Method | Explanation |
| :--- | :--- |
| `ReturnCode(code codes.Code)` | Change status code. If it is `codes.OK`, the error message is removed. |
| `ReturnErrorMessage(msg string)` | Change error message. Tf the current status code is `codes.OK`, it's changed to `codes.Internal` |
| `ReturnError(code codes.Code, msg string)` | Change status code and error message. If the code is `codes.OK`, the error message is removed. |
| `ReturnErrorf(code codes.Code, format string, args ...interface{})` | Same as `ReturnError` but with the support of `fmt.Sprintf() |

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/codes"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream("grpctest.Service/CreateItems").
				ReturnError(codes.Internal, `server went away`)
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return a payload

There are 4 methods:

| Method | Explanation |
| :--- | :--- |
| `Return(v interface{})` | The response is a `string`, a `[]byte` or an object of the same type of the method. If it's a `string` or `[]byte`, the response will be unmarshalled to the object. |
| `Returnf(format string, args ...interface{})` | Same as `Return()`, but with support for formatting using `fmt.Sprintf()` |
| `ReturnFile(filePath string)` | The response is the content of given file, read by `io.ReadFile()` |
| `ReturnJSON(v interface{})` | The input is marshalled by `json.Marshal(v)` and then unmarshalled to an object of the same type of the method. |

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			// string,[]byte --json.Unmarshal()--> &CreateItemsResponse{}
			s.ExpectClientStream("grpctest.Service/CreateItems").
				Return(`{"num_items": 5}`)

			// string --json.Unmarshal()--> &CreateItemsResponse{}
			s.ExpectClientStream("grpctest.Service/CreateItems").
				Returnf(`{"num_items": %d}`, 5)

			s.ExpectClientStream("grpctest.Service/CreateItems").
				Return(&CreateItemsResponse{NumItems: 5})

			// filePath --io.ReadFile()--> []byte --json.Unmarshal()--> &CreateItemsResponse{}
			s.ExpectClientStream("grpctest.Service/CreateItems").
				ReturnFile("resources/fixtures/create_items_response.json")

			// map[string]interface{} --json.Marshal()--> []byte --json.Unmarshal()--> &CreateItemsResponse{}
			s.ExpectClientStream("grpctest.Service/CreateItems").
				ReturnJSON(map[string]interface{}{"num_items": 41}) // `{"num_items": 5}`
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return with a custom handler

You can write your own logic for handling the request, for example:

```go
package main

import (
	"context"
	"testing"

	"github.com/nhatthm/grpcmock"
	"github.com/nhatthm/grpcmock/stream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream("grpctest.Service/CreateItems").
				WithPayload(grpcmock.MatchClientStreamMsgCount(3)).
				Run(func(_ context.Context, s grpc.ServerStream) (interface{}, error) {
					out := make([]*Item, 0)

					if err := stream.RecvAll(s, &out); err != nil {
						return nil, err
					}

					cnt := int64(0)

					for _, msg := range out {
						if msg.Id > 40 {
							cnt++
						}
					}

					return &CreateItemsResponse{NumItems: cnt}, nil
				})
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Mock a Server-Stream Method

### Expect Header

There are 2 methods for matching the headers:

`ServerStreamRequest.WithHeader(header string, value interface{})`

It checks whether a header matches the given `value`. The `value` could be `string`, `[]byte`, or a [`matcher.Matcher`](#match-a-value). If the `value` is
a `string` or a `[]byte`, the header is checked by using the [`matcher.Exact`](#exact).

For example:

```go
package main

import (
	"regexp"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.Service/ListItems").
				WithHeader("locale", regexp.MustCompile(`-US$`)).
				WithHeader("country", "US")
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

`ServerStreamRequest.WithHeaders(headers map[string]interface{})`

Similar to `WithHeader()`, this method checks for multiple headers. For example:

```go
package main

import (
	"regexp"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.Service/ListItems").
				WithHeaders(map[string]interface{}{
					"locale":  regexp.MustCompile(`-US$`),
					"country": "US",
				})
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Expect Payload

There are 2 methods for matching the request payload:

| Method | Explanation |
| :--- | :--- |
| `WithPayload(in interface{})` | Match the incoming payload with an expectation. See the table below for the supported types. |
| `WithPayloadf(format string, args ...interface{})` | An old school `fmt.Sprintf()` call will be made with `format` and `args`. The result
will be passed to `WithPayload()` |

| `in` Type | Matcher | Explanation |
| :--- | :--- | :--- |
| `string`, `[]byte` | [`matcher.JSON`](#json) | Match the payload with a json string.
| `*regexp.Regexp` | [`matcher.Regex`](#regexp) | Match the payload using Regular Expressions. |
| [`matcher.Matcher`](#match-a-value) | The same matcher | Match the payload using the provided matcher. |
| `func(interface{}) (bool, error)` | The same matcher | Match the payload using a custom matcher. |
| Others | [`matcher.JSON`](#json) | `in` is marshaled to `string` and matched using `matcher.JSON`. |

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/go-matcher"
	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.Service/ListItems").
				WithPayload(`{"id": 41}`)

			s.ExpectServerStream("grpctest.Service/ListItems").
				WithPayload(&ListItemRequest{})

			s.ExpectServerStream("grpctest.Service/ListItems").
				WithPayload(func(actual interface{}) (bool, error) {
					if _, ok := actual.(*ListItemRequest); !ok {
						return false, nil
					}

					return true, nil
				})

			s.ExpectServerStream("grpctest.Service/ListItems").
				WithPayload(matcher.RegexPattern(`{.*}`))
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Return

By default, if you don't specify anything, the mocked gRPC server will return a `codes.Unimplemented` error. You can return an error, a payload or write a
custom handler to feed the test scenario.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return an error

There are 4 methods, they are straightforward:

| Method | Explanation |
| :--- | :--- |
| `ReturnCode(code codes.Code)` | Change status code. If it is `codes.OK`, the error message is removed. |
| `ReturnErrorMessage(msg string)` | Change error message. Tf the current status code is `codes.OK`, it's changed to `codes.Internal` |
| `ReturnError(code codes.Code, msg string)` | Change status code and error message. If the code is `codes.OK`, the error message is removed. |
| `ReturnErrorf(code codes.Code, format string, args ...interface{})` | Same as `ReturnError` but with the support of `fmt.Sprintf() |

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/codes"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.Service/ListItems").
				ReturnError(codes.Internal, `server went away`)
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return a payload

There are 4 methods:

| Method | Explanation
| :--- | :--- |
| `Return(v interface{})` | The response is a `string`, a `[]byte` or a slice of objects of the same type of the method. If it's a `string` or `[]byte`, the response will be unmarshalled to a slice. |
| `Returnf(format string, args ...interface{})` | Same as `Return()`, but with support for formatting using `fmt.Sprintf()` |
| `ReturnFile(filePath string)` | The response is the content of given file, read by `io.ReadFile()` |
| `ReturnJSON(v interface{})` | The input is marshalled by `json.Marshal(v)` and then unmarshalled to a slice of objects of the same type of the method. |

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			// string,[]byte --json.Unmarshal()--> []*Item{}
			s.ExpectUnary("grpctest.Service/ListItems").
				Return(`[{"id": 41}]`)

			// string --json.Unmarshal()--> []*Item{}
			s.ExpectUnary("grpctest.Service/ListItems").
				Returnf(`[{"id": %d}]`, 41)

			s.ExpectUnary("grpctest.Service/ListItems").
				Return([]*Item{{Id: 41}})

			// filePath --io.ReadFile()--> []byte --json.Unmarshal()--> []*Item{}
			s.ExpectUnary("grpctest.Service/ListItems").
				ReturnFile("resources/fixtures/items.json")

			// []map[string]interface{} --json.Marshal()--> []byte --json.Unmarshal()--> []*Item{}
			s.ExpectUnary("grpctest.Service/ListItems").
				ReturnJSON([]map[string]interface{}{{"id": 41}}) // [{"id": 41}]
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return with custom stream behaviors

With `ServerStreamRequest.ReturnStream()`, you can customize the behaviors of the stream. There are several step helpers:

| Step | Explanation
| :--- | :--- |
| `AddHeader(key, value string)`<br/>`SetHeader(header map[string]string)`| Set one or many header without sending to client. |
| `SendHeader()` | Send all set header to client. |
| `Send(v interface{})` | Send a single message to client. |
| `SendMany(v interface{})` | Send multiple messages to client. |
| `ReturnError(code codes.Code, msg string)`<br/>`ReturnErrorf(code codes.Code, msg string, args ...interface{})` | Return an error to client |

_* All the steps executes sequentially._

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/codes"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.Service/ListItems").
				ReturnStream().
				Send(&Item{Id: 41, Name: "Item #41"}). // Sent an item to client.
				SendMany([]*Item{
					// Sent multiple items to client.
					{Id: 42, Name: "Item #42"},
					{Id: 43, Name: "Item #43"},
				}).
				ReturnError(codes.Aborted, "server aborted the transaction") // Return an error to client.
		},
	)(t)

	// Your request and assertions.
}
```

#### Return with a custom handler

You can write your own logic for handling the request, for example:

```go
package main

import (
	"context"
	"testing"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.Service/ListItems").
				Run(func(_ context.Context, _ interface{}, s grpc.ServerStream) error {
					_ = s.SendMsg(&Item{Id: 41, Name: "Item #41"})
					_ = s.SendMsg(&Item{Id: 42, Name: "Item #42"})

					return nil
				})
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Mock a Bidirectional-Stream Method

### Expect Header

There are 2 methods for matching the headers:

`BidirectionalStreamRequest.WithHeader(header string, value interface{})`

It checks whether a header matches the given `value`. The `value` could be `string`, `[]byte`, or a [`matcher.Matcher`](#match-a-value). If the `value` is
a `string` or a `[]byte`, the header is checked by using the [`matcher.Exact`](#exact).

For example:

```go
package main

import (
	"regexp"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectBidirectionalStream("grpctest.Service/TransformItems").
				WithHeader("locale", regexp.MustCompile(`-US$`)).
				WithHeader("country", "US")
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

`BidirectionalStreamRequest.WithHeaders(headers map[string]interface{})`

Similar to `WithHeader()`, this method checks for multiple headers. For example:

```go
package main

import (
	"regexp"
	"testing"

	"github.com/nhatthm/grpcmock"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectBidirectionalStream("grpctest.Service/TransformItems").
				WithHeaders(map[string]interface{}{
					"locale":  regexp.MustCompile(`-US$`),
					"country": "US",
				})
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Return

By default, if you don't specify anything, the mocked gRPC server will return a `codes.Unimplemented` error. You can return an error, a payload or write a
custom handler to feed the test scenario.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return an error

There are 4 methods, they are straightforward:

| Method | Explanation |
| :--- | :--- |
| `ReturnCode(code codes.Code)` | Change status code. If it is `codes.OK`, the error message is removed. |
| `ReturnErrorMessage(msg string)` | Change error message. Tf the current status code is `codes.OK`, it's changed to `codes.Internal` |
| `ReturnError(code codes.Code, msg string)` | Change status code and error message. If the code is `codes.OK`, the error message is removed. |
| `ReturnErrorf(code codes.Code, format string, args ...interface{})` | Same as `ReturnError` but with the support of `fmt.Sprintf() |

For example:

```go
package main

import (
	"testing"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/codes"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectBidirectionalStream("grpctest.Service/TransformItems").
				ReturnError(codes.Internal, `server went away`)
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Return with a custom handler

You can write your own logic for handling the request, for example:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestServer(t *testing.T) {
	s, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectBidirectionalStream("grpctest.Service/TransformItems").
				Run(func(ctx context.Context, s grpc.ServerStream) error {
					for {
						item := &Item{}
						err := s.RecvMsg(item)

						if errors.Is(err, io.EOF) {
							return nil
						}

						if err != nil {
							return err
						}

						item.Name = fmt.Sprintf("Modified #%d", item.Id)

						if err := s.SendMsg(item); err != nil {
							return err
						}
					}
				})
		},
	)(t)

	// Your request and assertions.
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Execution Plan

The mocked gRPC server is created with the `github.com/nhatthm/grpcmock/planner.Sequence()` by default, and it matches incoming requests sequentially. You can
easily change this behavior to match your application execution by implementing the `planner.Planner` interface.

```go
package planner

import (
	"context"

	"github.com/nhatthm/grpcmock/request"
	"github.com/nhatthm/grpcmock/service"
)

type Planner interface {
	// IsEmpty checks whether the planner has no expectation.
	IsEmpty() bool
	// Expect adds a new expectation.
	Expect(expect request.Request)
	// Plan decides how a request matches an expectation.
	Plan(ctx context.Context, req service.Method, in interface{}) (request.Request, error)
	// Remain returns remain expectations.
	Remain() []request.Request
	// Reset removes all the expectations.
	Reset()
}
```

Then use it with `Server.WithPlanner(newPlanner)` (see the [`ExampleServer_WithPlanner`](server_example_test.go#L24))

When the `Server.Expect[METHOD]()` is called, the mocked server will prepare a request and sends it to the planner. If there is an incoming request, the server
will call `Planner.PLan()` to find the expectation that matches the request and executes it.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### First Match

`planner.FirstMatch` creates a new `planner.Planner` that finds the first expectation that matches the incoming request.

For example, there are 3 expectations in order:

```
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 40})
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
    Return(`{"id": 41, "name": "Item #41 - 1"}`)
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
    Return(`{"id": 41, "name": "Item #41 - 2"}`)
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 42})
```

When the server receives a request with payload `{"id": 41}`, the `planner.FirstMatch` looks up and finds the second expectation which is the first
expectation that matches all the criteria. After that, there are only 3 expectations left:

```
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 40})
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
   	Return(`{"id": 41, "name": "Item #41 - 2"}`)
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 42})
```

When the server receives another request with payload `{"id": 40}`, the `planner.FirstMatch` does the same thing and there are only 2 expectations left:

```
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
   	Return(`{"id": 41, "name": "Item #41 - 2"}`)
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 42})
```

When the server receives another request with payload `{"id": 100}`, the `planner.FirstMatch` can not match it with any expectations and the server returns
a `FailedPrecondition` result with error message `unexpected request received`.

Due to the nature of the matcher, pay extra attention when you use repeatability. For example, given these expectations:

```
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
   	UnlimitedTimes().
   	Return(`{"id": 41, "name": "Item #41 - 1"}`)
Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
   	Return(`{"id": 41, "name": "Item #41 - 2"}`)
```

The 2nd expectation is never taken in account because with the same criteria, the planner always picks the first match, which is the first expectation.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Examples

See:

- [mock_example_test.go](mock_example_test.go)
- [server_example_test.go](server_example_test.go)

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)
