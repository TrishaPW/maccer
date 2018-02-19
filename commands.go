package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// Command represents a public, private or administrative command
type Command struct {
	commandManager  *CommandManager
	Function        func(args string, message discordgo.Message, contextual bool) (bool, error)
	Source          CommandSource
	ParametersRange CommandParametersRange
	Description     string
	Usage           string
	Example         string
	RequireVerified bool
	RequireAdmin    bool
	Context         bool
}

// LoadCommands is called on initialisation and is responsible for registering
// all commands and binding them to functions.
func LoadCommands(app *App) map[string]Command {
	return map[string]Command{
		"verify": {
			Function:    app.commandVerify,
			Source:      CommandSourcePRIVATE,
			Description: "Verify you are the owner of a Bay Area Roleplay forum account",
			Usage:       "verify <profile page URL>\nYour profile page can be accessed here: https://i.imgur.com/htrHTvV.png",
			ParametersRange: CommandParametersRange{
				Minimum: 1,
				Maximum: 1,
			},
			RequireVerified: false,
			RequireAdmin:    false,
			Context:         true,
		},
		"whois": {
			Function:    app.commandWhoIs,
			Source:      CommandSourcePRIMARY,
			Description: "Get a Discord users' forum account",
			Usage:       "whois @user",
			ParametersRange: CommandParametersRange{
				Minimum: 1,
				Maximum: 1,
			},
			RequireVerified: true,
			RequireAdmin:    false,
			Context:         false,
		},
	}
}

// CommandSource represents the source of a command.
type CommandSource int8

const (
	// CommandSourceADMINISTRATIVE are commands in the administrator channel,
	// mainly for admin work that may clutter up the primary channel.
	CommandSourceADMINISTRATIVE CommandSource = iota
	// CommandSourcePRIMARY are primary channel commands visible to all users
	// and mainly used for fun and group activity commands.
	CommandSourcePRIMARY CommandSource = iota
	// CommandSourcePRIVATE are private channel commands for dealing with
	// sensitive information such as verification.
	CommandSourcePRIVATE CommandSource = iota
	// CommandSourceOTHER represents any other channel that does not fall into
	// the above sources.
	CommandSourceOTHER CommandSource = iota
)

// CommandManager stores command state
type CommandManager struct {
	App      *App
	Commands map[string]Command
}

// CommandParametersRange represents minimum value and maximum value number of parameters for a command
type CommandParametersRange struct {
	Minimum int
	Maximum int
}

// StartCommandManager creates a command manager for the app
func (app *App) StartCommandManager() {
	app.commandManager = &CommandManager{
		App:      app,
		Commands: make(map[string]Command),
	}

	app.commandManager.Commands = LoadCommands(app)
}

// Process is called on a command string to check whether it's a valid command
// and, if so, call the associated function.
// nolint:gocyclo
func (cm CommandManager) Process(message discordgo.Message) (exists bool, source CommandSource, err error) {
	source, err = cm.getCommandSource(message)
	if err != nil {
		return
	}

	commandAndParameters := strings.SplitN(message.Content, " ", 2)
	commandParametersCount := 0
	commandTrigger := strings.ToLower(commandAndParameters[0])
	commandArgument := ""

	if len(commandAndParameters) > 1 {
		commandArgument = commandAndParameters[1]
		commandParametersCount = strings.Count(commandArgument, " ") + 1
	}

	commandObject, exists := cm.Commands[commandTrigger]
	commandObject.commandManager = &cm

	if !exists {
		return
	}

	if source != commandObject.Source {
		return
	}

	switch source {
	case CommandSourceADMINISTRATIVE:
		if message.ChannelID != cm.App.config.AdministrativeChannel {
			return
		}
	case CommandSourcePRIMARY:
		if message.ChannelID != cm.App.config.PrimaryChannel {
			return
		}
	}

	// Check if the user is verified.
	// todo: write IsUserVerified based on roles rather than a DB
	// if commandObject.RequireVerified {
	// 	verified, err := cm.App.IsUserVerified(message.Author.ID)
	// 	if err != nil {
	// 		errs = append(errs, err)
	// 		return exists, source, errs
	// 	}
	// 	if !verified {
	// 		logger.Debug("ignoring command that requires verification from non-verified user", zap.String("command", commandTrigger))

	// 		_, err = cm.App.discordClient.ChannelMessageSend(message.ChannelID, cm.App.locale.GetLangString("en", "CommandRequireVerification", message.Author.ID))
	// 		if err != nil {
	// 			errs = append(errs, err)
	// 		}
	// 		return exists, source, errs
	// 	}
	// }

	// Check if we have the required number of parameters.
	if commandObject.ParametersRange.Minimum > -1 && commandParametersCount < commandObject.ParametersRange.Minimum {
		logger.Debug("ignoring ignoring command with incorrect parameter count", zap.String("command", commandTrigger))

		_, err = cm.App.discordClient.ChannelMessageSend(message.ChannelID,
			fmt.Sprintf("%s\n%s\n%s", commandObject.Usage, commandObject.Description, commandObject.Example))

		return
	} else if commandObject.ParametersRange.Maximum > -1 && commandParametersCount > commandObject.ParametersRange.Maximum {
		logger.Debug("ignoring ignoring command with incorrect parameter count", zap.String("command", commandTrigger))

		_, err = cm.App.discordClient.ChannelMessageSend(message.ChannelID,
			fmt.Sprintf("Too many parameters, command requires %d", commandObject.ParametersRange.Maximum))

		return
	}

	err = cm.App.discordClient.ChannelTyping(message.ChannelID)
	if err != nil {
		return
	}

	success, err := commandObject.Function(commandArgument, message, false)
	if err != nil {
		return
	}

	if !success {
		_, err = cm.App.discordClient.ChannelMessageSend(
			message.ChannelID,
			fmt.Sprintf("%s\n%s\n%s", commandObject.Usage, commandObject.Description, commandObject.Example))
		return
	}

	return
}

func (cm CommandManager) getCommandSource(message discordgo.Message) (CommandSource, error) {
	if message.ChannelID == cm.App.config.AdministrativeChannel {
		return CommandSourceADMINISTRATIVE, nil
	} else if message.ChannelID == cm.App.config.PrimaryChannel {
		return CommandSourcePRIMARY, nil
	} else {
		ch, err := cm.App.discordClient.Channel(message.ChannelID)
		if err != nil {
			return CommandSourceOTHER, err
		}

		if ch.Type == discordgo.ChannelTypeDM {
			return CommandSourcePRIVATE, nil
		}
	}

	return CommandSourceOTHER, nil
}
