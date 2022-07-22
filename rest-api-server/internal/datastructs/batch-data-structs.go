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
package datastructs

import (
	"fmt"

	"hopsworks.ai/rdrs/version"
)

const DBS_OPS_EP_GROUP = "/" + version.API_VERSION + "/"
const BATCH_OPERATION = "batch"
const BATCH_HTTP_VERB = "POST"

// Request
type BatchOperation struct {
	Operations *[]BatchSubOperation `json:"operations" binding:"required,min=1,max=4096,unique,dive"`
}

type BatchSubOperation struct {
	Method      *string     `json:"method"        binding:"required,oneof=POST"`
	RelativeURL *string     `json:"relative-url"  binding:"required,min=1"`
	Body        *PKReadBody `json:"body"          binding:"required,min=1"`
}

// Response
type BatchResponse interface {
	Init()
	CreateNewSubResponse() interface{}
	AppendSubResponse(subResp interface{}) error
}

var _ BatchResponse = (*BatchResponseGRPC)(nil)
var _ BatchResponse = (*BatchResponseJSON)(nil)

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

func (b *BatchResponseJSON) CreateNewSubResponse() interface{} {
	subResponse := PKReadResponseWithCodeJSON{}
	subResponse.Init()
	return &subResponse
}

func (b *BatchResponseJSON) AppendSubResponse(subResp interface{}) error {
	subRespJson, ok := subResp.(*PKReadResponseWithCodeJSON)
	if !ok {
		return fmt.Errorf("Wrong object type. Expecting PKReadResponseJSON ")
	}

	newList := append(*b.Result, subRespJson)
	b.Result = &newList
	return nil
}

func (b *BatchResponseGRPC) Init() {
	subResponses := []*PKReadResponseWithCodeGRPC{}
	b.Result = &subResponses
}

func (b *BatchResponseGRPC) CreateNewSubResponse() interface{} {
	subResponse := PKReadResponseWithCodeGRPC{}
	subResponse.Init()
	return &subResponse
}

func (b *BatchResponseGRPC) AppendSubResponse(subResp interface{}) error {
	subRespGRPC, ok := subResp.(*PKReadResponseWithCodeGRPC)
	if !ok {
		return fmt.Errorf("Wrong object type. Expecting PKReadResponseGRPC ")
	}

	newList := append(*b.Result, subRespGRPC)
	b.Result = &newList
	return nil
}

// data structs for testing
type BatchSubOperationTestInfo struct {
	SubOperation BatchSubOperation
	Table        string
	DB           string
	HttpCode     int
	BodyContains string
	RespKVs      []interface{}
}

type BatchOperationTestInfo struct {
	Operations []BatchSubOperationTestInfo
	HttpCode   int
}
