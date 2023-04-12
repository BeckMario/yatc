// Package media provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.5-0.20230403173426-fd06f5aed350 DO NOT EDIT.
package media

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

// MediaUpload defines model for MediaUpload.
type MediaUpload struct {
	Media openapi_types.File `json:"media"`
}

// MediaUploadResponse defines model for MediaUploadResponse.
type MediaUploadResponse struct {
	MediaId string `json:"mediaId"`
}

// DownloadMediaParams defines parameters for DownloadMedia.
type DownloadMediaParams struct {
	// Compressed boolean which indicates if the compressed version is requested
	Compressed *bool `form:"compressed,omitempty" json:"compressed,omitempty"`
}

// UploadMediaMultipartRequestBody defines body for UploadMedia for multipart/form-data ContentType.
type UploadMediaMultipartRequestBody = MediaUpload

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// upload a media by mediaId
	// (POST /media)
	UploadMedia(w http.ResponseWriter, r *http.Request)
	// download a media by mediaId
	// (GET /media/{mediaId})
	DownloadMedia(w http.ResponseWriter, r *http.Request, mediaId string, params DownloadMediaParams)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// UploadMedia operation middleware
func (siw *ServerInterfaceWrapper) UploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UploadMedia(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// DownloadMedia operation middleware
func (siw *ServerInterfaceWrapper) DownloadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "mediaId" -------------
	var mediaId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "mediaId", runtime.ParamLocationPath, chi.URLParam(r, "mediaId"), &mediaId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "mediaId", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params DownloadMediaParams

	// ------------- Optional query parameter "compressed" -------------

	err = runtime.BindQueryParameter("form", true, false, "compressed", r.URL.Query(), &params.Compressed)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "compressed", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DownloadMedia(w, r, mediaId, params)
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
		r.Post(options.BaseURL+"/media", wrapper.UploadMedia)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/media/{mediaId}", wrapper.DownloadMedia)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/5yTzW7bOhCFX4WYe5dq5NabgLsW3XgRoCjQVeEFLY4sBhLJcEYJDEPvXgz1Y1txgaAr",
	"USKH55tzRmeoQheDR88E+gxUNdiZvHxC68yv2AZj5TWmEDGxw7zZyaYs6pA6w6Dh4LxJJyiATxFBA3Fy",
	"/gjDUEDCl94ltKB/j4VPyMYaNlBMF+2XsnB4xophKK71fyLF4An/wrHLgB/Q3dk7SnLS+TrIHRapSi6y",
	"Cx70SCAdOW5xfldff+yggFdMNJ76LLAhojfRgYbtw+ZhCwVEw01GLBevYiCWpzRgREPAYWxx1hJmJP4W",
	"7EmOVsEz+lzV9S27aBKX4vmn7N+SmKz+T1iDhv/KS6TllGd5HeZw6w2nHvOH0ePM/GWzWambGFtXZery",
	"maTxf5BecswIt2ZTX1VIVPet6mfOAqjvOhkrDeNHZVS2Ux1Oao60ADZHWkKGvRSOrpfn6dAglEe8Y//3",
	"8OavA4gmmQ4Zk9y4nghnVahHYZCpAZ1zhgK86WRELky3BhdXZr0b1LXKIYQWjVdvjasa5bwV45GUqxU3",
	"qMTjhERo1TSFypGaBgftDPbSY/4fJ7JLFdyBmSRhGParUdhutu//jDaMoyBuCJKdTFR9ale5LVsfS05q",
	"Mb3O9st9GhrmqMtSVNsmEOvHzeMWBHWqP9/4D8N++BMAAP//9xXOLNkEAAA=",
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
