# gRPC Test Utilities for Golang

[![GitHub Releases](https://img.shields.io/github/v/release/nhatthm/grpcmock)](https://github.com/nhatthm/grpcmock/releases/latest)
[![Build Status](https://github.com/nhatthm/grpcmock/actions/workflows/test.yaml/badge.svg)](https://github.com/nhatthm/grpcmock/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/nhatthm/grpcmock/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/nhatthm/grpcmock)
[![Go Report Card](https://goreportcard.com/badge/github.com/nhatthm/grpcmock)](https://goreportcard.com/report/github.com/nhatthm/grpcmock)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/nhatthm/grpcmock)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

Test gRPC service and client like a pro.

## Prerequisites

- `Go >= 1.16`

## Install

```bash
go get github.com/nhatthm/grpcmock
```

## Usage

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

func getItem(l *bufconn.Listener, id int32) (Item, error) {
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

### Client-Stream Method

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/test/bufconn"
)

func createItems(l *bufconn.Listener, items []*Item) (CreateItemsResponse, error) {
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

func createItems(l *bufconn.Listener, items []*Item) (CreateItemsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := &CreateItemsResponse{}
	err := grpcmock.InvokeClientStream(ctx, "myservice/CreateItems",
		func(stream grpc.ClientStream) error {
			// Handle the stream here.
		},
		out,
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	// return
}
```

### Server-Stream Method

```go
package main

import (
	"context"
	"time"

	"github.com/nhatthm/grpcmock"
	"google.golang.org/grpc/test/bufconn"
)

func listItems(l *bufconn.Listener) ([]Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	out := make([]Item, 0)
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

func listItems(l *bufconn.Listener) ([]Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	err := grpcmock.InvokeServerStream(ctx, "myservice/ListItems",
		&ListItemsRequest{},
		func(stream grpc.ClientStream) error {
			// Handle the stream here.
		},
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)

	// return
}
```

## Donation

If this project help you reduce time to develop, you can give me a cup of coffee :)

### Paypal donation

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;or scan this

<img src="https://user-images.githubusercontent.com/1154587/113494222-ad8cb200-94e6-11eb-9ef3-eb883ada222a.png" width="147px" />
