package middlewares

import (
	"net/http"

	"timedrop/models"
)

//CleanUpGamesMiddleware cleans up games
func CleanUpGamesMiddleware(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var game models.Game
	var lifeRequests models.LifeRequest
	go game.CleanUp()
	go lifeRequests.CleanUp()

	next(res, req)
	return
}
