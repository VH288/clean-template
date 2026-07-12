package service

import (
	"context"
	"fmt"
)

type HealthcheckService struct {
	externalHTTP ExternalHTTPClient
	externalGRPC ExternalGRPCClient
}

func NewHealthcheckService(externalHTTP ExternalHTTPClient, externalGRPC ExternalGRPCClient) *HealthcheckService {
	return &HealthcheckService{
		externalHTTP: externalHTTP,
		externalGRPC: externalGRPC,
	}
}

func (s *HealthcheckService) Check(ctx context.Context) (string, error) {
	httpMsg, err := s.externalHTTP.FetchExample(ctx)
	if err != nil {
		return "", err
	}

	grpcMsg, err := s.externalGRPC.GetReference(ctx, 1)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("service healthy, external http: %s, external grpc: %s", httpMsg, grpcMsg), nil
}
