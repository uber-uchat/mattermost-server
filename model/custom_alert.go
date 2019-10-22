// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type CustomAlertRequest struct {
	UserIds               *StringArray     `json:"user_ids"`
	Props                 *StringInterface `json:"props"`
	SendEmailNotification *bool            `json:"send_email_notification"`
	SendPushNotification  *bool            `json:"send_push_notification"`
	PostId                *string          `json:"post_id"`
	CustomMessage         *string          `json:"custom_message"`
	NotifyAllActiveUsers  *bool            `json:"notify_all_active_users"`
}

func CustomAlertRequestFromJson(data io.Reader) *CustomAlertRequest {
	var o *CustomAlertRequest
	json.NewDecoder(data).Decode(&o)
	return o
}
