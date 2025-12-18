package protobuf

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"google.golang.org/protobuf/types/descriptorpb"
)

type ctx struct {
	rootType reflect.Type
	rootName string
	root     *descriptorpb.DescriptorProto
	built    map[reflect.Type]*descriptorpb.DescriptorProto
	byName   map[string]*descriptorpb.DescriptorProto
	nested   []*descriptorpb.DescriptorProto
	anon     int
}

// DescriptorProtoFromStruct untest
func DescriptorProtoFromStruct(v any) (*descriptorpb.DescriptorProto, error) {
	t := reflect.TypeOf(v)
	if t == nil {
		return nil, fmt.Errorf("nil value")
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not struct: %s", t.Kind())
	}
	c := &ctx{rootType: t, rootName: typeName(t), built: map[reflect.Type]*descriptorpb.DescriptorProto{}, byName: map[string]*descriptorpb.DescriptorProto{}}
	m := buildMessage(t, c)
	m.NestedType = append([]*descriptorpb.DescriptorProto{}, c.nested...)
	return m, nil
}

func buildMessage(t reflect.Type, c *ctx) *descriptorpb.DescriptorProto {
	if d, ok := c.built[t]; ok {
		return d
	}
	name := typeNameWithCtx(t, c)
	d := &descriptorpb.DescriptorProto{}
	d.Name = strPtr(name)
	c.built[t] = d
	if c.root == nil {
		c.root = d
	} else {
		if _, ok := c.byName[name]; !ok {
			c.nested = append(c.nested, d)
			c.byName[name] = d
		}
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		fn := deriveFieldName(f)
		num := int32(len(d.Field) + 1)
		lab := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
		typ := descriptorpb.FieldDescriptorProto_TYPE_STRING
		var typeName string
		ft := f.Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		switch ft.Kind() {
		case reflect.Bool:
			typ = descriptorpb.FieldDescriptorProto_TYPE_BOOL
		case reflect.Int8, reflect.Int16, reflect.Int32:
			typ = descriptorpb.FieldDescriptorProto_TYPE_INT32
		case reflect.Int, reflect.Int64:
			typ = descriptorpb.FieldDescriptorProto_TYPE_INT64
		case reflect.Uint8, reflect.Uint16, reflect.Uint32:
			typ = descriptorpb.FieldDescriptorProto_TYPE_UINT32
		case reflect.Uint, reflect.Uint64:
			typ = descriptorpb.FieldDescriptorProto_TYPE_UINT64
		case reflect.Float32:
			typ = descriptorpb.FieldDescriptorProto_TYPE_FLOAT
		case reflect.Float64:
			typ = descriptorpb.FieldDescriptorProto_TYPE_DOUBLE
		case reflect.String:
			typ = descriptorpb.FieldDescriptorProto_TYPE_STRING
		case reflect.Slice:
			if ft.Elem().Kind() == reflect.Uint8 {
				typ = descriptorpb.FieldDescriptorProto_TYPE_BYTES
			} else {
				lab = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
				et := ft.Elem()
				for et.Kind() == reflect.Ptr {
					et = et.Elem()
				}
				switch et.Kind() {
				case reflect.Bool:
					typ = descriptorpb.FieldDescriptorProto_TYPE_BOOL
				case reflect.Int8, reflect.Int16, reflect.Int32:
					typ = descriptorpb.FieldDescriptorProto_TYPE_INT32
				case reflect.Int, reflect.Int64:
					typ = descriptorpb.FieldDescriptorProto_TYPE_INT64
				case reflect.Uint8, reflect.Uint16, reflect.Uint32:
					typ = descriptorpb.FieldDescriptorProto_TYPE_UINT32
				case reflect.Uint, reflect.Uint64:
					typ = descriptorpb.FieldDescriptorProto_TYPE_UINT64
				case reflect.Float32:
					typ = descriptorpb.FieldDescriptorProto_TYPE_FLOAT
				case reflect.Float64:
					typ = descriptorpb.FieldDescriptorProto_TYPE_DOUBLE
				case reflect.String:
					typ = descriptorpb.FieldDescriptorProto_TYPE_STRING
				case reflect.Struct:
					if et == reflect.TypeOf(time.Time{}) {
						typ = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
						typeName = ".google.protobuf.Timestamp"
					} else {
						child := buildMessage(et, c)
						typ = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
						typeName = child.GetName()
					}
				default:
					typ = descriptorpb.FieldDescriptorProto_TYPE_STRING
				}
			}
		case reflect.Map:
			lab = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
			entry := buildMapEntry(c.rootName, fn, ft, c)
			typ = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
			typeName = entry.GetName()
		case reflect.Struct:
			if ft == reflect.TypeOf(time.Time{}) {
				typ = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
				typeName = ".google.protobuf.Timestamp"
			} else {
				child := buildMessage(ft, c)
				typ = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
				typeName = child.GetName()
			}
		default:
			typ = descriptorpb.FieldDescriptorProto_TYPE_STRING
		}
		fdp := &descriptorpb.FieldDescriptorProto{}
		fdp.Name = strPtr(fn)
		fdp.Number = int32Ptr(num)
		fdp.Label = &lab
		fdp.Type = &typ
		if typeName != "" {
			fdp.TypeName = strPtr(typeName)
		}
		d.Field = append(d.Field, fdp)
	}
	return d
}

