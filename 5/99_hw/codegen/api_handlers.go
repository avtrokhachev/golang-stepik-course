package main

import "fmt"
import "strconv"
import "strings"
import "encoding/json"
import "net/http"
import "net/url"


func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		h.wrapperProfile(w, r)
	case "/user/create":
		h.wrapperCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	

	

	var query url.Values
	if r.Method == "GET" {
		query = r.URL.Query()
	} else {
		_ = r.ParseForm()
		query = r.Form
	}

	request, err := validateProfileParams(query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	var response interface{}
	response, err = h.Profile(r.Context(), request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]interface{}{
		"response": response,
		"error":    "",
	})
	w.Write(data)
}

func (h *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	var query url.Values
	if r.Method == "GET" {
		query = r.URL.Query()
	} else {
		_ = r.ParseForm()
		query = r.Form
	}

	request, err := validateCreateParams(query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	var response interface{}
	response, err = h.Create(r.Context(), request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]interface{}{
		"response": response,
		"error":    "",
	})
	w.Write(data)
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		h.wrapperCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	var query url.Values
	if r.Method == "GET" {
		query = r.URL.Query()
	} else {
		_ = r.ParseForm()
		query = r.Form
	}

	request, err := validateOtherCreateParams(query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	var response interface{}
	response, err = h.Create(r.Context(), request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]interface{}{
		"response": response,
		"error":    "",
	})
	w.Write(data)
}

func validateProfileParams(query url.Values) (ProfileParams, error) {
	validatedRequest := ProfileParams{}
	
	validatedRequest.Login = query.Get("login")
	
	
	if validatedRequest.Login == "" {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("Required filed is empty")}
	}

	

	

	

	

	

	
	

	return validatedRequest, nil
}

func validateCreateParams(query url.Values) (CreateParams, error) {
	validatedRequest := CreateParams{}
	
	validatedRequest.Login = query.Get("login")
	
	
	if validatedRequest.Login == "" {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("Required filed is empty")}
	}

	

	

	

	

	
	if len(validatedRequest.Login) < 10 {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation maxValue")}
	}

	
	
	validatedRequest.Name = query.Get("full_name")
	
	

	

	

	

	

	

	
	
	validatedRequest.Status = query.Get("status")
	
	

	
	if validatedRequest.Status == "" {
		validatedRequest.Status = "user"
	}

	

	

	

	

	
	enumStatusValues := []string{  "user" , "moderator" , "admin" }
	isValid := false
	for _, val := range enumStatusValues {
		if val == validatedRequest.Status {
			isValid = true
			break
		}
	}
	if isValid {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("status must be one of [%s]", strings.Join(enumStatusValues, ", "))}
	}
	
	var err error
	validatedRequest.Age, err = strconv.Atoi(query.Get("age"))
	if err != nil {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on convertation")}
	}
	
	

	

	
	if validatedRequest.Age > 128 {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation maxValue")}
	}

	

	
	if validatedRequest.Age < 0 {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation minValue")}
	}

	

	
	

	return validatedRequest, nil
}

func validateOtherCreateParams(query url.Values) (OtherCreateParams, error) {
	validatedRequest := OtherCreateParams{}
	
	validatedRequest.Username = query.Get("username")
	
	
	if validatedRequest.Username == "" {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("Required filed is empty")}
	}

	

	

	

	

	
	if len(validatedRequest.Username) < 3 {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation maxValue")}
	}

	
	
	validatedRequest.Name = query.Get("account_name")
	
	

	

	

	

	

	

	
	
	validatedRequest.Class = query.Get("class")
	
	

	
	if validatedRequest.Class == "" {
		validatedRequest.Class = "warrior"
	}

	

	

	

	

	
	enumClassValues := []string{  "warrior" , "sorcerer" , "rouge" }
	isValid := false
	for _, val := range enumClassValues {
		if val == validatedRequest.Class {
			isValid = true
			break
		}
	}
	if isValid {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("class must be one of [%s]", strings.Join(enumClassValues, ", "))}
	}
	
	var err error
	validatedRequest.Level, err = strconv.Atoi(query.Get("level"))
	if err != nil {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on convertation")}
	}
	
	

	

	
	if validatedRequest.Level > 50 {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation maxValue")}
	}

	

	
	if validatedRequest.Level < 1 {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation minValue")}
	}

	

	
	

	return validatedRequest, nil
}
