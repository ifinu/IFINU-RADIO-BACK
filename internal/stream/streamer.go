package stream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ifinu/radio-api/internal/config"
	"github.com/rs/zerolog/log"
)

type Streamer struct {
	cfg    *config.Config
	client *http.Client
}

func NewStreamer(cfg *config.Config) *Streamer {
	return &Streamer{
		cfg: cfg,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        cfg.MaxIdleConns,
				MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
				IdleConnTimeout:     cfg.IdleConnTimeout,
				DisableKeepAlives:   false,
			},
			Timeout: 0, // No timeout for streaming
		},
	}
}

// Stream proxies radio stream to client with buffering and retry logic
func (s *Streamer) Stream(ctx context.Context, w http.ResponseWriter, streamURL string) error {
	var lastErr error

	for attempt := 1; attempt <= s.cfg.StreamRetryAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		log.Info().
			Str("url", streamURL).
			Int("attempt", attempt).
			Msg("Connecting to radio stream")

		err := s.streamWithRetry(ctx, w, streamURL)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Warn().
			Err(err).
			Str("url", streamURL).
			Int("attempt", attempt).
			Msg("Stream attempt failed")

		if attempt < s.cfg.StreamRetryAttempts {
			// Exponential backoff
			backoff := s.cfg.StreamRetryDelay * time.Duration(attempt)
			log.Info().
				Dur("backoff", backoff).
				Msg("Retrying after backoff")

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("stream failed after %d attempts: %w", s.cfg.StreamRetryAttempts, lastErr)
}

func (s *Streamer) streamWithRetry(ctx context.Context, w http.ResponseWriter, streamURL string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", streamURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Set headers to get stream
	req.Header.Set("User-Agent", "IFINU-Radio/1.0")
	req.Header.Set("Icy-MetaData", "1")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("connect to stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Set response headers
	s.setStreamHeaders(w, resp)

	// Create buffer for streaming
	buffer := make([]byte, s.cfg.StreamBufferSize)

	// Stream with buffered copy
	log.Info().
		Str("url", streamURL).
		Int("buffer_size", s.cfg.StreamBufferSize).
		Msg("Starting buffered stream")

	// Use io.CopyBuffer for efficient streaming
	written, err := io.CopyBuffer(w, resp.Body, buffer)

	log.Info().
		Int64("bytes_written", written).
		Err(err).
		Msg("Stream ended")

	if err != nil {
		// Check if it's a client disconnect (not an error from our perspective)
		if isClientDisconnect(err) {
			log.Info().Msg("Client disconnected")
			return nil
		}
		return fmt.Errorf("stream copy: %w", err)
	}

	return nil
}

func (s *Streamer) setStreamHeaders(w http.ResponseWriter, resp *http.Response) {
	// Copy content type
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	} else {
		w.Header().Set("Content-Type", "audio/mpeg")
	}

	// Set streaming headers
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Connection", "keep-alive")

	// Copy ICY headers if present
	for key, values := range resp.Header {
		if len(key) > 4 && key[:4] == "Icy-" {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
	}
}

func isClientDisconnect(err error) bool {
	if err == nil {
		return false
	}

	// Check for common client disconnect errors
	errStr := err.Error()
	return errStr == "context canceled" ||
		errStr == "write: broken pipe" ||
		errStr == "write: connection reset by peer"
}
