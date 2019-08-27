// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"github.com/mattermost/mattermost-server/services/timezones"
)

type OooUser struct {
	UserId             string    `json:"user_id"`
	CreateAt           int64     `json:"create_at"`
	DeleteAt           int64     `json:"delete_at"`
	RequestNotifyProps StringMap `json:"notify_props,omitempty"`
	Timezone           StringMap `json:"timezone"`
}

// PreSave will set the Id and Username if missing.  It will also fill
// in the CreateAt, UpdateAt times.  It will also hash the password.  It should
// be run before saving the user to the db.
func (u *OooUser) PreSave() {
	//u.CreateAt = GetMillis()

	if u.Timezone == nil {
		u.Timezone = timezones.DefaultUserTimezone()
	}
}

func (u *OooUser) IsValid() *AppError {
	if len(u.UserId) != 26 {
		return InvalidUserError("id", "")
	}

	if u.CreateAt == 0 {
		return InvalidUserError("create_at", u.UserId)
	}

	return nil
}
