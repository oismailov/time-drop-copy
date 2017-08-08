package v2

import (
	"timedrop/api"
)

func InitApi() {
	r := api.Srv.Router.PathPrefix("/api/v2").Subrouter()
	InitAuth(r)
	InitFriend(r)
	InitGames(r)
	InitProfile(r)
	InitSearch(r)
	InitStatistics(r)
	InitLifeReques(r)
	InitUser(r)
}
