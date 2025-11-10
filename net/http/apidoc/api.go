package apidoc

import (
	"net/http"
	"reflect"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/hopeio/gox/reflect/structtag"
)

type APIOpts func(*API)

// WithApplyCustomSchemaToType enables customisation of types in the OpenAPI specification.
// Apply customisation to a specific type by checking the t parameter.
// Apply customisations to all types by ignoring the t parameter.
func WithApplyCustomSchemaToType(f func(t reflect.Type, s *openapi3.Schema)) APIOpts {
	return func(api *API) {
		api.ApplyCustomSchemaToType = f
	}
}

// NewAPI creates a new API from the router.
func NewAPI(name string, opts ...APIOpts) *API {
	api := &API{
		Name:       name,
		KnownTypes: defaultKnownTypes,
		Routes:     make(map[Pattern]MethodToRoute),
		// map of model name to schema.
		models:   make(map[string]*openapi3.Schema),
		comments: make(map[string]map[string]string),
	}
	for _, o := range opts {
		o(api)
	}
	return api
}

var defaultKnownTypes = map[reflect.Type]openapi3.Schema{
	reflect.TypeOf(time.Time{}):  *openapi3.NewDateTimeSchema(),
	reflect.TypeOf(&time.Time{}): *openapi3.NewDateTimeSchema().WithNullable(),
}

// Route models a single API route.
type Route struct {
	// Method is the HTTP method of the route, e.g. http.MethodGet
	Method Method
	// Pattern of the route, e.g. /posts/list, or /users/{id}
	Pattern Pattern
	// Params of the route.
	Params Params
	// Models used in the route.
	Models Models
	// Tags used in the route.
	Tags []string
	// OperationID for the route.
	OperationID string
	// Description for the route.
	Description string
}

// Params is a route parameter.
type Params struct {
	// Path parameters are used in the path of the URL, e.g. /users/{id} would
	// have a name of "id".
	Path map[string]Parameter
	// Query parameters are used in the querystring of the URL, e.g. /users/?sort={sortOrder} would
	// have a name of "sort".
	Query  map[string]Parameter
	Header map[string]Parameter
}

type Parameter struct {
	// Description of the param.
	Description string
	// Regexp is a regular expression used to validate the param.
	// An empty string means that no validation is applied.
	Regexp string
	// Required sets whether the querystring parameter must be present in the URL.
	Required bool
	// AllowEmpty sets whether the querystring parameter can be empty.
	AllowEmpty bool
	// Type of the param (string, number, integer, boolean).
	Type string
	// ApplyCustomSchema customises the OpenAPI schema for the query parameter.
	ApplyCustomSchema func(s *openapi3.Parameter)
}

func Type(fieldType reflect.Type) string {
	var typ string
	switch fieldType.Kind() {
	case reflect.Struct:
		typ = openapi3.TypeObject
	case reflect.Ptr:
		typ = openapi3.TypeObject
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		typ = openapi3.TypeInteger
	case reflect.Array, reflect.Slice:
		typ = openapi3.TypeArray
	case reflect.Float32, reflect.Float64:
		typ = openapi3.TypeNumber
	case reflect.String:
		typ = openapi3.TypeString
	case reflect.Bool:
		typ = openapi3.TypeBoolean

	}
	return typ
}

// MethodToRoute maps from a HTTP method to a Route.
type MethodToRoute map[Method]*Route

// Method is the HTTP method of the route, e.g. http.MethodGet
type Method string

// Pattern of the route, e.g. /posts/list, or /users/{id}
type Pattern string

