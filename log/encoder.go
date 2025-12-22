/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"io"

	jsonx "github.com/hopeio/gox/encoding/json"
	"go.uber.org/zap/zapcore"
)

func DefaultReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := jsonx.NewEncoder(w)
	// For consistency with our custom JSON encoder.
	enc.SetEscapeHTML(false)
	return enc
}
