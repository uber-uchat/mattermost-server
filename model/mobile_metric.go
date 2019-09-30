// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"time"
)

type DeviceDetails struct {
	Country        string `json:"country"`
	APILevel       string `json:"api_level"`
	BuildNumber    string `json:"build_number"`
	BundleId       string `json:"bundle_id"`
	DeviceId       string `json:"device_id"`
	DeviceLocale   string `json:"device_locale"`
	DeviceUniqueId string `json:"device_unique_id"`
	Height         string `json:"height"`
	Width          string `json:"width"`
	IsEmulator     string `json:"is_emulator"`
	IsTablet       string `json:"is_tablet"`
	Manufacturer   string `json:"manufacturer"`
	MaxMemory      string `json:"max_memory"`
	Model          string `json:"model"`
	SystemName     string `json:"system_name"`
	TimeZone       string `json:"timezone"`
	AppVersion     string `json:"app_version"`
	ServerVersion  string `json:"server_version"`
}

type Event struct {
	MetricType string        `json:"name"`
	Category   string        `json:"cat"`
	TimeStamp  int64         `json:"ts"`
	PID        string        `json:"pid"`
	Duration   time.Duration `json:"dur"`
}

type MobileMetrics struct {
	Events        []Event       `json:"trace_events"`
	DeviceDetails DeviceDetails `json:"device_info"`
}

func MobileMetricsFromJson(data io.Reader) *MobileMetrics {
	var me *MobileMetrics
	json.NewDecoder(data).Decode(&me)
	if me != nil {
		return me
	}

	return nil
}

func (metric *MobileMetrics) ToJson() string {
	b, _ := json.Marshal(metric)
	return string(b)
}
