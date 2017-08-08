package v2

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"timedrop/api"
	"timedrop/helpers"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func InitAuth(r *mux.Router) {
	l4g.Debug("Initializing v2 auth api routes")
	sr := r.PathPrefix("/auth").Subrouter()
	sr.Handle("/createUser", api.ApiHandler(createUser)).Methods("POST")
	sr.Handle("/overwrite", api.ApiHandler(overwriteUser)).Methods("POST")
	sr.Handle("/getUser/{userID:[0-9]+}", api.ApiHandler(getUser)).Methods("GET")
	sr.Handle("/updateUser", api.ApiTokenRequired(updateUser)).Methods("POST")
	sr.Handle("/changeUser", api.ApiHandler(changeUser)).Methods("POST")
	sr.Handle("/authToken", api.ApiHandler(authToken)).Methods("POST")
}

type createUserResponse struct {
	User  models.User      `json:"user"`
	Token models.AuthToken `json:"token"`
}

type VerifyCode struct {
	Token string `json:"verifyCode"`
}

func createUser(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	user := models.User{
		Username: models.GetGuestUsername(),
		Guest:    true,
	}

	tokenString, err := helpers.GenerateJWTToken()
	if err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	token := models.AuthToken{
		Token:     tokenString,
		UserRefer: user.ID,
	}

	user.AppendAuthToken(token)
	//add 100 points
	user.Score = 100
	if err := user.Save(); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	response := createUserResponse{
		User:  user,
		Token: token,
	}

	r.JSON(res, 200, response)
}

func updateUser(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	db := models.GetDatabaseSession()

	currentUser := req.Context().Value("user").(models.User)

	decoder := json.NewDecoder(req.Body)
	var user models.User
	if err := decoder.Decode(&user); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var resultUser models.User
	currentUser.UpdateUserUpdatedAt()
	// db.Where("facebook_id = ?", user.FacebookID).First(&resultUser)
	if err := resultUser.FindByID(currentUser.ID); err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var userByEmail *models.User
	userByEmail.FindByEmail(user.Email)
	if userByEmail != nil {
		if len(user.Email) > 0 && userByEmail.ID != resultUser.ID {
			r.JSON(res, 404, helpers.GenerateErrorResponse("email is already connected to an existing account", req.Header))
			return
		}
	}

	if user.Email != "" && !govalidator.IsEmail(user.Email) {
		r.JSON(res, 404, helpers.GenerateErrorResponse("email is not valid", req.Header))
		return
	}

	var userByLoginCode models.User
	isValidLoginCode := false
	if user.VerifyCode != "" {
		isValidLoginCode = userByLoginCode.ValidateUserLoginCode(user.VerifyCode, currentUser.ID)
		if isValidLoginCode == false {
			r.JSON(res, 200, map[string]string{"errors": "invalid_code"})
			return
		}
	}
	if (isValidLoginCode == true) && (user.Email != "") {
		resultUser.IsVerified = true
	}

	if user.Email != "" {
		var userByEmail models.User
		userByEmail.FindByEmail(user.Email)
		if userByEmail.ID != 0 && userByEmail.ID != currentUser.ID && isValidLoginCode == true {
			s := strconv.Itoa(int(userByEmail.ID))
			r.JSON(res, 200, map[string]string{"errors": "email_connected", "userId": s})
			return
		}
		resultUser.Email = user.Email
	}

	//do ONLY save "email" if a valid verifyCode is included as well.
	if isValidLoginCode == false {
		if user.Username != "" {
			var userByUsername models.User
			userByUsername.FindByUsername(user.Username)
			if userByUsername.ID != 0 && userByUsername.ID != currentUser.ID {
				r.JSON(res, 200, helpers.GenerateErrorResponse("username_taken", req.Header))
				return
			}

			if len(strings.Trim(user.Username, " ")) < 2 {
				r.JSON(res, 200, helpers.GenerateErrorResponse("username_lenght", req.Header))
				return
			}

			resultUser.Username = user.Username
		}

		if user.FacebookID != "" {
			var userByFacebookID models.User
			userByFacebookID.FindOtherUserByFacebookIDAndID(currentUser.ID, user.FacebookID)
			if userByFacebookID.ID != 0 && userByFacebookID.ID != currentUser.ID {
				//s := strconv.Itoa(int(userByFacebookID.ID))
				response := make(map[string]interface{})
				response["errors"] = "fb_connected"
				response["user"] = userByFacebookID

				r.JSON(res, 200, response)
				return
			}
			resultUser.IsVerified = true

			resultUser.FacebookID = user.FacebookID
		}

		if user.ExtraData != "" {
			resultUser.ExtraData = user.ExtraData
		}

		if user.FbImageUrl != "" {
			resultUser.FbImageUrl = user.FbImageUrl
		}

		if user.FbName != "" {
			resultUser.FbName = user.FbName
		}

		if user.IsVerified != false {
			resultUser.IsVerified = user.IsVerified
		}

		if user.Score != 0 {
			resultUser.Score = user.Score
		}

		if user.GamesPlayedCount != 0 {
			resultUser.GamesPlayedCount = user.GamesPlayedCount
		}

		if user.GamesWonCount != 0 {
			resultUser.GamesWonCount = user.GamesWonCount
		}

		if user.CurrentLevel != 0 {
			resultUser.CurrentLevel = user.CurrentLevel
		}

		if user.Level != "" {
			resultUser.Level = user.Level
		}

		if user.LevelRefer != 0 {
			resultUser.LevelRefer = user.LevelRefer
		}

		if user.LevelData != "" {
			resultUser.LevelData = user.LevelData
		}

		if user.TopLevel != "" {
			resultUser.TopLevel = user.TopLevel
		}

		if user.TopLevelRefer != 0 {
			resultUser.TopLevelRefer = user.TopLevelRefer
		}

		if user.AchievementsData != "" {
			resultUser.AchievementsData = user.AchievementsData
		}

		if user.TreasureData != "" {
			resultUser.TreasureData = user.TreasureData
		}

		if user.PergamentData != "" {
			resultUser.PergamentData = user.PergamentData
		}

		if user.Coins != 0 {
			resultUser.Coins = user.Coins
		}

		if user.Avatar != 0 {
			resultUser.Avatar = user.Avatar
		}
	}

	db.Save(&resultUser)

	r.JSON(res, 200, resultUser)
}

