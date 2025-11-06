/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package apidoc

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"

	"github.com/hopeio/gox/log"
)

var Doc *openapi3.T

// 参数为路径和格式
func GetDoc(realPath, modName string) *openapi3.T {
	if Doc != nil {
		return Doc
	}
	if realPath == "" {
		realPath = "."
	}

	realPath = realPath + modName
	err := os.MkdirAll(realPath, os.ModePerm)
	if err != nil {
		log.Error(err)
	}

	realPath = filepath.Join(realPath, modName+OpenapiEXT)

	apiType := filepath.Ext(realPath)

	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		return generate()
	} else {
		file, err := os.Open(realPath)
		if err != nil {
			log.Error(err)
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			log.Error(err)
		}
		/*var buf bytes.Buffer
		err = json.Compact(&buf, data)
		if err != nil {
			ulog.Error(err)
		}*/
		if apiType == ".json" {
			err = json.Unmarshal(data, &Doc)
			if err != nil {
				log.Error(err)
			}
		} else {
			//var v map[string]interface{}//子类型 json: unsupported type: map[interface{}]interface{}
			//var v interface{} //json: unsupported type: map[interface{}]interface{}
			err = yaml.Unmarshal(data, &Doc)
			if err != nil {
				log.Error(err)
			}
		}
	}
	return Doc
}

func generate() *openapi3.T {
	Doc = &openapi3.T{}
	info := new(openapi3.Info)
	Doc.Info = info

	Doc.OpenAPI = "3.1.0"
	Doc.Paths = openapi3.NewPaths()

	info.Title = "Title"
	info.Description = "Description"
	info.Version = "0.01"
	info.TermsOfService = "TermsOfService"

	var contact openapi3.Contact
	contact.Name = "Contact Name"
	contact.Email = "Contact Mail"
	contact.URL = "Contact URL"
	info.Contact = &contact

	var license openapi3.License
	license.Name = "License Name"
	license.URL = "License URL"
	info.License = &license

	Doc.Servers = []*openapi3.Server{{
		URL: "localhost:80",
	}}
	return Doc
}

func WriteToFile(realPath, modName string) {
	if Doc == nil {
		return
	}
	if realPath == "" {
		realPath = "."
	}

	realPath = realPath + modName
	err := os.MkdirAll(realPath, os.ModePerm)
	if err != nil {
		log.Error(err)
	}

	realPath = filepath.Join(realPath, modName+OpenapiEXT)

	if _, err := os.Stat(realPath); err == nil {
		os.Remove(realPath)
	}
	var file *os.File
	file, err = os.Create(realPath)
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(Doc)
	if err != nil {
		log.Error(err)
	}

	/*b, err := yaml.Marshal(swag.ToDynamicJSON(Doc))
	  if err != nil {
	  	log.Error(err)
	  }
	  if _, err := file.Write(b); err != nil {
	  	log.Error(err)
	  }*/

	Doc = nil
}

func NilDoc() {
	Doc = nil
}
