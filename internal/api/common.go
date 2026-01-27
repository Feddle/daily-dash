package api

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// NewHTTPClient creates a new HTTP client with the given configuration
func NewHTTPClient(timeout time.Duration, retryAttempts int, logger *zap.Logger) *resty.Client {
	client := resty.New().
		SetTimeout(timeout).
		SetRetryCount(retryAttempts).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryMaxWaitTime(10 * time.Second).
		SetLogger(&restyLogger{logger: logger})

	// Add retry condition
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		// Retry on network errors or 5xx status codes
		if err != nil {
			return true
		}
		return r.StatusCode() >= 500
	})

	return client
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, operation func() error, logger *zap.Logger, service string) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 500 * time.Millisecond
	b.MaxInterval = 10 * time.Second
	b.MaxElapsedTime = 30 * time.Second

	retryWithContext := backoff.WithContext(b, ctx)

	attempt := 0
	return backoff.Retry(func() error {
		attempt++
		err := operation()
		if err != nil {
			logger.Warn("operation failed, retrying",
				zap.String("service", service),
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
			return err
		}
		return nil
	}, retryWithContext)
}

// restyLogger adapts zap.Logger to resty's Logger interface
type restyLogger struct {
	logger *zap.Logger
}

func (l *restyLogger) Errorf(format string, v ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, v...))
}

func (l *restyLogger) Warnf(format string, v ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, v...))
}

func (l *restyLogger) Debugf(format string, v ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, v...))
}
