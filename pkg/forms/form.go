package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

//Regular expression to validate email
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

//Create a custom Form struct, which anonymously embeds a url.Values object (to hold the form data)
// and an Errors field to hold any validation errors for the form data.
type Form struct {
	url.Values
	Errors errors
}

//Define a New function to initialize a custom Form struct.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

//Implement a Required method to check that specific fields in the form
//data are present and not blank. If any fields fail this check, add the
//appropriate message to the form errors.
func (f *Form) Required(field ...string) {
	for _, field := range field {
		value := f.Get(field)
		if strings.TrimSpace(value) == " " {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

//Implement a MaxLength method to check that a specific field in the form
//contains a maximum number of characters. If the check fails then add the
//appropriate message to the form errors
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == " " {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (max is %d characters)", d))
	}
}

//Implement a PermittedValues method to check that a specific field in the form
//matches one of a set of specific permitted values. If the check fails
//then add the appropriate message to the form errors.
func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == " " {
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This field is invalid")
}

//Implement a Valid method which returns true if there are no errors.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

//Implement a MinLength method to check that a specific field in the form
//contains a minimum number of characters.
func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == " " {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d characters)", d))
	}
}

//Implement a MatchesPattern method to check that a specific field in the form
//matches a regular expression.
func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == " " {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}
