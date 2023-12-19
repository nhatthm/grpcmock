package stream

var _ Receiver = (*teeReceiver)(nil)

type teeReceiver struct {
	r Receiver
	s Sender
}

func (r *teeReceiver) RecvMsg(m any) error {
	if err := r.r.RecvMsg(m); err != nil {
		return err
	}

	return r.s.SendMsg(m)
}

// TeeReceiver returns a Receiver that sends to s what it receives from r.
func TeeReceiver(r Receiver, s Sender) Receiver {
	return &teeReceiver{
		r: r,
		s: s,
	}
}
