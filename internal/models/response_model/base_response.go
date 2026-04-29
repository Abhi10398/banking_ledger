package response_model

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"awesomeProject/service_errors"
)

type ResponseV2 struct {
	Errors []*service_errors.Error `json:"errors"`
	Data   any                     `json:"data"`
}

func (r *ResponseV2) SetData(data any) *ResponseV2 {
	r.Data = data
	return r
}

func (r *ResponseV2) AddError(err error) *ResponseV2 {
	if err == nil {
		r.Errors = append(r.Errors, nil)
		return r
	} else if e, ok := err.(service_errors.Error); ok {
		r.Errors = append(r.Errors, &e)
		return r
	} else if e, ok := err.(validator.ValidationErrors); ok {
		appendErr := service_errors.New(e.Error())
		r.Errors = append(r.Errors, &appendErr)
		return r
	}
	appendErr := service_errors.New(err.Error())
	r.Errors = append(r.Errors, &appendErr)
	return r
}

func NewResponseV2(data any) *ResponseV2 {
	r := EmptyResponseV2()
	r.SetData(data)
	return r
}

func EmptyResponseV2() *ResponseV2 {
	r := new(ResponseV2)
	r.Errors = make([]*service_errors.Error, 0)
	return r
}

func GetResponseV2(c *fiber.Ctx, code int, err error, data any) error {
	responseV2 := NewResponseV2(data)
	if err != nil && err != (service_errors.Error{}) {
		responseV2.AddError(err)
	}

	return c.Status(code).JSON(responseV2)
}

func GetResponseV2ForMultipleErrors(c *fiber.Ctx, code int, errs []error, data any) error {
	responseV2 := NewResponseV2(data)
	for _, err := range errs {
		if err != nil && err != (service_errors.Error{}) {
			responseV2.AddError(err)
		}
	}
	return c.Status(code).JSON(responseV2)
}
