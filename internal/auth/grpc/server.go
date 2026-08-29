package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/wxvn/grpc-mesh/internal/auth/service"
	pb "github.com/wxvn/grpc-mesh/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Register(ctx context.Context, email string, password string) (service.RegisterResult, error)
	Login(ctx context.Context, email string, password string) (service.LoginResult, error)
	RefreshTokens(ctx context.Context, refreshToken string) (service.RefreshTokensResult, error)
	Logout(ctx context.Context, refreshToken string) error

	ValidateToken(ctx context.Context, accessToken string) (uuid.UUID, error)
}

type AuthServer struct {
	pb.UnimplementedAuthServiceServer

	service AuthService
}

func NewServer(service AuthService) *AuthServer {
	return &AuthServer{
		service: service,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	result, err := s.service.Register(
		ctx,
		req.Email,
		req.Password,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.RegisterResponse{
		UserId:       result.UserID.String(),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	result, err := s.service.Login(
		ctx,
		req.Email,
		req.Password,
	)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (s *AuthServer) RefreshTokens(ctx context.Context, req *pb.RefreshTokensRequest) (*pb.RefreshTokensResponse, error) {
	result, err := s.service.RefreshTokens(
		ctx,
		req.RefreshToken,
	)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.RefreshTokensResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := s.service.Logout(
		ctx,
		req.RefreshToken,
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.LogoutResponse{
		Success: true,
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	userID, err := s.service.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return nil, status.Error(
			codes.Unauthenticated,
			err.Error(),
		)
	}

	return &pb.ValidateTokenResponse{
		UserId: userID.String(),
	}, nil
}
