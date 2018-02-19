package main

import (
	"fmt"
	"time"

	"github.com/Southclaws/invision-community-go"
	"github.com/bwmarrin/discordgo"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"
)

// App stores program state
type App struct {
	config         Config
	discordClient  *discordgo.Session
	mongodb        *mgo.Session
	users          *mgo.Collection
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

	app.mongodb, err = mgo.Dial(fmt.Sprintf("%s:%s", config.MongoHost, config.MongoPort))
	if err != nil {
		logger.Fatal("failed to connect to database",
			zap.Error(err))
	}

	if config.MongoPass != "" {
		err = app.mongodb.Login(&mgo.Credential{
			Source:   config.MongoName,
			Username: config.MongoUser,
			Password: config.MongoPass,
		})
		if err != nil {
			logger.Fatal("failed to authenticate to database",
				zap.Error(err))
		}
	}

	exists, err := app.CollectionExists(config.MongoName, "users")
	if err != nil {
		logger.Fatal("failed to check collection", zap.Error(err))
	}
	if !exists {
		err = app.mongodb.DB(config.MongoName).C("users").Create(&mgo.CollectionInfo{})
		if err != nil {
			logger.Fatal("failedto create collection", zap.Error(err))
		}
	}
	app.users = app.mongodb.DB(config.MongoName).C("users")

	err = app.users.EnsureIndex(mgo.Index{
		Name:   "UNIQUE_DISCORD",
		Key:    []string{"discord_id"},
		Unique: true,
	})
	if err != nil {
		logger.Fatal("failed to ensure index",
			zap.Error(err))
	}

	err = app.users.EnsureIndex(mgo.Index{
		Name:   "UNIQUE_FORUM",
		Key:    []string{"forum_id"},
		Unique: true,
	})
	if err != nil {
		logger.Fatal("failed to ensure index",
			zap.Error(err))
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

// CollectionExists checks if a collection exists in MongoDB
func (app App) CollectionExists(db, wantCollection string) (bool, error) {
	collections, err := app.mongodb.DB(db).CollectionNames()
	if err != nil {
		return false, err
	}

	for _, collection := range collections {
		if collection == wantCollection {
			return true, nil
		}
	}

	return false, nil
}
