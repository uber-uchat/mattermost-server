// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
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
	var pm = &ClientMetrics{
		mobileMetrics: nil,
	}

	histogramMetrics := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mobile_metrics_seconds",
			Help:    "Mobile metrics",
			Buckets: prometheus.ExponentialBuckets(0.01, 1.3, 36),
		},
		[]string{"event", "country", "app_version", "server_version", "model"},
	)
	prometheus.MustRegister(histogramMetrics)

	defer func() {
		if r := recover(); r != nil {
			mlog.Error(fmt.Sprintf("Recovering from ClientMetrics creation panic. Panic was: %v", r))
			pm.mobileMetrics = nil
		}
	}()

	pm.mobileMetrics = histogramMetrics

	return pm
}

func (c *ClientMetrics) CollectMobileMetrics(metrics *model.MobileMetrics) error {
	if metrics == nil {
		mlog.Error("Mobile metric is nil.")
		return nil
	}

	if c.mobileMetrics == nil {
		mlog.Error("Prometheus metric collector is nil.")
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
