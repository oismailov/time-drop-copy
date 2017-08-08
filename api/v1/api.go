package v1

import (
	"timedrop/api"
)

func InitApi() {
	r := api.Srv.Router.PathPrefix("/api/v1").Subrouter()
	InitAuth(r)
	InitFriend(r)
	InitGames(r)
	InitProfile(r)
	InitSearch(r)
	InitStatistics(r)
}
