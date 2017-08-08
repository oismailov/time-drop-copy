package v1

import (
	"encoding/json"
	// "fmt"
	"net/http"
	"strconv"

	"timedrop/api"
	"timedrop/helpers"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func InitAuth(r *mux.Router) {
	l4g.Debug("Initializing v1 auth api routes")
	authController := AuthCtrl{}
	r.Handle("/auth/login", api.ApiHandler(authController.Login)).Methods("POST")
	r.Handle("/auth/verifycode", api.ApiHandler(authController.VerifyCode)).Methods("POST")
	r.Handle("/auth/register", api.ApiHandler(authController.Register)).Methods("POST")
}

//AuthCtrl is the controller for /auth
type AuthCtrl struct{}

type userTokenResponseData struct {
	User  models.User `json:"user"`
	Token string      `json:"token"`
}

//Register handels /auth/register
func (authCtrl AuthCtrl) Register(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var user models.User
	if err := decoder.Decode(&user); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	user.Score = 100

	if user.Email == "" {
		user.Guest = true
	} else {
		if isUnique := user.IsEmailUnique(user.Email); !isUnique {
			r.JSON(res, 422, helpers.GenerateErrorResponse("register_email_already_assigned", req.Header))
			return
		}
	}

	if err := user.Save(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse("register_username_already_assigned", req.Header))
		return
	}

	var token string

	//Send login code
	if !user.Guest {
		go user.SendLoginEmail()
	} else {
		// Guest login generates token without auth
		jwtToken, _ := helpers.GenerateJWTToken()
		authToken := models.AuthToken{
			UserRefer: user.ID,
			Token:     jwtToken,
		}
		user.AppendAuthToken(authToken)
		user.Save()

		token = jwtToken
	}

	resUser := userTokenResponseData{
		User:  user,
		Token: token,
	}

	r.JSON(res, 201, resUser)

	return
}

type verifyCodeRequestData struct {
	UserID string `json:"userId" valid:"required"`
	Token  string `json:"code" valid:"required"`
}

//VerifyCode handels /auth/verifycode
func (authCtrl AuthCtrl) VerifyCode(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var verifyCodeRequest verifyCodeRequestData
	if err := decoder.Decode(&verifyCodeRequest); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// validate login request
	_, err := govalidator.ValidateStruct(verifyCodeRequest)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var user models.User
	user.FindByID(verifyCodeRequest.UserID)

	if err := user.ValidateLoginCode(verifyCodeRequest.Token); err != true {
		r.JSON(res, 422, helpers.GenerateErrorResponse("invalid_code", req.Header))
		return
	}

	userID, convErr := strconv.Atoi(verifyCodeRequest.UserID)
	if convErr != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse("invalid_user_id", req.Header))
		return
	}

	jwtToken, _ := helpers.GenerateJWTToken()
	authToken := models.AuthToken{
		UserRefer: uint(userID),
		Token:     jwtToken,
	}
	user.AppendAuthToken(authToken)
	user.Save()

	verifyCodeRequestResponse := userTokenResponseData{
		User:  user,
		Token: jwtToken,
	}

	r.JSON(res, 200, verifyCodeRequestResponse)
	return
}

type loginRequestData struct {
	Data string `json:"data" valid:"required"`
}

//Login handels /auth/login
func (authCtrl AuthCtrl) Login(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var loginRequest loginRequestData
	if err := decoder.Decode(&loginRequest); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// validate login request
	_, err := govalidator.ValidateStruct(loginRequest)
	if err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var user models.User
	var findErr error

	// Check if param is a email address
	if govalidator.IsEmail(loginRequest.Data) {
		if findErr = user.FindByEmail(loginRequest.Data); findErr != nil {
			r.JSON(res, 400, helpers.GenerateErrorResponse(findErr.Error(), req.Header))
			return
		}
	} else {
		if findErr = user.FindByUsername(loginRequest.Data); findErr != nil {
			r.JSON(res, 400, helpers.GenerateErrorResponse(findErr.Error(), req.Header))
			return
		}
	}

	if user.Guest {
		r.JSON(res, 422, helpers.GenerateErrorResponse("guest_login_not_supported", req.Header))
		return
	}

	var resUser models.User

	// If user is found proceed with the login process if not return no error
	// to make bruteforce attacks harder
	if user.ID != 0 {
		go user.SendLoginEmail()
		resUser.ID = user.ID
		resUser.Username = user.Username
	}

	r.JSON(res, 200, resUser)
	return
}
