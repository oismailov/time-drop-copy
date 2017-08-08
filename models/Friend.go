package models

import "errors"

//Friend handels friend (due to a gorm bug)
type Friend struct {
	BaseModel

	RequesterRefer uint `gorm:"unique_index:idx_friend_ids"`
	ReceiverRefer  uint `gorm:"unique_index:idx_friend_ids"`
}

//FindByUserID find friends by a user id
func (friend *Friend) FindByUserID(userID interface{}) (friends []Friend, err error) {
	db := GetDatabaseSession()
	result := db.Where("requester_refer = ? OR receiver_refer = ?", userID, userID).Find(&friends)
	return friends, result.Error
}

//Delete (unfriends) a friendship
func (friend *Friend) Delete(userID, friendID interface{}) {
	db := GetDatabaseSession()

	var game Game
	game.UnfriendCleanUp(userID, friendID)
	db.Exec("DELETE FROM friends WHERE (requester_refer = ? OR receiver_refer = ?) AND (requester_refer = ? OR receiver_refer = ?)",
		userID, userID, friendID, friendID)
}

//IsAlreadyFriendsWith checks if a friendship would be a duplicate
//e.g user1 -> user2 and user2 -> user1
func (friend *Friend) IsAlreadyFriendsWith(userID, friendID interface{}) bool {
	db := GetDatabaseSession()
	sqlQuery := "(requester_refer = ? OR receiver_refer = ?) AND (requester_refer = ? OR receiver_refer = ?)"

	var count int
	db.Model(&friend).Where(sqlQuery, userID, userID, friendID, friendID).Count(&count)
	return count > 0
}

//FriendRequest handels friend requests
type FriendRequest struct {
	BaseModel

	RequesterRefer uint `gorm:"unique_index:idx_freq_ids"`
	ReceiverRefer  uint `gorm:"unique_index:idx_freq_ids"`
}

//Save friend request
func (friendRequest *FriendRequest) Save() error {
	db := GetDatabaseSession()
	result := db.Save(&friendRequest)
	return result.Error
}

//FindOpenFriendRequestsByUserID friend request
func (friendRequest *FriendRequest) FindOpenFriendRequestsByUserID(userID interface{}) (friendRequests []FriendRequest, err error) {
	db := GetDatabaseSession()
	result := db.Where("receiver_refer = ?", userID).Find(&friendRequests)
	return friendRequests, result.Error
}

//FindPendingFriendRequestsByUserID friend request
func (friendRequest *FriendRequest) FindPendingFriendRequestsByUserID(userID interface{}) (friendRequests []FriendRequest, err error) {
	db := GetDatabaseSession()
	result := db.Where("requester_refer = ?", userID).Find(&friendRequests)
	return friendRequests, result.Error
}

//FindAndAcceptFriendRequest handels friend creation
func (friendRequest FriendRequest) FindAndAcceptFriendRequest(friendRequestID, receiverID interface{}) bool {
	db := GetDatabaseSession()
	result := db.First(&friendRequest, friendRequestID)
	if result.Error != nil {
		return false
	}
	if friendRequest.ID == 0 {
		return false
	}

	if friendRequest.ReceiverRefer != receiverID {
		return false
	}

	var friend Friend
	if areFriends := friend.IsAlreadyFriendsWith(friendRequest.ReceiverRefer, friendRequest.RequesterRefer); areFriends {
		return false
	}

	var friendUser User
	friendUser.FindByID(friendRequest.RequesterRefer)
	if friendUser.ID == 0 {
		return false
	}

	friend = Friend{
		ReceiverRefer:  friendRequest.ReceiverRefer,
		RequesterRefer: friendRequest.RequesterRefer,
	}
	result = db.Save(&friend)
	if result.Error != nil {
		return false
	}

	var pushNotification PushNotification
	go pushNotification.SendFriendRequestAcceptedPush(friendUser)

	// If user a sent a friend request to user b then the
	// request for user b -> user a should also be deleted
	sqlQuery := "requester_refer = ? AND receiver_refer = ?"
	db.Where(sqlQuery, friendRequest.ReceiverRefer, friendRequest.RequesterRefer).Delete(FriendRequest{})

	// Delete actual friend request
	db.Unscoped().Delete(&friendRequest)

	return true
}

//FindAndDecclineFriendRequest handels friend request denial
func (friendRequest FriendRequest) FindAndDecclineFriendRequest(friendRequestID interface{}, currentUserID uint) error {
	db := GetDatabaseSession()

	if result := db.First(&friendRequest, friendRequestID); result.Error != nil {
		return errors.New("friend_request_not_found")
	}

	if friendRequest.ReceiverRefer != currentUserID {
		return errors.New("friend_request_not_receiver")
	}

	if result := db.Unscoped().Delete(&friendRequest); result.Error != nil {
		return result.Error
	}

	return nil
}
