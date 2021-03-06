package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Southclaws/invision-community-go"
	"github.com/Southclaws/maccer/types"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// MatchURL matches a user's profile URL and captures the ID
var MatchURL = regexp.MustCompile(`https:\/\/forum\.bayarearoleplay\.com\/profile\/([0-9]*)-(\w+)(\/)?`)

func (app *App) commandVerify(args string, message discordgo.Message, contextual bool) (success bool, err error) {
	logger.Debug("verification request received",
		zap.String("url", args),
		zap.String("userID", message.Author.ID))

	match := MatchURL.FindStringSubmatch(args)

	if len(match) < 2 {
		_, err = app.discordClient.ChannelMessageSend(
			message.ChannelID,
			fmt.Sprintf(`
That is not a valid URL to a user page, it should be in the format:

%s

For more help, please read: https://forum.bayarearoleplay.com/topic/705-how-to-verify-your-discord-account/	`, "`https://forum.bayarearoleplay.com/profile/21-southclaws/`"))
		return false, err
	}

	userID := match[1]

	_, err = app.ipsClient.GetMember(userID)
	if err != nil {
		return false, err
	}

	code := uuid.New().String()

	_, err = app.discordClient.ChannelMessageSend(message.ChannelID, "***-- Verification --***\nPlease paste this unique token into the **Discord** > **Verification Code** section of your profile:")
	if err != nil {
		return false, err
	}
	_, err = app.discordClient.ChannelMessageSend(message.ChannelID, fmt.Sprintf("`%s`", code))
	if err != nil {
		return false, err
	}
	_, err = app.discordClient.ChannelMessageSend(message.ChannelID, "You can find this section at the bottom of the **Edit Profile** menu:\n\nhttps://i.imgur.com/JJMC0KZ.png\n\nhttps://i.imgur.com/n8vfO2N.png")
	if err != nil {
		return false, err
	}

	ticker := time.NewTicker(time.Second * 5)
	timer := time.NewTimer(time.Minute * 5)
	go func() {
		var (
			member    ips.Member
			inlineErr error
		)

	loop:
		for {
			select {
			case <-ticker.C:
				member, inlineErr = app.ipsClient.GetMember(userID)
				if inlineErr != nil {
					inlineErr = errors.Wrap(err, "failed to get member data from forum API")
					break loop
				}

				fieldGroups, ok := member.CustomFields["Discord"]
				if !ok {
					inlineErr = errors.New("no Discord field in member custom fields")
					break loop
				}

				gotCode, ok := fieldGroups["Verification Code"]
				if ok && len(gotCode) >= 8 {
					if gotCode == code {
						user := types.User{
							DiscordID: message.Author.ID,
							ForumID:   userID,
						}

						inlineErr = app.CreateUser(user)
						if inlineErr != nil {
							inlineErr = errors.Wrap(err, "failed to update user in database")
							break loop
						}

						inlineErr = app.discordClient.GuildMemberRoleAdd(
							app.config.GuildID,
							message.Author.ID,
							app.config.VerifiedRole,
						)
						if inlineErr != nil {
							inlineErr = errors.Wrap(err, "failed to add member to role")
							break loop
						}

						_, inlineErr = app.discordClient.ChannelMessageSend(message.ChannelID, "Your accounts have been linked and you have been verified!")
						if inlineErr != nil {
							inlineErr = errors.Wrap(err, "failed to send private message")
							break loop
						}

						break loop
					} /* else {
						_, inlineErr = app.discordClient.ChannelMessageSend(
							message.ChannelID,
							fmt.Sprintf("The codes did not match, the code you were given was '%s' and the code on your profile was '%s'",
								code, gotCode))
						if inlineErr != nil {
							break loop
						}
					}*/
				}

				logger.Debug("no code yet",
					zap.Any("customFields", fieldGroups))

			case <-timer.C:
				_, inlineErr = app.discordClient.ChannelMessageSend(
					message.ChannelID,
					"Your time has expired, please try again.")
				break loop
			}
		}

		if inlineErr != nil {
			app.ChannelLogError(inlineErr)
		}
	}()

	return true, nil
}
