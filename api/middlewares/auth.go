package middlewares

import (
	"net/http"

	"errors"

	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/unrolled/render"
	"timedrop/helpers"
	"timedrop/models"
	"golang.org/x/net/context"
)

//GetUserFromContext ensures a valid user is received from the context
func GetUserFromContext(_ http.ResponseWriter, req *http.Request) (models.User, error) {
	user := req.Context().Value("user").(models.User)
	if user.ID != 0 {
		return user, nil
	}
	return models.User{}, errors.New("couldn't find user")
}

//GetUserByTokenMiddlewareWithValidation passes user to the request context
func GetUserByTokenMiddlewareWithValidation(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	/*if rv := context.Get(req, "jwtToken"); rv != nil {
		token := rv.(*jwt.Token)
		var user models.User
		user.FindByToken(token.Raw)

		if user.ID != 0 {
			context.Set(req, "user", user)
			next(res, req)
			return
		}
	}*/

	r := render.New(render.Options{})
	r.JSON(res, 401, helpers.GenerateErrorResponse("invalid_token", req.Header))
	return
}

//GetUserByTokenMiddleware passes user to the request context
func GetUserByTokenMiddleware(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	req = req.WithContext(context.Background())

	token, err := jwtmiddleware.FromAuthHeader(req)
	if token != "" && err == nil {
		var user models.User
		user.FindByToken(token)

		fmt.Println(user)

		if user.ID != 0 {
			ctx := context.WithValue(req.Context(), "user", user)
			req = req.WithContext(ctx)
			next(res, req)
			return
		}
	}

	r := render.New(render.Options{})
	r.JSON(res, 401, helpers.GenerateErrorResponse("invalid_token", req.Header))
	return
}
