package reference

import (
	"context"
	"fmt"

	"clean-template/internal/config"
	externalv1 "clean-template/proto/external/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client externalv1.ReferenceServiceClient
}

func NewClient(cfg config.Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.ExternalGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial external grpc: %w", err)
	}

	return &Client{
		conn:   conn,
		client: externalv1.NewReferenceServiceClient(conn),
	}, nil
}

func (c *Client) GetReference(ctx context.Context, id int32) (string, error) {
	resp, err := c.client.GetReference(ctx, &externalv1.GetReferenceRequest{Id: id})
	if err != nil {
		return "", fmt.Errorf("call external grpc GetReference: %w", err)
	}
	return resp.GetLabel(), nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
