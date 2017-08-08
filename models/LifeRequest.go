package models

//LifeRequest struct handels life_requests
type LifeRequest struct {
	BaseModel

	RequesterRefer string `json:"requesterRefer"`
	ReceiverRefer  string `json:"receiverRefer"`
}

//AddUserFriends
func (lr *LifeRequest) CreateLifeRecords(users []uint, userId uint) {
	db := GetDatabaseSession()
	refers := make(map[uint]uint)
	rows, _ := db.Raw(`
    SELECT receiver_refer
    FROM life_requests
    WHERE receiver_refer in (?)
		AND requester_refer = ? AND approved = 'false' AND  collected = 'false'`, users, userId).Rows()
	defer rows.Close()
	for rows.Next() {
		var user uint
		rows.Scan(&user)
		refers[user] = user
	}

	var requesterUserModel User
	requesterUserModel.FindByID(userId)

	for _, user := range users {
		if refers[user] == 0 {
			var push PushNotification
			var userModel User
			userModel.FindByID(user)
			go push.SendLifeRequestPush(userModel, requesterUserModel.Username)
			db.Exec(`
      INSERT IGNORE INTO
        life_requests (requester_refer, receiver_refer, created_at)
      VALUES (?, ?, NOW())`, userId, user)
		}
	}
}

//AddUserFriends
func (lr *LifeRequest) GetLifeRequests(userId uint) []int {
	db := GetDatabaseSession()
	userIds := []int{}
	rows, _ := db.Raw(`SELECT requester_refer FROM life_requests WHERE receiver_refer = ? AND approved = 'false' AND collected = 'false'`, userId).Rows()
	defer rows.Close()
	for rows.Next() {
		var user int
		rows.Scan(&user)
		userIds = append(userIds, user)
	}
	return userIds
}

//GetAppovedRequests
func (lr *LifeRequest) GetAppovedRequests(userId uint) []int {
	db := GetDatabaseSession()
	userIds := []int{}
	rows, _ := db.Raw(`SELECT id, receiver_refer FROM life_requests WHERE requester_refer = ? AND approved = 'true' AND collected = 'false'`, userId).Rows()
	defer rows.Close()

	for rows.Next() {
		var user int
		var id int
		rows.Scan(&id, &user)
		userIds = append(userIds, user)
		db.Exec(`UPDATE life_requests SET collected = 'true' WHERE id = ?`, id)
	}

	return userIds
}

//GiveLife
func (lr *LifeRequest) GiveLife(userIds []uint, receiverId uint) {
	db := GetDatabaseSession()
	for _, user := range userIds {
		var push PushNotification
		var userModel User
		userModel.FindByID(user)
		go push.GiveLifeRequestPush(userModel)
	}

	db.Exec("UPDATE life_requests SET approved = 'true' WHERE requester_refer IN (?) AND receiver_refer = ? ", userIds, receiverId)
}

//CreateInstallRequest
func (lr *LifeRequest) CreateInstallRequest(guid string, userId uint) {
	db := GetDatabaseSession()
	var requesterRefer uint
	sql := "SELECT requester_refer FROM install_requests WHERE guid = ? AND requester_refer = ?"
	db.Raw(sql, guid, userId).Row().Scan(&requesterRefer)
	if requesterRefer == 0 {
		db.Exec(`INSERT INTO install_requests (requester_refer, guid, created_at) VALUES (?, ?, NOW())`, userId, guid)
	}
}

//CreateInstallFromRequest
func (lr *LifeRequest) CreateInstallFromRequest(guid string, userId uint) {
	db := GetDatabaseSession()
	var requesterRefer uint
	sql := "SELECT requester_refer FROM install_requests WHERE guid = ? AND requester_refer != ?"
	db.Raw(sql, guid, userId).Row().Scan(&requesterRefer)
	if requesterRefer != 0 {
		db.Exec(`INSERT INTO install_responses (receiver_refer, requester_refer, guid, created_at) VALUES (?, ?, ?, NOW())`, userId, requesterRefer, guid)
	}
}

//ShowInstallRequests
func (lr *LifeRequest) ShowInstallRequests(userId uint) []uint {
	db := GetDatabaseSession()
	sql := "SELECT receiver_refer FROM install_responses WHERE requester_refer = ? AND collected = 'false'"
	rows, _ := db.Raw(sql, userId).Rows()
	var userIds []uint
	for rows.Next() {
		var user uint
		rows.Scan(&user)
		userIds = append(userIds, user)
		db.Exec("UPDATE install_responses SET collected = 'true' WHERE requester_refer = ? AND  receiver_refer = ?", userId, user)
	}

	return userIds
}

//CleanUp
func (lr LifeRequest) CleanUp() {
	db := GetDatabaseSession()
	sql := `SELECT id FROM life_requests WHERE approved = 'false' AND created_at <= now() - INTERVAL 1 DAY`
	rows, _ := db.Raw(sql).Rows()
	var requests []int
	for rows.Next() {
		var requestId int
		rows.Scan(&requestId)
		requests = append(requests, requestId)
	}
	delete := `DELETE FROM life_requests WHERE id IN (?)`
	db.Exec(delete, requests)

}
