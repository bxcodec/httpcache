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

package cacheheader_test

import (
	"fmt"
	cacheControl "github.com/bxcodec/httpcache/helper/cacheheader"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestMaxAge(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl("")
	require.NoError(t, err)
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(-1))

	cd, err = cacheControl.ParseResponseCacheControl("max-age")
	require.Error(t, err)

	cd, err = cacheControl.ParseResponseCacheControl("max-age=20")
	require.NoError(t, err)
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(20))

	cd, err = cacheControl.ParseResponseCacheControl("max-age=0")
	require.NoError(t, err)
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(0))

	cd, err = cacheControl.ParseResponseCacheControl("max-age=-1")
	require.Error(t, err)
}

func TestSMaxAge(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl("")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, cacheControl.DeltaSeconds(-1))

	cd, err = cacheControl.ParseResponseCacheControl("s-maxage")
	require.Error(t, err)

	cd, err = cacheControl.ParseResponseCacheControl("s-maxage=20")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, cacheControl.DeltaSeconds(20))

	cd, err = cacheControl.ParseResponseCacheControl("s-maxage=0")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, cacheControl.DeltaSeconds(0))

	cd, err = cacheControl.ParseResponseCacheControl("s-maxage=-1")
	require.Error(t, err)
}

func TestResNoCache(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl("")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, cacheControl.DeltaSeconds(-1))

	cd, err = cacheControl.ParseResponseCacheControl("no-cache")
	require.NoError(t, err)
	require.Equal(t, cd.NoCachePresent, true)
	require.Equal(t, len(cd.NoCache), 0)

	cd, err = cacheControl.ParseResponseCacheControl("no-cache=MyThing")
	require.NoError(t, err)
	require.Equal(t, cd.NoCachePresent, true)
	require.Equal(t, len(cd.NoCache), 1)
}

func TestResSpaceOnly(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(" ")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, cacheControl.DeltaSeconds(-1))
}

func TestResTabOnly(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl("\t")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, cacheControl.DeltaSeconds(-1))
}

func TestResPrivateExtensionQuoted(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`private="Set-Cookie,Request-Id" public`)
	require.NoError(t, err)
	require.Equal(t, cd.Public, true)
	require.Equal(t, cd.PrivatePresent, true)
	require.Equal(t, len(cd.Private), 2)
	require.Equal(t, len(cd.Extensions), 0)
	require.Equal(t, cd.Private["Set-Cookie"], true)
	require.Equal(t, cd.Private["Request-Id"], true)
}

func TestResCommaFollowingBare(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`public, max-age=500`)
	require.NoError(t, err)
	require.Equal(t, cd.Public, true)
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(500))
	require.Equal(t, cd.PrivatePresent, false)
	require.Equal(t, len(cd.Extensions), 0)
}

func TestResCommaFollowingKV(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`max-age=500, public`)
	require.NoError(t, err)
	require.Equal(t, cd.Public, true)
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(500))
	require.Equal(t, cd.PrivatePresent, false)
	require.Equal(t, len(cd.Extensions), 0)
}

func TestResPrivateTrailingComma(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`private=Set-Cookie, public`)
	require.NoError(t, err)
	require.Equal(t, cd.Public, true)
	require.Equal(t, cd.PrivatePresent, true)
	require.Equal(t, len(cd.Private), 1)
	require.Equal(t, len(cd.Extensions), 0)
	require.Equal(t, cd.Private["Set-Cookie"], true)
}

func TestResPrivateExtension(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`private=Set-Cookie,Request-Id public`)
	require.NoError(t, err)
	require.Equal(t, cd.Public, true)
	require.Equal(t, cd.PrivatePresent, true)
	require.Equal(t, len(cd.Private), 2)
	require.Equal(t, len(cd.Extensions), 0)
	require.Equal(t, cd.Private["Set-Cookie"], true)
	require.Equal(t, cd.Private["Request-Id"], true)
}

