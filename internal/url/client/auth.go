package client

import (
	"context"

	pb "github.com/wxvn/grpc-mesh/proto/auth"
	"google.golang.org/grpc"
)

type AuthClient struct {
	client pb.AuthServiceClient
}

func NewAuthClient(conn grpc.ClientConnInterface) *AuthClient {
	return &AuthClient{
		client: pb.NewAuthServiceClient(conn),
	}
}

func (c *AuthClient) ValidateToken(ctx context.Context, accessToken string) (string, error) {

	response, err := c.client.ValidateToken(
		ctx,
		&pb.ValidateTokenRequest{
			AccessToken: accessToken,
		},
	)
	if err != nil {
		return "", err
	}

	return response.UserId, nil
}
