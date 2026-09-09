package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	appcompany "github.com/fragoulis/xmercise/internal/application/company"
	httpadapter "github.com/fragoulis/xmercise/internal/port/http"
)

func TestJSONHandlerReturnsJSONErrors(t *testing.T) {
	handler := httpadapter.NewJSONHandler(
		httpadapter.NewJSONStrictHandler(
			httpadapter.NewServer(appcompany.NewService(nil)),
		),
	)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		wantStatus     int
		wantField      string
		wantViolations []violation
	}{
		{
			name:       "invalid company ID",
			method:     http.MethodGet,
			path:       "/v1/companies/1",
			wantStatus: http.StatusBadRequest,
			wantField:  "id",
		},
		{
			name:       "missing create fields",
			method:     http.MethodPost,
			path:       "/v1/companies",
			body:       "{}",
			wantStatus: http.StatusBadRequest,
			wantField:  "name",
			wantViolations: []violation{
				{
					Field:   "name",
					Message: "is required",
				},
				{
					Field:   "type",
					Message: "is required",
				},
			},
		},
		{
			name:       "invalid company type",
			method:     http.MethodPost,
			path:       "/v1/companies",
			body:       `{"name":"Acme","employees_count":1,"registered":false,"type":"LLC"}`,
			wantStatus: http.StatusBadRequest,
			wantField:  "type",
		},
		{
			name:       "malformed JSON",
			method:     http.MethodPost,
			path:       "/v1/companies",
			body:       "{",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "patch with curl default content type",
			method:      http.MethodPatch,
			path:        "/v1/companies/c9f77fb9-5410-4086-a17d-9a90f31c132e",
			body:        `{"description":"foo bar"}`,
			contentType: "application/x-www-form-urlencoded",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			path:       "/unknown",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			} else if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", contentType)
			}

			var body errorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode JSON response: %v", err)
			}
			if body.Status != test.wantStatus || body.Title == "" {
				t.Fatalf("error response = %#v", body)
			}
			if test.wantField != "" && !containsViolation(body.Violations, test.wantField) {
				t.Fatalf("violations = %#v, want %q", body.Violations, test.wantField)
			}
			if test.wantViolations != nil && !reflect.DeepEqual(body.Violations, test.wantViolations) {
				t.Fatalf("violations = %#v, want %#v", body.Violations, test.wantViolations)
			}
		})
	}
}

type errorResponse struct {
	Status     int         `json:"status"`
	Title      string      `json:"title"`
	Violations []violation `json:"violations"`
}

type violation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func containsViolation(violations []violation, field string) bool {
	for _, violation := range violations {
		if violation.Field == field {
			return true
		}
	}

	return false
}