// API is a model of a REST API's routes, along with their
// request and response types.
type API struct {
	// Name of the API.
	Name string
	// Routes of the API.
	// From patterns, to methods, to route.
	Routes map[Pattern]MethodToRoute
	// StripPkgPaths to strip from the type names in the OpenAPI output to avoid
	// leaking internal implementation details such as internal repo names.
	//
	// This increases the risk of type clashes in the OpenAPI output, i.e. two types
	// in different namespaces that are set to be stripped, and have the same type Name
	// could clash.
	//
	// Example values could be "github.com/a-h/rest".
	StripPkgPaths []string

	// Models are the models that are in use in the API.
	// It's possible to customise the models prior to generation of the OpenAPI specification
	// by editing this value.
	models map[string]*openapi3.Schema

	// KnownTypes are added to the OpenAPI specification output.
	// The default implementation:
	//   Maps time.Time to a string.
	KnownTypes map[reflect.Type]openapi3.Schema

	// comments from the package. This can be cleared once the spec has been created.
	comments map[string]map[string]string

	// ApplyCustomSchemaToType callback to customise the OpenAPI specification for a given type.
	// Apply customisation to a specific type by checking the t parameter.
	// Apply customisations to all types by ignoring the t parameter.
	ApplyCustomSchemaToType func(t reflect.Type, s *openapi3.Schema)
}

// Merge route data into the existing configuration.
// This is typically used by adapters, such as the chiadapter
// to take information that the router already knows and add it
// to the specification.
func (api *API) Merge(r Route) {
	toUpdate := api.Route(string(r.Method), string(r.Pattern))
	mergeMap(toUpdate.Params.Path, r.Params.Path)
	mergeMap(toUpdate.Params.Query, r.Params.Query)
	if toUpdate.Models.Request.Type == nil {
		toUpdate.Models.Request = r.Models.Request
	}
	mergeMap(toUpdate.Models.Responses, r.Models.Responses)
}

func mergeMap[TKey comparable, TValue any](into, from map[TKey]TValue) {
	for kf, vf := range from {
		_, ok := into[kf]
		if !ok {
			into[kf] = vf
		}
	}
}

// Spec creates an OpenAPI 3.0 specification document for the API.
func (api *API) Spec() (spec *openapi3.T, err error) {
	spec, err = api.createOpenAPI()
	if err != nil {
		return
	}
	return
}

// Route upserts a route to the API definition.
func (api *API) Route(method, pattern string) (r *Route) {
	methodToRoute, ok := api.Routes[Pattern(pattern)]
	if !ok {
		methodToRoute = make(MethodToRoute)
		api.Routes[Pattern(pattern)] = methodToRoute
	}
	route, ok := methodToRoute[Method(method)]
	if !ok {
		route = &Route{
			Method:  Method(method),
			Pattern: Pattern(pattern),
			Models: Models{
				Responses: make(map[int]Model),
			},
			Params: Params{
				Path:   make(map[string]Parameter),
				Query:  make(map[string]Parameter),
				Header: make(map[string]Parameter),
			},
		}
		methodToRoute[Method(method)] = route
	}
	return route
}

// Get defines a GET request route for the given pattern.
func (api *API) Get(pattern string) (r *Route) {
	return api.Route(http.MethodGet, pattern)
}

// Head defines a HEAD request route for the given pattern.
func (api *API) Head(pattern string) (r *Route) {
	return api.Route(http.MethodHead, pattern)
}

// Post defines a POST request route for the given pattern.
func (api *API) Post(pattern string) (r *Route) {
	return api.Route(http.MethodPost, pattern)
}

// Put defines a PUT request route for the given pattern.
func (api *API) Put(pattern string) (r *Route) {
	return api.Route(http.MethodPut, pattern)
}

// Patch defines a PATCH request route for the given pattern.
func (api *API) Patch(pattern string) (r *Route) {
	return api.Route(http.MethodPatch, pattern)
}

// Delete defines a DELETE request route for the given pattern.
func (api *API) Delete(pattern string) (r *Route) {
	return api.Route(http.MethodDelete, pattern)
}

// Connect defines a CONNECT request route for the given pattern.
func (api *API) Connect(pattern string) (r *Route) {
	return api.Route(http.MethodConnect, pattern)
}

// Options defines an OPTIONS request route for the given pattern.
func (api *API) Options(pattern string) (r *Route) {
	return api.Route(http.MethodOptions, pattern)
}

// Trace defines an TRACE request route for the given pattern.
func (api *API) Trace(pattern string) (r *Route) {
	return api.Route(http.MethodTrace, pattern)
}

// HasResponseModel configures a response for the route.
// Example:
//
//	api.Get("/user").HasResponseModel(http.StatusOK, rest.ModelOf[User]())
func (rm *Route) HasResponseModel(status int, response Model) *Route {
	rm.Models.Responses[status] = response
	return rm
}

