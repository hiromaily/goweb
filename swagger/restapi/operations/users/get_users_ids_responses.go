package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/hiromaily/go-gin-wrapper/swagger/models"
)

/*GetUsersIdsOK An array of user's ID.

swagger:response getUsersIdsOK
*/
type GetUsersIdsOK struct {

	/*
	  In: Body
	*/
	Payload *models.UserIds `json:"body,omitempty"`
}

// NewGetUsersIdsOK creates GetUsersIdsOK with default headers values
func NewGetUsersIdsOK() *GetUsersIdsOK {
	return &GetUsersIdsOK{}
}

// WithPayload adds the payload to the get users ids o k response
func (o *GetUsersIdsOK) WithPayload(payload *models.UserIds) *GetUsersIdsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users ids o k response
func (o *GetUsersIdsOK) SetPayload(payload *models.UserIds) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIdsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetUsersIdsDefault Unexpected error.

swagger:response getUsersIdsDefault
*/
type GetUsersIdsDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersIdsDefault creates GetUsersIdsDefault with default headers values
func NewGetUsersIdsDefault(code int) *GetUsersIdsDefault {
	if code <= 0 {
		code = 500
	}

	return &GetUsersIdsDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get users ids default response
func (o *GetUsersIdsDefault) WithStatusCode(code int) *GetUsersIdsDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get users ids default response
func (o *GetUsersIdsDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get users ids default response
func (o *GetUsersIdsDefault) WithPayload(payload *models.Error) *GetUsersIdsDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users ids default response
func (o *GetUsersIdsDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIdsDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
