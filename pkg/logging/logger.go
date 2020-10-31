// Package logging provides an application logging solution by wrapping an existing
// logging library. The functionality is provided as an interface so that applications
// only need to depend on this logic.
// The underlying logging implementation can be easily switched if this is a requirement
package logging

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	gl "github.com/bihe/go-gelf/gelf" // this is the forked version of the library "gopkg.in/Graylog2/go-gelf.v2/gelf"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.binggl.net/monorepo/pkg/config"
)

// --------------------------------------------------------------------------
// interface definition and instance creation
// --------------------------------------------------------------------------

// KeyValue defines a combination of Key and Value
type KeyValue interface {
	Read() []string
}

// optionFunc wraps a func so it satisfies the KeyValue interface.
type keyvalueFunc func() []string

func (f keyvalueFunc) Read() []string {
	return f()
}

// LogV create a KeyValue pair
func LogV(key, value string) KeyValue {
	return keyvalueFunc(func() []string {
		return []string{key, value}
	})
}

// ErrV is a specific error value
func ErrV(err error) KeyValue {
	return keyvalueFunc(func() []string {
		if err == nil {
			return nil
		}
		return []string{ErrorKey, err.Error()}
	})
}

// Logger defines functions to
type Logger interface {
	// implement the io.Closer interface
	io.Closer
	// Info logs a message and values with logging-level INFO
	Info(msg string, keyvals ...KeyValue)
	// Info logs a message and values with logging-level DEBUG
	Debug(msg string, keyvals ...KeyValue)
	// Warn logs a message and values with logging-level WARNING
	Warn(msg string, keyvals ...KeyValue)
	// Error logs a message and values with logging-level ERROR
	Error(msg string, keyvals ...KeyValue)
	// InfoRequest is used to log a *http.Request (with a Trace-ID, if available)
	InfoRequest(msg string, req *http.Request, keyvals ...KeyValue)
	// ErrorRequest is used to log a *http.Request (with a Trace-ID, if available)
	ErrorRequest(msg string, req *http.Request, keyvals ...KeyValue)
}

// NewNop is useful for testing
func NewNop() Logger {
	l := zap.NewNop()
	s := l.Sugar()
	return &zapLogger{
		logger: l,
		sugar:  s,
		gw:     nil,
	}
}

// New returns a Logger instance which is used to perform the logging
func New(conf LogConfig, env config.Environment) Logger {
	var (
		l           *zap.Logger
		s           *zap.SugaredLogger
		err         error
		development bool
		level       zapcore.Level
		gw          gl.Writer
		paths       []string
		logEncoding string
	)

	// the output paths of the logger
	paths = append(paths, conf.FilePath)
	logEncoding = "json"

	if env == config.Development {
		development = true
		paths = append(paths, "stdout")
		logEncoding = "console"
	}

	switch conf.LogLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warning":
		level = zap.WarnLevel
	case "error":
		level = zap.WarnLevel
	default:
		level = zap.WarnLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// if a Graylog server is defined a new writer is created and a new sink is registered
	if conf.GrayLogServer != "" {
		// the value of 8 for the call-depth was determined by experimenting with
		// values from 1 to 10
		gw, err = gl.NewUDPWriter(conf.GrayLogServer, gl.WriterOptions{CallDepth: 8})
		if err != nil {
			panic(fmt.Sprintf("could not create a new gelf UDP writer: %s", err))
		}

		// to send data to the graylog server a new scheme is introduced gelf
		zap.RegisterSink("gelf", func(*url.URL) (zap.Sink, error) {
			return gelfWriterSink{
				Writer: gw,
			}, nil
		})
		// the gelf://server-uri is added to the out-put paths
		paths = append(paths, fmt.Sprintf("gelf://%s", conf.GrayLogServer))
	}

	loggerConfig := zap.Config{
		Level:         zap.NewAtomicLevelAt(level),
		Development:   development,
		Encoding:      logEncoding,
		EncoderConfig: encoderConfig,
		OutputPaths:   paths,
	}

	l, err = loggerConfig.Build(zap.AddCallerSkip(1)) // show the real caller, and not the wrapper for zap!
	if err != nil {
		panic(fmt.Sprintf("could not create new logger instance, %v", err))
	}
	s = l.Sugar()
	// all calls of the golang std logging library are redirected to zap
	zap.RedirectStdLog(l)

	return &zapLogger{
		logger: l,
		sugar:  s,
		conf:   conf,
		gw:     gw,
	}
}

var _ Logger = &zapLogger{} // compiler interface guard

// --------------------------------------------------------------------------
// logger implementation
// --------------------------------------------------------------------------

type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	conf   LogConfig
	gw     gl.Writer
}

func (l *zapLogger) Close() error {
	if l.gw != nil {
		return l.gw.Close()
	}
	return nil
}

func (l *zapLogger) Info(msg string, keyvals ...KeyValue) {
	l.sugar.Infow(msg, l.appendKeys(keyvals...)...)
}

func (l *zapLogger) Debug(msg string, keyvals ...KeyValue) {
	l.sugar.Debugw(msg, l.appendKeys(keyvals...)...)
}

func (l *zapLogger) Warn(msg string, keyvals ...KeyValue) {
	l.sugar.Warnw(msg, l.appendKeys(keyvals...)...)
}

func (l *zapLogger) Error(msg string, keyvals ...KeyValue) {
	l.sugar.Errorw(msg, l.appendKeys(keyvals...)...)
}

func (l *zapLogger) InfoRequest(msg string, req *http.Request, keyvals ...KeyValue) {
	kvs := l.appendKeys(keyvals...)
	reqID := middleware.GetReqID(req.Context())
	if reqID != "" {
		kvs = append(kvs, RequestIDKey)
		kvs = append(kvs, reqID)
	}
	l.sugar.Infow(msg, kvs...)
}

func (l *zapLogger) ErrorRequest(msg string, req *http.Request, keyvals ...KeyValue) {
	kvs := l.appendKeys(keyvals...)
	reqID := middleware.GetReqID(req.Context())
	if reqID != "" {
		kvs = append(kvs, RequestIDKey)
		kvs = append(kvs, reqID)
	}
	l.sugar.Errorw(msg, kvs...)
}

func (l *zapLogger) appendKeys(keyvals ...KeyValue) []interface{} {
	var kvs []interface{}
	kvs = append(kvs, l.stdKeys()...)
	if keyvals != nil && len(keyvals) > 0 {
		for _, kv := range keyvals {
			v := kv.Read()
			for _, e := range v {
				kvs = append(kvs, e)
			}
		}

	}
	return kvs
}

// add common keys to log-messages
func (l *zapLogger) stdKeys() []interface{} {
	var kvs []interface{}
	std := []interface{}{
		ApplicationNameKey, l.conf.Trace.AppName,
		HostIDKey, l.conf.Trace.HostID,
	}
	for _, e := range std {
		kvs = append(kvs, e)
	}
	return kvs
}

// --------------------------------------------------------------------------
// zap-sink implementation
// --------------------------------------------------------------------------

// the gelfWriterSink is used to send log-messages also to a defined graylog server
type gelfWriterSink struct {
	// gelf.Writer is a io.Writer interface already
	gl.Writer
}

// Sync is the method which is defined in the zap WriteSyncer interface
func (gelfWriterSink) Sync() error {
	return nil
}
