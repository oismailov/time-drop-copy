package middlewares

import (
	"net/http"

	"timedrop/models"
)

//CleanUpGamesMiddleware cleans up games
func CleanUpGamesMiddleware(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var game models.Game
	go game.CleanUp()
	next(res, req)
	return
}
