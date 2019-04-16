// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package utils

import (
	"fmt"
	"strings"
	"regexp"
	"io/ioutil"

	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

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
			filePath := *config.TeamSettings.AllowedMembersFilePath
			if config.TeamSettings.AllowedMembersFilePath != nil && len(filePath) > 0 {
				if file := FindConfigFile(filePath); file != "" {
					if data, err := ioutil.ReadFile(file); err == nil {
						isAllowedToJoin = strings.Contains(string(data), fmt.Sprintf(",%s,", user.Username))
						mlog.Error(":::::::::::::")
						mlog.Error(":::::DM debug:isMemberAllowedToJoin:Processed file:", mlog.Any("isAllowedToJoin", isAllowedToJoin), mlog.Any("user.Username", user.Username))
					}
				}
			}
		}
	}

	mlog.Error(":::::::::::::")
	mlog.Error(":::::DM debug:isMemberAllowedToJoin:", mlog.Any("isReadOnlyChannel", isReadOnlyChannel), mlog.Any("isAllowedToJoin", isAllowedToJoin))

	return isAllowedToJoin
}
