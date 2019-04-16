// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package utils

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	// "github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

var allowedMembers = ""

func IsMemberAllowedToJoin(channel *model.Channel, user *model.User, config *model.Config) bool {
	isReadOnlyChannel := false

	for _, channelName := range config.TeamSettings.ReadOnlyChannels {
		if channel.Name == channelName {
			isReadOnlyChannel = true
		}
	}

	isAllowedToJoin := true

	if isReadOnlyChannel {
		isAllowedToJoin, _ = regexp.MatchString("\\w+@uber.com", user.Email)
		if isAllowedToJoin {
			//if allowedMembers == "" {
			filePath := *config.TeamSettings.AllowedMembersFilePath
			if config.TeamSettings.AllowedMembersFilePath != nil && len(filePath) > 0 {
				if file := FindConfigFile(filePath); file != "" {
					if data, err := ioutil.ReadFile(file); err == nil {
						allowedMembers = string(data)
						// mlog.Error(":::::::::::::")
						// mlog.Error(":::::DM debug::Processing file:", mlog.Any("allowedMembers", allowedMembers))
					}
				}
			}
			//}
			isAllowedToJoin = strings.Contains(allowedMembers, fmt.Sprintf(",%s,", user.Username))
			// mlog.Error(":::::::::::::")
			// mlog.Error(":::::DM debug:isMemberAllowedToJoin:string ready:", mlog.Any("isAllowedToJoin", isAllowedToJoin), mlog.Any("user.Username", user.Username))
		}
	}

	return isAllowedToJoin
}
