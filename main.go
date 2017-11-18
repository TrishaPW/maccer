package main

import (
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	// loads environment variables from .env
	_ "github.com/joho/godotenv/autoload"
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
	DiscordToken          string // discord API token
	BotID                 string // the bot's client ID
	GuildID               string // the discord server ID
	VerifiedRole          string // ID of the role for verified members
	DebugUser             string // When set, only this user can interact with the bot
	AdministrativeChannel string // administrative channel where someone can speak as bot
	PrimaryChannel        string // main channel the bot hangs out in
	LogChannel            string // logging channel for errors etc
	ForumEndpoint         string // Forum URL
	ForumKey              string // API key
}

func main() {
	Start(Config{
		DiscordToken:          configStrFromEnv("DISCORD_TOKEN"),
		BotID:                 configStrFromEnv("BOT_ID"),
		GuildID:               configStrFromEnv("GUILD_ID"),
		VerifiedRole:          configStrFromEnv("VERIFIED_ROLE"),
		DebugUser:             os.Getenv("DEBUG_USER"),
		AdministrativeChannel: configStrFromEnv("ADMINISTRATIVE_CHANNEL"),
		PrimaryChannel:        configStrFromEnv("PRIMARY_CHANNEL"),
		LogChannel:            configStrFromEnv("LOG_CHANNEL"),
		ForumEndpoint:         configStrFromEnv("FORUM_ENDPOINT"),
		ForumKey:              configStrFromEnv("FORUM_KEY"),
	})
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
