/*
 * This file is part of the RonDB REST API Server
 * Copyright (c) 2022 Hopsworks AB
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */
package api

import (
	"fmt"
)

// Request
type BatchOpRequest struct {
	Operations *[]BatchSubOp `json:"operations" binding:"required,min=1,max=4096,unique,dive"`
}

type BatchSubOp struct {
	Method      *string     `json:"method"        binding:"required,oneof=POST"`
	RelativeURL *string     `json:"relative-url"  binding:"required,min=1"`
	Body        *PKReadBody `json:"body"          binding:"required,min=1"`
}

// Response
type BatchOpResponse interface {
	Init()
	CreateNewSubResponse() PKReadResponseWithCode
	AppendSubResponse(subResp PKReadResponseWithCode) error
}

var _ BatchOpResponse = (*BatchResponseJSON)(nil)
var _ BatchOpResponse = (*BatchResponseGRPC)(nil)

type BatchResponseJSON struct {
	Result *[]*PKReadResponseWithCodeJSON `json:"result" binding:"required"`
}

type BatchResponseGRPC struct {
	Result *[]*PKReadResponseWithCodeGRPC `json:"result" binding:"required"`
}

func (b *BatchResponseJSON) Init() {
	subResponses := []*PKReadResponseWithCodeJSON{}
	b.Result = &subResponses
}

func (b *BatchResponseJSON) CreateNewSubResponse() PKReadResponseWithCode {
	subResponse := PKReadResponseWithCodeJSON{}
	subResponse.Init()
	return &subResponse
}

func (b *BatchResponseJSON) AppendSubResponse(subResp PKReadResponseWithCode) error {
	subRespJson, ok := subResp.(*PKReadResponseWithCodeJSON)
	if !ok {
		return fmt.Errorf("Wrong object type. Expecting PKReadResponseWithCodeJSON ")
	}

	newList := append(*b.Result, subRespJson)
	b.Result = &newList
	return nil
}

func (b *BatchResponseGRPC) Init() {
	subResponses := []*PKReadResponseWithCodeGRPC{}
	b.Result = &subResponses
}

func (b *BatchResponseGRPC) CreateNewSubResponse() PKReadResponseWithCode {
	subResponse := PKReadResponseWithCodeGRPC{}
	subResponse.Init()
	return &subResponse
}

func (b *BatchResponseGRPC) AppendSubResponse(subResp PKReadResponseWithCode) error {
	subRespGRPC, ok := subResp.(*PKReadResponseWithCodeGRPC)
	if !ok {
		return fmt.Errorf("Wrong object type. Expecting PKReadResponseWithCodeGRPC ")
	}

	newList := append(*b.Result, subRespGRPC)
	b.Result = &newList
	return nil
}

// data structs for testing
type BatchSubOperationTestInfo struct {
	SubOperation BatchSubOp
	Table        string
	DB           string
	HttpCode     int
	RespKVs      []interface{}
}

type BatchOperationTestInfo struct {
	Operations     []BatchSubOperationTestInfo
	HttpCode       int
	ErrMsgContains string
}
