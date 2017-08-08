package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/asaskevich/govalidator.v4"

	"timedrop/helpers"
	"timedrop/middlewares"
	"timedrop/models"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
)

//GameCtrl is the controller for /games
type GameCtrl struct{}

//List handels /games (GET)
func (gameCtrl GameCtrl) List(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	db := models.GetDatabaseSession()

	var games []models.Game

	sqlQuery := "opponent_refer = ? AND state_creator = ? AND state_opponent != ? AND completed != 1"
	db.Where(sqlQuery, currentUser.ID, models.GameStateCompleted, models.GameStateCompleted).Find(&games)

	var parsedGames []models.Game
	for _, game := range games {
		game.Creator.FindByID(game.CreatorRefer)
		game.Opponent.FindByID(game.OpponentRefer)
		parsedGames = append(parsedGames, game)
	}

	r.JSON(res, 200, parsedGames)
	return
}

type createGameRequestData struct {
	FriendID string `json:"friendId"`
}

//ListHistory handels /games/history (GET)
func (gameCtrl GameCtrl) ListHistory(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var requestData createGameRequestData
	if err := decoder.Decode(&requestData); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
	}

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	db := models.GetDatabaseSession()
	var games []models.Game

	sqlQuery := "(creator_refer = ? AND state_creator = ?) OR (opponent_refer = ? AND state_opponent = ?)"
	sqlStmt := db.Where(sqlQuery, currentUser.ID, models.GameStateCompleted, currentUser.ID, models.GameStateCompleted)

	if requestData.FriendID != "" {
		sqlQuery = "(creator_refer = ? AND opponent_refer = ?) OR (creator_refer = ? AND opponent_refer = ?) AND completed = 1"
		sqlStmt = db.Where(sqlQuery, currentUser.ID, requestData.FriendID, requestData.FriendID, currentUser.ID)
	}

	sqlStmt.Order("updated_at desc").Find(&games)

	var parsedGames []models.Game
	for _, game := range games {
		game.Creator.FindByID(game.CreatorRefer)
		game.Opponent.FindByID(game.OpponentRefer)
		parsedGames = append(parsedGames, game)
	}

	r.JSON(res, 200, parsedGames)
	return
}

