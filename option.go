package schedulerx

import "log/slog"

type Option func(t *Parser) error

func WithLogger(logger slog.Logger) Option {
	return func(t *Parser) error {
		t.logger = logger
		return nil
	}
}
