package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetUsersIDHandlerFunc turns a function with the right signature into a get users ID handler
type GetUsersIDHandlerFunc func(GetUsersIDParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetUsersIDHandlerFunc) Handle(params GetUsersIDParams) middleware.Responder {
	return fn(params)
}

// GetUsersIDHandler interface for that can handle valid get users ID params
type GetUsersIDHandler interface {
	Handle(GetUsersIDParams) middleware.Responder
}

// NewGetUsersID creates a new http.Handler for the get users ID operation
func NewGetUsersID(ctx *middleware.Context, handler GetUsersIDHandler) *GetUsersID {
	return &GetUsersID{Context: ctx, Handler: handler}
}

/*GetUsersID swagger:route GET /users/{id} Users getUsersId

Get specific user data

Get specific user data.

*/
type GetUsersID struct {
	Context *middleware.Context
	Handler GetUsersIDHandler
}

func (o *GetUsersID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, _ := o.Context.RouteInfo(r)
	var Params = NewGetUsersIDParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
