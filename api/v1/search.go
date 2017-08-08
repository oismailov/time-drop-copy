package v1

import (
	"encoding/json"
	"net/http"

	"timedrop/api"
	"timedrop/helpers"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"
	"gopkg.in/asaskevich/govalidator.v4"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func InitSearch(r *mux.Router) {
	l4g.Debug("Initializing v1 search api routes")
	searchController := SearchCtrl{}
	r.Handle("/search", api.ApiTokenRequired(searchController.Search)).Methods("POST")
}

//SearchCtrl is the controller for /auth
type SearchCtrl struct{}

type searchRequestData struct {
	Data string `json:"data" valid:"required"`
}

type searchResult struct {
	ID       uint
	Username string
	Level    string
	Score    int
}

//Search handels /search
func (searchCtrl SearchCtrl) Search(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var searchRequest searchRequestData
	if err := decoder.Decode(&searchRequest); err != nil {
		panic(err)
	}

	// validate login request
	_, err := govalidator.ValidateStruct(searchRequest)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if len(searchRequest.Data) < 2 {
		r.JSON(res, 422, helpers.GenerateErrorResponse("search_min_length", req.Header))
		return
	}

	var userRes []models.User
	db := models.GetDatabaseSession()
	db.Where("email LIKE ? OR username LIKE ?", "%"+searchRequest.Data+"%", "%"+searchRequest.Data+"%").Find(&userRes)

	var parsedUsers []searchResult
	for _, user := range userRes {
		parsedUser := searchResult{
			Username: user.Username,
			ID:       user.ID,
			Level:    user.Level,
			Score:    user.Score,
		}
		parsedUsers = append(parsedUsers, parsedUser)
	}

	r.JSON(res, 200, parsedUsers)
}
