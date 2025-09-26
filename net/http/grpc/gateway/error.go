/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gateway

import (
	"fmt"
	"net/http"
	"net/textproto"
	"slices"

	http2 "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc/metadata"
)

func InComingHeaderMatcher(key string) (string, bool) {
	if slices.Contains(InComingHeader, key) {
		return key, true
	}
	return "", false
}

func OutgoingHeaderMatcher(key string) (string, bool) {
	if slices.Contains(OutgoingHeader, key) {
		return key, true
	}
	return "", false
}

func HandleForwardResponseServerMetadata(w http.ResponseWriter, md metadata.MD) {
	for _, k := range OutgoingHeader {
		if vs, ok := md[k]; ok {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
	}
}

func HandleForwardResponseTrailerHeader(w http.ResponseWriter, md metadata.MD) {
	for k := range md {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", grpc.MetadataTrailerPrefix, k))
		w.Header().Add(http2.HeaderTrailer, tKey)
	}
}

func HandleForwardResponseTrailer(w http.ResponseWriter, md metadata.MD) {
	for k, vs := range md {
		tKey := fmt.Sprintf("%s%s", grpc.MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}
