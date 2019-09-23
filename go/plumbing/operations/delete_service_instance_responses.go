// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/netlify/open-api/go/models"
)

// DeleteServiceInstanceReader is a Reader for the DeleteServiceInstance structure.
type DeleteServiceInstanceReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteServiceInstanceReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewDeleteServiceInstanceNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewDeleteServiceInstanceDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewDeleteServiceInstanceNoContent creates a DeleteServiceInstanceNoContent with default headers values
func NewDeleteServiceInstanceNoContent() *DeleteServiceInstanceNoContent {
	return &DeleteServiceInstanceNoContent{}
}

/*DeleteServiceInstanceNoContent handles this case with default header values.

Deleted
*/
type DeleteServiceInstanceNoContent struct {
}

func (o *DeleteServiceInstanceNoContent) Error() string {
	return fmt.Sprintf("[DELETE /sites/{site_id}/services/{addon}/instances][%d] deleteServiceInstanceNoContent ", 204)
}

func (o *DeleteServiceInstanceNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteServiceInstanceDefault creates a DeleteServiceInstanceDefault with default headers values
func NewDeleteServiceInstanceDefault(code int) *DeleteServiceInstanceDefault {
	return &DeleteServiceInstanceDefault{
		_statusCode: code,
	}
}

/*DeleteServiceInstanceDefault handles this case with default header values.

error
*/
type DeleteServiceInstanceDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the delete service instance default response
func (o *DeleteServiceInstanceDefault) Code() int {
	return o._statusCode
}

func (o *DeleteServiceInstanceDefault) Error() string {
	return fmt.Sprintf("[DELETE /sites/{site_id}/services/{addon}/instances][%d] deleteServiceInstance default  %+v", o._statusCode, o.Payload)
}

func (o *DeleteServiceInstanceDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *DeleteServiceInstanceDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
