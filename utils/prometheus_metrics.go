// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package utils

import (
	"fmt"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/prometheus/client_golang/prometheus"
)

type ClientMetrics struct {
	mobileMetrics *prometheus.HistogramVec
}

func NewClientMetrics() *ClientMetrics {
	histogramMetrics := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mobile_metrics_seconds",
			Help:    "Mobile metrics",
			Buckets: prometheus.ExponentialBuckets(0.01, 1.3, 36),
		},
		[]string{"event", "country", "app_version", "server_version", "model"},
	)
	prometheus.MustRegister(histogramMetrics)
	pm := &ClientMetrics{
		mobileMetrics: histogramMetrics,
	}

	return pm
}

func (c *ClientMetrics) CollectMobileMetrics(metrics *model.MobileMetrics) error {
	if metrics == nil {
		return nil
	}

	for _, event := range metrics.Events {
		c.mobileMetrics.WithLabelValues(
			event.MetricType,
			metrics.DeviceDetails.Country,
			metrics.DeviceDetails.AppVersion,
			metrics.DeviceDetails.ServerVersion,
			metrics.DeviceDetails.Model,
		).Observe(float64(event.Duration))

		mlog.Info(fmt.Sprint("Telemetry details from client: ", metrics.ToJson()))
	}

	return nil
}
