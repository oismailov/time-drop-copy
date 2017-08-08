package api

import (
	"context"
	"fmt"
	"net/http"
	"timedrop/helpers"
	"timedrop/models"

	l4g "github.com/alecthomas/log4go"
	"github.com/auth0/go-jwt-middleware"
	"github.com/unrolled/render"
)

type handler struct {
	handleFunc  func(http.ResponseWriter, *http.Request)
	requireUser bool
}

func ApiHandler(h func(http.ResponseWriter, *http.Request)) http.Handler {
	return &handler{h, false}
}

func ApiTokenRequired(h func(http.ResponseWriter, *http.Request)) http.Handler {
	return &handler{h, true}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l4g.Debug("%v", r.URL.Path)

	//if api requires user
	if h.requireUser {
		req := r.WithContext(context.Background())

		token, err := jwtmiddleware.FromAuthHeader(req)
		if token != "" && err == nil {
			var user models.User
			user.FindByToken(token)

			fmt.Println(user)

			if user.ID != 0 {
				ctx := context.WithValue(req.Context(), "user", user)
				req = req.WithContext(ctx)
				// next(res, req)
				r = req
				// return
				h.handleFunc(w, r)
				return
			}
		}
		renderer := render.New(render.Options{})
		renderer.JSON(w, 401, helpers.GenerateErrorResponse("invalid_token", r.Header))
		return
	}

	h.handleFunc(w, r)
}
