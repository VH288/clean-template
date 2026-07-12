package service

import "context"

type ExternalHTTPClient interface {
	FetchExample(ctx context.Context) (string, error)
}

type ExternalGRPCClient interface {
	GetReference(ctx context.Context, id int32) (string, error)
}
