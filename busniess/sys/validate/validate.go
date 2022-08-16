// Package validate contains the support for validating models
package validate

import (
	"github.com/go-playground/validator/v10"
)

// validate holds the settings and caches for validating request struct values
var validate *validator.Validate
