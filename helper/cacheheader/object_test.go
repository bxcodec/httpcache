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
	"github.com/stretchr/testify/require"

	cacheControl "github.com/bxcodec/httpcache/helper/cacheheader"
	"net/http"
	"testing"
	"time"
)

func TestCachableStatusCode(t *testing.T) {
	ok := []int{200, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501}
	for _, v := range ok {
		require.True(t, cacheControl.CachableStatusCode(v), "status code should be cacheable: %d", v)
	}

	notok := []int{201, 429, 500, 504}
	for _, v := range notok {
		require.False(t, cacheControl.CachableStatusCode(v), "status code should not be cachable: %d", v)
	}
}

func fill(t *testing.T, now time.Time) cacheControl.Object {
	RespDirectives, err := cacheControl.ParseResponseCacheControl("")
	require.NoError(t, err)
	ReqDirectives, err := cacheControl.ParseRequestCacheControl("")
	require.NoError(t, err)

	obj := cacheControl.Object{
		RespDirectives: RespDirectives,
		RespHeaders:    http.Header{},
		RespStatusCode: 200,
		RespDateHeader: now,

		ReqDirectives: ReqDirectives,
		ReqHeaders:    http.Header{},
		ReqMethod:     "GET",

		NowUTC: now,
	}

	return obj
}

func TestGETPrivate(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	RespDirectives, err := cacheControl.ParseResponseCacheControl("private")
	require.NoError(t, err)

	obj.RespDirectives = RespDirectives

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonResponsePrivate)
}

func TestGETPrivateWithPrivateCache(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	RespDirectives, err := cacheControl.ParseResponseCacheControl("private")
	require.NoError(t, err)

	obj.CacheIsPrivate = true
	obj.RespDirectives = RespDirectives

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)
}

func TestUncachableMethods(t *testing.T) {
	type methodPair struct {
		m string
		r cacheControl.Reason
	}

	tc := []methodPair{
		{"PUT", cacheControl.ReasonRequestMethodPUT},
		{"DELETE", cacheControl.ReasonRequestMethodDELETE},
		{"CONNECT", cacheControl.ReasonRequestMethodCONNECT},
		{"OPTIONS", cacheControl.ReasonRequestMethodOPTIONS},
		{"CONNECT", cacheControl.ReasonRequestMethodCONNECT},
		{"TRACE", cacheControl.ReasonRequestMethodTRACE},
		{"MADEUP", cacheControl.ReasonRequestMethodUnkown},
	}

	for _, mp := range tc {
		now := time.Now().UTC()

		obj := fill(t, now)
		obj.ReqMethod = mp.m

		rv := cacheControl.ObjectResults{}
		cacheControl.CachableObject(&obj, &rv)
		require.NoError(t, rv.OutErr)
		require.Len(t, rv.OutReasons, 1)
		require.Contains(t, rv.OutReasons, mp.r)
	}
}

func TestHEAD(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "HEAD"
	obj.RespLastModifiedHeader = now.Add(time.Hour * -1)

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)

	cacheControl.ExpirationObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)
	require.False(t, rv.OutExpirationTime.IsZero())
}

const twentyFourHours = time.Hour * 24

func TestHEADLongLastModified(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "HEAD"
	obj.RespLastModifiedHeader = now.Add(time.Hour * -70000)

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)

	cacheControl.ExpirationObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)
	require.False(t, rv.OutExpirationTime.IsZero())
	require.WithinDuration(t, now.Add(twentyFourHours), rv.OutExpirationTime, time.Second*60)
}

func TestNonCachablePOST(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "POST"

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonRequestMethodPOST)
}

func TestCachablePOSTExpiresHeader(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "POST"
	obj.RespExpiresHeader = now.Add(time.Hour * 1)

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)
}

func TestCachablePOSTSMax(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "POST"
	obj.RespDirectives.SMaxAge = cacheControl.DeltaSeconds(900)

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)
}

func TestNonCachablePOSTSMax(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "POST"
	obj.CacheIsPrivate = true
	obj.RespDirectives.SMaxAge = cacheControl.DeltaSeconds(900)

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonRequestMethodPOST)
}

