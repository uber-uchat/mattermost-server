// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	l4g "github.com/alecthomas/log4go"

	"github.com/mattermost/mattermost-server/model"
)

func (a *App) SendAutoResponse(channel *model.Channel, receiver *model.User) {
	active, message :=
		receiver.NotifyProps["auto_reply_active"] == "true",
		receiver.NotifyProps["auto_reply_message"]

	if active && message != "" {
		autoResponsePost := &model.Post{
			ChannelId: channel.Id,
			Message:   message,
			Type:      model.POST_AUTO_RESPONSE,
			UserId:    receiver.Id,
		}

		if _, err := a.CreatePost(autoResponsePost, channel, false); err != nil {
			l4g.Error(err.Error())
		}
	}
}

func (a *App) SetAutoResponseStatus(user *model.User, oldNotifyProps model.StringMap) {
	active := user.NotifyProps["auto_reply_active"] == "true"
	oldActive := oldNotifyProps["auto_reply_active"] == "true"

	autoResponseEnabled := !oldActive && active
	autoResponseDisabled := oldActive && !active

	if autoResponseEnabled {
		a.SetStatusOutOfOffice(user.Id)
	} else if autoResponseDisabled {
		a.SetStatusOnline(user.Id, "", true)
	}
}

func (a *App) DisableAutoResponse(userId string, asAdmin bool) *model.AppError {
	user, err := a.GetUser(userId)
	if err != nil {
		return err
	}

	active := user.NotifyProps["auto_reply_active"] == "true"

	if active {
		patch := &model.UserPatch{}
		patch.NotifyProps = user.NotifyProps
		patch.NotifyProps["auto_reply_active"] = "false"

		a.PatchUser(userId, patch, asAdmin)
	}

	return nil
}
