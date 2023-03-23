// Package statuses provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package statuses

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

// CreateStatusRequest defines model for CreateStatusRequest.
type CreateStatusRequest struct {
	Content string             `json:"content"`
	UserId  openapi_types.UUID `json:"userId"`
}

// StatusResponse defines model for StatusResponse.
type StatusResponse struct {
	Content string             `json:"content"`
	Id      openapi_types.UUID `json:"id"`
	UserId  openapi_types.UUID `json:"userId"`
}

// CreateStatusJSONRequestBody defines body for CreateStatus for application/json ContentType.
type CreateStatusJSONRequestBody = CreateStatusRequest

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// get all statuses of a user
	// (GET /statuses)
	GetStatuses(w http.ResponseWriter, r *http.Request)
	// create a status
	// (POST /statuses)
	CreateStatus(w http.ResponseWriter, r *http.Request)
	// delete a status by id
	// (DELETE /statuses/{statusId})
	DeleteStatus(w http.ResponseWriter, r *http.Request, statusId openapi_types.UUID)
	// get a status by id
	// (GET /statuses/{statusId})
	GetStatus(w http.ResponseWriter, r *http.Request, statusId openapi_types.UUID)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetStatuses operation middleware
func (siw *ServerInterfaceWrapper) GetStatuses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetStatuses(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// CreateStatus operation middleware
func (siw *ServerInterfaceWrapper) CreateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CreateStatus(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// DeleteStatus operation middleware
func (siw *ServerInterfaceWrapper) DeleteStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "statusId" -------------
	var statusId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "statusId", runtime.ParamLocationPath, chi.URLParam(r, "statusId"), &statusId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "statusId", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeleteStatus(w, r, statusId)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetStatus operation middleware
func (siw *ServerInterfaceWrapper) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "statusId" -------------
	var statusId openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "statusId", runtime.ParamLocationPath, chi.URLParam(r, "statusId"), &statusId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "statusId", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetStatus(w, r, statusId)
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
		r.Get(options.BaseURL+"/statuses", wrapper.GetStatuses)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/statuses", wrapper.CreateStatus)
	})
	r.Group(func(r chi.Router) {
		r.Delete(options.BaseURL+"/statuses/{statusId}", wrapper.DeleteStatus)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/statuses/{statusId}", wrapper.GetStatus)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8yUT2/bPAzGv4rA9z06tZOuQ+bb/gBDbsN6HHKQZbpWYVuqRHcLAn/3gVLiNE6QtkBR",
	"7GbQosjn95DagjKtNR125CHfglc1tjJ8fnUoCW9JUu9/4kOPnjhsnbHoSGM4pExH2IUf+Ee2tkHIwYcc",
	"sf+XAG1sjDvd3cGQQO/RrcrjLI6J37VWtVChtBdUo4iXQQKVca0kPtjr8vTSIQGHD712WEL+Cw7Fd7XW",
	"Y4Yp7lERt7EX563pPL6ZOj1Rtsg+VmUxX86wKBazD6qQs09LxNlcqlJWy2KxvLl5XuB7UAtHXoCO83RX",
	"Ge6lRK+ctqRNB/kOKRfSFBqMAfH5xwoSeETn47k5CzIWO2k15HB9lV1dQwJWUh3Qp1FB9OEOgwdsjeQ6",
	"DAG+I93uz7CKaGI4v8iyiX3S2karkJzee+5gP+v8pQnbkPi/wwpy+C89bEW6W4l0MizDCEU6JzeRyTEL",
	"3yuF3ld9I8bWA3Lft610G8hZmZBNI/ZqhamEFMydEco7z66MKNZDAtb4MzCeLitET9HTF1NuXgXikv5z",
	"78FwPEDkehxOvJi/WQtTCy4hbza7hSgnyGNUyMOOnOE8JIcJTLfxa1UOcd4bJDy14FuIjxZY6WSLhI6v",
	"nm4JLyNbPbagOcrDDwl0sj28M6sSpoSTJ7SeW+/1+c149ZxG0SM0UWxErHduRC+v67/O5t0n9fLj8ALi",
	"IQvd455n7xrIoSayeZo2RsmmNp7yZZZlwKp3V2yPaaKHYT38DQAA//8JrvUiEQgAAA==",
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
