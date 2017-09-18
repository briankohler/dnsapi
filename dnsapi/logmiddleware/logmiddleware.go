package logmiddleware

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"net/http"
	"strings"
	"time"
)

// Middleware is a middleware handler that logs the request as it goes in and the response as it goes out.
type Middleware struct {
	// Logger is the log.Logger instance used to log messages with the Logger middleware
	Logger *logrus.Logger
	// Name is the name of the application as recorded in latency metrics
	Name string
}

// NewMiddleware returns a new *Middleware, yay!
func NewMiddleware() *Middleware {
	return NewCustomMiddleware(logrus.InfoLevel, &logrus.TextFormatter{}, "web")
}

// NewCustomMiddleware builds a *Middleware with the given level and formatter
func NewCustomMiddleware(level logrus.Level, formatter logrus.Formatter, name string) *Middleware {
	log := logrus.New()
	log.Level = level
	log.Formatter = formatter

	return &Middleware{Logger: log, Name: name}
}

// NewMiddlewareFromLogger returns a new *Middleware which writes to a given logrus logger.
func NewMiddlewareFromLogger(logger *logrus.Logger, name string) *Middleware {
	return &Middleware{Logger: logger, Name: name}
}

func (l *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	entry := l.Logger.WithFields(logrus.Fields{
		"http_request": r.RequestURI,
		"http_method":  r.Method,
		"remote_host":  strings.Split(r.RemoteAddr, ":")[0],
	})

	// X-Forwarded-For
	if fwdfor := r.Header.Get("X-Forwarded-For"); fwdfor != "" {
		entry = entry.WithField("forwardedfor", fwdfor)
	}

	if reqID := r.Header.Get("X-Request-Id"); reqID != "" {
		entry = entry.WithField("request_id", reqID)
	}

	next(rw, r)

	latency := time.Since(start)
	res := rw.(negroni.ResponseWriter)
	entry.WithFields(logrus.Fields{
		"http_status_code":  res.Status(),
		"request_time_mili": (latency.Nanoseconds() / 1000000),
		"type":              "access_log",
	}).Info("completed handling request")
}
