// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package sqlstore

import (
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
	"net/http"
)

type SqlOooRequestStore struct {
	SqlStore
}

func NewSqlOooRequestStore(sqlStore SqlStore) store.OooRequestStore {
	ors := &SqlOooRequestStore{SqlStore: sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.OooUser{}, "OooRequestUsers").SetKeys(false, "UserId")
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("Timezone").SetMaxSize(2000)
		table.ColMap("RequestNotifyProps").SetMaxSize(2000)
	}

	return ors
}

func (us SqlOooRequestStore) CreateIndexesIfNotExists() {
	us.CreateIndexIfNotExists("idx_ooo_request_create_at", "OooRequestUsers", "CreateAt")
	us.CreateIndexIfNotExists("idx_ooo_request_delete_at", "OooRequestUsers", "DeleteAt")
	us.CreateIndexIfNotExists("idx_ooo_request_user_id", "OooRequestUsers", "UserId")
}

func (s SqlOooRequestStore) Save(user *model.OooUser) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		user.PreSave()

		if err := s.GetMaster().Insert(user); err != nil {
			result.Err = model.NewAppError("SqlOooRequestStore.Save", "store.sql_ooo_user.save.app_error", nil, "user_id="+user.UserId+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = user
		}
	})
}

func (us SqlOooRequestStore) Update(userId string, createAt, deleteAt int64, requestNotifyProps model.StringMap) store.StoreChannel {

	return store.Do(func(result *store.StoreResult) {

		if _, err := us.GetMaster().Exec("UPDATE OooRequestUsers SET CreateAt = :CreateAt, DeleteAt = :DeleteAt, RequestNotifyProps = :RequestNotifyProps WHERE UserId = :UserId", map[string]interface{}{"CreateAt": createAt, "DeleteAt": deleteAt, "RequestNotifyProps": model.MapToJson(requestNotifyProps), "UserId": userId}); err != nil {
			fmt.Println("Error Found : " + err.Error())
			result.Err = model.NewAppError("SqlOooRequestStore.Update", "store.sql_ooo_user.update.app_error", nil, "id="+userId+", "+err.Error(), http.StatusInternalServerError)
		}
	})
}

func (us SqlOooRequestStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var count int
		err := us.GetMaster().SelectOne(&count, "Select COUNT(*) from OooRequestUsers")
		if err != nil {
			result.Err = model.NewAppError("SqlOooRequestStore.Get", store.MISSING_ACCOUNT_ERROR, nil, "user_id="+id, http.StatusNotFound)
			return
		}
		if count == 0 {
			result.Data = nil
			return
		}

		if user, err := us.GetMaster().Get(model.OooUser{}, id); err != nil {
			result.Err = model.NewAppError("SqlOooRequestStore.Get", store.MISSING_ACCOUNT_ERROR, nil, "user_id="+id, http.StatusNotFound)
		} else {
			if user == nil {
				result.Err = model.NewAppError("SqlOooRequestStore.Get", store.MISSING_ACCOUNT_ERROR, nil, "user_id="+id, http.StatusInternalServerError)
			} else {
				result.Data = user.(*model.OooUser)
			}
		}
	})
}

func (us SqlOooRequestStore) GetAllBefore(time int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var users []*model.OooUser

		query := "Select * FROM OooRequestUsers WHERE CreateAt <= :Time ORDER BY CreateAt ASC"
		if _, err := us.GetReplica().Select(&users, query, map[string]interface{}{"Time": time}); err != nil {
			result.Err = model.NewAppError("SqlOooRequestStore.GetAllBefore", "store.sql_ooo_user.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
		result.Data = users
	})
}

func (us SqlOooRequestStore) GetAllExpiredBefore(time int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var users []*model.OooUser

		query := "Select * FROM OooRequestUsers WHERE DeleteAt <= :Time ORDER BY DeleteAt ASC"
		if _, err := us.GetReplica().Select(&users, query, map[string]interface{}{"Time": time}); err != nil {
			result.Err = model.NewAppError("SqlOooRequestStore.GetAllBefore", "store.sql_ooo_user.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
		result.Data = users
	})
}

func (us SqlOooRequestStore) PermanentDelete(userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := us.GetMaster().Exec("DELETE FROM OooRequestUsers WHERE UserId = :UserId", map[string]interface{}{"UserId": userId}); err != nil {
			result.Err = model.NewAppError("SqlOooRequestStore.PermanentDelete", "store.sql_ooo_user.permanent_delete.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (us SqlOooRequestStore) PermanentDeleteBefore(time int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := us.GetMaster().Exec("DELETE FROM OooRequestUsers WHERE DeleteAt < :Time AND DeleteAt != 0", map[string]interface{}{"Time": time}); err != nil {
			result.Err = model.NewAppError("SqlOooRequestStore.PermanentDelete", "store.sql_ooo_user.permanent_delete.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	})
}
