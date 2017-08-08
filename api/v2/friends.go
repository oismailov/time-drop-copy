package v2

import (
	"encoding/json"
	"net/http"

	"timedrop/api"
	"timedrop/helpers"
	"timedrop/middlewares"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func InitFriend(r *mux.Router) {
	l4g.Debug("Initializing v2 friends api routes")
	friendsController := FriendsCtrl{}
	sr := r.PathPrefix("/friends").Subrouter()
	sr.Handle("/", api.ApiTokenRequired(friendsController.List)).Methods("GET")
	sr.Handle("/", api.ApiTokenRequired(friendsController.AddFriend)).Methods("POST")
	sr.Handle("/{friendID:[0-9]+}", api.ApiTokenRequired(friendsController.RemoveFriend)).Methods("DELETE")
	sr.Handle("/request", api.ApiTokenRequired(friendsController.SendFriendRequest)).Methods("POST")
	sr.Handle("/request", api.ApiTokenRequired(friendsController.ListFriendRequest)).Methods("GET")
	sr.Handle("/request/{friendID:[0-9]+}", api.ApiTokenRequired(friendsController.DeclineFriendRequest)).Methods("DELETE")
	sr.Handle("/updateFriends", api.ApiTokenRequired(friendsController.UpdateFriends)).Methods("POST")
	sr2 := r.PathPrefix("/").Subrouter()
	sr2.Handle("/updateFriends", api.ApiTokenRequired(friendsController.UpdateFacebookFriends)).Methods("POST")
}

//FriendsCtrl is the controller for /friends
type FriendsCtrl struct{}

//List handels /friends
func (friendsCtrl FriendsCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	user, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var friend models.Friend
	friends, _ := friend.FindByUserID(user.ID)

	// var parsedFriends []models.User
	parsedFriends := []models.User{}
	friendIds := make(map[uint]uint, 0)
	for _, friend := range friends {
		var friendID uint
		friendID = friend.RequesterRefer
		if friend.ReceiverRefer != user.ID {
			friendID = friend.ReceiverRefer
		}

		if len(friendIds) > 0 {
			if ok := friendIds[friendID]; ok == 0 {
				friendIds[friendID] = friendID
				var tmpUser models.User
				tmpUser.FindByID(friendID)

				if tmpUser.ID != 0 {
					tmpUser.Email = ""
					parsedFriends = append(parsedFriends, tmpUser)
				}
			}
		} else {
			friendIds[friendID] = friendID
			var tmpUser models.User
			tmpUser.FindByID(friendID)

			if tmpUser.ID != 0 {
				tmpUser.Email = ""
				parsedFriends = append(parsedFriends, tmpUser)
			}
		}

	}

	r.JSON(res, 200, parsedFriends)

	return
}

type friendRequestOpenReponse struct {
	ID        uint        `json:"id"`
	Requester models.User `json:"requester"`
}

type friendRequestPendingReponse struct {
	ID       uint        `json:"id"`
	Receiver models.User `json:"receiver"`
}

//ListFriendRequest handels /friends/request (GET)
func (friendsCtrl FriendsCtrl) ListFriendRequest(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	user, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var friendRequest models.FriendRequest
	friendRequests, _ := friendRequest.FindOpenFriendRequestsByUserID(user.ID)

	var parsedFriendRequests []friendRequestOpenReponse
	for _, fReq := range friendRequests {
		var tmpUser models.User
		tmpUser.FindByID(fReq.RequesterRefer)

		if tmpUser.ID != 0 {
			tmpUser.Email = ""
			frOpenRes := friendRequestOpenReponse{
				ID:        fReq.ID,
				Requester: tmpUser,
			}
			parsedFriendRequests = append(parsedFriendRequests, frOpenRes)
		}
	}

	pendingFriendRequests, _ := friendRequest.FindPendingFriendRequestsByUserID(user.ID)

	var parsedPendingFriendRequests []friendRequestPendingReponse
	for _, fReq := range pendingFriendRequests {
		var tmpUser models.User
		tmpUser.FindByID(fReq.RequesterRefer)

		if tmpUser.ID != 0 {
			tmpUser.Email = ""
			frPendingRes := friendRequestPendingReponse{
				ID:       fReq.ID,
				Receiver: tmpUser,
			}
			parsedPendingFriendRequests = append(parsedPendingFriendRequests, frPendingRes)
		}
	}

	jsonResMap := make(map[string]interface{})
	jsonResMap["pending"] = parsedPendingFriendRequests
	jsonResMap["open"] = parsedFriendRequests

	r.JSON(res, 200, jsonResMap)
	return
}

type friendRequestRequestData struct {
	FriendID string `json:"friendId" valid:"required"`
}

//SendFriendRequest handels /friends/request
func (friendsCtrl FriendsCtrl) SendFriendRequest(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var friendRequestRequest friendRequestRequestData
	if err := decoder.Decode(&friendRequestRequest); err != nil {
		panic(err)
	}

	// validate login request
	_, err := govalidator.ValidateStruct(friendRequestRequest)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var receiver models.User
	receiver.FindByID(friendRequestRequest.FriendID)
	if receiver.ID == 0 {
		r.JSON(res, 404, helpers.GenerateErrorResponse("receiver_not_found", req.Header))
		return
	}

	// requester, err := middlewares.GetUserFromContext(res, req)
	requester := req.Context().Value("user").(models.User)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var friend models.Friend
	if areFriends := friend.IsAlreadyFriendsWith(requester.ID, receiver.ID); areFriends {
		r.JSON(res, 422, helpers.GenerateErrorResponse("already_friends", req.Header))
		return
	}

	if receiver.ID == requester.ID {
		r.JSON(res, 422, helpers.GenerateErrorResponse("can_not_friend_yourself", req.Header))
		return
	}

	friendRequest := models.FriendRequest{
		RequesterRefer: requester.ID,
		ReceiverRefer:  receiver.ID,
	}
	if saveErr := friendRequest.Save(); saveErr != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse("friend_request_pending", req.Header))
		return
	}

	var pushNotification models.PushNotification
	go pushNotification.SendFriendRequestPush(receiver)

	r.Text(res, 201, "")
	return
}

