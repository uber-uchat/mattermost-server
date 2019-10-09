package app

import (
	"fmt"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

type CustomAlertsJob struct {
	Channel chan model.CustomAlertRequest
}

func (s *Server) InitCustomAlerts() {
	if *s.Config().EmailSettings.EnableCustomAlerts {

		alertReqsChannel := make(chan model.CustomAlertRequest)
		customAlertsJob := CustomAlertsJob{
			Channel: alertReqsChannel,
		}
		s.FakeApp().StartCustomAlertsWorker(alertReqsChannel)
		s.CustomAlertsJob = customAlertsJob
	}
}

func (a *App) StartCustomAlertsWorker(alertRequests chan model.CustomAlertRequest) {
	mlog.Debug("Starting the worker for custom push alerts")
	a.Srv.Go(func() { a.customAlertsWorker(alertRequests) })
}

func (a *App) GetChannelForCustomPushAlerts() chan model.CustomAlertRequest {
	return a.Srv.CustomAlertsJob.Channel;
}

func (s *Server) StopCustomAlertsJob() {
	channel := s.CustomAlertsJob.Channel
	close(channel)
	mlog.Debug("Stopped the worker for custom push alerts")
}

func (a *App) customAlertsWorker(customAlertRequests chan model.CustomAlertRequest) {
	mlog.Debug("Worker initialized for custom alerts job")

	for alertReq := range customAlertRequests {
		mlog.Info("processing new custom alert request")
		postInfo, channelInfo := a.validateNewAlertRequest(alertReq)

		if !*alertReq.SendPushNotification || postInfo == nil || channelInfo == nil {
			continue
		}

		a.processAlert(alertReq, postInfo, channelInfo)
	}
}

func (a *App) validateNewAlertRequest(alertReq model.CustomAlertRequest) (*model.Post, *model.Channel) {

	if alertReq.PostId == nil || *alertReq.PostId == "" {
		mlog.Error("Invalid Request, Mandatory fields are missing - 'postId' ")
		return nil, nil
	}

	postInfo, err := a.GetSinglePost(*alertReq.PostId)
	if err != nil {
		mlog.Error(fmt.Sprintf("No valid Post data found for the postId: %v", *alertReq.PostId))
		return nil, nil
	}

	channelInfo, err := a.GetChannel(postInfo.ChannelId)
	if err != nil {
		mlog.Error(fmt.Sprintf("No valid Channel data found for the channelId: %v", postInfo.ChannelId))
		return postInfo, nil
	}

	if channelInfo.Type != model.CHANNEL_OPEN {
		mlog.Error(fmt.Sprintf("Channel with Id: '%v', displayName: '%v' associated with the postId: %v is not open for all ",
			channelInfo.Name, channelInfo.DisplayName, *alertReq.PostId))
		return postInfo, nil
	}

	return postInfo, channelInfo
}

func (a *App) processAlert(alertRequest model.CustomAlertRequest, postInfo *model.Post, channelInfo *model.Channel) {

	if *alertRequest.NotifyAllActiveUsers {

		limit := 500
		offset := 0

		for userSessions, err := a.GetAllSessionsWithActiveDeviceIds(limit, offset); err == nil && len(userSessions) > 0; {
			mlog.Info(fmt.Sprintf("processing %v active device sessions in current batch", len(userSessions)))
			a.sendPushNotificationForActiveSessions(userSessions, alertRequest, channelInfo, postInfo)
			offset = offset + limit
		}

	} else {

		for _, userId := range *alertRequest.UserIds {
			userSessions, err := a.GetMobileAppSessions(userId)
			if err != nil {
				mlog.Error("Unable to get mobile sessions for userId:"+userId, mlog.Err(err))
				continue
			}

			a.sendPushNotificationForActiveSessions(userSessions, alertRequest, channelInfo, postInfo)

		}
	}

}

func (a *App) sendPushNotificationForActiveSessions(userSessions []*model.Session, notificationReq model.CustomAlertRequest,
	channelInfo *model.Channel, postInfo *model.Post) {

	for _, session := range userSessions {
		mlog.Info(fmt.Sprintf("processing the active session for userId: %v", session.UserId))

		if session.IsExpired() {
			continue
		}

		deviceId := session.DeviceId
		userInfo, _ := a.GetUser(session.UserId)

		if deviceId != "" && *notificationReq.SendPushNotification {
			a.SendCustomAlertToPushNotificationHub(postInfo, userInfo, channelInfo, notificationReq.CustomMessage, session)
		}

	}
}
