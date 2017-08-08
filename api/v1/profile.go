package v1

import (
	"encoding/json"
	"net/http"

	"gopkg.in/asaskevich/govalidator.v4"

	"timedrop/api"
	"timedrop/helpers"
	"timedrop/middlewares"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func InitProfile(r *mux.Router) {
	l4g.Debug("Initializing v1 games api routes")
	profileController := ProfileCtrl{}
	r.Handle("/profile", api.ApiTokenRequired(profileController.List)).Methods("GET")
	r.Handle("/profile", api.ApiTokenRequired(profileController.Update)).Methods("PUT")
	r.Handle("/profile/language", api.ApiTokenRequired(profileController.SetLanguage)).Methods("PUT")
	r.Handle("/profile/verifyemail", api.ApiTokenRequired(profileController.VerifyEmail)).Methods("POST")
	r.Handle("/profile/pushtoken", api.ApiTokenRequired(profileController.SetPushToken)).Methods("POST")
	r.Handle("/profile/pushtoken", api.ApiTokenRequired(profileController.DeletePushToken)).Methods("PUT")

}

//ProfileCtrl is the controller for /profile
type ProfileCtrl struct{}

//List /profile (GET) handler
func (profileCtrl ProfileCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, currentUser)
	return
}

type profileUpdateRequestData struct {
	Email string `json:"email" valid:"email"`
}

//Update /profile (PUT) handler
func (profileCtrl ProfileCtrl) Update(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var profileUpdateRequest profileUpdateRequestData
	if err := decoder.Decode(&profileUpdateRequest); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// validate login request
	_, err := govalidator.ValidateStruct(profileUpdateRequest)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if isUnique := currentUser.IsEmailUnique(profileUpdateRequest.Email); !isUnique {
		r.JSON(res, 422, helpers.GenerateErrorResponse("register_email_already_assigned", req.Header))
		return
	}

	if err := currentUser.SendEmailVerification(profileUpdateRequest.Email); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
	return
}

type verifyEmailRequestData struct {
	Code string `json:"code" valid:"required"`
}

//VerifyEmail /profile/verifyemail (POST) handler
func (profileCtrl ProfileCtrl) VerifyEmail(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var verifyEmailRequest verifyEmailRequestData
	if err := decoder.Decode(&verifyEmailRequest); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// validate email request
	_, err := govalidator.ValidateStruct(verifyEmailRequest)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	email, err := currentUser.ValidateEmailCode(verifyEmailRequest.Code)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse("invalid_code", req.Header))
		return
	}

	if isUnique := currentUser.IsEmailUnique(email); !isUnique {
		r.JSON(res, 422, helpers.GenerateErrorResponse("register_email_already_assigned", req.Header))
		return
	}

	currentUser.Email = email
	currentUser.Guest = false
	if err := currentUser.Save(); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, currentUser)
	return
}

//SetPushToken /profile/pushtoken (POST) handler
func (profileCtrl ProfileCtrl) SetPushToken(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var setPushTokenRequestData models.PushToken
	if err := decoder.Decode(&setPushTokenRequestData); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// validate email request
	_, err := govalidator.ValidateStruct(setPushTokenRequestData)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if err := currentUser.AddPushToken(setPushTokenRequestData); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 200, "")
	return
}

type deletePushTokenRequestTemplate struct {
	Token string `json:"token" valid:"required"`
}

//DeletePushToken /profile/pushtoken (DELETE) handler
func (profileCtrl ProfileCtrl) DeletePushToken(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var deletePushTokenRequestData deletePushTokenRequestTemplate
	if err := decoder.Decode(&deletePushTokenRequestData); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// validate email request
	_, err := govalidator.ValidateStruct(deletePushTokenRequestData)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if err := currentUser.DeletePushToken(deletePushTokenRequestData.Token); err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.Text(res, 204, "")
	return
}

type setLanguageRequestTemplate struct {
	Language string `json:"language" valid:"required"`
}

//SetLanguage /profile/language (PUT) handler
func (profileCtrl ProfileCtrl) SetLanguage(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var setLanguageRequestData setLanguageRequestTemplate
	if err := decoder.Decode(&setLanguageRequestData); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// validate email request
	_, err := govalidator.ValidateStruct(setLanguageRequestData)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	currentUser.Language = setLanguageRequestData.Language
	if err := currentUser.Save(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, currentUser)
	return
}
