package main

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func commandVerify(cm CommandManager, args string, message discordgo.Message, contextual bool) (success bool, enterContext bool, err error) {
	logger.Info("verification request received",
		zap.String("userID", message.Author.ID))

	return
}
