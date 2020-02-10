/**
 *  Copyright 2015 Paul Querna
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package cacheheader

import (
	"errors"
	"math"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

// TODO: more extensions from here: http://www.iana.org/assignments/http-cache-directives/http-cache-directives.xhtml

// Directive errors
var (
	ErrQuoteMismatch         = errors.New("missing closing quote")
	ErrMaxAgeDeltaSeconds    = errors.New("failed to parse delta-seconds in `max-age`")
	ErrSMaxAgeDeltaSeconds   = errors.New("failed to parse delta-seconds in `s-maxage`")
	ErrMaxStaleDeltaSeconds  = errors.New("failed to parse delta-seconds in `min-fresh`")
	ErrMinFreshDeltaSeconds  = errors.New("failed to parse delta-seconds in `min-fresh`")
	ErrNoCacheNoArgs         = errors.New("unexpected argument to `no-cache`")
	ErrNoStoreNoArgs         = errors.New("unexpected argument to `no-store`")
	ErrNoTransformNoArgs     = errors.New("unexpected argument to `no-transform`")
	ErrOnlyIfCachedNoArgs    = errors.New("unexpected argument to `only-if-cached`")
	ErrMustRevalidateNoArgs  = errors.New("unexpected argument to `must-revalidate`")
	ErrPublicNoArgs          = errors.New("unexpected argument to `public`")
	ErrProxyRevalidateNoArgs = errors.New("unexpected argument to `proxy-revalidate`")
	// Experimental
	ErrImmutableNoArgs                  = errors.New("unexpected argument to `immutable`")
	ErrStaleIfErrorDeltaSeconds         = errors.New("failed to parse delta-seconds in `stale-if-error`")
	ErrStaleWhileRevalidateDeltaSeconds = errors.New("failed to parse delta-seconds in `stale-while-revalidate`")
)

// Cacheable HTTP header directives
const (
	HeaderNoCache      = "no-cache"
	HeaderNoStore      = "no-store"
	HeaderMaxAge       = "max-age"
	HeaderMaxStale     = "max-stale"
	HeaderMinFresh     = "min-fresh"
	HeaderNoTransform  = "no-transform"
	HeaderOnlyIfCached = "only-if-cached"
)

func whitespace(b byte) bool {
	if b == '\t' || b == ' ' {
		return true
	}
	return false
}

func parse(value string, cd cacheDirective) error {
	var err error = nil
	i := 0

	for i < len(value) && err == nil {
		// eat leading whitespace or commas
		if whitespace(value[i]) || value[i] == ',' {
			i++
			continue
		}

		j := i + 1

		for j < len(value) {
			if !isToken(value[j]) {
				break
			}
			j++
		}

		token := strings.ToLower(value[i:j])
		tokenHasFields := hasFieldNames(token)
		/*
			println("GOT TOKEN:")
			println("	i -> ", i)
			println("	j -> ", j)
			println("	token -> ", token)
		*/

		if j+1 < len(value) && value[j] == '=' {
			k := j + 1
			// minimum size two bytes of "", but we let httpUnquote handle it.
			if k < len(value) && value[k] == '"' {
				eaten, result := httpUnquote(value[k:])
				if eaten == -1 {
					return ErrQuoteMismatch
				}
				i = k + eaten

				err = cd.addPair(token, result)
			} else {
				z := k
				for z < len(value) {
					if tokenHasFields {
						if whitespace(value[z]) {
							break
						}
					} else {
						if whitespace(value[z]) || value[z] == ',' {
							break
						}
					}
					z++
				}
				i = z

				result := value[k:z]
				if result != "" && result[len(result)-1] == ',' {
					result = result[:len(result)-1]
				}

				err = cd.addPair(token, result)
			}
		} else {
			if token != "," {
				err = cd.addToken(token)
			}
			i = j
		}
	}

	return err
}

// DeltaSeconds specifies a non-negative integer, representing
// time in seconds: http://tools.ietf.org/html/rfc7234#section-1.2.1
//
// When set to -1, this means unset.
//
type DeltaSeconds int32

// ParseDeltaSeconds is parser for delta-seconds, a uint31, more or less:
// http://tools.ietf.org/html/rfc7234#section-1.2.1
func ParseDeltaSeconds(v string) (DeltaSeconds, error) {
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				return DeltaSeconds(math.MaxInt32), nil
			}
		}
		return DeltaSeconds(-1), err
	}

	if n > math.MaxInt32 {
		return DeltaSeconds(math.MaxInt32), nil
	}
	return DeltaSeconds(n), nil
}

