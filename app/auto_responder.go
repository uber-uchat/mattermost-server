// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

func (a *App) SendAutoResponse(channel *model.Channel, receiver, sender *model.User) {
	if receiver == nil || receiver.NotifyProps == nil {
		return
	}
	active := receiver.NotifyProps[model.AUTO_RESPONDER_ACTIVE_NOTIFY_PROP] == "true"
	message := receiver.NotifyProps[model.AUTO_RESPONDER_MESSAGE_NOTIFY_PROP]

	status, err := a.GetStatus(receiver.Id)
	if err != nil {
		status = &model.Status{UserId: receiver.Id, Status: model.STATUS_ONLINE, Manual: false, LastActivityAt: model.GetMillis(), ActiveChannel: ""}
	}

	if status.Status != model.STATUS_OUT_OF_OFFICE {
		return
	}

	channelMember, err := a.Srv.Store.Channel().GetMember(channel.Id, sender.Id)
	if err != nil {
		mlog.Error(err.Error())
		return
	}

	result := <-a.Srv.Store.OooRequestUser().Get(receiver.Id)
	if result.Err != nil {
		mlog.Error(result.Err.Error())
		return
	}
	oooUser := result.Data.(*model.OooUser)

	if active && message != "" && status.Status == model.STATUS_OUT_OF_OFFICE && (channelMember.LastAutoReplyPostAt < oooUser.CreateAt || channelMember.LastAutoReplyPostAt > oooUser.DeleteAt) {
		autoResponderPost := &model.Post{
			ChannelId: channel.Id,
			Message:   message,
			RootId:    "",
			ParentId:  "",
			Type:      model.POST_AUTO_RESPONDER,
			UserId:    receiver.Id,
		}

		if _, err := a.CreatePost(autoResponderPost, channel, false); err != nil {
			mlog.Error(err.Error())
		}
		channelMember.LastAutoReplyPostAt = model.GetMillis()
		result = <-a.Srv.Store.Channel().UpdateMember(channelMember)
	}
}

func (a *App) SetAutoResponderStatus(user *model.User, oldNotifyProps model.StringMap) {
	active := user.NotifyProps[model.AUTO_RESPONDER_ACTIVE_NOTIFY_PROP] == "true"
	oldActive := oldNotifyProps[model.AUTO_RESPONDER_ACTIVE_NOTIFY_PROP] == "true"

	autoResponderEnabled := !oldActive && active
	autoResponderDisabled := oldActive && !active

	if autoResponderEnabled {
		a.SetStatusOutOfOffice(user.Id)
	} else if autoResponderDisabled {
		a.SetStatusOnline(user.Id, true)
	}
}

func (a *App) DisableAutoResponder(userId string, asAdmin bool) *model.AppError {
	user, err := a.GetUser(userId)
	if err != nil {
		return err
	}

	props := user.NotifyProps
	props[model.AUTO_RESPONDER_FROM_DATE] = ""
	props[model.AUTO_RESPONDER_FROM_TIME] = ""
	props[model.AUTO_RESPONDER_TO_DATE] = ""
	props[model.AUTO_RESPONDER_TO_TIME] = ""

	user, err = a.UpdateUserNotifyProps(userId, props)
	if err != nil {
		return err
	}

	active := user.NotifyProps[model.AUTO_RESPONDER_ACTIVE_NOTIFY_PROP] == "true"
	if active {
		patch := &model.UserPatch{}
		patch.NotifyProps = user.NotifyProps
		patch.NotifyProps[model.AUTO_RESPONDER_ACTIVE_NOTIFY_PROP] = "false"

		_, err := a.PatchUser(userId, patch, asAdmin)
		if err != nil {
			return err
		}
	}

	return nil
}
