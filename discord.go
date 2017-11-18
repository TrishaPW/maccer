package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ConnectDiscord sets up the Discord API and event listeners
func (app *App) ConnectDiscord() {
	var err error

	app.discordClient, err = discordgo.New("Bot " + app.config.DiscordToken)
	if err != nil {
		logger.Fatal("failed to connect to Discord API",
			zap.Error(err))
	}

	app.discordClient.AddHandler(app.onReady)
	app.discordClient.AddHandler(app.onMessage)
	app.discordClient.AddHandler(app.onJoin)

	err = app.discordClient.Open()
	if err != nil {
		logger.Fatal("failed to start Discord client",
			zap.Error(err))
	}
}

func (app *App) onReady(s *discordgo.Session, event *discordgo.Ready) {
	logger.Debug("discord ready")

	roles, err := s.GuildRoles(app.config.GuildID)
	if err != nil {
		logger.Fatal("failed to get guild roles",
			zap.Error(err))
	}

	found := 0
	for _, role := range roles {
		if role.ID == app.config.VerifiedRole {
			found++
		}
	}
	if found != 1 {
		logger.Fatal("role not found.",
			zap.String("role", app.config.VerifiedRole))
	}

	logger.Debug("app ready")

	app.ready <- true
}

func (app *App) onMessage(s *discordgo.Session, event *discordgo.MessageCreate) {
	if len(app.ready) > 0 {
		<-app.ready
	}

	if event.Message.Author.ID == app.config.BotID {
		return
	}

	if app.config.DebugUser != "" {
		if event.Message.Author.ID != app.config.DebugUser {
			logger.Debug("ignoring command from non debug user")
			return
		}
		logger.Debug("accepting command from debug user")
	}

	_, _, err := app.commandManager.Process(*event.Message)
	if err != nil {
		app.ChannelLogError(err)
	}

	// if source != CommandSourcePRIVATE && source != CommandSourceADMINISTRATIVE {
	// 	for i := range event.Message.Mentions {
	// 		if event.Message.Mentions[i].ID == app.config.BotID {
	// 			// todo: summon
	// 		}
	// 	}
	// }
}

func (app *App) onJoin(s *discordgo.Session, event *discordgo.GuildMemberAdd) {
	// todo: IsUserVerified
	// if verified {
	// 	err = app.discordClient.GuildMemberRoleAdd(app.config.GuildID, event.Member.User.ID, app.config.VerifiedRole)
	// 	if err != nil {
	// 		logger.Warn("failed to add verified role to member", zap.Error(err))
	// 	}
	// } else {
	// 	ch, err := s.UserChannelCreate(event.Member.User.ID)
	// 	if err != nil {
	// 		logger.Warn("failed to create user channel", zap.Error(err))
	// 		return
	// 	}
	// 	_, err = app.discordClient.ChannelMessageSend(ch.ID, app.locale.GetLangString("en", "AskUserVerify"))
	// 	if err != nil {
	// 		logger.Warn("failed to send message", zap.Error(err))
	// 	}
	// }
}

// ChannelLogError sends an error to the logging channel, exiting on failure
func (app *App) ChannelLogError(err error) {
	_, err = app.discordClient.ChannelMessageSend(app.config.LogChannel, fmt.Sprint(errors.WithStack(err)))
	if err != nil {
		logger.Fatal("failed to log error", zap.Error(err))
	}
}
