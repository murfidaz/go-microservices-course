package logging

import (
	"io"
	"log/slog"
)

type Options struct {
	Service string
	Version string
	Level   slog.Level
	Format  string // "text" untuk manusia, "json" untuk log collectors
}

func New(w io.Writer, o Options) *slog.Logger {
	opts := &slog.HandlerOptions{Level: o.Level}
	var h slog.Handler
	if o.Format == "json" {
		h = slog.NewJSONHandler(w, opts)
	} else {
		opts.ReplaceAttr = shortTime
		h = slog.NewTextHandler(w, opts)
	}
	return slog.New(contextHandler{h}).With(
		slog.String("service", o.Service),
		slog.String("version", o.Version),
	)
}

func shortTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		a.Value = slog.StringValue(a.Value.Time().Format("15:04:05.000"))
	}
	return a
}

func ParseLevel(s string) (slog.Level, error) {
	var l slog.Level
	err := l.UnmarshalText([]byte(s))
	return l, err
}
