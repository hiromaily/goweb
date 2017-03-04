package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
)

// Jwt jwt
// swagger:model Jwt
type Jwt struct {

	// return code.
	Code int32 `json:"code,omitempty"`

	// token for API call.
	Token string `json:"token,omitempty"`
}

// Validate validates this jwt
func (m *Jwt) Validate(formats strfmt.Registry) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
