// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// Application defines model for Application.
type Application struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// Error defines model for Error.
type Error struct {
	Msg string `json:"msg"`
}

// CreateApplicationJSONBody defines parameters for CreateApplication.
type CreateApplicationJSONBody struct {
	Script string `json:"script"`
}

// CreateApplicationJSONRequestBody defines body for CreateApplication for application/json ContentType.
type CreateApplicationJSONRequestBody CreateApplicationJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// GetApplication request
	GetApplication(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateApplication request with any body
	CreateApplicationWithBody(ctx context.Context, name string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateApplication(ctx context.Context, name string, body CreateApplicationJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetApplication(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetApplicationRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateApplicationWithBody(ctx context.Context, name string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateApplicationRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateApplication(ctx context.Context, name string, body CreateApplicationJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateApplicationRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetApplicationRequest generates requests for GetApplication
func NewGetApplicationRequest(server string, name string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/steady/application/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateApplicationRequest calls the generic CreateApplication builder with application/json body
func NewCreateApplicationRequest(server string, name string, body CreateApplicationJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateApplicationRequestWithBody(server, name, "application/json", bodyReader)
}

// NewCreateApplicationRequestWithBody generates requests for CreateApplication with any type of body
func NewCreateApplicationRequestWithBody(server string, name string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/steady/application/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetApplication request
	GetApplicationWithResponse(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*GetApplicationResponse, error)

	// CreateApplication request with any body
	CreateApplicationWithBodyWithResponse(ctx context.Context, name string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateApplicationResponse, error)

	CreateApplicationWithResponse(ctx context.Context, name string, body CreateApplicationJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateApplicationResponse, error)
}

type GetApplicationResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Application
	JSONDefault  *Error
}

// Status returns HTTPResponse.Status
func (r GetApplicationResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetApplicationResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateApplicationResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *Application
	JSONDefault  *Error
}

// Status returns HTTPResponse.Status
func (r CreateApplicationResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateApplicationResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetApplicationWithResponse request returning *GetApplicationResponse
func (c *ClientWithResponses) GetApplicationWithResponse(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*GetApplicationResponse, error) {
	rsp, err := c.GetApplication(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetApplicationResponse(rsp)
}

// CreateApplicationWithBodyWithResponse request with arbitrary body returning *CreateApplicationResponse
func (c *ClientWithResponses) CreateApplicationWithBodyWithResponse(ctx context.Context, name string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateApplicationResponse, error) {
	rsp, err := c.CreateApplicationWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateApplicationResponse(rsp)
}

func (c *ClientWithResponses) CreateApplicationWithResponse(ctx context.Context, name string, body CreateApplicationJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateApplicationResponse, error) {
	rsp, err := c.CreateApplication(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateApplicationResponse(rsp)
}

// ParseGetApplicationResponse parses an HTTP response from a GetApplicationWithResponse call
func ParseGetApplicationResponse(rsp *http.Response) (*GetApplicationResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetApplicationResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Application
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseCreateApplicationResponse parses an HTTP response from a CreateApplicationWithResponse call
func ParseCreateApplicationResponse(rsp *http.Response) (*CreateApplicationResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateApplicationResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest Application
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /steady/application/{name})
	GetApplication(c *gin.Context, name string)

	// (POST /steady/application/{name})
	CreateApplication(c *gin.Context, name string)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

// GetApplication operation middleware
func (siw *ServerInterfaceWrapper) GetApplication(c *gin.Context) {

	var err error

	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameter("simple", false, "name", c.Param("name"), &name)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter name: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetApplication(c, name)
}

// CreateApplication operation middleware
func (siw *ServerInterfaceWrapper) CreateApplication(c *gin.Context) {

	var err error

	// ------------- Path parameter "name" -------------
	var name string

	err = runtime.BindStyledParameter("simple", false, "name", c.Param("name"), &name)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter name: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.CreateApplication(c, name)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL      string
	Middlewares  []MiddlewareFunc
	ErrorHandler func(*gin.Context, error, int)
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router *gin.Engine, si ServerInterface) *gin.Engine {
	return RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router *gin.Engine, si ServerInterface, options GinServerOptions) *gin.Engine {

	errorHandler := options.ErrorHandler

	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandler:       errorHandler,
	}

	router.GET(options.BaseURL+"/steady/application/:name", wrapper.GetApplication)

	router.POST(options.BaseURL+"/steady/application/:name", wrapper.CreateApplication)

	return router
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8xSTY/TQAz9K5HhGDVZuOXGlxBnjqs9eGfc7CzNzKztIKoq/x150tK0jYSQEOI0yfjZ",
	"8/zeO4BLQ06Rogp0BxD3RAOWz3c574JDDSnab+aUiTVQKUYcyE7dZ4IORDnEHqYacmJdFEJU6olhmmpg",
	"ehkDk4fufu4/oh/qEzo9PpNTG/OJOfHts4P0K69ezTbQ7UxDhbhNZSi6b9hb1SMNKUINGnRnF1+V0O+r",
	"j6f778RSFIB2027ujFvKFDEH6ODtpt20tgbqU+HXSGlv8Kxdc7BdJ6v2VKSxhUrpi4cOPpMulbZVJKco",
	"88Jv2tYOl6JSLN3L0c8yezO7Zl+vmbbQwavmbGtz9LRZPlPk8CSOQ54dXg6uThygoLY47vSv0ZitXSEw",
	"RvqRySn5is6YjIwDKbFAd3+4alHspdJUbcNOiavHPZjJ0BVHoD7G9JS2c0iUR6oXfK8D9WDRlBW3PjCh",
	"0rVhLyOJvk9+/0ciXUZ7Xuv36T7i1gN+qc6CpomE3tvxK/KXckw3ybv7V8mT0TkS+X9SN03TzwAAAP//",
	"mig8VhkFAAA=",
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