func TestResMultipleNoCacheTabExtension(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl("no-cache " + "\t" + "no-cache=Mything aasdfdsfa")
	require.NoError(t, err)
	require.Equal(t, cd.NoCachePresent, true)
	require.Equal(t, len(cd.NoCache), 1)
	require.Equal(t, len(cd.Extensions), 1)
	require.Equal(t, cd.NoCache["Mything"], true)
}

func TestResExtensionsEmptyQuote(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`foo="" bar="hi"`)
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, cacheControl.DeltaSeconds(-1))
	require.Equal(t, len(cd.Extensions), 2)
	require.Contains(t, cd.Extensions, "bar=hi")
	require.Contains(t, cd.Extensions, "foo=")
}

func TestResQuoteMismatch(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`foo="`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrQuoteMismatch)
}

func TestResMustRevalidateNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`must-revalidate=234`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrMustRevalidateNoArgs)
}

func TestResNoTransformNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`no-transform="xxx"`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrNoTransformNoArgs)
}

func TestResNoStoreNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`no-store=""`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrNoStoreNoArgs)
}

func TestResProxyRevalidateNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`proxy-revalidate=23432`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrProxyRevalidateNoArgs)
}

func TestResPublicNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`public=999Vary`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrPublicNoArgs)
}

func TestResMustRevalidate(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`must-revalidate`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.MustRevalidate, true)
}

func TestResNoTransform(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`no-transform`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.NoTransform, true)
}

func TestResNoStore(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`no-store`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.NoStore, true)
}

func TestResProxyRevalidate(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`proxy-revalidate`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.ProxyRevalidate, true)
}

func TestResPublic(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`public`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.Public, true)
}

func TestResPrivate(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`private`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Len(t, cd.Private, 0)
	require.Equal(t, cd.PrivatePresent, true)
}

func TestResImmutable(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`immutable`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.Immutable, true)
}

func TestResStaleIfError(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`stale-if-error=99999`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.StaleIfError, cacheControl.DeltaSeconds(99999))
}

func TestResStaleWhileRevalidate(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`stale-while-revalidate=99999`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.StaleWhileRevalidate, cacheControl.DeltaSeconds(99999))
}

func TestParseDeltaSecondsZero(t *testing.T) {
	ds, err := cacheControl.ParseDeltaSeconds("0")
	require.NoError(t, err)
	require.Equal(t, ds, cacheControl.DeltaSeconds(0))
}

func TestParseDeltaSecondsLarge(t *testing.T) {
	ds, err := cacheControl.ParseDeltaSeconds(fmt.Sprintf("%d", int64(math.MaxInt32)*2))
	require.NoError(t, err)
	require.Equal(t, ds, cacheControl.DeltaSeconds(math.MaxInt32))
}

func TestParseDeltaSecondsVeryLarge(t *testing.T) {
	ds, err := cacheControl.ParseDeltaSeconds(fmt.Sprintf("%d", int64(math.MaxInt64)))
	require.NoError(t, err)
	require.Equal(t, ds, cacheControl.DeltaSeconds(math.MaxInt32))
}

func TestParseDeltaSecondsNegative(t *testing.T) {
	ds, err := cacheControl.ParseDeltaSeconds("-60")
	require.Error(t, err)
	require.Equal(t, cacheControl.DeltaSeconds(-1), ds)
}

func TestReqNoCacheNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`no-cache=234`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrNoCacheNoArgs)
}

func TestReqNoStoreNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`no-store=,,x`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrNoStoreNoArgs)
}

func TestReqNoTransformNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`no-transform=akx`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrNoTransformNoArgs)
}

func TestReqOnlyIfCachedNoArgs(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`only-if-cached=no-store`)
	require.Error(t, err)
	require.Nil(t, cd)
	require.Equal(t, err, cacheControl.ErrOnlyIfCachedNoArgs)
}

func TestReqMaxAge(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`max-age=99999`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(99999))
	require.Equal(t, cd.MaxStale, cacheControl.DeltaSeconds(-1))
}

