package v2

import (
	"encoding/json"
	"net/http"
	"timedrop/api"
	"timedrop/helpers"
	"timedrop/middlewares"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type LifeRequestCtrl struct{}

func InitLifeReques(r *mux.Router) {
	l4g.Debug("Initializing v2 LifeReques api routes")
	lifeRequestController := LifeRequestCtrl{}
	r.Handle("/lifeRequest", api.ApiTokenRequired(lifeRequestController.create)).Methods("POST")
	r.Handle("/lifeRequest", api.ApiTokenRequired(lifeRequestController.show)).Methods("GET")
	r.Handle("/installRequest", api.ApiTokenRequired(lifeRequestController.showInstallRequests)).Methods("GET")
	r.Handle("/giveLife", api.ApiTokenRequired(lifeRequestController.giveLife)).Methods("POST")
	r.Handle("/installRequest", api.ApiTokenRequired(lifeRequestController.createInstallRequest)).Methods("POST")
	r.Handle("/installFromRequest", api.ApiTokenRequired(lifeRequestController.createInstallFromRequest)).Methods("POST")

}

type createLifeReques struct {
	Receivers []uint `json:"receiverRefers"`
}

type giveLifeRequest struct {
	Requesters []uint `json:"requesterRefers"`
}

type showLifeReques struct {
	IncomingRequests []interface{} `json:"incomingRequests"`
	ApprovedRequests []interface{} `json:"approvedRequests"`
}

type installRequests struct {
	Guid string `json:"guid"`
}

func (LifeRequestCtrl LifeRequestCtrl) create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	user, _ := middlewares.GetUserFromContext(res, req)

	var data createLifeReques
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&data); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var lifeRequest models.LifeRequest
	lifeRequest.CreateLifeRecords(data.Receivers, user.ID)

	r.JSON(res, 200, map[string]string{})
	return
}

func (LifeRequestCtrl LifeRequestCtrl) show(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	user, _ := middlewares.GetUserFromContext(res, req)

	var lifeRequest models.LifeRequest
	var userIds []int
	var approvedUserIds []int
	var incomingRequests []interface{}
	var approvedRequests []interface{}
	//get user ids
	userIds = lifeRequest.GetLifeRequests(user.ID)
	approvedUserIds = lifeRequest.GetAppovedRequests(user.ID)

	for _, item := range userIds {
		var userModel models.User
		userModel.FindByID(item)
		incomingRequests = append(incomingRequests, &userModel)
	}

	for _, item := range approvedUserIds {
		var userModel models.User
		userModel.FindByID(item)
		approvedRequests = append(approvedRequests, &userModel)
	}

	responseObj := showLifeReques{
		IncomingRequests: incomingRequests,
		ApprovedRequests: approvedRequests,
	}

	r.JSON(res, 200, responseObj)

	return
}

func (LifeRequestCtrl LifeRequestCtrl) giveLife(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	user, _ := middlewares.GetUserFromContext(res, req)

	var data giveLifeRequest
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&data); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var lifeRequest models.LifeRequest
	lifeRequest.GiveLife(data.Requesters, user.ID)

	r.JSON(res, 200, map[string]string{})
	return
}

func (LifeRequestCtrl LifeRequestCtrl) createInstallRequest(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	user, _ := middlewares.GetUserFromContext(res, req)

	var data installRequests
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&data); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var lifeRequest models.LifeRequest
	lifeRequest.CreateInstallRequest(data.Guid, user.ID)

	r.JSON(res, 200, map[string]string{})
	return
}

//createInstallFromRequest
func (LifeRequestCtrl LifeRequestCtrl) createInstallFromRequest(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	user, _ := middlewares.GetUserFromContext(res, req)

	var data installRequests
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&data); err != nil {
		r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	var lifeRequest models.LifeRequest
	lifeRequest.CreateInstallFromRequest(data.Guid, user.ID)

	r.JSON(res, 200, map[string]string{})
	return
}

//showInstallRequests
func (LifeRequestCtrl LifeRequestCtrl) showInstallRequests(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	user, _ := middlewares.GetUserFromContext(res, req)

	var lifeRequest models.LifeRequest
	users := make([]interface{}, 0)

	userIds := lifeRequest.ShowInstallRequests(user.ID)
	for _, item := range userIds {
		var userModel models.User
		userModel.FindByID(item)
		users = append(users, &userModel)
	}

	r.JSON(res, 200, users)
	return
}
