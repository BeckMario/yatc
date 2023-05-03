// Package statuses provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.5-0.20230403173426-fd06f5aed350 DO NOT EDIT.
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
	Content  string                `json:"content"`
	MediaIds *[]openapi_types.UUID `json:"mediaIds,omitempty"`
}

// StatusResponse defines model for StatusResponse.
type StatusResponse struct {
	Content  string                `json:"content"`
	Id       openapi_types.UUID    `json:"id"`
	MediaIds *[]openapi_types.UUID `json:"mediaIds,omitempty"`
	UserId   openapi_types.UUID    `json:"userId"`
}

// StatusesResponse defines model for StatusesResponse.
type StatusesResponse struct {
	Statuses []StatusResponse `json:"statuses"`
}

// CreateStatusParams defines parameters for CreateStatus.
type CreateStatusParams struct {
	// XUser supplied from api gateway if authenticated, can be set manually locally.
	XUser openapi_types.UUID `json:"X-user"`
}

// CreateStatusJSONRequestBody defines body for CreateStatus for application/json ContentType.
type CreateStatusJSONRequestBody = CreateStatusRequest

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// create a status
	// (POST /statuses)
	CreateStatus(w http.ResponseWriter, r *http.Request, params CreateStatusParams)
	// delete a status by id
	// (DELETE /statuses/{statusId})
	DeleteStatus(w http.ResponseWriter, r *http.Request, statusId openapi_types.UUID)
	// get a status by id
	// (GET /statuses/{statusId})
	GetStatus(w http.ResponseWriter, r *http.Request, statusId openapi_types.UUID)
	// get all statuses of a user
	// (GET /users/{userId}/statuses)
	GetStatuses(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// CreateStatus operation middleware
func (siw *ServerInterfaceWrapper) CreateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params CreateStatusParams

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
		siw.Handler.CreateStatus(w, r, params)
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

// GetStatuses operation middleware
func (siw *ServerInterfaceWrapper) GetStatuses(w http.ResponseWriter, r *http.Request) {
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
		siw.Handler.GetStatuses(w, r, userId)
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
		r.Post(options.BaseURL+"/statuses", wrapper.CreateStatus)
	})
	r.Group(func(r chi.Router) {
		r.Delete(options.BaseURL+"/statuses/{statusId}", wrapper.DeleteStatus)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/statuses/{statusId}", wrapper.GetStatus)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/users/{userId}/statuses", wrapper.GetStatuses)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/9RVT2/bOgz/KgLfOzp1kr4+ZL7tDzDkNmyXAUUPtEzHKmxLleh2RuDvPkh2ktpJ2wAL",
	"hu0UgxFF/v6Q2oLUldE11ewg2YKTBVUYPj9aQqZvjNy4r/TQkGMfNlYbsqwoHJK6ZqrDH/QDK1MSJOBC",
	"jtj9FwG3po9bVW+gi6CiTOE6C1copip85NpWyJBA06jsVNYQQGuxha6LwNJDoyxlkNzuO7nbn9PpPUn2",
	"iTsQzuja0cVQqGycsZz/n2fpYjWjNF3O/pMpzt6tiGYLlBnmq3S5urmB6G2cl2AngsaRXU869DHxVChZ",
	"CBnkdYILEj3Ut1ubUB6OHPgZCr4sAL0igRtOjDD/aymHBP6JDyaNB4fGE03fssf+/uP2/FFV59pXzMhJ",
	"qwwrXUMytO15UBz46wPi/Zc1RPBI1vXnFr68NlSjUZDA9dX86hoiMMhFwBE/R2d0P0gePfpCXqTRtIVU",
	"ixUxWQfJ7bQt1xhTKspEbnUl0CixQaYnbIXKBTZcUM1KIlMWCYm1SEk4YlFh3WBZtqLU0v9egccNCRSE",
	"GVmIoMbKo/w+81LCcwLZNhQN2+EMN3Z3fTI5/qCzdjJj6NuXAXt87zyk7bOrXxP91E7qxlL7TkOgN0ag",
	"fDlfXKyFqe989ak8UpJzeeO57scsC350TVWhbSGBPirwMHmMGzf2qc/YGyfe9l/rrOttWhLTsYs+hfh5",
	"LvLKCZ0fWghe8J49OGFX9Ne9MFJjfjxqB9LEHtKEtB70njSRtqKvd0xdBBs6MWSfif8Obn67U18kfUN8",
	"FuPerH5ruHjbvwPdaOm9LgedLciwmE7IMbw+f7YYdBE5ylLsuPWc4I6VU7L4bLKPO1YbW/qFz2ySOA7v",
	"QKEdJ6v5fA4e/3DFdmxyctDddT8DAAD//ybTZnksCgAA",
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
