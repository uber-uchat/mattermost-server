// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package utils

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils/fileutils"
)

var allowedMembers = ""
var configWatcher *ConfigWatcher

func IsReadOnlyChannel(channel *model.Channel, config *model.Config) bool {
	for _, channelName := range config.TeamSettings.ReadOnlyChannels {
		if channel.Name == channelName {
			return true
		}
	}
	return false
}

func IsMemberAllowedToJoin(channel *model.Channel, user *model.User, config *model.Config) bool {
	isReadOnlyChannel := IsReadOnlyChannel(channel, config)

	isAllowedToJoin := true

	if isReadOnlyChannel {
		isAllowedToJoin, _ = regexp.MatchString("\\w+@uber.com", user.Email)
		if isAllowedToJoin {
			if allowedMembers == "" {
				if !LoadAllowedMembers(config.TeamSettings.AllowedMembersFilePath) {
					return true
				}
			}
			isAllowedToJoin = strings.Contains(allowedMembers, fmt.Sprintf(",%s,", user.Username))
		}
	}

	return isAllowedToJoin
}

func LoadAllowedMembers(filePath *string) bool {
	if filePath != nil && len(*filePath) > 0 {
		if file := fileutils.FindConfigFile(*filePath); file != "" {
			if data, err := ioutil.ReadFile(file); err == nil {
				allowedMembers = string(data)
				EnableFileWatch(filePath)
				return true
			}
		}
	}
	return false
}

func DisableFileWatch() {
	if configWatcher != nil {
		close(configWatcher.close)
		<-configWatcher.closed
		configWatcher = nil
	}
}

func EnableFileWatch(cfgFileName *string) {
	if configWatcher == nil {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			mlog.Info(fmt.Sprintf("failed to create config watcher for file: %v", *cfgFileName))
			return
		}

		configFile := fileutils.FindConfigFile(filepath.Clean(*cfgFileName))
		configDir, _ := filepath.Split(configFile)

		watcher.Add(configDir)

		ret := &ConfigWatcher{
			watcher: watcher,
			close:   make(chan struct{}),
			closed:  make(chan struct{}),
		}

		go func() {
			defer close(ret.closed)
			defer watcher.Close()
			for {
				select {
				case event := <-watcher.Events:
					// we only care about the config file
					if filepath.Clean(event.Name) == configFile {
						if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
							LoadAllowedMembers(cfgFileName)
						}
					}
				case err := <-watcher.Errors:
					mlog.Error(fmt.Sprintf("Failed while watching config file at %v with err=%v", *cfgFileName, err.Error()))
				case <-ret.close:
					return
				}
			}
		}()

		configWatcher = ret
	}
}
