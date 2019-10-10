// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAutoResponderStatus(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	th.App.SetStatusOnline(user.Id, true)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "true"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated1, _ := th.App.PatchUser(user.Id, patch, true)

	// autoResponder is enabled, status should be OOO
	th.App.SetAutoResponderStatus(userUpdated1, user.NotifyProps)

	status, err := th.App.GetStatus(userUpdated1.Id)
	require.Nil(t, err)
	assert.Equal(t, model.STATUS_OUT_OF_OFFICE, status.Status)

	patch2 := &model.UserPatch{}
	patch2.NotifyProps = make(map[string]string)
	patch2.NotifyProps["auto_responder_active"] = "false"
	patch2.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated2, _ := th.App.PatchUser(user.Id, patch2, true)

	// autoResponder is disabled, status should be ONLINE
	th.App.SetAutoResponderStatus(userUpdated2, userUpdated1.NotifyProps)

	status, err = th.App.GetStatus(userUpdated2.Id)
	require.Nil(t, err)
	assert.Equal(t, model.STATUS_ONLINE, status.Status)

}

func TestDisableAutoResponder(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	th.App.SetStatusOnline(user.Id, true)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "true"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	th.App.PatchUser(user.Id, patch, true)

	th.App.DisableAutoResponder(user.Id, true)

	userUpdated1, err := th.App.GetUser(user.Id)
	require.Nil(t, err)
	assert.Equal(t, userUpdated1.NotifyProps["auto_responder_active"], "false")

	th.App.DisableAutoResponder(user.Id, true)

	userUpdated2, err := th.App.GetUser(user.Id)
	require.Nil(t, err)
	assert.Equal(t, userUpdated2.NotifyProps["auto_responder_active"], "false")
}

func TestSendAutoResponseSuccess(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "true"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated1, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)
	err = th.App.InsertOooRequestUser(user.Id, 1, 1, map[string]string{}, map[string]string{})
	require.Nil(t, err)
	th.App.SetStatusOutOfOffice(user.Id)

	user = th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	patch = &model.UserPatch{}
	userUpdated2, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)
	_, err = th.App.AddTeamMember(th.BasicTeam.Id, userUpdated2.Id)
	require.Nil(t, err)
	_, err = th.App.AddUserToChannel(userUpdated2, th.BasicChannel)
	require.Nil(t, err)

	th.App.CreatePost(&model.Post{
		ChannelId: th.BasicChannel.Id,
		Message:   "zz" + model.NewId() + "a",
		UserId:    th.BasicUser.Id},
		th.BasicChannel,
		false)

	th.App.SendAutoResponse(th.BasicChannel, userUpdated1, userUpdated2)

	if list, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1); err != nil {
		require.Nil(t, err)
	} else {
		autoResponderPostFound := false
		for _, post := range list.Posts {
			if post.Type == model.POST_AUTO_RESPONDER {
				autoResponderPostFound = true
			}
		}
		assert.True(t, autoResponderPostFound)
	}
}

func TestSendAutoResponseFailure(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "false"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated1, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)

	user = th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	patch = &model.UserPatch{}
	userUpdated2, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)

	_, err = th.App.AddTeamMember(th.BasicTeam.Id, userUpdated2.Id)
	require.Nil(t, err)
	_, err = th.App.AddUserToChannel(userUpdated2, th.BasicChannel)
	require.Nil(t, err)
	th.App.CreatePost(&model.Post{
		ChannelId: th.BasicChannel.Id,
		Message:   "zz" + model.NewId() + "a",
		UserId:    th.BasicUser.Id},
		th.BasicChannel,
		false)

	th.App.SendAutoResponse(th.BasicChannel, userUpdated1, userUpdated2)

	if list, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1); err != nil {
		require.Nil(t, err)
	} else {
		autoResponderPostFound := false
		for _, post := range list.Posts {
			if post.Type == model.POST_AUTO_RESPONDER {
				autoResponderPostFound = true
			}
		}
		assert.False(t, autoResponderPostFound)
	}
}

func TestAutoResponseConsolidation(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "true"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated1, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)
	err = th.App.InsertOooRequestUser(user.Id, 1, 1, map[string]string{}, map[string]string{})
	require.Nil(t, err)
	th.App.SetStatusOutOfOffice(user.Id)

	user = th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	patch = &model.UserPatch{}
	userUpdated2, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)
	_, err = th.App.AddTeamMember(th.BasicTeam.Id, userUpdated2.Id)
	require.Nil(t, err)
	_, err = th.App.AddUserToChannel(userUpdated2, th.BasicChannel)
	require.Nil(t, err)

	th.App.CreatePost(&model.Post{
		ChannelId: th.BasicChannel.Id,
		Message:   "zz" + model.NewId() + "a",
		UserId:    th.BasicUser.Id},
		th.BasicChannel,
		false)

	th.App.SendAutoResponse(th.BasicChannel, userUpdated1, userUpdated2)

	if _, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1); err != nil {
		require.Nil(t, err)
	} else {
		th.App.SendAutoResponse(th.BasicChannel, userUpdated1, userUpdated2)

		if list, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1); err != nil {
			require.Nil(t, err)
		} else {
			autoResponderPostCount := 0
			for _, post := range list.Posts {
				if post.Type == model.POST_AUTO_RESPONDER {
					autoResponderPostCount += 1
				}
			}
			assert.True(t, autoResponderPostCount == 1)
		}
	}

}
