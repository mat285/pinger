package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/blendlabs/go-util/env"
)

const (
	// DefaultBufferPoolSize is the default buffer pool size.
	DefaultBufferPoolSize = 1 << 8 // 256

	// DefaultTextTimeFormat is the default time format.
	DefaultTextTimeFormat = time.RFC3339

	// DefaultTextWriterUseColor is a default setting for writers.
	DefaultTextWriterUseColor = true
	// DefaultTextWriterShowTimestamp is a default setting for writers.
	DefaultTextWriterShowTimestamp = true
	// DefaultTextWriterShowLabel is a default setting for writers.
	DefaultTextWriterShowLabel = true
)

// TextWritable is a type with a custom formater for text writing.
type TextWritable interface {
	WriteText(formatter TextFormatter, buf *bytes.Buffer)
}

// FlagTextColorProvider is a function types can implement to provide a color.
type FlagTextColorProvider interface {
	FlagTextColor() AnsiColor
}

// TextFormatter formats text.
type TextFormatter interface {
	Colorize(value string, color AnsiColor) string
	ColorizeStatusCode(code int) string
	ColorizeByStatusCode(code int, value string) string
}

// NewTextWriter returns a new text writer for an output.
func NewTextWriter(output io.Writer) *TextWriter {
	return &TextWriter{
		output:        NewInterlockedWriter(output),
		showLabel:     DefaultTextWriterShowLabel,
		showTimestamp: DefaultTextWriterShowTimestamp,
		useColor:      DefaultTextWriterUseColor,
		timeFormat:    DefaultTextTimeFormat,
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
}

// NewTextWriterFromEnv returns a new text writer from the environment.
func NewTextWriterFromEnv() *TextWriter {
	return &TextWriter{
		output:        NewInterlockedWriter(os.Stdout),
		errorOutput:   NewInterlockedWriter(os.Stderr),
		showTimestamp: env.Env().Bool(EnvVarShowTimestamp, DefaultTextWriterShowTimestamp),
		useColor:      env.Env().Bool(EnvVarUseColor, DefaultTextWriterUseColor),
		showLabel:     env.Env().Bool(EnvVarShowLabel, DefaultTextWriterShowLabel),
		label:         env.Env().String(EnvVarLabel),
		timeFormat:    env.Env().String(EnvVarTextTimeFormat, DefaultTextTimeFormat),
		bufferPool:    NewBufferPool(DefaultBufferPoolSize),
	}
}

// TextWriter handles outputting logging events to given writer streams as textual columns.
type TextWriter struct {
	output      io.Writer
	errorOutput io.Writer

	showTimestamp bool
	showLabel     bool
	useColor      bool

	timeFormat string
	label      string

	bufferPool *BufferPool
}

// UseColor is a formatting option.
func (wr *TextWriter) UseColor() bool {
	return wr.useColor
}

// WithUseColor sets a formatting option.
func (wr *TextWriter) WithUseColor(useColor bool) *TextWriter {
	wr.useColor = useColor
	return wr
}

// ShowTimestamp is a formatting option.
func (wr *TextWriter) ShowTimestamp() bool {
	return wr.showTimestamp
}

// WithShowTimestamp sets a formatting option.
func (wr *TextWriter) WithShowTimestamp(showTimestamp bool) *TextWriter {
	wr.showTimestamp = showTimestamp
	return wr
}

// ShowLabel is a formatting option.
func (wr *TextWriter) ShowLabel() bool {
	return wr.showLabel
}

// WithShowLabel sets a formatting option.
func (wr *TextWriter) WithShowLabel(showLabel bool) *TextWriter {
	wr.showLabel = showLabel
	return wr
}

// Label is a formatting option.
func (wr *TextWriter) Label() string {
	return wr.label
}

// WithLabel sets a formatting option.
func (wr *TextWriter) WithLabel(label string) Writer {
	wr.label = label
	return wr
}

// TimeFormat is a formatting option.
func (wr *TextWriter) TimeFormat() string {
	return wr.timeFormat
}

// WithTimeFormat sets a formatting option.
func (wr *TextWriter) WithTimeFormat(timeFormat string) *TextWriter {
	wr.timeFormat = timeFormat
	return wr
}

// Output returns the output.
func (wr *TextWriter) Output() io.Writer {
	return wr.output
}

// WithOutput sets the primary output.
func (wr *TextWriter) WithOutput(output io.Writer) *TextWriter {
	wr.output = NewInterlockedWriter(output)
	return wr
}

// ErrorOutput returns an io.Writer for the error stream.
func (wr *TextWriter) ErrorOutput() io.Writer {
	if wr.errorOutput != nil {
		return wr.errorOutput
	}
	return wr.output
}

// WithErrorOutput sets the error output.
func (wr *TextWriter) WithErrorOutput(errorOutput io.Writer) *TextWriter {
	wr.errorOutput = NewInterlockedWriter(errorOutput)
	return wr
}

// Colorize (optionally) applies a color to a string.
func (wr *TextWriter) Colorize(value string, color AnsiColor) string {
	if wr.useColor {
		return color.Apply(value)
	}
	return value
}

// ColorizeStatusCode adds color to a status code.
func (wr *TextWriter) ColorizeStatusCode(statusCode int) string {
	if wr.useColor {
		return ColorizeStatusCode(statusCode)
	}
	return strconv.Itoa(statusCode)
}

// ColorizeByStatusCode colorizes a string by a status code (green, yellow, red).
func (wr *TextWriter) ColorizeByStatusCode(statusCode int, value string) string {
	if wr.useColor {
		return ColorizeByStatusCode(statusCode, value)
	}
	return value
}

// FormatFlag formats an event flag.
func (wr *TextWriter) FormatFlag(flag Flag, color AnsiColor) string {
	return fmt.Sprintf("[%s]", wr.Colorize(string(flag), color))
}

// FormatLabel returns the app name.
func (wr *TextWriter) FormatLabel() string {
	return fmt.Sprintf("[%s]", wr.Colorize(wr.label, ColorBlue))
}

// FormatTimestamp returns a new timestamp string.
func (wr *TextWriter) FormatTimestamp(optionalTime ...time.Time) string {
	timeFormat := DefaultTextTimeFormat
	if len(wr.timeFormat) > 0 {
		timeFormat = wr.timeFormat
	}
	if len(optionalTime) > 0 {
		return wr.Colorize(optionalTime[0].Format(timeFormat), ColorGray)
	}
	return wr.Colorize(time.Now().UTC().Format(timeFormat), ColorGray)
}

// GetBuffer returns a leased buffer from the buffer pool.
func (wr *TextWriter) GetBuffer() *bytes.Buffer {
	return wr.bufferPool.Get()
}

// PutBuffer adds the leased buffer back to the pool.
// It Should be called in conjunction with `GetBuffer`.
func (wr *TextWriter) PutBuffer(buffer *bytes.Buffer) {
	wr.bufferPool.Put(buffer)
}

// Write writes to stdout.
func (wr *TextWriter) Write(e Event) error {
	return wr.write(wr.Output(), e)
}

// WriteError writes to stderr (or stdout if .errorOutput is unset).
func (wr *TextWriter) WriteError(e Event) error {
	return wr.write(wr.ErrorOutput(), e)
}

func (wr *TextWriter) write(output io.Writer, e Event) error {
	buf := wr.bufferPool.Get()
	defer wr.bufferPool.Put(buf)

	if wr.showTimestamp {
		buf.WriteString(wr.FormatTimestamp(e.Timestamp()))
		buf.WriteRune(RuneSpace)
	}

	if wr.showLabel && len(wr.label) > 0 {
		buf.WriteString(wr.FormatLabel())
		buf.WriteRune(RuneSpace)
	}

	if typed, isTyped := e.(FlagTextColorProvider); isTyped {
		if flagColor := typed.FlagTextColor(); len(flagColor) > 0 {
			buf.WriteString(wr.FormatFlag(e.Flag(), flagColor))
		} else {
			buf.WriteString(wr.FormatFlag(e.Flag(), GetFlagTextColor(e.Flag())))
		}
	} else {
		buf.WriteString(wr.FormatFlag(e.Flag(), GetFlagTextColor(e.Flag())))
	}
	buf.WriteRune(RuneSpace)

	if typed, isTyped := e.(TextWritable); isTyped {
		typed.WriteText(wr, buf)
	} else if typed, isTyped := e.(fmt.Stringer); isTyped {
		buf.WriteString(typed.String())
	}

	buf.WriteRune(RuneNewline)
	_, err := buf.WriteTo(output)
	return err
}
