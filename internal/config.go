package internal

import (
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

type Config struct {
	Port     string     `yaml:"port" env:"PORT" env-required:"true"`
	Dapr     DaprConfig `yaml:"dapr"`
	Database string     `yaml:"database" env:"DATABASE"`
}

type DaprConfig struct {
	AppIds struct {
		Status   string `yaml:"status" env:"DAPR_STATUS_APP_ID" env-default:"status-service"`
		User     string `yaml:"user" env:"DAPR_USER_APP_ID" env-default:"user-service"`
		Timeline string `yaml:"timeline" env:"DAPR_TIMELINE_APP_ID" env-default:"timeline-service"`
		Media    string `yaml:"media" env:"DAPR_MEDIA_APP_ID" env-default:"media-service"`
	} `yaml:"app-ids"`
	Host       string           `yaml:"host" env:"DAPR_HOST" env-default:"http://localhost"`
	HttpPort   string           `yaml:"http-port" env:"DAPR_HTTP_PORT" env-default:"3500"`
	GrpcPort   string           `yaml:"grpc-port" env:"DAPR_GRPC_PORT" env-default:"50001"`
	PubSub     PubSubConfig     `yaml:"pubsub"`
	S3         S3BindingConfig  `yaml:"s3-binding"`
	StateStore StateStoreConfig `yaml:"state-store"`
}

type StateStoreConfig struct {
	Name string `yaml:"name" env:"DAPR_STATE_STORE_NAME" env-default:"pubsub"`
}

type PubSubConfig struct {
	Name  string `yaml:"name" env:"DAPR_PUBSUB_NAME" env-default:"pubsub"`
	Topic string `yaml:"topic" env:"DAPR_TOPIC_NAME"`
}

type S3BindingConfig struct {
	Name string `yaml:"name" env:"DAPR_S3_BINDING_NAME" env-default:"s3"`
}

func NewConfig(path string, logger *zap.Logger) *Config {
	var config Config
	err := cleanenv.ReadConfig(path, &config)

	if err != nil {
		description, _ := cleanenv.GetDescription(&config, nil)
		logger.Info("Config usage" + description)
		logger.Info("couldn't read config, using env as fallback", zap.Error(err))
		err := cleanenv.ReadEnv(&config)
		if err != nil {
			logger.Fatal("couldn't init config with config.yaml or env", zap.Error(err))
		}
	}

	return &config
}
