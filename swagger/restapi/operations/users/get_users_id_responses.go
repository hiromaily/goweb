package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/hiromaily/go-gin-wrapper/swagger/models"
)

/*GetUsersIDOK An user.

swagger:response getUsersIdOK
*/
type GetUsersIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.Users `json:"body,omitempty"`
}

// NewGetUsersIDOK creates GetUsersIDOK with default headers values
func NewGetUsersIDOK() *GetUsersIDOK {
	return &GetUsersIDOK{}
}

// WithPayload adds the payload to the get users Id o k response
func (o *GetUsersIDOK) WithPayload(payload *models.Users) *GetUsersIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users Id o k response
func (o *GetUsersIDOK) SetPayload(payload *models.Users) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetUsersIDDefault Unexpected error.

swagger:response getUsersIdDefault
*/
type GetUsersIDDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersIDDefault creates GetUsersIDDefault with default headers values
func NewGetUsersIDDefault(code int) *GetUsersIDDefault {
	if code <= 0 {
		code = 500
	}

	return &GetUsersIDDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get users ID default response
func (o *GetUsersIDDefault) WithStatusCode(code int) *GetUsersIDDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get users ID default response
func (o *GetUsersIDDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get users ID default response
func (o *GetUsersIDDefault) WithPayload(payload *models.Error) *GetUsersIDDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users ID default response
func (o *GetUsersIDDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIDDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
