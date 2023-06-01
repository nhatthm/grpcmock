package stream

// SendReceiver is an interface wrapper around grpc.ClientStream and grpc.ServerStream.
type SendReceiver interface {
	Sender
	Receiver
}

// SendAndRecvAll sends and receives messages until getting io.EOF.
func SendAndRecvAll(sr SendReceiver, in interface{}, out interface{}) error {
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		if err := RecvAll(sr, out); err != nil {
			errCh <- err

			return
		}
	}()

	if err := SendAll(sr, in); err != nil {
		return err
	}

	if err := CloseSend(sr); err != nil {
		return err
	}

	return <-errCh
}
