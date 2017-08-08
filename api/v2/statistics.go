package v2

import (
	"net/http"

	"timedrop/api"
	"timedrop/helpers"
	"timedrop/middlewares"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func InitStatistics(r *mux.Router) {
	l4g.Debug("Initializing v1 statistics api routes")
	statisticsController := StatisticsCtrl{}
	sr := r.PathPrefix("/statistics").Subrouter()
	sr.Handle("/toplist", api.ApiTokenRequired(statisticsController.TopList)).Methods("GET")
	sr.Handle("/rank", api.ApiTokenRequired(statisticsController.Rank)).Methods("GET")
}

//StatisticsCtrl handels /statistics
type StatisticsCtrl struct{}

//TopList returns the top 10 for the same level
func (statisticsCtrl StatisticsCtrl) TopList(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	var topListRankRefer uint

	rankNameParam := req.FormValue("rank")
	if rankNameParam != "" {
		var paramRank models.Level
		paramRank.FindByName(rankNameParam)
		topListRankRefer = paramRank.ID
	} else {
		currentUser, err := middlewares.GetUserFromContext(res, req)
		if err != nil {
			r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
		topListRankRefer = currentUser.LevelRefer
	}

	db := models.GetDatabaseSession()
	var users []models.User
	db.Where("level_refer = ?", topListRankRefer).Order("score desc").Limit(20).Find(&users)

	r.JSON(res, 200, users)
}

type userRankResult struct {
	models.User
	Rank int `json:"rank"`
}

//Rank returns users rank within his level
func (statisticsCtrl StatisticsCtrl) Rank(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	db := models.GetDatabaseSession()
	sqlQuery := "SELECT *, @curRank := @curRank + 1 AS rank FROM users p, (SELECT @curRank := 0) r WHERE `level_refer` = ? ORDER BY score DESC;"

	var results []userRankResult
	db.Raw(sqlQuery, currentUser.LevelRefer).Scan(&results)

	for _, user := range results {
		if user.ID == currentUser.ID {
			r.JSON(res, 200, user)
			return
		}
	}

	r.JSON(res, 404, helpers.GenerateErrorResponse("user_not_found", req.Header))
	return
}
