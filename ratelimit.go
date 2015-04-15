// Based https://github.com/PuerkitoBio/throttled/blob/master/rate.go:
//
// Copyright (c) 2014, Martin Angers
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
// * list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// * Neither the name of the author nor the names of its contributors may be used
//   to endorse or promote products derived from this software without specific
//   prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// --
// Modifications by Dmitry Chestnykh are in public domain.

// Package webrate contains a modified RateLimit function for
// http://github.com/PuerkitoBio/throttled, which rate limits only specified
// HTTP methods and doesn't add headers to response.
package webrate

import (
	"net/http"
	"time"

	"github.com/PuerkitoBio/throttled"
)

// Static check to ensure that rateLimiter implements Limiter.
var _ throttled.Limiter = (*rateLimiter)(nil)

// RateLimit creates a throttler that limits the number of requests allowed
// in a certain time window defined by the Quota q. The q parameter specifies
// the requests per time window, and it is silently set to at least 1 request
// and at least a 1 second window if it is less than that. The time window
// starts when the first request is made outside an existing window. Fractions
// of seconds are not supported, they are truncated.
//
// The vary parameter indicates what criteria should be used to group requests
// for which the limit must be applied (ex.: rate limit based on the remote address).
// See varyby.go for the various options.
//
// The specified store is used to keep track of the request count and the
// time remaining in the window. The throttled package comes with some stores
// in the throttled/store package. Custom stores can be created too, by implementing
// the Store interface.
//
// Requests that bust the rate limit are denied access and go through the denied handler,
// which may be specified on the Throttler and that defaults to the package-global
// variable DefaultDeniedHandler.
func RateLimit(q throttled.Quota, methods []string, vary *throttled.VaryBy, store throttled.Store) *throttled.Throttler {
	// Extract requests and window
	reqs, win := q.Quota()

	// Create and return the throttler
	return throttled.Custom(&rateLimiter{
		reqs:    reqs,
		window:  win,
		vary:    vary,
		store:   store,
		methods: methods,
	})
}

// The rate limiter implements limiting the request to a certain quota
// based on the vary-by criteria. State is saved in the store.
type rateLimiter struct {
	reqs    int
	window  time.Duration
	methods []string
	vary    *throttled.VaryBy
	store   throttled.Store
}

func (r *rateLimiter) hasMethod(method string) bool {
	for _, s := range r.methods {
		if s == method {
			return true
		}
	}
	return false
}

// Start initializes the limiter for execution.
func (r *rateLimiter) Start() {
	if r.reqs < 1 {
		r.reqs = 1
	}
	if r.window < time.Second {
		r.window = time.Second
	}
}

// Limit is called for each request to the throttled handler. It checks if
// the request can go through and signals it via the returned channel.
// It returns an error if the operation fails.
func (r *rateLimiter) Limit(w http.ResponseWriter, req *http.Request) (<-chan bool, error) {
	// Create return channel and initialize
	ch := make(chan bool, 1)
	ok := true
	key := r.vary.Key(req)

	if r.hasMethod(req.Method) {
		// Get the current count and remaining seconds
		cnt, secs, err := r.store.Incr(key, r.window)
		// Handle the possible situations: error, begin new window, or increment current window.
		switch {
		case err != nil && err != throttled.ErrNoSuchKey:
			// An unexpected error occurred
			return nil, err
		case err == throttled.ErrNoSuchKey || secs <= 0:
			// Reset counter
			if err := r.store.Reset(key, r.window); err != nil {
				return nil, err
			}
			cnt = 1
			secs = int(r.window.Seconds())
		default:
			// If the limit is reached, deny access
			if cnt > r.reqs {
				ok = false
			}
		}
	}
	// Send response via the return channel
	ch <- ok
	return ch, nil
}
