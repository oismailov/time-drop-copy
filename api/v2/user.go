package v2

import (
	"encoding/json"
	"net/http"
	"timedrop/api"
	"timedrop/helpers"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type UserCtrl struct {
	ChangeTo       string `json:"changeTo"`
	InvalidateUser string `json:"invalidateUser"`
	FacebookId     string `json:"facebookId"`
	FbName         string `json:"fbName"`
	FbImage        string `json:"fbImage"`
	Email          string `json:"email"`
	VerifyCode     string `json:"verifyCode"`
}

type ResultChangeUser struct {
	User  models.User      `json:"user"`
	Token models.AuthToken `json:"token"`
}

func InitUser(r *mux.Router) {
	l4g.Debug("Initializing v2 User api routes")
	userController := UserCtrl{}
	r.Handle("/changeUser", api.ApiTokenRequired(userController.change)).Methods("POST")

}

func (User UserCtrl) change(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	//decode post data to UserCtrl type
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&User); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if User.ChangeTo == "" || User.InvalidateUser == "" {
		r.JSON(res, 400, helpers.GenerateErrorResponse("invalid request", req.Header))
		return
	}

	var resultUser ResultChangeUser
	if err := resultUser.User.FindByID(User.ChangeTo); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var invalidateUser models.User
	if err := invalidateUser.FindByID(User.InvalidateUser); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if User.FacebookId != "" {
		//delete by facebook id
		invalidateUser.DeleteFacebookData()
	} else if User.Email != "" && User.VerifyCode != "" {
		//delete by email
		invalidateUser.DeleteEmailData()
	}

	var authToken models.AuthToken
	tokenId := resultUser.User.FindTokenIdByUserId()

	if err := authToken.FindByUserRefer(tokenId); err != nil {

		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	response := ResultChangeUser{
		User:  resultUser.User,
		Token: authToken,
	}
	r.JSON(res, 200, response)
	return
}
