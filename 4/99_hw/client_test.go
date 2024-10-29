package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// код писать тут
const correctAuthToken = "someCorrectAuthToken"

type UserData struct {
	Id        int    `xml:"id"`
	Guid      string `xml:"guid"`
	IsActive  bool   `xml:"osActive"`
	Balance   string `xml:"balance"`
	Picture   string `xml:"picture"`
	Age       int    `xml:"age"`
	EyeColor  string `xml:"eyeColor"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Gender    string `xml:"gender"`
	Company   string `xml:"company"`
	Email     string `xml:"email"`
	Phone     string `xml:"phone"`
	Address   string `xml:"address"`
	About     string `xml:"about"`
}

type UsersData struct {
	Row []UserData `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// 1. Check authToken and return error if it's not correct
	authToken := r.Header.Get("AccessToken")
	if authToken != correctAuthToken {
		http.Error(w, fmt.Sprintf("Provided access %s token is incorrect!", authToken), http.StatusUnauthorized)
		return
	}

	// 2. Read all Users from dataset.xml
	data, err := os.ReadFile("dataset.xml")
	if err != nil {
		http.Error(w, "An error occurred while trying to read dataset.xml", http.StatusInternalServerError)
		return
	}

	users := &UsersData{}
	err = xml.Unmarshal(data, users)
	if err != nil {
		http.Error(w, "An error occurred while trying to parse dataset.xml", http.StatusInternalServerError)
		return
	}

	var limit int
	limit, err = strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		http.Error(w, "An error occurred while trying to parse limit", http.StatusBadRequest)
		return
	}

	var offset int
	offset, err = strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		http.Error(w, "An error occurred while trying to parse offset", http.StatusBadRequest)
		return
	}

	var orderBy int
	orderBy, err = strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		http.Error(w, "An error occurred while trying to parse order_by", http.StatusBadRequest)
		return
	}

	query := r.FormValue("query")

	order_field := r.FormValue("order_field")
	if order_field == "" {
		order_field = "Name"
	}

	// 3. Search users with filtering by Name and About fields
	result := make([]UserData, 0)
	if query == "" { // unnecessary, only for performance improvements
		result = append(result, users.Row...)
	} else {
		for _, user := range users.Row {
			userName := user.FirstName + user.LastName
			if strings.Contains(userName, query) || strings.Contains(user.About, query) {
				result = append(result, user)
			}
		}
	}

	// 4. Order the result
	if orderBy != OrderByAsIs {
		switch order_field {
		case "Id":
			sort.Slice(result, func(i int, j int) bool {
				return result[i].Id < result[j].Id
			})
		case "Age":
			sort.Slice(result, func(i int, j int) bool {
				return result[i].Age < result[j].Age
			})
		case "Name":
			sort.Slice(result, func(i int, j int) bool {
				return result[i].FirstName < result[j].FirstName
			})
		}
	}
	if orderBy == OrderByDesc {
		slices.Reverse(result)
	}

	// 5. offset + limit
	if offset >= len(result) {
		result = make([]UserData, 0)
	}
	result = result[offset:]

	if limit > len(result) {
		limit = len(result)
	}
	result = result[:limit]

	response := make([]User, len(result))
	for i, user := range result {
		response[i] = User{
			Id:     user.Id,
			Name:   user.FirstName + user.LastName,
			Age:    user.Age,
			Gender: user.Gender,
			About:  user.About,
		}
	}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		http.Error(w, "An error occurred while trying to write result", http.StatusInternalServerError)
		return
	}
}

// Tests for FindUsers
func TestRaisesErrorInValidation(t *testing.T) {
	type ErrorFindUsers struct {
		SearchReq SearchRequest
		Error     string
	}

	var testsData = []ErrorFindUsers{
		{
			SearchReq: SearchRequest{
				Limit: -1,
			},
			Error: "limit must be > 0",
		},
		{
			SearchReq: SearchRequest{
				Offset: -1,
			},
			Error: "offset must be > 0",
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	testClient := &SearchClient{URL: testServer.URL, AccessToken: correctAuthToken}
	for _, test := range testsData {
		_, err := testClient.FindUsers(test.SearchReq)
		assert.EqualError(t, err, test.Error)
	}
}

func TestRaisesErrorInSearchServer(t *testing.T) {
	type ErrorFindUsers struct {
		Handler func(w http.ResponseWriter, r *http.Request)
		Error   string
	}

	var testsData = []ErrorFindUsers{
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(10 * time.Second)
			},
			Error: "timeout for limit=1&offset=0&order_by=0&order_field=&query=",
		},
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				err, _ := json.Marshal(&SearchErrorResponse{Error: "Some unknown error"})
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(err)
			},
			Error: "unknown bad request error: Some unknown error",
		},
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			Error: "Bad AccessToken",
		},
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			Error: "SearchServer fatal error",
		},
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("garbage"))
			},
			Error: "cant unpack error json: invalid character 'g' looking for beginning of value",
		},
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("garbage"))
			},
			Error: "cant unpack error json: invalid character 'g' looking for beginning of value",
		},
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				err, _ := json.Marshal(&SearchErrorResponse{Error: "ErrorBadOrderField"})
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(err)
			},
			Error: "OrderFeld  invalid",
		},
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				err, _ := json.Marshal(&SearchErrorResponse{Error: "UnknownError"})
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(err)
			},
			Error: "unknown bad request error: UnknownError",
		},
		{
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("garbage"))
			},
			Error: "cant unpack result json: invalid character 'g' looking for beginning of value",
		},
	}

	for _, test := range testsData {
		testServer := httptest.NewServer(http.HandlerFunc(test.Handler))
		testClient := &SearchClient{URL: testServer.URL, AccessToken: correctAuthToken}
		_, err := testClient.FindUsers(SearchRequest{})
		assert.EqualError(t, err, test.Error)
	}
}
