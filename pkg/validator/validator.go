package validator

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"gorm.io/gorm"
)

// Validator is an interface for validating data
type Validator interface {
	Validate(ctx context.Context, data interface{}) error
}

// GoValidator is a struct that implements the Validator interface using go-playground/validator
type GoValidator struct {
	validate *validator.Validate
	uni      ut.Translator
	DB       *gorm.DB
}

// ValidationError is a custom error type for validation errors
type ValidationError struct {
	ErrorFields map[string]string `json:"errorFields,omitempty"`
}

// NewGoValidator creates a new instance of GoValidator with custom validators and messages
func NewGoValidator(db *gorm.DB) *GoValidator {
	v := validator.New()
	eng := en.New()
	uni := ut.New(eng, eng)
	trans, _ := uni.GetTranslator("en")

	en_translations.RegisterDefaultTranslations(v, trans)

	// register custom tag name function to use json tag instead of struct field name
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	gv := &GoValidator{validate: v, uni: trans, DB: db}
	gv.registerCustomValidators()
	gv.registerCustomMessages(trans)

	return gv
}

// registerCustomMessages registers custom error messages for validation tags
func (v *GoValidator) registerCustomMessages(trans ut.Translator) {
	// Override Required
	v.validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is a required field", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return fmt.Sprintf("%s is a required field", toProperCase(fe.Field()))
	})

	// Override Email
	v.validate.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} must be a valid email address", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return fmt.Sprintf("%s must be a valid email address", toProperCase(fe.Field()))
	})

	// Override Min
	v.validate.RegisterTranslation("min", trans, func(ut ut.Translator) error {
		return ut.Add("min", "{0} must be at least {1} characters long", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return fmt.Sprintf("%s must be at least %s characters long", toProperCase(fe.Field()), fe.Param())
	})
}

// registerCustomValidators registers custom validation functions and their corresponding error messages
func (v *GoValidator) registerCustomValidators() {
	v.validate.RegisterValidation("unique", v.uniqueValidator)
	v.validate.RegisterValidation("incolumn", v.incolumnValidator)
	v.validate.RegisterValidation("complexpassword", v.complexPasswordValidator)
	v.validate.RegisterValidation("whitespace", v.whiteSpaceValidator)
	v.validate.RegisterValidation("username", v.usernameValidator)

	// Message untuk unique
	v.validate.RegisterTranslation("unique", v.uni, func(ut ut.Translator) error {
		return ut.Add("unique", "{0} already exists", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return fmt.Sprintf("%s already exists", fe.Value().(string))
	})

	// Message untuk complexpassword
	v.validate.RegisterTranslation("complexpassword", v.uni, func(ut ut.Translator) error {
		return ut.Add("complexpassword", "{0} not valid", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		return "Password must be 8-64 characters long, contain: uppercase, lowercase, special character and number"
	})

	// Message untuk username
	v.validate.RegisterTranslation("username", v.uni, func(ut ut.Translator) error {
        return ut.Add("username", "{0} can only contain letters, numbers, underscores, hyphens, and dots", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        return fmt.Sprintf("%s can only contain letters, numbers, underscores, hyphens, and dots", toProperCase(fe.Field()))
    })
}

// Custom validation functions
func (v *GoValidator) uniqueValidator(fl validator.FieldLevel) bool {
	params := strings.Split(fl.Param(), "->")
	if len(params) != 2 {
		return false
	}

	tableName := params[0]
	fieldName := params[1]
	fieldValue := fl.Field().String()

	var count int64 // GORM menggunakan int64 untuk Count()
	// GORM raw table query yang sangat aman dari SQL Injection
	err := v.DB.Table(tableName).Where(fmt.Sprintf("%s = ?", fieldName), fieldValue).Count(&count).Error
	if err != nil {
		return false
	}

	return count == 0
}

// incolumnValidator digunakan untuk validasi bahwa nilai yang diberikan harus ada di kolom tertentu di database (mirip dengan "exists" tapi untuk update data agar tidak memvalidasi dirinya sendiri)
func (v *GoValidator) incolumnValidator(fl validator.FieldLevel) bool {
	params := strings.Split(fl.Param(), "->")
	if len(params) != 2 {
		return false
	}

	tableName := params[0]
	fieldName := params[1]
	fieldValue := fl.Field().String()

	if fieldValue == "" {
		return true
	}

	var count int64
	err := v.DB.Table(tableName).Where(fmt.Sprintf("%s = ?", fieldName), fieldValue).Count(&count).Error
	if err != nil {
		return false
	}

	return count != 0
}

// complexPasswordValidator digunakan untuk validasi password yang kompleks (8-12 karakter, harus ada huruf besar, huruf kecil, angka, dan karakter khusus)
func (v *GoValidator) complexPasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 || len(password) > 64 {
		return false
	}
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_\-+={[}\]|\\:;"'<,>.?/~]`).MatchString(password)
	return hasLower && hasUpper && hasNumber && hasSpecial
}

// whiteSpaceValidator digunakan untuk validasi bahwa string tidak boleh mengandung spasi (baik spasi biasa maupun tab)
func (v *GoValidator) whiteSpaceValidator(fl validator.FieldLevel) bool {
	text := fl.Field().String()
	return strings.IndexFunc(text, unicode.IsSpace) == -1
}

// username validator
func (v *GoValidator) usernameValidator(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	
	return usernameRegex.MatchString(username)
}

// Validate implements the Validator interface, validating the given data and returning a ValidationError if there are any validation errors
func (v *GoValidator) Validate(ctx context.Context, data interface{}) error {
	err := v.validate.StructCtx(ctx, data)
	if err == nil {
		return nil
	}

	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err
	}

	validationErrors := err.(validator.ValidationErrors)
	if len(validationErrors) > 0 {
		errorFields := make(map[string]string)
		for _, err := range validationErrors {
			errorFields[err.Field()] = err.Translate(v.uni)
		}
		return &ValidationError{ErrorFields: errorFields}
	}
	return nil
}

func (ve *ValidationError) Error() string {
	var errMsgs []string
	for field, msg := range ve.ErrorFields {
		errMsgs = append(errMsgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(errMsgs, "; ")
}

// Helper function to convert snake_case to Proper Case (for better error messages)
func toProperCase(input string) string {
	// snake case
	input = strings.ReplaceAll(input, "_", " ")

	// camel case
	var spacedString strings.Builder
	for i, r := range input {
		if i > 0 && unicode.IsUpper(r) && input[i-1] != ' ' {
			spacedString.WriteRune(' ')
		}
		spacedString.WriteRune(r)
	}

	// make first word capital
	words := strings.Fields(spacedString.String())
	for i, word := range words {
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}

	return strings.Join(words, " ")
}