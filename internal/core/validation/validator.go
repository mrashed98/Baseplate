package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(data map[string]interface{}, schema map[string]interface{}) error {
	if schema == nil || len(schema) == 0 {
		// No schema defined, allow any data
		return nil
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	documentLoader := gojsonschema.NewBytesLoader(dataJSON)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		var validationErrors []ValidationError
		for _, desc := range result.Errors() {
			validationErrors = append(validationErrors, ValidationError{
				Field:   desc.Field(),
				Message: desc.Description(),
			})
		}
		return &ValidationErrors{Errors: validationErrors}
	}

	return nil
}

func removeRequiredDeep(schema map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range schema {
		if k == "required" {
			continue
		}
		if nestedMap, ok := v.(map[string]interface{}); ok {
			result[k] = removeRequiredDeep(nestedMap)
		} else if nestedArray, ok := v.([]interface{}); ok {
			// Handle arrays (e.g., allOf, anyOf, oneOf, items)
			cleanedArray := make([]interface{}, len(nestedArray))
			for i, item := range nestedArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					cleanedArray[i] = removeRequiredDeep(itemMap)
				} else {
					cleanedArray[i] = item
				}
			}
			result[k] = cleanedArray
		} else {
			result[k] = v
		}
	}
	return result
}

func (v *Validator) ValidatePartial(data map[string]interface{}, schema map[string]interface{}) error {
	// For partial updates, remove required constraint
	if len(schema) == 0 {
		return nil
	}

	// Create a copy of schema without required fields
	partialSchema := removeRequiredDeep(schema)

	return v.Validate(data, partialSchema)
}

func IsValidationError(err error) bool {
	var ve *ValidationErrors
	return errors.As(err, &ve)
}

func GetValidationErrors(err error) *ValidationErrors {
	var ve *ValidationErrors
	if errors.As(err, &ve) {
		return ve
	}
	return nil
}