func TestCachablePOSTMax(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "POST"
	obj.RespDirectives.MaxAge = cacheControl.DeltaSeconds(9000)

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)
}

func TestPUTs(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "PUT"

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonRequestMethodPUT)
}

func TestPUTWithExpires(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqMethod = "PUT"
	obj.RespExpiresHeader = now.Add(time.Hour * 1)

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonRequestMethodPUT)
}

func TestAuthorization(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqHeaders.Set("Authorization", "bearer random")

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonRequestAuthorizationHeader)
}

func TestCachableAuthorization(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqHeaders.Set("Authorization", "bearer random")
	obj.RespDirectives.Public = true
	obj.RespDirectives.MaxAge = cacheControl.DeltaSeconds(300)

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.NoError(t, rv.OutErr)
	require.Len(t, rv.OutReasons, 0)
}

func TestRespNoStore(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.RespDirectives.NoStore = true

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonResponseNoStore)
}

func TestReqNoStore(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.ReqDirectives.NoStore = true

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonRequestNoStore)
}

func TestResp500(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.RespStatusCode = 500

	rv := cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &rv)
	require.Len(t, rv.OutReasons, 1)
	require.Contains(t, rv.OutReasons, cacheControl.ReasonResponseUncachableByDefault)
}

func TestExpirationSMaxShared(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.RespDirectives.SMaxAge = cacheControl.DeltaSeconds(60)

	rv := cacheControl.ObjectResults{}
	cacheControl.ExpirationObject(&obj, &rv)
	require.Len(t, rv.OutWarnings, 0)
	require.WithinDuration(t, now.Add(time.Second*60), rv.OutExpirationTime, time.Second*1)
}

func TestExpirationSMaxPrivate(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.CacheIsPrivate = true
	obj.RespDirectives.SMaxAge = cacheControl.DeltaSeconds(60)

	rv := cacheControl.ObjectResults{}
	cacheControl.ExpirationObject(&obj, &rv)
	require.Len(t, rv.OutWarnings, 0)
	require.True(t, rv.OutExpirationTime.IsZero())
}

func TestExpirationMax(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	obj.RespDirectives.MaxAge = cacheControl.DeltaSeconds(60)

	rv := cacheControl.ObjectResults{}
	cacheControl.ExpirationObject(&obj, &rv)
	require.Len(t, rv.OutWarnings, 0)
	require.WithinDuration(t, now.Add(time.Second*60), rv.OutExpirationTime, time.Second*1)
}

func TestExpirationMaxAndSMax(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	// cache should select the SMax age since this is a shared cache.
	obj.RespDirectives.MaxAge = cacheControl.DeltaSeconds(60)
	obj.RespDirectives.SMaxAge = cacheControl.DeltaSeconds(900)

	rv := cacheControl.ObjectResults{}
	cacheControl.ExpirationObject(&obj, &rv)
	require.Len(t, rv.OutWarnings, 0)
	require.WithinDuration(t, now.Add(time.Second*900), rv.OutExpirationTime, time.Second*1)
}

func TestExpirationExpires(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	// cache should select the SMax age since this is a shared cache.
	obj.RespExpiresHeader = now.Add(time.Second * 1500)

	rv := cacheControl.ObjectResults{}
	cacheControl.ExpirationObject(&obj, &rv)
	require.Len(t, rv.OutWarnings, 0)
	require.WithinDuration(t, now.Add(time.Second*1500), rv.OutExpirationTime, time.Second*1)
}

func TestExpirationExpiresNoServerDate(t *testing.T) {
	now := time.Now().UTC()

	obj := fill(t, now)
	// cache should select the SMax age since this is a shared cache.
	obj.RespDateHeader = time.Time{}
	obj.RespExpiresHeader = now.Add(time.Second * 1500)

	rv := cacheControl.ObjectResults{}
	cacheControl.ExpirationObject(&obj, &rv)
	require.Len(t, rv.OutWarnings, 0)
	require.WithinDuration(t, now.Add(time.Second*1500), rv.OutExpirationTime, time.Second*1)
}
