package gohelpertools

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// RoundTripFunc is used to satisfy the interface requirements for http.Client
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip is used to satisfy the interface requirements for http.Client
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

var jsonTests = []struct {
	name          string
	json          string
	errorExpected bool
	maxSize       int
	allowUnknown  bool
	contentType   string
}{
	{name: "good json", json: `{"foo": "bar"}`, errorExpected: false, maxSize: 1024, allowUnknown: false},
	{name: "badly formatted json", json: `{"foo":"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "incorrect type", json: `{"foo": 1}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "incorrect type", json: `{1: 1}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "empty body", json: ``, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "syntax error in json", json: `{"foo": 1"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "unknown field in json", json: `{"fooo": "bar"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "incorrect type for field", json: `{"foo": 10.2}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "allow unknown field in json", json: `{"fooo": "bar"}`, errorExpected: false, maxSize: 1024, allowUnknown: true},
	{name: "missing field name", json: `{jack: "bar"}`, errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "not json", json: `Hello, world`, errorExpected: true, maxSize: 1024, allowUnknown: false},
}

func TestTools_ReadJSON(t *testing.T) {
	for _, e := range jsonTests {
		var testTools Tools
		// set max file size
		testTools.MaxJSONSize = e.maxSize

		// allow/disallow unknown fields.
		testTools.AllowUnknownFields = e.allowUnknown

		// declare a variable to read the decoded json into.
		var decodedJSON struct {
			Foo string `json:"foo"`
		}

		// create a request with the body.
		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(e.json)))
		if err != nil {
			t.Log("Error", err)
		}
		if e.contentType != "" {
			req.Header.Add("Content-Type", e.contentType)
		} else {
			req.Header.Add("Content-Type", "application/json")
		}

		// create a test response recorder, which satisfies the requirements
		// for a ResponseWriter.
		rr := httptest.NewRecorder()

		// call ReadJSON and check for an error.
		err = testTools.ReadJSON(rr, req, &decodedJSON)

		// if we expect an error, but do not get one, something went wrong.
		if e.errorExpected && err == nil {
			t.Errorf("%s: error expected, but none received", e.name)
		}

		// if we do not expect an error, but get one, something went wrong.
		if !e.errorExpected && err != nil {
			t.Errorf("%s: error not expected, but one received: %s \n%s", e.name, err.Error(), e.json)
		}
		req.Body.Close()
	}
}

func TestTools_ReadJSONAndMarshal(t *testing.T) {
	// set max file size
	var testTools Tools

	// create a request with the body
	req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(`{"foo": "bar"}`)))
	if err != nil {
		t.Log("Error", err)
	}

	// create a test response recorder, which satisfies the requirements
	// for a ResponseWriter
	rr := httptest.NewRecorder()

	// call readJSON and check for an error; since we are using nil for the final parameter,
	// we should get an error
	err = testTools.ReadJSON(rr, req, nil)

	// we expect an error, but did not get one, so something went wrong
	if err == nil {
		t.Error("error expected, but none received")
	}

	req.Body.Close()
}

var writeJSONTests = []struct {
	name          string
	payload       any
	errorExpected bool
}{
	{
		name: "valid",
		payload: JSONResponse{
			Error:   false,
			Message: "foo",
		},
		errorExpected: false,
	},
	{
		name:          "invalid",
		payload:       make(chan int),
		errorExpected: true,
	},
}

func TestTools_WriteJSON(t *testing.T) {
	for _, e := range writeJSONTests {
		// create a variable of type toolbox.Tools, and just use the defaults.
		var testTools Tools

		rr := httptest.NewRecorder()

		headers := make(http.Header)
		headers.Add("FOO", "BAR")
		err := testTools.WriteJSON(rr, http.StatusOK, e.payload, headers)
		if err == nil && e.errorExpected {
			t.Errorf("%s: expected error, but did not get one", e.name)
		}
		if err != nil && !e.errorExpected {
			t.Errorf("%s: did not expect error, but got one: %v", e.name, err)
		}
	}
}

func TestTools_ErrorJSON(t *testing.T) {
	var testTools Tools

	rr := httptest.NewRecorder()
	err := testTools.ErrorJSON(rr, errors.New("some error"), http.StatusServiceUnavailable)
	if err != nil {
		t.Error(err)
	}

	var requestPayload JSONResponse
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&requestPayload)
	if err != nil {
		t.Error("received error when decoding ErrorJSON payload:", err)
	}

	if !requestPayload.Error {
		t.Error("error set to false in response from ErrorJSON, and should be set to true")
	}

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("wrong status code returned; expected 503, but got %d", rr.Code)
	}
}

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string returned")
	}
}

var slugTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid string", s: "now is the time", expected: "now-is-the-time", errorExpected: false},
	{name: "empty string", s: "", expected: "", errorExpected: true},
	{name: "complex string", s: "Now is the time for all GOOD men! + Fish & such &^?123", expected: "now-is-the-time-for-all-good-men-fish-such-123", errorExpected: false},
	{name: "japanese string", s: "こんにちは世界", expected: "", errorExpected: true},
	{name: "japanese string plus roman characters", s: "こんにちは世界 hello world", expected: "hello-world", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {
	var testTool Tools

	for _, e := range slugTests {
		slug, err := testTool.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: error received when none expected: %s", e.name, err.Error())
		}

		if !e.errorExpected && slug != e.expected {
			t.Errorf("%s: wrong slug returned; expected %s but got %s", e.name, e.expected, slug)
		}
	}
}