//Create handels /games (POST)
func (gameCtrl GameCtrl) Create(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	decoder := json.NewDecoder(req.Body)
	var createGameRequest createGameRequestData
	if err := decoder.Decode(&createGameRequest); err != nil {
		if err != io.EOF {
			r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
	}

	var game models.Game

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// try to match create request with an already open game
	if createGameRequest.FriendID == "" {
		if matchedGame, err := game.FindMatchNew(currentUser.ID); err == nil && matchedGame.ID != 0 {
			matchedGame.OpponentRefer = currentUser.ID
			matchedGame.Opponent = currentUser

			var creatorUser models.User
			if err := creatorUser.FindByID(matchedGame.CreatorRefer); err != nil {
				r.JSON(res, 400, helpers.GenerateErrorResponse(err.Error(), req.Header))
				return
			}
			matchedGame.Creator = creatorUser

			if err := matchedGame.Save(); err != nil {
				r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
				return
			}

			r.JSON(res, 200, matchedGame)
			return
		}
		log.Error(err)
	}

	game = models.Game{
		CreatorRefer:  currentUser.ID,
		Creator:       currentUser,
		LevelRefer:    currentUser.LevelRefer,
		StateCreator:  models.GameStatePending,
		StateOpponent: models.GameStatePending,
	}

	friendID, _ := strconv.Atoi(createGameRequest.FriendID)
	if createGameRequest.FriendID != "" {
		if err := game.Opponent.FindByID(friendID); err != nil {
			r.JSON(res, 404, helpers.GenerateErrorResponse("game_opponent_not_found", req.Header))
			return
		}
		game.OpponentRefer = uint(friendID)
		game.FromFriendRequest = true
	}

	if game.OpponentRefer != 0 {
		if isOpen := game.GameIsAlreadyOpen(game.CreatorRefer, game.OpponentRefer); isOpen {
			r.JSON(res, 422, helpers.GenerateErrorResponse("game_already_open", req.Header))
			return
		}
	}

	game.SetRandomGameType()
	game.SetRandomMapID()

	if err := game.Save(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, game)
	return
}

//StartGame starts the game and checks its status
func (gameCtrl GameCtrl) StartGame(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	gameID := vars["gameID"]

	var game models.Game
	if err := game.FindByID(gameID); err != nil {
		if err == gorm.ErrRecordNotFound {
			r.JSON(res, 404, helpers.GenerateErrorResponse("game_not_found", req.Header))
			return
		}

		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if game.CreatorRefer != currentUser.ID && game.OpponentRefer != currentUser.ID {
		r.JSON(res, 422, helpers.GenerateErrorResponse("game_not_related", req.Header))
		return
	}

	isCreator := game.CreatorRefer == currentUser.ID
	if (isCreator && game.StateCreator != models.GameStatePending) || (!isCreator && game.StateOpponent != models.GameStatePending) {
		r.JSON(res, 422, helpers.GenerateErrorResponse("game_not_pending", req.Header))
		return
	}

	timeNow := time.Now()
	if isCreator {
		game.StartTimeCreator = &timeNow
		game.StateCreator = models.GameStateStarted
	} else {
		game.StartTimeOpponent = &timeNow
		game.StateOpponent = models.GameStateStarted

		if game.FromFriendRequest {
			game.FriendRequestAccepted = true
			game.FriendRequestAcceptedTime = &timeNow
		}
	}

	if err := game.Save(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	r.JSON(res, 200, game)
	return
}

type resultGameRequestData struct {
	Data int `json:"data" valid:"required"`
}

//Result saves the result
func (gameCtrl GameCtrl) Result(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})

	vars := mux.Vars(req)
	gameID := vars["gameID"]

	var game models.Game
	if err := game.FindByID(gameID); err != nil {
		if err == gorm.ErrRecordNotFound {
			r.JSON(res, 404, helpers.GenerateErrorResponse("game_not_found", req.Header))
			return
		}

		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	decoder := json.NewDecoder(req.Body)
	var resultGameRequest resultGameRequestData
	if err := decoder.Decode(&resultGameRequest); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	// validate login request
	_, err := govalidator.ValidateStruct(resultGameRequest)
	if err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	currentUser, err := middlewares.GetUserFromContext(res, req)
	if err != nil {
		r.JSON(res, 500, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	isCreator := game.CreatorRefer == currentUser.ID
	if isCreator {
		if game.StateCreator != models.GameStateStarted {
			r.JSON(res, 422, helpers.GenerateErrorResponse("game_not_started", req.Header))
			return
		}
		if game.ScoreCreator != 0 {
			r.JSON(res, 422, helpers.GenerateErrorResponse("score_already_saved", req.Header))
			return
		}
		game.ScoreCreator = resultGameRequest.Data
		game.StateCreator = models.GameStateCompleted

		game.Opponent.FindByID(game.OpponentRefer)
		if game.OpponentRefer != 0 {
			var pushNotification models.PushNotification
			go pushNotification.SendGameRequestPush(game.Opponent)
		}
	} else {
		if game.StateOpponent != models.GameStateStarted {
			r.JSON(res, 422, helpers.GenerateErrorResponse("game_not_started", req.Header))
			return
		}
		if game.ScoreOpponent != 0 {
			r.JSON(res, 422, helpers.GenerateErrorResponse("score_already_saved", req.Header))
			return
		}
		game.ScoreOpponent = resultGameRequest.Data
		game.StateOpponent = models.GameStateCompleted
	}

	if err := game.Save(); err != nil {
		r.JSON(res, 422, helpers.GenerateErrorResponse(err.Error(), req.Header))
		return
	}

	if (game.StateCreator == models.GameStateCompleted) && (game.StateOpponent == models.GameStateCompleted) {
		if err := game.Complete(); err != nil {
			r.JSON(res, 401, helpers.GenerateErrorResponse(err.Error(), req.Header))
			return
		}
	}

	r.JSON(res, 200, game)
	return
}
