// +build integration

package main

// TODO: refactoring

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/hiromaily/go-gin-wrapper/pkg/auth/jwts"
	"github.com/hiromaily/go-gin-wrapper/pkg/config"
	"github.com/hiromaily/go-gin-wrapper/pkg/server/httpheader"
	"github.com/hiromaily/go-gin-wrapper/pkg/token"
)

var errRedirect = errors.New("redirect")

func setup() {
	flag.Parse()
}

func teardown() {
	// TODO: drop test database
}

func TestMain(m *testing.M) {
	setup()

	code := m.Run()

	teardown()

	os.Exit(code)
}

func getloginReferer() (string, error) {
	// config
	conf, err := config.GetEnvConf()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"http://%s:%d/login",
		conf.Server.Host,
		conf.Server.Port,
	), nil
}

func getServer(authMode jwts.JWTAlgo) (*gin.Engine, error) {
	// this code is related to main() in main.go

	// config
	conf, err := config.GetEnvConf()
	if err != nil {
		return nil, err
	}

	// overwrite config by args
	if *portNum != 0 {
		conf.Server.Port = *portNum
	}
	// overwrite jwt mode
	conf.API.JWT.Mode = authMode

	// server
	regi := NewRegistry(conf, true) // run as test mode
	return regi.NewServer().Start()
}

func getClientServer(authMode jwts.JWTAlgo, isCookie bool) (*http.Client, *httptest.Server, error) {
	ginEngine, err := getServer(authMode)
	if err != nil {
		return nil, nil, err
	}

	ts := httptest.NewServer(ginEngine)
	// defer ts.Close()

	client := &http.Client{
		Timeout: time.Duration(3) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errRedirect
		},
	}
	if isCookie {
		jar, _ := cookiejar.New(nil) // cookie
		client.Jar = jar
	}
	return client, ts, nil
}

func createPostData(email, pass, ginToken string) url.Values {
	data := make(url.Values)

	data.Add("inputEmail", email)
	data.Add("inputPassword", pass)
	if ginToken != "" {
		data.Add("gintoken", ginToken)
	}
	return data
}

func TestGetRequest(t *testing.T) {
	basicAuthHeaders := map[string]string{"Authorization": "Basic d2ViOnRlc3Q="}

	type args struct {
		url     string
		method  string
		headers []map[string]string
	}
	type want struct {
		statusCode int
		nextPage   string
		err        error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "top page",
			args: args{
				url:     "/",
				method:  "GET",
				headers: nil,
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "top page 2",
			args: args{
				url:     "/index",
				method:  "GET",
				headers: nil,
			},
			want: want{
				statusCode: http.StatusMovedPermanently,
				nextPage:   "/",
				err:        errRedirect,
			},
		},
		{
			name: "login page",
			args: args{
				url:     "/login",
				method:  "GET",
				headers: nil,
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "logout page",
			args: args{
				url:     "/logout",
				method:  "GET",
				headers: nil,
			},
			want: want{
				statusCode: http.StatusNotFound,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "accounts page",
			args: args{
				url:     "/accounts/",
				method:  "GET",
				headers: nil,
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				nextPage:   "/login",
				err:        errRedirect,
			},
		},
		{
			name: "admin page without basic auth header",
			args: args{
				url:     "/admin/",
				method:  "GET",
				headers: nil,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "admin page",
			args: args{
				url:     "/admin/",
				method:  "GET",
				headers: []map[string]string{basicAuthHeaders},
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "",
				err:        nil,
			},
		},
	}

	client, ts, err := getClientServer(jwts.AlgoHMAC, false)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("[%s] %s %s", tt.name, tt.args.method, tt.args.url)

			req, _ := http.NewRequest(tt.args.method, fmt.Sprintf("%s%s", ts.URL, tt.args.url), nil)
			if tt.args.headers != nil {
				httpheader.SetHTTPHeaders(req, tt.args.headers)
			}
			// request
			res, err := client.Do(req)
			defer func() {
				if res.Body != nil {
					res.Body.Close()
				}
			}()

			// handle response
			urlError, isURLErr := err.(*url.Error)
			if isURLErr && urlError.Err.Error() != tt.want.err.Error() {
				t.Errorf("%s request: actual error: %v, want error: %v", tt.args.url, err, tt.want.err)
				return
			}
			if isURLErr && urlError.Err.Error() == errRedirect.Error() {
				if tt.want.nextPage != res.Header["Location"][0] {
					t.Errorf("%s request: actual next page: %v, want: %v", tt.args.url, res.Header["Location"][0], tt.want.nextPage)
				}
			}
			if err != nil {
				return
			}
			if res.StatusCode != tt.want.statusCode {
				t.Errorf("%s request: actual status code: %v, want: %v", tt.args.url, res.StatusCode, tt.want.statusCode)
			}
		})
	}
}

