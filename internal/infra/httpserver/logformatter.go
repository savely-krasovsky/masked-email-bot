package httpserver

import (
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
)

type logFormatter struct {
	logger *zap.Logger
}

func (l *logFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &logEntry{
		logger:  l.logger,
		request: r,
	}
}

type logEntry struct {
	logger  *zap.Logger
	request *http.Request
}

func (l *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.logger.Info(
		l.request.Method+" "+l.request.RequestURI+" "+l.request.Proto,
		zap.Int("status_code", status),
		zap.Int("bytes", bytes),
		zap.Object("header", headers(header)),
		zap.Duration("elapsed", elapsed),
	)
}

func (l *logEntry) Panic(v interface{}, stack []byte) {
	middleware.PrintPrettyStack(v)
}

type headers map[string][]string

func (h headers) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	for key, value := range h {
		encoder.AddString("key", key)
		if err := encoder.AddArray("value", values(value)); err != nil {
			return err
		}
	}

	return nil
}

type values []string

func (v values) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, value := range v {
		encoder.AppendString(value)
	}

	return nil
}
