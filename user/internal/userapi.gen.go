// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.5-0.20230403173426-fd06f5aed350 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

// CreateFollowerRequest defines model for CreateFollowerRequest.
type CreateFollowerRequest struct {
	Id openapi_types.UUID `json:"id"`
}

// CreateUserRequest defines model for CreateUserRequest.
type CreateUserRequest struct {
	Username string `json:"username"`
}

// UserResponse defines model for UserResponse.
type UserResponse struct {
	Id       openapi_types.UUID `json:"id"`
	Username string             `json:"username"`
}

// UsersResponse defines model for UsersResponse.
type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

// DeleteUserParams defines parameters for DeleteUser.
type DeleteUserParams struct {
	// XUser supplied from api gateway if authenticated, can be set manually locally. Has to be equal to userId in path
	XUser openapi_types.UUID `json:"X-user"`
}

// FollowUserParams defines parameters for FollowUser.
type FollowUserParams struct {
	// XUser supplied from api gateway if authenticated, can be set manually locally. Has to be equal to userId in path
	XUser openapi_types.UUID `json:"X-user"`
}

// UnfollowUserParams defines parameters for UnfollowUser.
type UnfollowUserParams struct {
	// XUser supplied from api gateway if authenticated, can be set manually locally. Has to be equal to userId in path
	XUser openapi_types.UUID `json:"X-user"`
}

// CreateUserJSONRequestBody defines body for CreateUser for application/json ContentType.
type CreateUserJSONRequestBody = CreateUserRequest

// FollowUserJSONRequestBody defines body for FollowUser for application/json ContentType.
type FollowUserJSONRequestBody = CreateFollowerRequest

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// create a user
	// (POST /users)
	CreateUser(w http.ResponseWriter, r *http.Request)
	// delete a user by id
	// (DELETE /users/{userId})
	DeleteUser(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID, params DeleteUserParams)
	// get a user by id
	// (GET /users/{userId})
	GetUser(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID)
	// get all users which given user follows
	// (GET /users/{userId}/followees)
	GetFollowees(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID)
	// get all followers of a user
	// (GET /users/{userId}/followers)
	GetFollowers(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID)
	// follow a user
	// (POST /users/{userId}/followers)
	FollowUser(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID, params FollowUserParams)
	// unfollow a user
	// (DELETE /users/{userId}/followers/{followerUserId})
	UnfollowUser(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID, followerUserId openapi_types.UUID, params UnfollowUserParams)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// CreateUser operation middleware
