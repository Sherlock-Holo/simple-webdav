package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testHandle(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Test_basicAuth(t *testing.T) {
	const (
		user     = "test"
		password = "test"
	)

	type args struct {
		set      bool
		user     string
		password string
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
	}{
		{
			name: "200 OK",
			args: args{
				set:      true,
				user:     "test",
				password: "test",
			},
			wantCode: 200,
		},
		{
			name:     "403 forbidden not set",
			wantCode: 403,
		},
		{
			name: "403 forbidden set",
			args: args{
				set: true,
			},
			wantCode: 403,
		},
	}

	handler := basicAuth(user, password, http.HandlerFunc(testHandle))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "http://test", nil)

			if tt.args.set {
				request.SetBasicAuth(tt.args.user, tt.args.password)
			}

			handler.ServeHTTP(recorder, request)

			response := recorder.Result()

			if response.StatusCode != tt.wantCode {
				t.Errorf("http status code %d, want %d", response.StatusCode, tt.wantCode)
			}
		})
	}
}