// FieldNames that present in a header.
type FieldNames map[string]bool

// internal interface for shared methods of RequestCacheDirectives and ResponseCacheDirectives
type cacheDirective interface {
	addToken(s string) error
	addPair(s string, v string) error
}

// RequestCacheDirectives is representation of possible request directives in a `Cache-Control` header: http://tools.ietf.org/html/rfc7234#section-5.2.1
//
// Note: Many fields will be `nil` in practice.
//
type RequestCacheDirectives struct {

	// max-age(delta seconds): http://tools.ietf.org/html/rfc7234#section-5.2.1.1
	//
	// The "max-age" request directive indicates that the client is
	// unwilling to accept a response whose age is greater than the
	// specified number of seconds.  Unless the max-stale request directive
	// is also present, the client is not willing to accept a stale
	// response.
	MaxAge DeltaSeconds

	// max-stale(delta seconds): http://tools.ietf.org/html/rfc7234#section-5.2.1.2
	//
	// The "max-stale" request directive indicates that the client is
	// willing to accept a response that has exceeded its freshness
	// lifetime.  If max-stale is assigned a value, then the client is
	// willing to accept a response that has exceeded its freshness lifetime
	// by no more than the specified number of seconds.  If no value is
	// assigned to max-stale, then the client is willing to accept a stale
	// response of any age.
	MaxStale DeltaSeconds

	// min-fresh(delta seconds): http://tools.ietf.org/html/rfc7234#section-5.2.1.3
	//
	// The "min-fresh" request directive indicates that the client is
	// willing to accept a response whose freshness lifetime is no less than
	// its current age plus the specified time in seconds.  That is, the
	// client wants a response that will still be fresh for at least the
	// specified number of seconds.
	MinFresh DeltaSeconds

	// no-cache(bool): http://tools.ietf.org/html/rfc7234#section-5.2.1.4
	//
	// The "no-cache" request directive indicates that a cache MUST NOT use
	// a stored response to satisfy the request without successful
	// validation on the origin server.
	NoCache bool

	// no-store(bool): http://tools.ietf.org/html/rfc7234#section-5.2.1.5
	//
	// The "no-store" request directive indicates that a cache MUST NOT
	// store any part of either this request or any response to it.  This
	// directive applies to both private and shared caches.
	NoStore bool

	// no-transform(bool): http://tools.ietf.org/html/rfc7234#section-5.2.1.6
	//
	// The "no-transform" request directive indicates that an intermediary
	// (whether or not it implements a cache) MUST NOT transform the
	// payload, as defined in Section 5.7.2 of RFC7230.
	NoTransform bool

	// only-if-cached(bool): http://tools.ietf.org/html/rfc7234#section-5.2.1.7
	//
	// The "only-if-cached" request directive indicates that the client only
	// wishes to obtain a stored response.
	OnlyIfCached bool

	// Extensions: http://tools.ietf.org/html/rfc7234#section-5.2.3
	//
	// The Cache-Control header field can be extended through the use of one
	// or more cache-extension tokens, each with an optional value.  A cache
	// MUST ignore unrecognized cache directives.
	Extensions []string
}

func (cd *RequestCacheDirectives) addToken(token string) error {
	var err error

	switch token {
	case HeaderMaxAge:
		err = ErrMaxAgeDeltaSeconds
	case HeaderMaxStale:
		err = ErrMaxStaleDeltaSeconds
	case HeaderMinFresh:
		err = ErrMinFreshDeltaSeconds
	case HeaderNoCache:
		cd.NoCache = true
	case HeaderNoStore:
		cd.NoStore = true
	case HeaderNoTransform:
		cd.NoTransform = true
	case "only-if-cached":
		cd.OnlyIfCached = true
	default:
		cd.Extensions = append(cd.Extensions, token)
	}
	return err
}

func (cd *RequestCacheDirectives) addPair(token string, v string) error {
	var err error = nil

	switch token {
	case HeaderMaxAge:
		cd.MaxAge, err = ParseDeltaSeconds(v)
		if err != nil {
			err = ErrMaxAgeDeltaSeconds
		}
	case HeaderMaxStale:
		cd.MaxStale, err = ParseDeltaSeconds(v)
		if err != nil {
			err = ErrMaxStaleDeltaSeconds
		}
	case HeaderMinFresh:
		cd.MinFresh, err = ParseDeltaSeconds(v)
		if err != nil {
			err = ErrMinFreshDeltaSeconds
		}
	case HeaderNoCache:
		err = ErrNoCacheNoArgs
	case HeaderNoStore:
		err = ErrNoStoreNoArgs
	case HeaderNoTransform:
		err = ErrNoTransformNoArgs
	case "only-if-cached":
		err = ErrOnlyIfCachedNoArgs
	default:
		// TODO(pquerna): this sucks, making user re-parse
		cd.Extensions = append(cd.Extensions, token+"="+v)
	}

	return err
}

