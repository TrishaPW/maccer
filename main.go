package main

import (
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	// loads environment variables from .env
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
)

var logger *zap.Logger

func init() {
	var config zap.Config
	debug := os.Getenv("DEBUG")

	if os.Getenv("TESTING") != "" {
		config = zap.NewDevelopmentConfig()
		config.DisableCaller = true
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.MessageKey = "@message"
		config.EncoderConfig.TimeKey = "@timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		if debug != "0" && debug != "" {
			dyn := zap.NewAtomicLevel()
			dyn.SetLevel(zap.DebugLevel)
			config.Level = dyn
		}
	}
	_logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	logger = _logger.With(
		zap.String("@version", os.Getenv("GIT_HASH")),
		zap.Namespace("@fields"),
	)
}

// Config stores configuration variables
type Config struct {
	DiscordToken          string `required:"true",split_words:"true"` // discord API token
	BotID                 string `required:"true",split_words:"true"` // the bot's client ID
	GuildID               string `required:"true",split_words:"true"` // the discord server ID
	VerifiedRole          string `required:"true",split_words:"true"` // ID of the role for verified members
	DebugUser             string `required:"true",split_words:"true"` // When set, only this user can interact with the bot
	AdministrativeChannel string `required:"true",split_words:"true"` // administrative channel where someone can speak as bot
	PrimaryChannel        string `required:"true",split_words:"true"` // main channel the bot hangs out in
	LogChannel            string `required:"true",split_words:"true"` // logging channel for errors etc
	ForumEndpoint         string `required:"true",split_words:"true"` // Forum URL
	ForumKey              string `required:"true",split_words:"true"` // API key
	MongoHost             string `split_words:"true" required:"true"` // MongoDB host address
	MongoPort             string `split_words:"true" required:"true"` // MongoDB host port
	MongoName             string `split_words:"true" required:"true"` // MongoDB database name
	MongoUser             string `split_words:"true" required:"true"` // MongoDB user name
	MongoPass             string `split_words:"true"`                 // MongoDB password
}

func main() {
	var config Config
	err := envconfig.Process("maccer", &config)
	if err != nil {
		logger.Fatal("failed to load config",
			zap.Error(err))
	}
	Start(config)
}

func configStrFromEnv(name string) (value string) {
	value = os.Getenv(name)
	if value == "" {
		logger.Fatal("environment variable not set",
			zap.String("name", name))
	}
	return
}

func configIntFromEnv(name string) (value int) {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		logger.Fatal("environment variable not set",
			zap.String("name", name))
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		logger.Fatal("failed to convert environment variable to int",
			zap.Error(err),
			zap.String("name", name))
	}
	return
}
