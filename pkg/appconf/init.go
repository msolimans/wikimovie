package appconf

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/msolimans/wikimovie/pkg/es"
	"github.com/msolimans/wikimovie/pkg/sqs"
	"github.com/spf13/viper"
)

type Environment string

const (
	Env_Dev  Environment = "dev"
	Env_Prod Environment = "prod"
)

type ServiceRateLimiter struct {
	Max              int //max number of reqs
	ExpirationInSecs int //(like within 60 seconds)
}
type ServiceConfig struct {
	Port      int
	Timeout   int
	RateLimit ServiceRateLimiter
}

// Configuration struct used for configuration
type Configuration struct {
	Service       ServiceConfig
	Env           Environment
	ElasticSearch es.ESConfig

	Bucket string
	Aws    *aws.Config
	Worker sqs.WorkerConfig
}

var config = &Configuration{}

func (*Configuration) IsDevelopment() bool {
	return config.Env == Env_Dev
}

func (*Configuration) IsProduction() bool {
	return config.Env == Env_Prod
}

func init() {
	viper.SetDefault("Env", Env_Dev)

	_ = viper.BindEnv("Env", "ENV")

	// bindings (Viper key to a ENV variable)
	_ = viper.BindEnv("Service.Port", "SERVICE_PORT")
	_ = viper.BindEnv("Service.Timeout", "SERVICE_TIMEOUT")
	_ = viper.BindEnv("Service.RateLimit.Max", "SERVICE_RATE_LIMIT_MAX")
	_ = viper.BindEnv("Service.RateLimit.ExpirationInSecs", "SERVICE_RATE_LIMIT_EXPIRATIONINSECS")

	_ = viper.BindEnv("ElasticSearch.IdleConnTimeout", "ELASTICSEARCH_IDLE_CONN_TIMEOUT")
	_ = viper.BindEnv("ElasticSearch.MaxIdleConnsPerHost", "ELASTICSEARCH_MAX_IDLE_CONNS_PER_HOST")
	_ = viper.BindEnv("ElasticSearch.MaxIdleConns", "ELASTICSEARCH_MAX_IDLE_CONNS")
	_ = viper.BindEnv("ElasticSearch.Urls", "ELASTICSEARCH_URLS")

	_ = viper.BindEnv("Aws.Region", "AWS_REGION")
	_ = viper.BindEnv("Aws.Endpoint", "AWS_ENDPOINT")

	_ = viper.BindEnv("Worker.WaitTimeSeconds", "WORKER_WAIT_TIME_SECONDS")
	_ = viper.BindEnv("Worker.MaxMessages", "WORKER_MAX_MESSAGES")
	_ = viper.BindEnv("Worker.RetryIntervals", "WORKER_RETRY_INTERVALS")
	_ = viper.BindEnv("Worker.Queue", "WORKER_QUEUE")

	_ = viper.BindEnv("Bucket", "BUCKET")

}
