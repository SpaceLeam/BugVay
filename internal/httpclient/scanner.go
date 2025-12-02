package httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/time/rate"
)

type Scanner struct {
	client  *http.Client
	limiter *rate.Limiter
}

func NewScanner(rps int, timeout time.Duration) *Scanner {
	return &Scanner{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		},
		limiter: rate.NewLimiter(rate.Limit(rps), rps),
	}
}

func (s *Scanner) DoRequest(ctx context.Context, req *http.Request) (status int, body []byte, err error) {
	// Rate limiting
	if err := s.limiter.Wait(ctx); err != nil {
		return 0, nil, err
	}

	// Retry policy with exponential backoff
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 500 * time.Millisecond
	bo.MaxElapsedTime = 10 * time.Second

	var resp *http.Response
	op := func() error {
		r, err := s.client.Do(req)
		if err != nil {
			return err
		}
		resp = r

		// Treat 5xx as retryable
		if resp.StatusCode >= 500 {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			return errors.New("server error")
		}
		return nil
	}

	err = backoff.Retry(op, bo)
	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // Limit 1MB
	return resp.StatusCode, b, nil
}
