/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package _go

import (
	"os"
	"strings"

	execx "github.com/hopeio/gox/os/exec"
)

const GoListDir = `go list -m -f {{.Dir}} `
const GOPATHKey = "GOPATH"

var gopath, modPath string

func init() {
	if gopath == "" {
		gopath = os.Getenv(GOPATHKey)
	}
	if gopath != "" && !strings.HasSuffix(gopath, "/") {
		gopath = gopath + "/"
	}
	modPath = gopath + "pkg/mod/"
}

func GetDepDir(dep string) string {
	if !strings.Contains(dep, "@") {
		return modDepDir(dep)
	}
	depPath := modPath + dep
	_, err := os.Stat(depPath)
	if os.IsNotExist(err) {
		depPath = modDepDir(dep)
	}
	return depPath
}

func modDepDir(dep string) string {
	depPath, err := execx.RunGetOut(GoListDir + dep)
	if err != nil || depPath == "" {
		execx.RunGetOut("go get " + dep)
		depPath, _ = execx.RunGetOut(GoListDir + dep)
	}
	return depPath
}
