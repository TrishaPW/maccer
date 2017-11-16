package main

import (
	"time"

	"github.com/Southclaws/invision-community-go"

	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// App stores program state
type App struct {
	config         Config
	discordClient  *discordgo.Session
	ipsClient      *ips.Client
	ready          chan bool
	cache          *cache.Cache
	commandManager *CommandManager
}

// Start starts the app with the specified config and blocks until fatal error
func Start(config Config) {
	var err error

	app := App{
		config: config,
		cache:  cache.New(5*time.Minute, 30*time.Second),
	}

	app.ipsClient, err = ips.NewClient(config.ForumEndpoint, config.ForumKey)
	if err != nil {
		logger.Fatal("failed to create IPS client",
			zap.Error(err))
	}

	logger.Debug("started with debug logging enabled",
		zap.Any("config", app.config))

	app.StartCommandManager()
	app.ConnectDiscord()

	done := make(chan bool)
	<-done
}
