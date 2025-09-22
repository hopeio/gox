package apidoc

import (
	"github.com/getkin/kin-openapi/openapi3"
	reflectx "github.com/hopeio/gox/reflect"
	"reflect"
	"strings"
)

type ComponentType int

const (
	ComponentTypeParameters ComponentType = iota
	ComponentTypeHeaders
	ComponentTypeRequestBodies
	ComponentTypeResponses
	ComponentTypeSecuritySchemes
	ComponentTypeExamples
	ComponentTypeLinks
	ComponentTypeCallbacks
)

func AddComponent(name string, v interface{}) {
	schema := openapi3.NewSchema()
	schemaRef := openapi3.NewSchemaRef("", schema)
	typs := openapi3.Types{"object"}
	schema.Type = &typs
	schema.Properties = make(map[string]*openapi3.SchemaRef)

	body := reflect.TypeOf(v).Elem()
	var typ, subFieldName string
	for i := range body.NumField() {
		json := strings.Split(body.Field(i).Tag.Get("json"), ",")[0]
		if json == "-" {
			continue
		}
		if json == "" {
			json = body.Field(i).Name
		}
		fieldType := body.Field(i).Type
		switch fieldType.Kind() {
		case reflect.Struct:
			typ = "object"
			v = reflect.ValueOf(v).Elem().Field(i).Addr().Interface()
			subFieldName = fieldType.Name()
		case reflect.Ptr:
			typ = "object"
			v = reflect.New(fieldType.Elem()).Interface()
			subFieldName = fieldType.Elem().Name()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			typ = "integer"
		case reflect.Array, reflect.Slice:
			typ = "array"
			v = reflect.New(reflectx.DerefType(fieldType)).Interface()
			subFieldName = reflectx.DerefType(fieldType).Name()
		case reflect.Float32, reflect.Float64:
			typ = "number"
		case reflect.String:
			typ = "string"
		case reflect.Bool:
			typ = "boolean"

		}

		subSchema := openapi3.SchemaRef{Value: new(openapi3.Schema)}
		if typ == "object" {
			subSchema.Ref = "#/components/schemas/" + subFieldName
			AddComponent(subFieldName, v)
		}
		if typ == "array" {
			subSchema.Value.Type = &openapi3.Types{"array"}
			subSchema.Value.Items = new(openapi3.SchemaRef)
			subSchema.Value.Items.Ref = "#/components/schemas/" + subFieldName
			AddComponent(subFieldName, v)
		}
		schema.Properties[json] = &subSchema
	}
	if Doc.Components == nil {
		Doc.Components = &openapi3.Components{Schemas: make(map[string]*openapi3.SchemaRef)}
	}
	Doc.Components.Schemas[name] = schemaRef
}
