/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/errors"
	"github.com/hopeio/gox/log"
	httpx "github.com/hopeio/gox/net/http"
)

func SetLog(app *gin.Engine, logger2 *log.Logger, errHandle bool) {
	app.Use(LoggerWithFormatter(httpx.DefaultLogFormatter, logger2, errHandle))

}

// ErrorLogger returns a handlerfunc for any error type.
func ErrorLogger() gin.HandlerFunc {
	return ErrorLoggerT(gin.ErrorTypeAny)
}

// ErrorLoggerT returns a handlerfunc for a given error type.
func ErrorLoggerT(typ gin.ErrorType) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		errs := c.Errors.ByType(typ)
		if len(errs) > 0 {
			data, err := httpx.DefaultCodec.Marshal(errors.InvalidArgument.Msg(strings.Join(errs.Errors(), "\n")))
			if err != nil {
				c.Status(http.StatusInternalServerError)
				c.Abort()
				return
			}
			c.Writer.Write(data)
			return
		}
	}
}

// Logger instances a Logger middleware that will write the logs to gin.DefaultWriter.
// By default gin.DefaultWriter = os.Stdout.
func Logger() gin.HandlerFunc {
	return LoggerWithConfig(httpx.LoggerConfig{})
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f httpx.Formatter, logger *log.Logger, hasErr bool) gin.HandlerFunc {
	return LoggerWithConfig(httpx.LoggerConfig{
		Formatter: f,
		Logger:    logger,
		ErrHandle: hasErr,
	})
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf httpx.LoggerConfig) gin.HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = httpx.DefaultLogFormatter
	}

	logger := conf.Logger
	if logger == nil {
		logger = log.DefaultLogger()
	}

	notlogged := conf.SkipPaths

	errHandle := conf.ErrHandle

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// ErrorLog only when path is not being skipped
		for _, ext := range notlogged {
			if strings.HasSuffix(path, ext) {
				return
			}
		}
		param := httpx.FormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		// Stop timer
		param.TimeStamp = time.Now()
		param.Latency = param.TimeStamp.Sub(start)

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

		param.BodySize = c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		param.Path = path

		if errHandle {
			logger.Warn(formatter(param))
		} else {
			logger.Info(formatter(param))
		}
	}
}