func (siw *ServerInterfaceWrapper) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CreateUser(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// DeleteUser operation middleware
func (siw *ServerInterfaceWrapper) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "userId" -------------
	var userId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "userId", runtime.ParamLocationPath, chi.URLParam(r, "userId"), &userId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userId", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params DeleteUserParams

	headers := r.Header

	// ------------- Required header parameter "X-user" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-user")]; found {
		var XUser openapi_types.UUID
		n := len(valueList)
		if n != 1 {
			siw.ErrorHandlerFunc(w, r, &TooManyValuesForParamError{ParamName: "X-user", Count: n})
			return
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "X-user", runtime.ParamLocationHeader, valueList[0], &XUser)
		if err != nil {
			siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "X-user", Err: err})
			return
		}

		params.XUser = XUser

	} else {
		err := fmt.Errorf("Header parameter X-user is required, but not found")
		siw.ErrorHandlerFunc(w, r, &RequiredHeaderError{ParamName: "X-user", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeleteUser(w, r, userId, params)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetUser operation middleware
func (siw *ServerInterfaceWrapper) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "userId" -------------
	var userId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "userId", runtime.ParamLocationPath, chi.URLParam(r, "userId"), &userId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userId", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetUser(w, r, userId)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetFollowees operation middleware
func (siw *ServerInterfaceWrapper) GetFollowees(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "userId" -------------
	var userId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "userId", runtime.ParamLocationPath, chi.URLParam(r, "userId"), &userId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userId", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetFollowees(w, r, userId)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetFollowers operation middleware
func (siw *ServerInterfaceWrapper) GetFollowers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "userId" -------------
	var userId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "userId", runtime.ParamLocationPath, chi.URLParam(r, "userId"), &userId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userId", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetFollowers(w, r, userId)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// FollowUser operation middleware
func (siw *ServerInterfaceWrapper) FollowUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "userId" -------------
	var userId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "userId", runtime.ParamLocationPath, chi.URLParam(r, "userId"), &userId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userId", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params FollowUserParams

	headers := r.Header

	// ------------- Required header parameter "X-user" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-user")]; found {
		var XUser openapi_types.UUID
		n := len(valueList)
		if n != 1 {
			siw.ErrorHandlerFunc(w, r, &TooManyValuesForParamError{ParamName: "X-user", Count: n})
			return
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "X-user", runtime.ParamLocationHeader, valueList[0], &XUser)
		if err != nil {
			siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "X-user", Err: err})
			return
		}

		params.XUser = XUser

	} else {
		err := fmt.Errorf("Header parameter X-user is required, but not found")
		siw.ErrorHandlerFunc(w, r, &RequiredHeaderError{ParamName: "X-user", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.FollowUser(w, r, userId, params)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// UnfollowUser operation middleware
func (siw *ServerInterfaceWrapper) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "userId" -------------
	var userId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "userId", runtime.ParamLocationPath, chi.URLParam(r, "userId"), &userId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userId", Err: err})
		return
	}

	// ------------- Path parameter "followerUserId" -------------
	var followerUserId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "followerUserId", runtime.ParamLocationPath, chi.URLParam(r, "followerUserId"), &followerUserId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "followerUserId", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params UnfollowUserParams

	headers := r.Header

	// ------------- Required header parameter "X-user" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-user")]; found {
		var XUser openapi_types.UUID
		n := len(valueList)
		if n != 1 {
			siw.ErrorHandlerFunc(w, r, &TooManyValuesForParamError{ParamName: "X-user", Count: n})
			return
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "X-user", runtime.ParamLocationHeader, valueList[0], &XUser)
		if err != nil {
			siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "X-user", Err: err})
			return
		}

		params.XUser = XUser

	} else {
		err := fmt.Errorf("Header parameter X-user is required, but not found")
		siw.ErrorHandlerFunc(w, r, &RequiredHeaderError{ParamName: "X-user", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UnfollowUser(w, r, userId, followerUserId, params)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshallingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshallingParamError) Error() string {
	return fmt.Sprintf("Error unmarshalling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshallingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/users", wrapper.CreateUser)
	})
	r.Group(func(r chi.Router) {
		r.Delete(options.BaseURL+"/users/{userId}", wrapper.DeleteUser)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/users/{userId}", wrapper.GetUser)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/users/{userId}/followees", wrapper.GetFollowees)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/users/{userId}/followers", wrapper.GetFollowers)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/users/{userId}/followers", wrapper.FollowUser)
	})
	r.Group(func(r chi.Router) {
		r.Delete(options.BaseURL+"/users/{userId}/followers/{followerUserId}", wrapper.UnfollowUser)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xXSW/bOBT+K8SbOcqW7UwGHt2mLdIGvRUIUCDwgZKeLAYSyXBJahj67wVJWba8xEbS",
	"LAhyiRSK5Pu2J9FLyEQtBUduNCRL0FmJNfW3nxVSgxeiqsQ9qh94a1Eb90AqIVEZhn4ay91f/EVrWSEk",
	"MBn9W+TpeDrANJ0M/slSOvhvijgY0yynxTSdTM/PIYJCqJoaSMBalkMEZiHdam0U43NomggU3lqmMIfk",
	"2hWZdXNEeoOZgSZqIV7pB+BZjYrTGvsg/9fldzRZaeujpbsN9gEIpbUUXONLSRM9nZPf9igxfZiZWxso",
	"Gqz9zd8KC0jgr3idp7gNU9xTqemqUaXoYq/eGmaNG2e8EG7zHHWmmDRMcEgCNseRGU989f8dKh1mjF0V",
	"IZFTySCBs+FoeAYRSGpKDzbu8EsRQuPIUbf/ZQ7JRq4ggENtPol84WZmghvkfhGVsmKZXxbfaFd51UDH",
	"FNkNbtPXwSiLfiCo5rFORuM/BqBviavd11jbLEOtC1tVC5J5tLm3Stu6pmoBCYRRQokNOhk61z0Do1bn",
	"eOkul3kTrKzQ4K7kX/x4K7mkitZovEfXyx1ojjXmpFCiJlQyMqcG7+mCsIJQa0rkxomCeUQyykmKRKMh",
	"NeWWOjaVyNx1SL5RTYxwz/HW0srdB6CEceKyAi6BkECJNPe4QtPBz4FdR2PtV7Sh/dHX2zYrN4+IYiWm",
	"r9tiaKsGbE+qOttJ1Gi3vdbWk86iLeuDia31JF2QUG0rABHMcU9zfUVzis2vKcgLt9hBnedojom822Vx",
	"ET7ZgdEhCy66Se/Bh5M/Q3rTkUdbUlVeDk3uS5aVZM7ukAebgvZ6w6jWjSNmqVPMUh9mPc6sTmSnzc4H",
	"a9Oh6MCJIDjwZt5b0bv8KM6e87C1/UPmpAPX6HUOXG0i8yFRaKzimpgSu9EQrX7Sw7OHs/3Q2yderm6v",
	"jp/VrnjxxtthVXVFan/lPuWPhny286LlJ+XTrUF1twqSVZWjZYxM4tjrUQptkuloeg4OWrvHcttODc6O",
	"jXRpaGbN7wAAAP//1PzGCeoQAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