// ParseRequestCacheControl Parses a Cache Control Header from a Request into a set of directives.
func ParseRequestCacheControl(value string) (*RequestCacheDirectives, error) {
	cd := &RequestCacheDirectives{
		MaxAge:   -1,
		MaxStale: -1,
		MinFresh: -1,
	}

	err := parse(value, cd)
	if err != nil {
		return nil, err
	}
	return cd, nil
}

// ResponseCacheDirectives  Representation of possible response directives in a `Cache-Control` header: http://tools.ietf.org/html/rfc7234#section-5.2.2
//
// Note: Many fields will be `nil` in practice.
//
type ResponseCacheDirectives struct {

	// must-revalidate(bool): http://tools.ietf.org/html/rfc7234#section-5.2.2.1
	//
	// The "must-revalidate" response directive indicates that once it has
	// become stale, a cache MUST NOT use the response to satisfy subsequent
	// requests without successful validation on the origin server.
	MustRevalidate bool

	// no-cache(FieldName): http://tools.ietf.org/html/rfc7234#section-5.2.2.2
	//
	// The "no-cache" response directive indicates that the response MUST
	// NOT be used to satisfy a subsequent request without successful
	// validation on the origin server.
	//
	// If the no-cache response directive specifies one or more field-names,
	// then a cache MAY use the response to satisfy a subsequent request,
	// subject to any other restrictions on caching.  However, any header
	// fields in the response that have the field-name(s) listed MUST NOT be
	// sent in the response to a subsequent request without successful
	// revalidation with the origin server.
	NoCache FieldNames

	// no-cache(cast-to-bool): http://tools.ietf.org/html/rfc7234#section-5.2.2.2
	//
	// While the RFC defines optional field-names on a no-cache directive,
	// many applications only want to know if any no-cache directives were
	// present at all.
	NoCachePresent bool

	// no-store(bool): http://tools.ietf.org/html/rfc7234#section-5.2.2.3
	//
	// The "no-store" request directive indicates that a cache MUST NOT
	// store any part of either this request or any response to it.  This
	// directive applies to both private and shared caches.
	NoStore bool

	// no-transform(bool): http://tools.ietf.org/html/rfc7234#section-5.2.2.4
	//
	// The "no-transform" response directive indicates that an intermediary
	// (regardless of whether it implements a cache) MUST NOT transform the
	// payload, as defined in Section 5.7.2 of RFC7230.
	NoTransform bool

	// public(bool): http://tools.ietf.org/html/rfc7234#section-5.2.2.5
	//
	// The "public" response directive indicates that any cache MAY store
	// the response, even if the response would normally be non-cacheable or
	// cacheable only within a private cache.
	Public bool

	// private(FieldName): http://tools.ietf.org/html/rfc7234#section-5.2.2.6
	//
	// The "private" response directive indicates that the response message
	// is intended for a single user and MUST NOT be stored by a shared
	// cache.  A private cache MAY store the response and reuse it for later
	// requests, even if the response would normally be non-cacheable.
	//
	// If the private response directive specifies one or more field-names,
	// this requirement is limited to the field-values associated with the
	// listed response header fields.  That is, a shared cache MUST NOT
	// store the specified field-names(s), whereas it MAY store the
	// remainder of the response message.
	Private FieldNames

	// private(cast-to-bool): http://tools.ietf.org/html/rfc7234#section-5.2.2.6
	//
	// While the RFC defines optional field-names on a private directive,
	// many applications only want to know if any private directives were
	// present at all.
	PrivatePresent bool

	// proxy-revalidate(bool): http://tools.ietf.org/html/rfc7234#section-5.2.2.7
	//
	// The "proxy-revalidate" response directive has the same meaning as the
	// must-revalidate response directive, except that it does not apply to
	// private caches.
	ProxyRevalidate bool

	// max-age(delta seconds): http://tools.ietf.org/html/rfc7234#section-5.2.2.8
	//
	// The "max-age" response directive indicates that the response is to be
	// considered stale after its age is greater than the specified number
	// of seconds.
	MaxAge DeltaSeconds

	// s-maxage(delta seconds): http://tools.ietf.org/html/rfc7234#section-5.2.2.9
	//
	// The "s-maxage" response directive indicates that, in shared caches,
	// the maximum age specified by this directive overrides the maximum age
	// specified by either the max-age directive or the Expires header
	// field.  The s-maxage directive also implies the semantics of the
	// proxy-revalidate response directive.
	SMaxAge DeltaSeconds

	////
	// Experimental features
	// - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Extension_Cache-Control_directives
	// - https://www.fastly.com/blog/stale-while-revalidate-stale-if-error-available-today
	////

	// immutable(cast-to-bool): experimental feature
	Immutable bool

	// stale-if-error(delta seconds): experimental feature
	StaleIfError DeltaSeconds

	// stale-while-revalidate(delta seconds): experimental feature
	StaleWhileRevalidate DeltaSeconds

	// Extensions: http://tools.ietf.org/html/rfc7234#section-5.2.3
	//
	// The Cache-Control header field can be extended through the use of one
	// or more cache-extension tokens, each with an optional value.  A cache
	// MUST ignore unrecognized cache directives.
	Extensions []string
}