//DeclineFriendRequest handels /friends/request/{friendRequestID}
func (friendsCtrl FriendsCtrl) DeclineFriendRequest(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	friendRequestID := vars["friendRequestID"]

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var friendRequest models.FriendRequest
	friendRequest.FindAndDecclineFriendRequest(friendRequestID, currentUser.ID)

	r.Text(res, 201, "")
	return
}

type addFriendRequestData struct {
	FriendRequestID string `json:"friendRequestId" valid:"required"`
}

//AddFriend handels /friends (POST)
func (friendsCtrl FriendsCtrl) AddFriend(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var addFriendRequest addFriendRequestData
	if err := decoder.Decode(&addFriendRequest); err != nil {
		panic(err)
	}

	// validate login request
	_, err := govalidator.ValidateStruct(addFriendRequest)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	receiver, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var friendRequest models.FriendRequest
	if isAdded := friendRequest.FindAndAcceptFriendRequest(addFriendRequest.FriendRequestID, receiver.ID); !isAdded {
		r.JSON(res, 422, helpers.GenerateErrorResponse("invalid_friend_request", req.Header))
		return
	}
	var friend models.Friend
	friends, _ := friend.FindByUserID(receiver.ID)

	r.JSON(res, 201, friends)
	return
}

//RemoveFriend handels /friends (DELETE)
func (friendsCtrl FriendsCtrl) RemoveFriend(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	friendID := vars["friendID"]

	owner, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var friend models.Friend
	friend.Delete(owner.ID, friendID)

	r.Text(res, 204, "")
	return
}

type updateFriendRequest struct {
	Friends []uint `json:"friends"`
}

// v2/updateFriends
func (friendsCtrl FriendsCtrl) UpdateFriends(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var data updateFriendRequest
	if err := decoder.Decode(&data); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	user := req.Context().Value("user").(models.User)

	var friend models.Friend
	friends, err := friend.FindByUserID(user.ID)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var duplicate bool
	for _, dataFriend := range data.Friends {
		duplicate = false
		for _, userFriend := range friends {
			if dataFriend == userFriend.RequesterRefer || dataFriend == userFriend.ReceiverRefer {
				duplicate = true
				break
			}
		}

		if !duplicate {
			var friendRequest models.FriendRequest
			if isAdded := friendRequest.FindAndAcceptFriendRequest(dataFriend, user.ID); !isAdded {
				r.JSON(res, 422, helpers.GenerateErrorResponse("invalid_friend_request", req.Header))
				return
			}
		}
	}
}

// POST v2/updateFriends
func (friendsCtrl FriendsCtrl) UpdateFacebookFriends(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var data updateFriendRequest
	if err := decoder.Decode(&data); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	user := req.Context().Value("user").(models.User)

	var users models.User
	userIds := users.FindAllByFacebookIDs(data.Friends, user.ID)

	//update user friends
	user.AddUserFriends(userIds, user.ID)

	//show friends
	var friend models.Friend
	friends, _ := friend.FindByUserID(user.ID)

	// var parsedFriends []models.User
	parsedFriends := []models.User{}
	for _, friend := range friends {
		var friendID uint
		friendID = friend.RequesterRefer
		if friend.ReceiverRefer != user.ID {
			friendID = friend.ReceiverRefer
		}

		var tmpUser models.User
		tmpUser.FindByID(friendID)

		if tmpUser.ID != 0 {
			tmpUser.Email = ""
			parsedFriends = append(parsedFriends, tmpUser)
		}
	}

	r.JSON(res, 200, parsedFriends)

}
