// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAutoResponseStatus(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	th.App.SetStatusOnline(user.Id, "", true)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_reply_active"] = "true"
	patch.NotifyProps["auto_reply_message"] = "Hello, I'm unavailable today."

	userUpdated1, _ := th.App.PatchUser(user.Id, patch, true)

	// autoResponse is enabled, status should be OOO
	th.App.SetAutoResponseStatus(userUpdated1, user.NotifyProps)

	status, err := th.App.GetStatus(userUpdated1.Id)
	require.Nil(t, err)
	assert.Equal(t, model.STATUS_OUT_OF_OFFICE, status.Status)

	patch2 := &model.UserPatch{}
	patch2.NotifyProps = make(map[string]string)
	patch2.NotifyProps["auto_reply_active"] = "false"
	patch2.NotifyProps["auto_reply_message"] = "Hello, I'm unavailable today."

	userUpdated2, _ := th.App.PatchUser(user.Id, patch2, true)

	// autoResponse is disabled, status should be ONLINE
	th.App.SetAutoResponseStatus(userUpdated2, userUpdated1.NotifyProps)

	status, err = th.App.GetStatus(userUpdated2.Id)
	require.Nil(t, err)
	assert.Equal(t, model.STATUS_ONLINE, status.Status)

}

func TestDisableAutoResponse(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	th.App.SetStatusOnline(user.Id, "", true)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_reply_active"] = "true"
	patch.NotifyProps["auto_reply_message"] = "Hello, I'm unavailable today."

	th.App.PatchUser(user.Id, patch, true)

	th.App.DisableAutoResponse(user.Id, true)

	userUpdated1, err := th.App.GetUser(user.Id)
	require.Nil(t, err)
	assert.Equal(t, userUpdated1.NotifyProps["auto_reply_active"], "false")

	th.App.DisableAutoResponse(user.Id, true)

	userUpdated2, err := th.App.GetUser(user.Id)
	require.Nil(t, err)
	assert.Equal(t, userUpdated2.NotifyProps["auto_reply_active"], "false")
}

func TestSendAutoResponseSuccess(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_reply_active"] = "true"
	patch.NotifyProps["auto_reply_message"] = "Hello, I'm unavailable today."

	userUpdated1, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)

	th.App.SendAutoResponse(th.BasicChannel, userUpdated1)

	if list, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1); err != nil {
		require.Nil(t, err)
	} else {
		autoResponderPostFound := false
		for _, post := range list.Posts {
			if post.Type == model.POST_AUTO_RESPONSE {
				autoResponderPostFound = true
			}
		}
		assert.True(t, autoResponderPostFound)
	}
}

func TestSendAutoResponseFailure(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(user)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_reply_active"] = "false"
	patch.NotifyProps["auto_reply_message"] = "Hello, I'm unavailable today."

	userUpdated1, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)

	th.App.SendAutoResponse(th.BasicChannel, userUpdated1)

	if list, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1); err != nil {
		require.Nil(t, err)
	} else {
		autoResponderPostFound := false
		for _, post := range list.Posts {
			if post.Type == model.POST_AUTO_RESPONSE {
				autoResponderPostFound = true
			}
		}
		assert.False(t, autoResponderPostFound)
	}
}
