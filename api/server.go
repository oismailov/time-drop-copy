package api

import (
	"net/http"
	"time"
	l4g "github.com/alecthomas/log4go"
	"github.com/braintree/manners"
	"github.com/gorilla/mux"
)

type Server struct {
	Server *manners.GracefulServer
	// Store  store.Store
	Router *mux.Router
}

var Srv *Server

func NewServer(port string) {
	l4g.Info("Server is initializing...")

	var httpServer http.Server
	httpServer.Addr = port

	Srv = &Server{}
	Srv.Server = manners.NewWithServer(&httpServer)
	Srv.Router = mux.NewRouter()
}

func StartServer(port string) {
	l4g.Info("Starting server...")
	l4g.Info("Server is listening at" + port)

	var handler http.Handler = Srv.Router
	Srv.Server.Server.Handler = handler

	go func() {
		err := Srv.Server.ListenAndServe()
		if err != nil {
			l4g.Critical("Error starting server, err:%v", err)
			time.Sleep(time.Second)
			panic("Error starting server " + err.Error())
		}
	}()
}

func StopServer() {
	l4g.Info("Stopping server...")
}
