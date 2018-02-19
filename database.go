package main

import (
	"strings"

	"github.com/Southclaws/maccer/types"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

var (
	// ErrUserDiscordDuplicate is triggered when a discord ID is attempted to be registered twice
	ErrUserDiscordDuplicate = errors.New("discord ID already registered")
	// ErrUserForumDuplicate is triggered when a forum ID is attempted to be registered twice
	ErrUserForumDuplicate = errors.New("forum ID already registered")
)

// CreateUser inserts a new record for a user
func (app App) CreateUser(user types.User) (err error) {
	err = app.users.Insert(user)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE_DISCORD") {
			err = ErrUserDiscordDuplicate
		}
		if strings.Contains(err.Error(), "UNIQUE_DISCORD") {
			err = ErrUserForumDuplicate
		}
	}
	return
}

// GetUserByDiscord returns a user from the database via their discord ID
func (app App) GetUserByDiscord(id string) (user types.User, exists bool, err error) {
	err = app.users.Find(bson.M{"discord_id": id}).One(&user)
	if err != nil {
		if err.Error() == "not found" {
			err = nil
		} else {
			err = errors.Wrap(err, "failed to get user by discord ID")
		}
	} else {
		exists = true
	}
	return
}

// GetUserByForum returns a user from the database via their forum ID
func (app App) GetUserByForum(id string) (user types.User, exists bool, err error) {
	err = app.users.Find(bson.M{"forum_id": id}).One(&user)
	if err != nil {
		if err.Error() == "not found" {
			err = nil
		} else {
			err = errors.Wrap(err, "failed to get user by discord ID")
		}
	} else {
		exists = true
	}
	return
}

// UpdateUser updates the details for a user in the database
func (app App) UpdateUser(user types.User) (err error) {
	err = app.users.Update(bson.M{"discord_id": user.DiscordID}, user)
	return
}
