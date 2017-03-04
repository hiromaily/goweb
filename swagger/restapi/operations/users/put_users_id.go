package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// PutUsersIDHandlerFunc turns a function with the right signature into a put users ID handler
type PutUsersIDHandlerFunc func(PutUsersIDParams) middleware.Responder

// Handle executing the request and returning a response
func (fn PutUsersIDHandlerFunc) Handle(params PutUsersIDParams) middleware.Responder {
	return fn(params)
}

// PutUsersIDHandler interface for that can handle valid put users ID params
type PutUsersIDHandler interface {
	Handle(PutUsersIDParams) middleware.Responder
}

// NewPutUsersID creates a new http.Handler for the put users ID operation
func NewPutUsersID(ctx *middleware.Context, handler PutUsersIDHandler) *PutUsersID {
	return &PutUsersID{Context: ctx, Handler: handler}
}

/*PutUsersID swagger:route PUT /users/{id} Users putUsersId

Update specific user data

Update specific user data.

*/
type PutUsersID struct {
	Context *middleware.Context
	Handler PutUsersIDHandler
}

func (o *PutUsersID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, _ := o.Context.RouteInfo(r)
	var Params = NewPutUsersIDParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
