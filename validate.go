package gomc

import (
	//"fmt"
	"github.com/asaskevich/govalidator"
	"strings"
	"reflect"
	"regexp"
)

type ValidationRule struct {
	Rule string
	Message string
	Min int
	Max int
	ValidatedOnActions []string
}

type RequestError struct {
    Field string `json:"field"`
    Message string `json:"message"`
	Data interface{} `json:"data"`
}

type RequestErrorWrapper struct {
    Message string `json:"message"`
    Errors []RequestError `json:"errors"`
}

func ValidateField(field string, rule ValidationRule, value reflect.Value, errors *[]RequestError){

	empty := IsEmptyValue(value)
	valid := false

	//Not Empty
	if rule.Rule == "NotEmpty" {

		if !empty {
			valid = true
		}

		if !valid {
			error := RequestError{
				Field : field,
				Message : rule.Message,
			}
			*errors = append(*errors, error)
		}
	}

	//Then Check Rules
	if rule.Rule != "NotEmpty" {
		if !empty {

			valid = false

			//Rules
			if rule.Rule == "IsEmail"{
				valid = govalidator.IsEmail(value.String())
			}
			if rule.Rule == "IsAlpha"{
				valid = govalidator.IsAlpha(value.String())
			}
			if rule.Rule == "IsAlphanumeric"{
				valid = govalidator.IsAlphanumeric(value.String())
			}
			if rule.Rule == "IsByteLength"{
				valid = govalidator.IsByteLength(value.String(), rule.Min, rule.Max)
			}
			if rule.Rule == "IsURL"{
				valid = govalidator.IsURL(value.String())
			}
			if rule.Rule == "IsZipCode" {
				var validZip = regexp.MustCompile(`^\d{5}(?:[-\s]\d{4})?$`)
				valid = validZip.MatchString(value.String())
			}
			if rule.Rule == "IsStateCode" {
				var validStateCode = regexp.MustCompile(`^((A[LKSZR])|(C[AOT])|(D[EC])|(F[ML])|(G[AU])|(HI)|(I[DLNA])|(K[SY])|(LA)|(M[EHDAINSOT])|(N[EVHJMYCD])|(MP)|(O[HKR])|(P[WAR])|(RI)|(S[CD])|(T[NX])|(UT)|(V[TIA])|(W[AVIY]))$`)
				valid = validStateCode.MatchString(value.String())
			}

			if !valid {

				error := RequestError{
					Field : field,
					Message : rule.Message,
				}
				*errors = append(*errors, error)
			}
		}
	}
}

func ModelValidate(m *Model) []RequestError {


	data := reflect.ValueOf(m.Data)
	errors := m.ValidationErrors
	if len(m.ValidationRules) > 0 {
		for i := 0; i < data.NumField(); i++ {
			valueField := data.Field(i)
			nameField := data.Type().Field(i).Name
			if rules, ok := m.ValidationRules[nameField]; ok {
			    for _, rule := range rules{
			    	if StringInSlice(m.SaveAction, rule.ValidatedOnActions){
						ValidateField(JsonKeyFromStructKey(m.Schema, nameField), rule, valueField, &errors)
				    }
			    }
			}
		}
	}

    return errors
}

func FormatErrors(errorString string, model interface{}) ([]RequestError){

    var errors []RequestError

    errArray := strings.Split(errorString,";")
    for _, error := range errArray{
        if(error != ""){
            errorArray := strings.Split(error,": ")
            error := RequestError{
                Field : JsonKeyFromStructKey(model, errorArray[0]),
                Message : errorArray[1],
            }
            errors = append(errors, error)
        }
    }

    return errors
}