//ParseResponseCacheControl will Parses a Cache Control Header from a Response into a set of directives.
func ParseResponseCacheControl(value string) (*ResponseCacheDirectives, error) {
	cd := &ResponseCacheDirectives{
		MaxAge:  -1,
		SMaxAge: -1,
		// Exerimantal stale timeouts
		StaleIfError:         -1,
		StaleWhileRevalidate: -1,
	}

	err := parse(value, cd)
	if err != nil {
		return nil, err
	}
	return cd, nil
}

func (cd *ResponseCacheDirectives) addToken(token string) error {
	var err error
	switch token {
	case "must-revalidate":
		cd.MustRevalidate = true
	case HeaderNoCache:
		cd.NoCachePresent = true
	case HeaderNoStore:
		cd.NoStore = true
	case HeaderNoTransform:
		cd.NoTransform = true
	case "public":
		cd.Public = true
	case "private":
		cd.PrivatePresent = true
	case "proxy-revalidate":
		cd.ProxyRevalidate = true
	case HeaderMaxAge:
		err = ErrMaxAgeDeltaSeconds
	case "s-maxage":
		err = ErrSMaxAgeDeltaSeconds
	// Experimental
	case "immutable":
		cd.Immutable = true
	case "stale-if-error":
		err = ErrMaxAgeDeltaSeconds
	case "stale-while-revalidate":
		err = ErrMaxAgeDeltaSeconds
	default:
		cd.Extensions = append(cd.Extensions, token)
	}
	return err
}

func hasFieldNames(token string) bool {
	switch token {
	case HeaderNoCache:
		return true
	case "private":
		return true
	}
	return false
}

func (cd *ResponseCacheDirectives) addPair(token string, v string) error {
	var err error = nil

	switch token {
	case "must-revalidate":
		err = ErrMustRevalidateNoArgs
	case HeaderNoCache:
		cd.NoCachePresent = true
		tokens := strings.Split(v, ",")
		if cd.NoCache == nil {
			cd.NoCache = make(FieldNames)
		}
		for _, t := range tokens {
			k := http.CanonicalHeaderKey(textproto.TrimString(t))
			cd.NoCache[k] = true
		}
	case HeaderNoStore:
		err = ErrNoStoreNoArgs
	case HeaderNoTransform:
		err = ErrNoTransformNoArgs
	case "public":
		err = ErrPublicNoArgs
	case "private":
		cd.PrivatePresent = true
		tokens := strings.Split(v, ",")
		if cd.Private == nil {
			cd.Private = make(FieldNames)
		}
		for _, t := range tokens {
			k := http.CanonicalHeaderKey(textproto.TrimString(t))
			cd.Private[k] = true
		}
	case "proxy-revalidate":
		err = ErrProxyRevalidateNoArgs
	case HeaderMaxAge:
		cd.MaxAge, err = ParseDeltaSeconds(v)
	case "s-maxage":
		cd.SMaxAge, err = ParseDeltaSeconds(v)
	// Experimental
	case "immutable":
		err = ErrImmutableNoArgs
	case "stale-if-error":
		cd.StaleIfError, err = ParseDeltaSeconds(v)
	case "stale-while-revalidate":
		cd.StaleWhileRevalidate, err = ParseDeltaSeconds(v)
	default:
		// TODO(pquerna): this sucks, making user re-parse, and its technically not 'quoted' like the original,
		// but this is still easier, just a SplitN on "="
		cd.Extensions = append(cd.Extensions, token+"="+v)
	}

	return err
}
