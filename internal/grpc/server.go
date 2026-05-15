package grpc

import (
	"fmt"
	"net"

	"finance-tracker/pb"
	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the gRPC server that hosts all gRPC services.
type Server struct {
	pb.UnimplementedTransactionServiceServer
	transactionSvc *service.TransactionService
	grpcServer     *grpc.Server
	lis            net.Listener
	jwtSecret      string
}

// NewServer creates a new gRPC server.
func NewServer(transactionSvc *service.TransactionService, jwtSecret string) *Server {
	return &Server{
		transactionSvc: transactionSvc,
		jwtSecret:      jwtSecret,
	}
}

// Serve starts the gRPC server on the given address.
func (s *Server) Serve(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.lis = lis
	s.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(newAuthInterceptor(s.jwtSecret).UnaryInterceptor))
	pb.RegisterTransactionServiceServer(s.grpcServer, s)
	fmt.Printf("gRPC server listening on %s\n", addr)
	return s.grpcServer.Serve(lis)
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

// MapAppErrorToGRPC converts an apperror.Error to a gRPC status code.
func MapAppErrorToGRPC(err *apperror.Error) codes.Code {
	switch err.Code {
	case "VALIDATION_ERROR":
		return codes.InvalidArgument
	case "NOT_FOUND":
		return codes.NotFound
	case "INTERNAL_ERROR":
		return codes.Internal
	case "UNAUTHORIZED":
		return codes.Unauthenticated
	case "FORBIDDEN":
		return codes.PermissionDenied
	case "CONFLICT":
		return codes.AlreadyExists
	case "INSUFFICIENT_FUNDS":
		return codes.FailedPrecondition
	default:
		return codes.Internal
	}
}

// mapErr maps an apperror.Error to a gRPC error with the correct status code.
func mapErr(appErr *apperror.Error) error {
	return status.Error(MapAppErrorToGRPC(appErr), appErr.Message)
}
