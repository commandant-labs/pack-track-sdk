package packtrack

// Logger is a minimal leveled logger interface.
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// NoopLogger implements Logger and discards logs.
type NoopLogger struct{}

func (NoopLogger) Debugf(string, ...any) {}
func (NoopLogger) Infof(string, ...any)  {}
func (NoopLogger) Warnf(string, ...any)  {}
func (NoopLogger) Errorf(string, ...any) {}
