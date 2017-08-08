package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"timedrop/api"
	"timedrop/config"
	"timedrop/helpers"
	"timedrop/models"

	"timedrop/api/v1"
	"timedrop/api/v2"

	log "github.com/Sirupsen/logrus"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/nicksnyder/go-i18n/i18n"
)

func main() {
	// set logging options
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)

	// load translations
	i18n.MustLoadTranslationFile("assets/i18n/en-US.all.json")
	i18n.MustLoadTranslationFile("assets/i18n/de-DE.all.json")

	// Bootstrap tables
	models.Bootstrap()

	// Auth mw bootstrap
	jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return helpers.GetHMACSecret(), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
		UserProperty:  "jwtToken",
	})

	//parse cmd attributes
	var flagDevMode bool
	var port string
	flag.BoolVar(&flagDevMode, "dev_mode", false, "if true - load dev config")
	flag.StringVar(&port, "port", "80", "set listen port")
	flag.Parse()

	if flagDevMode {
		config.LoadConfig("config/config_dev.json")
	} else {
		config.LoadConfig("config/config_prod.json")
	}

	api.NewServer(port)
	v1.InitApi()
	v2.InitApi()
	api.StartServer()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c

	api.StopServer()

	// baseRouter := mux.NewRouter()
	// authController := controllers.AuthCtrl{}
	// friendsController := controllers.FriendsCtrl{}
	// searchController := controllers.SearchCtrl{}
	// gamesController := controllers.GameCtrl{}
	// statisticsController := controllers.StatisticsCtrl{}
	// profileController := controllers.ProfileCtrl{}
	//
	// // /auth group
	// authRouter := mux.NewRouter()
	// authRouter.HandleFunc("/api/v1/auth/login", authController.Login).Methods("POST")
	// authRouter.HandleFunc("/api/v1/auth/verifycode", authController.VerifyCode).Methods("POST")
	// authRouter.HandleFunc("/api/v1/auth/register", authController.Register).Methods("POST")
	// // authRouter.HandleFunc("/api/v2/auth/createUser", authController.CreateUser).Methods("GET")
	//
	// baseRouter.PathPrefix("/api/v1/auth").Handler(negroni.New(
	// 	negroni.Wrap(authRouter),
	// ))
	//
	// // /auth/v2
	//
	// authRouterV2 := mux.NewRouter()
	// authRouterV2.HandleFunc("/api/v2/auth/createUser", authController.CreateUser).Methods("POST")
	// baseRouter.PathPrefix("/api/v2/auth").Handler(negroni.New(
	// 	negroni.Wrap(authRouterV2),
	// ))
	//
	// // /profile group
	// profileRouter := mux.NewRouter()
	// profileRouter.HandleFunc("/api/v1/profile", profileController.List).Methods("GET")
	// profileRouter.HandleFunc("/api/v1/profile", profileController.Update).Methods("PUT")
	// profileRouter.HandleFunc("/api/v1/profile/language", profileController.SetLanguage).Methods("PUT")
	// profileRouter.HandleFunc("/api/v1/profile/verifyemail", profileController.VerifyEmail).Methods("POST")
	// profileRouter.HandleFunc("/api/v1/profile/pushtoken", profileController.SetPushToken).Methods("POST")
	// profileRouter.HandleFunc("/api/v1/profile/pushtoken", profileController.DeletePushToken).Methods("PUT")
	//
	// baseRouter.PathPrefix("/api/v1/profile").Handler(negroni.New(
	// 	//negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	// 	negroni.HandlerFunc(middlewares.GetUserByTokenMiddleware),
	// 	negroni.Wrap(profileRouter),
	// ))
	//
	// // /friends group
	// friends := mux.NewRouter()
	// friends.HandleFunc("/api/v1/friends", friendsController.List).Methods("GET")
	// friends.HandleFunc("/api/v1/friends", friendsController.AddFriend).Methods("POST")
	// friends.HandleFunc("/api/v1/friends/{friendID:[0-9]+}", friendsController.RemoveFriend).Methods("DELETE")
	// friends.HandleFunc("/api/v1/friends/request", friendsController.SendFriendRequest).Methods("POST")
	// friends.HandleFunc("/api/v1/friends/request", friendsController.ListFriendRequest).Methods("GET")
	// friends.HandleFunc("/api/v1/friends/request/{friendRequestID:[0-9]+}", friendsController.DeclineFriendRequest).Methods("DELETE")
	//
	// baseRouter.PathPrefix("/api/v1/friends").Handler(negroni.New(
	// 	//negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	// 	negroni.HandlerFunc(middlewares.GetUserByTokenMiddleware),
	// 	negroni.Wrap(friends),
	// ))
	//
	// // /search group
	// search := mux.NewRouter()
	// search.HandleFunc("/api/v1/search", searchController.Search).Methods("POST")
	//
	// baseRouter.PathPrefix("/api/v1/search").Handler(negroni.New(
	// 	//negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	// 	negroni.HandlerFunc(middlewares.GetUserByTokenMiddleware),
	// 	negroni.Wrap(search),
	// ))
	//
	// // /games group
	// games := mux.NewRouter()
	// games.HandleFunc("/api/v1/games", gamesController.List).Methods("GET")
	// games.HandleFunc("/api/v1/games/history", gamesController.ListHistory).Methods("POST")
	// games.HandleFunc("/api/v1/games", gamesController.Create).Methods("POST")
	// games.HandleFunc("/api/v1/games/{gameID:[0-9]+}/start", gamesController.StartGame).Methods("POST")
	// games.HandleFunc("/api/v1/games/{gameID:[0-9]+}/result", gamesController.Result).Methods("POST")
	//
	// baseRouter.PathPrefix("/api/v1/games").Handler(negroni.New(
	// 	//negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	// 	negroni.HandlerFunc(middlewares.CleanUpGamesMiddleware),
	// 	negroni.HandlerFunc(middlewares.GetUserByTokenMiddleware),
	// 	negroni.Wrap(games),
	// ))
	//
	// // /statistics group
	// statistics := mux.NewRouter()
	// statistics.HandleFunc("/api/v1/statistics/toplist", statisticsController.TopList).Methods("GET")
	// statistics.HandleFunc("/api/v1/statistics/rank", statisticsController.Rank).Methods("GET")
	//
	// baseRouter.PathPrefix("/api/v1/statistics").Handler(negroni.New(
	// 	//negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	// 	negroni.HandlerFunc(middlewares.GetUserByTokenMiddleware),
	// 	negroni.Wrap(statistics),
	// ))
	//
	// n := negroni.New(negroni.NewLogger())
	// n.Use(recovery.JSONRecovery(false))
	//
	// n.UseHandler(baseRouter)
	//
	// // Get PORT from env
	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "3000"
	// }
	//
	// n.Run(":" + port)
}
