/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package v8

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	jsonx "github.com/hopeio/gox/encoding/json"
)

type SearchResponse[T any] struct {
	Took     int     `json:"took"`
	TimedOut bool    `json:"timed_out"`
	Shards   Shards  `json:"_shards"`
	Hits     Hits[T] `json:"hits"`
}

type Shards struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

type Hits[T any] struct {
	Total    Total    `json:"total"`
	MaxScore any      `json:"max_score"`
	Hits     []Hit[T] `json:"hits"`
}

type Hit[T any] struct {
	Index  string      `json:"_index"`
	Id     string      `json:"_id"`
	Score  interface{} `json:"_score"`
	Source T           `json:"_source"`
	Sort   []int       `json:"sort"`
}

type Total struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

func GetResponseData[T any](response *esapi.Response, err error) (*T, error) {
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(string(data))
	}
	var res T
	err = jsonx.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func GetSearchResponseData[T any](response *esapi.Response, err error) (*SearchResponse[T], error) {
	return GetResponseData[SearchResponse[T]](response, err)
}

func CreateDocument[T any](ctx context.Context, es *elasticsearch.Client, index, id string, obj T) error {
	body, _ := jsonx.Marshal(obj)
	esreq := esapi.CreateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
	}
	resp, err := esreq.Do(ctx, es)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
