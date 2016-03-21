package gomc

import (
	//"fmt"
	"github.com/asaskevich/govalidator"
	"strings"
	"reflect"
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
}

type RequestErrorWrapper struct {
    Message string `json:"message"`
    Errors []RequestError `json:"errors"`
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

			    		empty := IsEmptyValue(valueField)
				    	valid := false
				    	
				    	//Not Empty
				    	if rule.Rule == "NotEmpty" {
				    		
				    		if !empty {
				    			valid = true
				    		}

				    		if !valid {
					    		error := RequestError{
					                Field : JsonKeyFromStructKey(m.Schema, nameField),
					                Message : rule.Message,
					            }
					            
					            errors = append(errors, error)
					    	}
				    	}

				    	//Then Check Rules
				    	if rule.Rule != "NotEmpty" {
					    	if !empty {
								
								valid = false
					    		
					    		//Rules
						    	if rule.Rule == "IsEmail"{
						    		valid = govalidator.IsEmail(valueField.String())
						    	}
						    	if rule.Rule == "IsAlpha"{
						    		valid = govalidator.IsAlpha(valueField.String())
						    	}
						    	if rule.Rule == "IsAlphanumeric"{
						    		valid = govalidator.IsAlphanumeric(valueField.String())
						    	}
						    	if rule.Rule == "IsByteLength"{
						    		valid = govalidator.IsByteLength(valueField.String(), rule.Min, rule.Max)
						    	}
						    	if rule.Rule == "IsURL"{
						    		valid = govalidator.IsURL(valueField.String())
						    	}
						    	
						    	if !valid {

						    		error := RequestError{
						                Field : JsonKeyFromStructKey(m.Schema, nameField),
						                Message : rule.Message,
						            }
						            errors = append(errors, error)
						    	}
					    	}
					    }
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

