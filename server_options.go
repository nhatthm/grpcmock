//go:build !testcoverage

package grpcmock

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/tap"
)

// WriteBufferSize is a wrapper of google.golang.org/grpc.WriteBufferSize().
func WriteBufferSize(sz int) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.WriteBufferSize(sz))
	}
}

// ReadBufferSize is a wrapper of google.golang.org/grpc.ReadBufferSize().
func ReadBufferSize(sz int) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.ReadBufferSize(sz))
	}
}

// InitialWindowSize is a wrapper of google.golang.org/grpc.InitialWindowSize().
func InitialWindowSize(sz int32) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.InitialWindowSize(sz))
	}
}

// InitialConnWindowSize is a wrapper of google.golang.org/grpc.InitialConnWindowSize().
func InitialConnWindowSize(sz int32) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.InitialConnWindowSize(sz))
	}
}

// KeepaliveParams is a wrapper of google.golang.org/grpc.KeepaliveParams().
func KeepaliveParams(kp keepalive.ServerParameters) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.KeepaliveParams(kp))
	}
}

// KeepaliveEnforcementPolicy is a wrapper of google.golang.org/grpc.KeepaliveEnforcementPolicy().
func KeepaliveEnforcementPolicy(kep keepalive.EnforcementPolicy) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.KeepaliveEnforcementPolicy(kep))
	}
}

// MaxRecvMsgSize is a wrapper of google.golang.org/grpc.MaxRecvMsgSize().
func MaxRecvMsgSize(m int) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.MaxRecvMsgSize(m))
	}
}

// MaxSendMsgSize is a wrapper of google.golang.org/grpc.MaxSendMsgSize().
func MaxSendMsgSize(m int) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.MaxSendMsgSize(m))
	}
}

// MaxConcurrentStreams is a wrapper of google.golang.org/grpc.MaxConcurrentStreams().
func MaxConcurrentStreams(n uint32) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.MaxConcurrentStreams(n))
	}
}

// Creds is a wrapper of google.golang.org/grpc.Creds().
func Creds(c credentials.TransportCredentials) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.Creds(c))
	}
}

// UnaryInterceptor is a wrapper of google.golang.org/grpc.ChainUnaryInterceptor().
func UnaryInterceptor(i grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.ChainUnaryInterceptor(i))
	}
}

// ChainUnaryInterceptor is a wrapper of google.golang.org/grpc.ChainUnaryInterceptor().
func ChainUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.ChainUnaryInterceptor(interceptors...))
	}
}

// StreamInterceptor is a wrapper of google.golang.org/grpc.ChainStreamInterceptor().
func StreamInterceptor(i grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.ChainStreamInterceptor(i))
	}
}

// ChainStreamInterceptor is a wrapper of google.golang.org/grpc.ChainStreamInterceptor().
func ChainStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.ChainStreamInterceptor(interceptors...))
	}
}

// InTapHandle is a wrapper of google.golang.org/grpc.InTapHandle().
func InTapHandle(h tap.ServerInHandle) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.InTapHandle(h))
	}
}

// StatsHandler is a wrapper of google.golang.org/grpc.StatsHandler().
func StatsHandler(h stats.Handler) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.StatsHandler(h))
	}
}

// UnknownServiceHandler is a wrapper of google.golang.org/grpc.UnknownServiceHandler().
func UnknownServiceHandler(streamHandler grpc.StreamHandler) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.UnknownServiceHandler(streamHandler))
	}
}

// ConnectionTimeout is a wrapper of google.golang.org/grpc.ConnectionTimeout().
func ConnectionTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.ConnectionTimeout(d))
	}
}

// MaxHeaderListSize is a wrapper of google.golang.org/grpc.MaxHeaderListSize().
func MaxHeaderListSize(sz uint32) ServerOption {
	return func(s *Server) {
		s.serverOpts = append(s.serverOpts, grpc.MaxHeaderListSize(sz))
	}
}
