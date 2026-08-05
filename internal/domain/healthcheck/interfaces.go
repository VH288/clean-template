package healthcheck

import "context"

type DependencyStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Report struct {
	Status       string             `json:"status"`
	Dependencies []DependencyStatus `json:"dependencies"`
}

type PostgresChecker interface {
	Ping(ctx context.Context) error
}

type RedisChecker interface {
	Ping(ctx context.Context) error
}

type MongoChecker interface {
	Ping(ctx context.Context) error
}

type GRPCChecker interface {
	CheckHealth(ctx context.Context, service string) error
}

type Usecase interface {
	CheckHTTP(ctx context.Context) *Report
	CheckGRPC(ctx context.Context) *Report
	CheckWebSocket(ctx context.Context) *Report
}
