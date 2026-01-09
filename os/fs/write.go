/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package fs

import (
	"io"
)

func WriteReader(reader io.Reader, filename string) (int, error) {
	f, _ := Create(filename)
	defer f.Close()
	n, err := io.Copy(f, reader)
	return int(n), err
}

func Write(data []byte, filename string) (int, error) {
	f, _ := Create(filename)
	defer f.Close()
	return f.Write(data)
}