// HasResponseModel configures the request model of the route.
// Example:
//
//	api.Post("/user").HasRequestModel(http.StatusOK, rest.ModelOf[User]())
func (rm *Route) HasRequestModel(request Model) *Route {
	rm.Models.Request = request
	return rm
}

// HasPathParameter configures a path parameter for the route.
func (rm *Route) HasPathParameter(name string, p Parameter) *Route {
	rm.Params.Path[name] = p
	return rm
}

// HasQueryParameter configures a query parameter for the route.
func (rm *Route) HasQueryParameter(name string, q Parameter) *Route {
	rm.Params.Query[name] = q
	return rm
}

func (rm *Route) HasRequest(request Model) *Route {
	var hasJson bool
	for j := 0; j < request.Type.NumField(); j++ {
		tags, err := structtag.Parse(string(request.Type.Field(j).Tag))
		if err != nil {
			rm.HasRequestModel(Model{Type: request.Type})
			return rm
		}
		if uri, ok := tags.Get("uri"); ok && uri.Value != "" && uri.Value != "-" {
			rm.HasPathParameter(uri.Value, Parameter{
				Description:       tags.TryGet("comment").Value,
				Regexp:            "",
				Type:              Type(request.Type.Field(j).Type),
				ApplyCustomSchema: nil,
			})
		}
		if uri, ok := tags.Get("path"); ok && uri.Value != "" && uri.Value != "-" {
			rm.HasPathParameter(uri.Value, Parameter{
				Description:       tags.TryGet("comment").Value,
				Regexp:            "",
				Type:              Type(request.Type.Field(j).Type),
				ApplyCustomSchema: nil,
			})
		}
		if query, ok := tags.Get("query"); ok && query.Value != "" && query.Value != "-" {
			rm.HasQueryParameter(query.Value, Parameter{
				Description:       tags.TryGet("comment").Value,
				Regexp:            "",
				Type:              Type(request.Type.Field(j).Type),
				ApplyCustomSchema: nil,
			})
		}
		if header, ok := tags.Get("header"); ok && header.Value != "" && header.Value != "-" {
			rm.HasPathParameter(header.Value, Parameter{
				Description:       tags.TryGet("comment").Value,
				Regexp:            "",
				Type:              Type(request.Type.Field(j).Type),
				ApplyCustomSchema: nil,
			})
		}
		if js := tags.MustGet("json"); !hasJson && js.Value != "" && js.Value != "-" {
			hasJson = true
			rm.HasRequestModel(Model{Type: request.Type})
		}
	}
	return rm
}

// HasTags sets the tags for the route.
func (rm *Route) HasTags(tags []string) *Route {
	rm.Tags = append(rm.Tags, tags...)
	return rm
}

// HasOperationID sets the OperationID for the route.
func (rm *Route) HasOperationID(operationID string) *Route {
	rm.OperationID = operationID
	return rm
}

// HasDescription sets the description for the route.
func (rm *Route) HasDescription(description string) *Route {
	rm.Description = description
	return rm
}

// Models defines the models used by a route.
type Models struct {
	Request   Model
	Responses map[int]Model
}

// ModelOf creates a model of type T.
func ModelOf[T any]() Model {
	var t T
	m := Model{
		Type: reflect.TypeOf(t),
	}
	if sm, ok := any(t).(CustomSchemaApplier); ok {
		m.s = sm.ApplyCustomSchema
	}
	return m
}

func modelFromType(t reflect.Type) Model {
	m := Model{
		Type: t,
	}
	if sm, ok := reflect.New(t).Interface().(CustomSchemaApplier); ok {
		m.s = sm.ApplyCustomSchema
	}
	return m
}

// CustomSchemaApplier is a type that customises its OpenAPI schema.
type CustomSchemaApplier interface {
	ApplyCustomSchema(s *openapi3.Schema)
}

var _ CustomSchemaApplier = Model{}

// Model is a model used in one or more routes.
type Model struct {
	Type reflect.Type
	s    func(s *openapi3.Schema)
}

func (m Model) ApplyCustomSchema(s *openapi3.Schema) {
	if m.s == nil {
		return
	}
	m.s(s)
}