func TestLoginRequest(t *testing.T) {
	loginReferer, err := getloginReferer()
	if err != nil {
		t.Fatal(err)
	}

	contentType := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	referer := map[string]string{"Referer": loginReferer}
	loginHeaders := []map[string]string{contentType, referer}

	type args struct {
		url     string
		method  string
		headers []map[string]string
		email   string
		pass    string
		isToken bool
	}
	type want struct {
		statusCode int
		nextPage   string
		err        error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "access login page without email, password",
			args: args{
				url:     "/login",
				method:  "POST",
				headers: loginHeaders,
				email:   "",
				pass:    "",
				isToken: true,
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "access login page without email",
			args: args{
				url:     "/login",
				method:  "POST",
				headers: loginHeaders,
				email:   "",
				pass:    "password",
				isToken: true,
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "access login page without password",
			args: args{
				url:     "/login",
				method:  "POST",
				headers: loginHeaders,
				email:   "foobar@gogin.com",
				pass:    "",
				isToken: true,
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "access login page with invalid email",
			args: args{
				url:     "/login",
				method:  "POST",
				headers: loginHeaders,
				email:   "wrong@gogin.com",
				pass:    "password",
				isToken: true,
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "access login page with invalid password",
			args: args{
				url:     "/login",
				method:  "POST",
				headers: loginHeaders,
				email:   "wrong@gogin.com",
				pass:    "hogehoge",
				isToken: true,
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "access login page without token",
			args: args{
				url:     "/login",
				method:  "POST",
				headers: loginHeaders,
				email:   "foobar@gogin.com",
				pass:    "password",
				isToken: false,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				nextPage:   "",
				err:        nil,
			},
		},
		{
			name: "happy path: access login page",
			args: args{
				url:     "/login",
				method:  "POST",
				headers: loginHeaders,
				email:   "foobar@gogin.com",
				pass:    "password",
				isToken: true,
			},
			want: want{
				statusCode: http.StatusOK,
				nextPage:   "/accounts/",
				err:        errRedirect,
			},
		},
	}

	client, ts, err := getClientServer(jwts.AlgoHMAC, true)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("[%s] %s %s", tt.name, tt.args.method, tt.args.url)

			var (
				ginToken string
				postData url.Values
				err      error
			)

			// get gin token
			if tt.args.isToken {
				ginToken, err = token.GetToken(client, fmt.Sprintf("%s%s", ts.URL, tt.args.url))
				if err != nil {
					t.Fatalf("fail to call token.GetToken() %v", err)
				}
			}
			if tt.args.method == "POST" {
				postData = createPostData(tt.args.email, tt.args.pass, ginToken)
				t.Logf("postData: %v", postData)
			}

			req, _ := http.NewRequest(tt.args.method, fmt.Sprintf("%s%s", ts.URL, tt.args.url), bytes.NewBuffer([]byte(postData.Encode())))
			if tt.args.headers != nil {
				httpheader.SetHTTPHeaders(req, tt.args.headers)
				httpheader.Debug(req)
			}
			// request
			res, err := client.Do(req)
			defer func() {
				if res.Body != nil {
					res.Body.Close()
				}
			}()

			urlError, isURLErr := err.(*url.Error)
			if isURLErr && urlError.Err.Error() != tt.want.err.Error() {
				t.Errorf("%s request: actual error: %v, want error: %v", tt.args.url, err, tt.want.err)
				return
			}
			if isURLErr && urlError.Err.Error() == errRedirect.Error() {
				if tt.want.nextPage != res.Header["Location"][0] {
					t.Errorf("%s request: actual next page: %v, want: %v", tt.args.url, res.Header["Location"][0], tt.want.nextPage)
				}
			}
			if err != nil {
				return
			}
			if res.StatusCode != tt.want.statusCode {
				t.Errorf("%s request: actual status code: %v, want: %v", tt.args.url, res.StatusCode, tt.want.statusCode)
			}
		})
	}
}

func TestJWTAPIRequest(t *testing.T) {
	ajaxHeader := map[string]string{"X-Requested-With": "XMLHttpRequest"}
	keyHeader := map[string]string{"X-Custom-Header-Gin": "key12345"}
	contentType := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	jwtHeaders := []map[string]string{ajaxHeader, keyHeader, contentType}

	type args struct {
		url      string
		method   string
		headers  []map[string]string
		email    string
		pass     string
		authMode jwts.JWTAlgo // TODO: how to handle // no, hmac, rsa
	}
	type want struct {
		statusCode int
		err        error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrong method",
			args: args{
				url:      "/api/jwts",
				method:   "GET",
				headers:  jwtHeaders,
				email:    "foobar@gogin.com",
				pass:     "password",
				authMode: jwts.AlgoHMAC,
			},
			want: want{
				statusCode: http.StatusNotFound,
				err:        nil,
			},
		},
		{
			name: "http header without Authorization",
			args: args{
				url:      "/api/jwts",
				method:   "POST",
				headers:  []map[string]string{ajaxHeader, keyHeader},
				email:    "foobar@gogin.com",
				pass:     "password",
				authMode: jwts.AlgoHMAC,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "password is wrong",
			args: args{
				url:      "/api/jwts",
				method:   "POST",
				headers:  jwtHeaders,
				email:    "foobar@gogin.com",
				pass:     "wrong-password",
				authMode: jwts.AlgoHMAC,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "happy path with hmac",
			args: args{
				url:      "/api/jwts",
				method:   "POST",
				headers:  jwtHeaders,
				email:    "foobar@gogin.com",
				pass:     "password",
				authMode: jwts.AlgoHMAC,
			},
			want: want{
				statusCode: http.StatusOK,
				err:        nil,
			},
		},
		{
			name: "happy path with rsa",
			args: args{
				url:      "/api/jwts",
				method:   "POST",
				headers:  jwtHeaders,
				email:    "foobar@gogin.com",
				pass:     "password",
				authMode: jwts.AlgoRSA,
			},
			want: want{
				statusCode: http.StatusOK,
				err:        nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("[%s] %s %s", tt.name, tt.args.method, tt.args.url)

			client, ts, err := getClientServer(tt.args.authMode, true)
			if err != nil {
				t.Fatal(err)
			}
			defer ts.Close()

			postData := createPostData(tt.args.email, tt.args.pass, "")

			req, _ := http.NewRequest(tt.args.method, fmt.Sprintf("%s%s", ts.URL, tt.args.url), bytes.NewBuffer([]byte(postData.Encode())))
			if tt.args.headers != nil {
				httpheader.SetHTTPHeaders(req, tt.args.headers)
			}
			// request
			res, err := client.Do(req)
			defer func() {
				if res.Body != nil {
					res.Body.Close()
				}
			}()

			urlError, isURLErr := err.(*url.Error)
			if isURLErr && urlError.Err.Error() != tt.want.err.Error() {
				t.Errorf("%s request: actual error: %v, want error: %v", tt.args.url, err, tt.want.err)
				return
			}
			if err != nil {
				return
			}
			if res.StatusCode != tt.want.statusCode {
				t.Errorf("%s request: actual status code: %v, want: %v", tt.args.url, res.StatusCode, tt.want.statusCode)
			}

			// get jwts for next request
			if res.StatusCode != http.StatusOK {
				return
			}
			if _, err := jwts.GetJWTResponseToken(res); err != nil {
				t.Errorf("fail to parse response: %v", err)
			}
		})
	}
}

func TestGetUserAPIRequest(t *testing.T) {
	ajaxHeader := map[string]string{"X-Requested-With": "XMLHttpRequest"}
	keyHeader := map[string]string{"X-Custom-Header-Gin": "key12345"}
	wrongKeyHeader := map[string]string{"X-Custom-Header-Gin": "wrong-key"}
	jwtAuthHeader := map[string]string{"Authorization": "Bearer %s"}

	type args struct {
		url     string
		method  string
		headers []map[string]string
		isJWT   bool
	}
	type want struct {
		statusCode int
		err        error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "happy path with user list",
			args: args{
				url:     "/api/users",
				method:  "GET",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusOK,
				err:        nil,
			},
		},
		{
			name: "no jwt token",
			args: args{
				url:     "/api/users",
				method:  "GET",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   false,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "wrong key http header",
			args: args{
				url:     "/api/users",
				method:  "GET",
				headers: []map[string]string{ajaxHeader, wrongKeyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "no key http header",
			args: args{
				url:     "/api/users",
				method:  "GET",
				headers: []map[string]string{ajaxHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "no ajax http header",
			args: args{
				url:     "/api/users",
				method:  "GET",
				headers: []map[string]string{keyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "no http header",
			args: args{
				url:     "/api/users",
				method:  "GET",
				headers: nil,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "body is not included for POST request", // TODO: body is required
			args: args{
				url:     "/api/users",
				method:  "POST",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "not defined method `PUT`",
			args: args{
				url:     "/api/users",
				method:  "PUT",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusNotFound,
				err:        nil,
			},
		},
		{
			name: "not defined method `DELETE`",
			args: args{
				url:     "/api/users",
				method:  "DELETE",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusNotFound,
				err:        nil,
			},
		},
		{
			name: "happy path with user ID",
			args: args{
				url:     "/api/users/id/1",
				method:  "GET",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusOK,
				err:        nil,
			},
		},
		{
			name: "no jwt with user ID",
			args: args{
				url:     "/api/users/id/1",
				method:  "GET",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   false,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		{
			name: "not defined method `POST`",
			args: args{
				url:     "/api/users/id/1",
				method:  "POST",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusNotFound,
				err:        nil,
			},
		},
		{
			name: "body is not included for PUT request", // TODO: body is required
			args: args{
				url:     "/api/users/id/1",
				method:  "PUT",
				headers: []map[string]string{ajaxHeader, keyHeader},
				isJWT:   true,
			},
			want: want{
				statusCode: http.StatusBadRequest,
				err:        nil,
			},
		},
		//{
		//	name: "happy path for DELETE request",
		//	args: args{
		//		url:     "/api/users/id/1",
		//		method:  "DELETE",
		//		headers: []map[string]string{ajaxHeader, keyHeader},
		//		isJWT:   true,
		//	},
		//	want: want{
		//		statusCode: http.StatusOK,
		//		err:        nil,
		//	},
		//},
	}

	client, ts, err := getClientServer(jwts.AlgoHMAC, false)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.Close()

	// get jwt token first
	postData := createPostData("foobar@gogin.com", "password", "")
	token, err := jwts.GetJWTToken(client, fmt.Sprintf("%s%s", ts.URL, "/api/jwts"), postData)
	if err != nil {
		t.Fatal(err)
	}
	jwtAuthHeader["Authorization"] = fmt.Sprintf(jwtAuthHeader["Authorization"], token)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("[%s] %s %s", tt.name, tt.args.method, tt.args.url)

			req, _ := http.NewRequest(tt.args.method, fmt.Sprintf("%s%s", ts.URL, tt.args.url), nil)
			if tt.args.headers != nil {
				httpheader.SetHTTPHeaders(req, tt.args.headers)
			}
			if tt.args.isJWT {
				httpheader.SetHTTPHeaders(req, []map[string]string{jwtAuthHeader})
			}

			// request
			res, err := client.Do(req)
			defer func() {
				if res.Body != nil {
					res.Body.Close()
				}
			}()

			// handle response
			urlError, isURLErr := err.(*url.Error)
			if isURLErr && urlError.Err.Error() != tt.want.err.Error() {
				t.Errorf("%s request: actual error: %v, want error: %v", tt.args.url, err, tt.want.err)
				return
			}
			if err != nil {
				return
			}
			if res.StatusCode != tt.want.statusCode {
				t.Errorf("%s request: actual status code: %v, want: %v", tt.args.url, res.StatusCode, tt.want.statusCode)
			}
		})
	}
}
