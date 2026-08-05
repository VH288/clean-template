package bootstrap

import (
	"log/slog"

	"clean-template/internal/app/lifecycle"
	"clean-template/internal/config"
	"clean-template/internal/infrastructure/apm"
	"clean-template/internal/infrastructure/external"
	grpcclient "clean-template/internal/infrastructure/grpc"
	"clean-template/internal/infrastructure/kafka"
	"clean-template/internal/infrastructure/mongodb"
	infraws "clean-template/internal/infrastructure/websocket"

	"github.com/jmoiron/sqlx"
	goredis "github.com/redis/go-redis/v9"
)

// Container is the root DI graph. Domain modules live in their own fields
// so bootstrap.go stays small as domains grow.
type Container struct {
	Config *config.Config
	Logger *slog.Logger

	Infra       *Infra
	Sample      *Sample
	Healthcheck *Healthcheck
	Lifecycle   *lifecycle.Manager
}

// Infra holds shared infrastructure dependencies.
type Infra struct {
	DB         *sqlx.DB
	Redis      *goredis.Client
	Mongo      *mongodb.Client
	KafkaProd  *kafka.Producer
	GRPCClient *grpcclient.Client
	WSHub      *infraws.Hub
	HTTPClient *external.HTTPClient
	APM        *apm.Provider
	Logger     *slog.Logger
	Config     *config.Config
}