func overwriteUser(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	authVal := req.Header.Get("authorization")

	if len(authVal) == 0 {
		err := errors.New("missing auth header value")
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var user models.User
	if err := decoder.Decode(&user); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

}

type getUserRequestData struct {
	UserID string `json:"userId" valid:"required"`
}

func getUser(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	userID := vars["userID"]

	var user models.User
	err := user.FindByID(userID)
	if err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, user)
}

type changeUserRequest struct {
	FacebookID string `json:"facebookId"`
	Email      string `json:"email"`
	AuthToken  string `json:"authToken"`
	DeleteFlag bool   `json:"deleteFlag"`
	DeleteId   int    `json:"deleteId"`
}

func changeUser(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var data changeUserRequest
	if err := decoder.Decode(&data); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var user models.User

	if data.DeleteFlag {
		err := user.FindByID(data.DeleteId)
		if err != nil {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
	}

	if len(data.FacebookID) > 0 {
		if err := user.FindByFacebookID(data.FacebookID); err != nil {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
	} else {
		if err := user.FindByEmail(data.Email); err != nil {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
	}

	r.JSON(res, 200, user)

}

type authTokenRequest struct {
	Email string `json:"email"`
}

func authToken(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var data changeUserRequest
	if err := decoder.Decode(&data); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var user models.User
	err := user.FindByEmail(data.Email)
	if err != nil {
		r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
	}

	if !user.DoesTokenExist() {
		tokenString, err := helpers.GenerateJWTToken()
		if err != nil {
			r.JSON(res, 404, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}

		authToken := models.AuthToken{
			Token:     tokenString,
			UserRefer: user.ID,
		}

		user.AppendAuthToken(authToken)
		user.SendAuthToken(tokenString)
	}

	r.JSON(res, 200, nil)
}