func TestReqMaxStale(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`max-stale=99999`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.MaxStale, cacheControl.DeltaSeconds(99999))
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(-1))
	require.Equal(t, cd.MinFresh, cacheControl.DeltaSeconds(-1))
}

func TestReqMaxAgeBroken(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`max-age`)
	require.Error(t, err)
	require.Equal(t, cacheControl.ErrMaxAgeDeltaSeconds, err)
	require.Nil(t, cd)
}

func TestReqMaxStaleBroken(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`max-stale`)
	require.Error(t, err)
	require.Equal(t, cacheControl.ErrMaxStaleDeltaSeconds, err)
	require.Nil(t, cd)
}

func TestReqMinFresh(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`min-fresh=99999`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.MinFresh, cacheControl.DeltaSeconds(99999))
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(-1))
	require.Equal(t, cd.MaxStale, cacheControl.DeltaSeconds(-1))
}

func TestReqMinFreshBroken(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`min-fresh`)
	require.Error(t, err)
	require.Equal(t, cacheControl.ErrMinFreshDeltaSeconds, err)
	require.Nil(t, cd)
}

func TestReqMinFreshJunk(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`min-fresh=a99a`)
	require.Equal(t, cacheControl.ErrMinFreshDeltaSeconds, err)
	require.Nil(t, cd)
}

func TestReqMinFreshBadValue(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`min-fresh=-1`)
	require.Equal(t, cacheControl.ErrMinFreshDeltaSeconds, err)
	require.Nil(t, cd)
}

func TestReqExtensions(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`min-fresh=99999 foobar=1 cats`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.MinFresh, cacheControl.DeltaSeconds(99999))
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(-1))
	require.Equal(t, cd.MaxStale, cacheControl.DeltaSeconds(-1))
	require.Len(t, cd.Extensions, 2)
	require.Contains(t, cd.Extensions, "foobar=1")
	require.Contains(t, cd.Extensions, "cats")
}

func TestReqMultiple(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`no-store no-transform`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.NoStore, true)
	require.Equal(t, cd.NoTransform, true)
	require.Equal(t, cd.OnlyIfCached, false)
	require.Len(t, cd.Extensions, 0)
}

func TestReqMultipleComma(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`no-cache,only-if-cached`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.NoCache, true)
	require.Equal(t, cd.NoStore, false)
	require.Equal(t, cd.NoTransform, false)
	require.Equal(t, cd.OnlyIfCached, true)
	require.Len(t, cd.Extensions, 0)
}

func TestReqLeadingComma(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`,no-cache`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Len(t, cd.Extensions, 0)
	require.Equal(t, cd.NoCache, true)
	require.Equal(t, cd.NoStore, false)
	require.Equal(t, cd.NoTransform, false)
	require.Equal(t, cd.OnlyIfCached, false)
}

func TestReqMinFreshQuoted(t *testing.T) {
	cd, err := cacheControl.ParseRequestCacheControl(`min-fresh="99999"`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.MinFresh, cacheControl.DeltaSeconds(99999))
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(-1))
	require.Equal(t, cd.MaxStale, cacheControl.DeltaSeconds(-1))
}

func TestNoSpacesIssue3(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`no-cache,no-store,max-age=0,must-revalidate`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.NoCachePresent, true)
	require.Equal(t, cd.NoStore, true)
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(0))
	require.Equal(t, cd.MustRevalidate, true)
}

func TestNoSpacesIssue3PrivateFields(t *testing.T) {
	cd, err := cacheControl.ParseResponseCacheControl(`no-cache, no-store, private=set-cookie,hello, max-age=0, must-revalidate`)
	require.NoError(t, err)
	require.NotNil(t, cd)
	require.Equal(t, cd.NoCachePresent, true)
	require.Equal(t, cd.NoStore, true)
	require.Equal(t, cd.MaxAge, cacheControl.DeltaSeconds(0))
	require.Equal(t, cd.MustRevalidate, true)
	require.Equal(t, true, cd.Private["Set-Cookie"])
	require.Equal(t, true, cd.Private["Hello"])
}
