# gRPC Test Utilities for Golang

[![GitHub Releases](https://img.shields.io/github/v/release/nhatthm/grpcmock)](https://github.com/nhatthm/grpcmock/releases/latest)
[![Build Status](https://github.com/nhatthm/grpcmock/actions/workflows/test.yaml/badge.svg)](https://github.com/nhatthm/grpcmock/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/nhatthm/grpcmock/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/nhatthm/grpcmock)
[![Go Report Card](https://goreportcard.com/badge/github.com/nhatthm/grpcmock)](https://goreportcard.com/report/github.com/nhatthm/grpcmock)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/nhatthm/grpcmock)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

Test gRPC service and client like a pro.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Install](#install)
- [Usage](#usage)
    - [Mock a gRPC server](#mock-a-grpc-server)
        - [Unary Method](#unary-method)
        - [Client-Stream Method](#client-stream-method)
        - [Server-Stream Method](#server-stream-method)
        - [Bidirectional-Stream Method](#bidirectional-stream-method)
    - [Invoke a gRPC method](#invoke-a-grpc-method)
        - [Unary Method](#unary-method-1)
        - [Client-Stream Method](#client-stream-method-1)
        - [Server-Stream Method](#server-stream-method-1)
        - [Bidirectional-Stream Method](#bidirectional-stream-method-1)

## Prerequisites

- `Go >= 1.17`

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Install

```bash
go get github.com/nhatthm/grpcmock
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Usage

### Mock a gRPC server

Read more about [mocking a gRPC server](SERVER.md)

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Unary Method

Read more about [mocking a Unary Method](SERVER.md#mock-a-unary-method)

```go
package main

import (
	"context"
	"testing"
	"time"

	"github.com/nhatthm/grpcmock"
	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/stretchr/testify/assert"
)

func TestGetItems(t *testing.T) {
	t.Parallel()

	expected := &Item{Id: 42, Name: "Item 42"}

	_, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectUnary("myservice/GetItem").
				WithHeader("locale", "en-US").
				WithPayload(&GetItemRequest{Id: 42}).
				Return(expected)
		},
	)(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := &Item{}

	err := grpcmock.InvokeUnary(ctx,
		"myservice/GetItem",
		&GetItemRequest{Id: 42}, out,
		grpcmock.WithHeader("locale", "en-US"),
		grpcmock.WithContextDialer(d),
		grpcmock.WithInsecure(),
	)

	grpcAssert.EqualMessage(t, expected, out)
	assert.NoError(t, err)
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Client-Stream Method

Read more about [mocking a Client-Stream Method](SERVER.md#mock-a-client-stream-method)

```go
package main

import (
	"context"
	"testing"
	"time"

	"github.com/nhatthm/grpcmock"
	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/stretchr/testify/assert"
)

func TestCreateItems(t *testing.T) {
	t.Parallel()

	expected := &CreateItemsResponse{NumItems: 1}

	_, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream("myservice/CreateItems").
				WithPayload([]*Item{{Id: 42}}).
				Return(expected)
		},
	)(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := &CreateItemsResponse{}
	err := grpcmock.InvokeClientStream(ctx, "myservice/CreateItems",
		grpcmock.SendAll([]*Item{{Id: 42}}), out,
		grpcmock.WithContextDialer(d),
		grpcmock.WithInsecure(),
	)

	grpcAssert.EqualMessage(t, expected, out)
	assert.NoError(t, err)
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Server-Stream Method

Read more about [mocking a Server-Stream Method](SERVER.md#mock-a-server-stream-method)

```go
package main

import (
	"context"
	"testing"
	"time"

	"github.com/nhatthm/grpcmock"
	"github.com/stretchr/testify/assert"
)

func TestListItems(t *testing.T) {
	t.Parallel()

	expected := []*Item{
		{Id: 41, Name: "Item 41"},
		{Id: 42, Name: "Item 42"},
	}

	_, d := grpcmock.MockAndStartServer(
		grpcmock.RegisterService(RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("myservice/ListItems").
				WithPayload(&ListItemsRequest{}).
				Return(expected)
		},
	)(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	actual := make([]*Item, 0)

	err := grpcmock.InvokeServerStream(ctx,
		"myservice/ListItems",
		&ListItemsRequest{},
		grpcmock.RecvAll(&out),
		grpcmock.WithContextDialer(d),
		grpcmock.WithInsecure(),
	)

	assert.NoError(t, err)
	assert.Len(t, actual, len(expected))

	for i := 0; i < len(expected); i++ {
		grpcAssert.EqualMessage(t, expected[i], actual[i])
	}
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Bidirectional-Stream Method

Coming soon by EOY 2021.

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Invoke a gRPC method

#### Unary Method

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/test/bufconn"
)

func getItem(l *bufconn.Listener, id int32) (*Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := &Item{}
	err := grpcmock.InvokeUnary(ctx, "myservice/GetItem",
		&GetItemRequest{Id: id}, out,
		grpcmock.WithHeader("Locale", "en-US"),
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	return out, err
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Client-Stream Method

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/test/bufconn"
)

func createItems(l *bufconn.Listener, items []*Item) (*CreateItemsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := &CreateItemsResponse{}
	err := grpcmock.InvokeClientStream(ctx, "myservice/CreateItems",
		grpcmock.SendAll(items), out,
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	return out, err
}
```

Or with a custom handler

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func createItems(l *bufconn.Listener, items []*Item) (*CreateItemsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := &CreateItemsResponse{}
	err := grpcmock.InvokeClientStream(ctx, "myservice/CreateItems",
		func(s grpc.ClientStream) error {
			// Handle the stream here.
			return nil
		},
		out,
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	return out, err
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Server-Stream Method

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/test/bufconn"
)

func listItems(l *bufconn.Listener) ([]*Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := make([]*Item, 0)
	err := grpcmock.InvokeServerStream(ctx, "myservice/ListItems",
		&ListItemsRequest{},
		grpcmock.RecvAll(&out),
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	return out, err
}
```

Or with a custom handler

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func listItems(l *bufconn.Listener) ([]*Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := make([]*Item, 0)
	err := grpcmock.InvokeServerStream(ctx, "myservice/ListItems",
		&ListItemsRequest{},
		func(s grpc.ClientStream) error {
			// Handle the stream here.
			return nil
		},
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	return out, err
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

#### Bidirectional-Stream Method

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/test/bufconn"
)

func transformItems(l *bufconn.Listener, in []*Item) ([]*Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := make([]*Item, 0)
	err := grpcmock.InvokeBidirectionalStream(ctx, "myservice/TransformItems",
		grpcmock.SendAndRecvAll(in, &out),
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	return out, err
}
```

Or with a custom handler

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func transformItems(l *bufconn.Listener, in []*Item) ([]*Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := make([]*Item, 0)
	err := grpcmock.InvokeBidirectionalStream(ctx, "myservice/TransformItems",
		func(s grpc.ClientStream) error {
			// Handle the stream here.
			return nil
		},
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	return out, err
}
```

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

## Donation

If this project help you reduce time to develop, you can give me a cup of coffee :)

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)

### Paypal donation

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;or scan this

<img src="https://user-images.githubusercontent.com/1154587/113494222-ad8cb200-94e6-11eb-9ef3-eb883ada222a.png" width="147px" />

[<sub><sup>[table of contents]</sup></sub>](#table-of-contents)
