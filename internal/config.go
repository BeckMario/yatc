package internal

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
	Host     string       `yaml:"host" env:"DAPR_HOST" env-default:"http://localhost"`
	HttpPort string       `yaml:"http-port" env:"DAPR_HTTP_PORT" env-default:"3500"`
	GrpcPort string       `yaml:"grpc-port" env:"DAPR_GRPC_PORT" env-default:"50001"`
	PubSub   PubSubConfig `yaml:"pubsub"`
}

type PubSubConfig struct {
	Name  string `yaml:"name" env:"DAPR_PUBSUB_NAME" env-default:"pubsub"`
	Topic string `yaml:"topic" env:"DAPR_TOPIC_NAME"`
}
