// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

var allowedMembers = ""
var configWatcher *watcher

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
		if file, err := resolveConfigFilePath(*filePath); err == nil {
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
		fsWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			mlog.Info(fmt.Sprintf("failed to create config watcher for file: %v", *cfgFileName))
			return
		}

		configFile, err := resolveConfigFilePath(filepath.Clean(*cfgFileName))
		if err != nil {
			mlog.Info(fmt.Sprintf("failed to clean and resolve config file: %v", *cfgFileName))
			return
		}

		configDir, _ := filepath.Split(configFile)
		if err := fsWatcher.Add(configDir); err != nil {
			if closeErr := fsWatcher.Close(); closeErr != nil {
				mlog.Error("failed to stop fsnotify watcher for %s", mlog.String("path", *cfgFileName), mlog.Err(closeErr))
			}
			mlog.Info(fmt.Sprintf("failed to create config watcher for file: %v", configFile))
			return
		}

		ret := &watcher{
			fsWatcher: fsWatcher,
			close:     make(chan struct{}),
			closed:    make(chan struct{}),
		}

		go func() {
			defer close(ret.closed)
			defer func() {
				if err := fsWatcher.Close(); err != nil {
					mlog.Error("failed to stop fsnotify watcher for %s", mlog.String("path", configFile))
				}
			}()

			for {
				select {
				case event := <-fsWatcher.Events:
					// We only care about the given file.
					if filepath.Clean(event.Name) == configFile {
						if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
							mlog.Info("Config file watcher detected a change", mlog.String("path", configFile))
							LoadAllowedMembers(cfgFileName)
						}
					}
				case err := <-fsWatcher.Errors:
					mlog.Error("Failed while watching config file", mlog.String("path", configFile), mlog.Err(err))
				case <-ret.close:
					return
				}
			}
		}()
		configWatcher = ret
	}
}
