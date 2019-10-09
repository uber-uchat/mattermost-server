package api4

import (
	"fmt"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"net/http"
)

func (api *API) InitCustomAlerts() {

	api.BaseRoutes.ApiRoot.Handle("/custom_alerts", api.ApiSessionRequired(sendCustomAlertNotification)).Methods("POST")
}

func sendCustomAlertNotification(c *Context, w http.ResponseWriter, r *http.Request) {

	if !c.App.SessionHasPermissionToSendAlerts(c.App.Session) {
		c.SetPermissionError(model.PERMISSION_MANAGE_JOBS)
		mlog.Error(fmt.Sprintf("UserId: %v does not have permission to send alerts", c.App.Session.UserId))
		return
	}

	if *c.App.Config().EmailSettings.SendPushNotifications {
		pushServer := *c.App.Config().EmailSettings.PushNotificationServer
		if license := c.App.License(); pushServer == model.MHPNS && (license == nil || !*license.Features.MHPNS) {
			mlog.Warn("Push notifications are disabled. Go to System Console > Notifications > Mobile Push to enable them.")
			return
		}
	}

	notificationReq := model.CustomAlertRequestFromJson(r.Body)
	if notificationReq == nil {
		return
	}

	alertReqsChannel := c.App.GetChannelForCustomPushAlerts()
	alertReqsChannel <- *notificationReq

	mlog.Info("Custom alert request has been submitted to the worker pool")

	ReturnStatusOK(w)
}
