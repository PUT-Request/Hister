// SPDX-License-Identifier: AGPL-3.0-or-later

// Package errors provides shared error-handling helpers for the server.
package errors

import (
	"io"

	"github.com/rs/zerolog/log"
)

// LogCloseBody closes an io.Closer and logs any error at warn level.
// It is intended for use with deferred HTTP response body closes:
//
//	defer errors.LogCloseBody(resp.Body)
func LogCloseBody(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Warn().Err(err).Msg("failed to close response body")
	}
}
