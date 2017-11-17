package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (app *App) commandVerify(args string, message discordgo.Message, contextual bool) (success bool, err error) {
	logger.Info("verification request received",
		zap.String("userID", message.Author.ID))

	_, err = app.ipsClient.GetMember(args)
	if err != nil {
		return false, err
	}

	code, err := GenerateRandomString(8)
	if err != nil {
		return false, err
	}

	app.discordClient.ChannelMessage(
		message.ChannelID,
		fmt.Sprintf(`Verification --
Please paste this unique token into the **Discord** > **Verification Code** section of your profile:

%s%s%s`,
			"`", code, "`")) // can't escape ` inside a multi-line string so gotta format it in!

	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(time.Minute)
	go func() {
		var inlineErr error
	loop:
		for {
			select {
			case <-ticker.C:
				member, err := app.ipsClient.GetMember(args)
				if err != nil {
					app.discordClient.ChannelMessage(
						message.ChannelID,
						fmt.Sprintf("There was an error while attempting to load profile information: %v please let Southclaws know!", err))
					break loop
				}

				fieldGroups, ok := member.CustomFields["Discord"]
				if !ok {
					inlineErr = errors.New("no Discord field in ")
					break loop
				}

				gotCode, ok := fieldGroups["Verification Code"]
				if ok {
					if gotCode == code {
						member.CustomFields["Discord"]["Discord Username"] = message.Author.Username
						member.CustomFields["Discord"]["Discord ID"] = message.Author.ID
						app.discordClient.ChannelMessage(message.ChannelID, "Your accounts have been linked and you have been verified!")
					} else {
						app.discordClient.ChannelMessage(
							message.ChannelID,
							fmt.Sprintf("The codes did not match, the code you were given was '%s' and the code on your profile was '%s'",
								code, gotCode))
					}
					break loop
				}

			case <-timer.C:
				app.discordClient.ChannelMessage(
					message.ChannelID,
					"Your time has expired, please try again.")
			}
		}

		if inlineErr != nil {
			app.ChannelLogError(inlineErr)
		}
	}()

	return true, nil
}

/*
Author: Matt Silverlock
Date: 2014-05-24
Accessed: 2017-02-22
https://elithrar.github.io/article/generating-secure-random-numbers-crypto-rand
*/

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
