/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package protogen

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

// 这泛型真别扭
type BaseType interface {
}

func GetOption[T any](desc protoreflect.Descriptor, xt protoreflect.ExtensionType) (T, bool) {
	if desc == nil {
		return *new(T), false
	}
	if !proto.HasExtension(desc.Options(), xt) {
		return *new(T), false
	}

	v, ok := proto.GetExtension(desc.Options(), xt).(T)
	return v, ok
}

func GetOptionWithDefault[T any](desc protoreflect.Descriptor, xt protoreflect.ExtensionType, def T) T {
	v, ok := GetOption[T](desc, xt)
	if !ok {
		return def
	}
	return v
}

func SetExtension[T any](desc protoreflect.Descriptor, extension *protoimpl.ExtensionInfo, value T) {
	if !proto.HasExtension(desc.Options(), extension) {
		return
	}
	proto.SetExtension(desc.Options(), extension, value)
}

func GenerateImport(name string, importPath string, g *protogen.GeneratedFile) string {
	return g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       name,
		GoImportPath: protogen.GoImportPath(importPath),
	})
}

func PrintComments(comments protogen.CommentSet, g *protogen.GeneratedFile) {
	for _, comment := range comments.LeadingDetached {
		g.P(comment)
	}
	g.P(comments.Leading)
	g.P(comments.LeadingDetached)
}