func buildMapEntry(rootName, fieldName string, ft reflect.Type, c *ctx) *descriptorpb.DescriptorProto {
	en := rootName + "_" + toPascal(fieldName) + "Entry"
	if d, ok := c.byName[en]; ok {
		return d
	}
	d := &descriptorpb.DescriptorProto{}
	d.Name = strPtr(en)
	b := true
	d.Options = &descriptorpb.MessageOptions{MapEntry: &b}
	ktyp := ft.Key()
	vtyp := ft.Elem()
	for vtyp.Kind() == reflect.Ptr {
		vtyp = vtyp.Elem()
	}
	kLab := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	kNum := int32(1)
	kFd := &descriptorpb.FieldDescriptorProto{}
	kFd.Name = strPtr("key")
	kFd.Number = int32Ptr(kNum)
	kFd.Label = &kLab
	kFd.Type = enumForScalar(ktyp)
	vLab := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	vNum := int32(2)
	vFd := &descriptorpb.FieldDescriptorProto{}
	vFd.Name = strPtr("value")
	vFd.Number = int32Ptr(vNum)
	vFd.Label = &vLab
	switch vtyp.Kind() {
	case reflect.Bool:
		tt := descriptorpb.FieldDescriptorProto_TYPE_BOOL
		vFd.Type = &tt
	case reflect.Int8, reflect.Int16, reflect.Int32:
		tt := descriptorpb.FieldDescriptorProto_TYPE_INT32
		vFd.Type = &tt
	case reflect.Int, reflect.Int64:
		tt := descriptorpb.FieldDescriptorProto_TYPE_INT64
		vFd.Type = &tt
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		tt := descriptorpb.FieldDescriptorProto_TYPE_UINT32
		vFd.Type = &tt
	case reflect.Uint, reflect.Uint64:
		tt := descriptorpb.FieldDescriptorProto_TYPE_UINT64
		vFd.Type = &tt
	case reflect.Float32:
		tt := descriptorpb.FieldDescriptorProto_TYPE_FLOAT
		vFd.Type = &tt
	case reflect.Float64:
		tt := descriptorpb.FieldDescriptorProto_TYPE_DOUBLE
		vFd.Type = &tt
	case reflect.String:
		tt := descriptorpb.FieldDescriptorProto_TYPE_STRING
		vFd.Type = &tt
	case reflect.Slice:
		if vtyp.Elem().Kind() == reflect.Uint8 {
			tt := descriptorpb.FieldDescriptorProto_TYPE_BYTES
			vFd.Type = &tt
		} else {
			tt := descriptorpb.FieldDescriptorProto_TYPE_STRING
			vFd.Type = &tt
		}
	case reflect.Struct:
		if vtyp == reflect.TypeOf(time.Time{}) {
			tt := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
			vFd.Type = &tt
			vFd.TypeName = strPtr(".google.protobuf.Timestamp")
		} else {
			child := buildMessage(vtyp, c)
			tt := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
			vFd.Type = &tt
			vFd.TypeName = strPtr(child.GetName())
		}
	default:
		tt := descriptorpb.FieldDescriptorProto_TYPE_STRING
		vFd.Type = &tt
	}
	d.Field = []*descriptorpb.FieldDescriptorProto{kFd, vFd}
	c.nested = append(c.nested, d)
	c.byName[en] = d
	return d
}

func enumForScalar(t reflect.Type) *descriptorpb.FieldDescriptorProto_Type {
	switch t.Kind() {
	case reflect.Bool:
		tt := descriptorpb.FieldDescriptorProto_TYPE_BOOL
		return &tt
	case reflect.Int8, reflect.Int16, reflect.Int32:
		tt := descriptorpb.FieldDescriptorProto_TYPE_INT32
		return &tt
	case reflect.Int, reflect.Int64:
		tt := descriptorpb.FieldDescriptorProto_TYPE_INT64
		return &tt
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		tt := descriptorpb.FieldDescriptorProto_TYPE_UINT32
		return &tt
	case reflect.Uint, reflect.Uint64:
		tt := descriptorpb.FieldDescriptorProto_TYPE_UINT64
		return &tt
	case reflect.Float32:
		tt := descriptorpb.FieldDescriptorProto_TYPE_FLOAT
		return &tt
	case reflect.Float64:
		tt := descriptorpb.FieldDescriptorProto_TYPE_DOUBLE
		return &tt
	case reflect.String:
		tt := descriptorpb.FieldDescriptorProto_TYPE_STRING
		return &tt
	default:
		tt := descriptorpb.FieldDescriptorProto_TYPE_STRING
		return &tt
	}
}

func typeName(t reflect.Type) string {
	if t.Name() != "" {
		return t.Name()
	}
	return "Struct"
}

func typeNameWithCtx(t reflect.Type, c *ctx) string {
	if t.Name() != "" {
		p := t.PkgPath()
		if p == "" {
			return t.Name()
		}
		s := p + "_" + t.Name()
		s = strings.NewReplacer("/", "_", ".", "_", "-", "_").Replace(s)
		return s
	}
	c.anon++
	return "AnonStruct" + fmt.Sprintf("%d", c.anon)
}

func deriveFieldName(f reflect.StructField) string {
	n := f.Name
	tag := f.Tag.Get("json")
	if tag != "" {
		p := strings.Split(tag, ",")[0]
		if p != "" && p != "-" {
			n = p
		}
	}
	return toSnake(n)
}

func toSnake(s string) string {
	var out []rune
	var prev rune
	for i, r := range s {
		if r == '-' || r == ' ' {
			out = append(out, '_')
			prev = '_'
			continue
		}
		if unicode.IsUpper(r) {
			if i > 0 && prev != '_' {
				out = append(out, '_')
			}
			out = append(out, unicode.ToLower(r))
		} else {
			out = append(out, r)
		}
		prev = r
	}
	return string(out)
}

func toPascal(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
	for i := range parts {
		p := parts[i]
		if p == "" {
			continue
		}
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		parts[i] = string(r)
	}
	return strings.Join(parts, "")
}

func strPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32 { return &i }
