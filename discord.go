package main

import (
	"github.com/bwmarrin/discordgo"
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

// nolint:gocyclo
func (app *App) onReady(s *discordgo.Session, event *discordgo.Ready) {
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

	app.ready <- true
}

// nolint:gocyclo
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

	_, source, errors := app.commandManager.Process(*event.Message)
	for _, err := range errors {
		if err != nil {
			err = app.WarnUserError(event.Message.ChannelID, err.Error())
			if err != nil {
				logger.Warn("failed to warn user of error", zap.Error(err))
			}
		}
	}

	if source != CommandSourcePRIVATE && source != CommandSourceADMINISTRATIVE {
		for i := range event.Message.Mentions {
			if event.Message.Mentions[i].ID == app.config.BotID {
				// todo: summon
				// err := app.HandleSummon(*event.Message)
				// if err != nil {
				// 	logger.Warn("failed to handle summon", zap.Error(err))
				// }
			}
		}
	}
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

// WarnUserError informs the user of an error and provides them with
// instructions for what to do next.
func (app App) WarnUserError(channelid string, errorString string) error {
	_, err := app.discordClient.ChannelMessageSend(channelid, errorString)
	return err
}
