package app

import (
	"fmt"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
	"strconv"
	"time"
)

func (a *App) GetOooRequestUser(userId string) (*model.OooUser, *model.AppError) {
	result := <-a.Srv.Store.OooRequestUser().Get(userId)
	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.OooUser), nil
}

func (a *App) InsertOooRequestUser(userId string, createAt, deleteAt int64, requestNotifyProps, timezone model.StringMap) *model.AppError {
	newUser := &model.OooUser{userId, createAt, deleteAt, requestNotifyProps, timezone}
	result := <-a.Srv.Store.OooRequestUser().Save(newUser)
	if result.Err != nil {
		return result.Err
	}

	return nil
}

func (a *App) UserExists(userId string) (bool, *model.AppError) {
	result := <-a.Srv.Store.OooRequestUser().Get(userId)
	if result.Err != nil {
		if result.Err.Id == store.MISSING_ACCOUNT_ERROR {
			return false, nil
		}

		return false, result.Err
	}

	if result.Data == nil {
		return false, nil
	}
	return true, nil
}

func (a *App) Update(userId string, createAt, deleteAt int64, requestNotifyProps model.StringMap) *model.AppError {
	result := <-a.Srv.Store.OooRequestUser().Update(userId, createAt, deleteAt, requestNotifyProps)
	if result.Err != nil {
		return result.Err
	}

	return nil
}

func (a *App) UpdateOooRequestUser(userId string, user *model.User, delayedUpdate bool) *model.AppError {
	userExists, err := a.UserExists(userId)
	if err != nil {
		return err
	}
	status, err := a.GetStatus(userId)
	if err != nil {
		return err
	}

	var fromDate time.Time
	if user.NotifyProps["fromDate"] == "" {
		fromDate = time.Now()
	} else {
		fromDate, _ = time.Parse("2006-1-2", user.NotifyProps["fromDate"])
	}

	fromTime := user.NotifyProps["fromTime"]
	toTime := user.NotifyProps["toTime"]

	var toDate time.Time
	if user.NotifyProps["toDate"] == "" {
		toDate = fromDate.AddDate(200, 0, 0)
	} else {
		toDate, _ = time.Parse("2006-1-2", user.NotifyProps["toDate"])
	}

	offset, _ := strconv.Atoi(user.NotifyProps["offset"])

	startDateMillis := model.GetStartOfDayMillis(fromDate, offset)
	if fromTime != "" {
		startDateMillis = AddTimeMillis(fromTime, startDateMillis)
	}

	endDateMillis := model.GetStartOfDayMillis(toDate, offset)
	if toTime != "" {
		endDateMillis = AddTimeMillis(toTime, endDateMillis)
	} else {
		endDateMillis = model.GetEndOfDayMillis(toDate, offset)
	}

	if userExists {
		if delayedUpdate && status.Status == model.STATUS_OUT_OF_OFFICE {
			a.SetStatusOnline(userId, true)
			a.DisableAutoResponder(userId, false)
		}
		err := a.Update(userId, startDateMillis, endDateMillis, user.NotifyProps)
		if err != nil {
			return err
		}
		return nil
	}

	err = a.InsertOooRequestUser(userId, startDateMillis, endDateMillis, user.NotifyProps, user.Timezone)
	if err != nil {
		return err
	}
	return nil
}

func AddTimeMillis(timeString string, dateMillis int64) int64 {
	fTime, _ := time.Parse("3:04 PM", timeString)
	h := fTime.Hour()
	m := fTime.Minute()
	dateMillis = dateMillis + (int64(time.Hour / time.Millisecond))*int64(h)
	dateMillis = dateMillis + (int64(time.Minute / time.Millisecond))*int64(m)
	return dateMillis
}

func (a *App) DoOutOfOfficeRequestHandle() {

	if !*a.Srv.Config().TeamSettings.ExperimentalEnableAutomaticReplies {
		return
	}

	mlog.Debug("Handling Out Of Office Request")

	t := time.Now().UTC()
	time := model.GetMillisForTime(t)

	result := <-a.Srv.Store.OooRequestUser().GetAllExpiredBefore(time)
	users := result.Data.([]*model.OooUser)
	for _, user := range users {
		a.SetStatusOnline(user.UserId, true)
		a.DisableAutoResponder(user.UserId, false)
	}

	result = <-a.Srv.Store.OooRequestUser().PermanentDeleteBefore(time)
	if result.Err != nil {
		mlog.Error("Unable to delete the users, error: " + result.Err.Error())
	}

	result = <-a.Srv.Store.OooRequestUser().GetAllBefore(time)
	users = result.Data.([]*model.OooUser)
	for _, user := range users {
		newUser, err := a.GetUser(user.UserId)
		if err == nil {
			if newUser.NotifyProps[model.AUTO_RESPONDER_ACTIVE_NOTIFY_PROP] == "true" {
				a.SetStatusOutOfOffice(user.UserId)
			} else {
				result := <-a.Srv.Store.OooRequestUser().PermanentDelete(user.UserId)
				if result.Err != nil {
					mlog.Error(fmt.Sprintf("Failed to delete OOO status for user_id=%v, err=%v", user.UserId, result.Err), mlog.String("user_id", user.UserId))
				}
			}
		}
	}
	mlog.Debug("Out Of Office Requests Processed")
}
