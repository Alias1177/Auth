package service

type Logger interface {
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Debugw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
	Close() error
}
