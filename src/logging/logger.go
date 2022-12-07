package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
	"strings"
	"time"
)

type message struct {
	text   string
	fields []field
}

type field struct {
	name  string
	value string
}

func convertToMessage(msg string, args ...string) *message {
	regex := regexp.MustCompile(`{(\w+)}`)
	matches := regex.FindAllStringSubmatch(msg, -1)

	var fields []field
	text := msg

	for i, v := range matches {
		fields = append(fields, field{
			name:  v[1],
			value: args[i],
		})

		text = strings.Replace(text, v[0], args[i], -1)
	}

	return &message{
		text:   text,
		fields: fields,
	}
}

type logger struct {
	innerLog zerolog.Logger
}

type nilLogger struct {
}

// region nilLogger

func (n *nilLogger) Error(error *error, message string, args ...string) {
}

func (n *nilLogger) Warning(message string, args ...string) {
}

func (n *nilLogger) Information(message string, args ...string) {
}

func (n *nilLogger) Debug(message string, args ...string) {
}

func (n *nilLogger) Trace(message string, args ...string) {
}

// endregion

func (l *logger) Error(error *error, message string, args ...string) {
	logger := l.innerLog.Error()

	if error != nil {
		logger = logger.Err(*error)
	}

	msg := convertToMessage(message, args...)
	for _, v := range msg.fields {
		logger = logger.Str(v.name, v.value)
	}

	logger.Msg(msg.text)
}

func (l *logger) Warning(message string, args ...string) {
	logger := l.innerLog.Warn()

	msg := convertToMessage(message, args...)
	for _, v := range msg.fields {
		logger = logger.Str(v.name, v.value)
	}

	logger.Msg(msg.text)
}

func (l *logger) Information(message string, args ...string) {
	logger := l.innerLog.Info()

	msg := convertToMessage(message, args...)
	for _, v := range msg.fields {
		logger = logger.Str(v.name, v.value)
	}

	logger.Msg(msg.text)
}

func (l *logger) Debug(message string, args ...string) {
	logger := l.innerLog.Debug()

	msg := convertToMessage(message, args...)
	for _, v := range msg.fields {
		logger = logger.Str(v.name, v.value)
	}

	logger.Msg(msg.text)
}

func (l *logger) Trace(message string, args ...string) {
	logger := l.innerLog.Trace()

	msg := convertToMessage(message, args...)
	for _, v := range msg.fields {
		logger = logger.Str(v.name, v.value)
	}

	logger.Msg(msg.text)
}

type Logger interface {
	Error(error *error, message string, args ...string)
	Warning(message string, args ...string)
	Information(message string, args ...string)
	Debug(message string, args ...string)
	Trace(message string, args ...string)
}

type LoggerOptions struct {
	IsProduction bool
	AppName      string
}

func NewLogger(options LoggerOptions) Logger {
	l := log.Logger

	// change to colorful console output when not in production
	if !options.IsProduction {
		l = l.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	l = l.With().
		Str("Application", options.AppName).
		Logger()

	return &logger{innerLog: l}
}

func NilLogger() Logger {
	return &nilLogger{}
}
