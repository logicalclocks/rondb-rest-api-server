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
package grpcsrv

import (
	"context"
	"net/http"

	"fmt"

	ds "hopsworks.ai/rdrs/internal/datastructs"
	"hopsworks.ai/rdrs/internal/handlers"
)

type GRPCServer struct {
}

var server GRPCServer
var pkReadHandler handlers.PKReader
var batchOpHandler handlers.Batcher

func GetGRPCServer() *GRPCServer {
	return &server

}

func (s *GRPCServer) RegisterPKReadHandler(handler handlers.PKReader) {
	pkReadHandler = handler
}

func (s *GRPCServer) RegisterBatchOpHandler(handler handlers.Batcher) {
	batchOpHandler = handler
}

func (s *GRPCServer) PKRead(c context.Context, reqProto *PKReadRequestProto) (*PKReadResponseProto, error) {
	req, apiKey := ConvertPKReadRequestProto(reqProto)

	var response ds.PKReadResponse = (ds.PKReadResponse)(&ds.PKReadResponseGRPC{})
	response.Init()

	status, err := pkReadHandler.PkReadHandler(req, &apiKey, response)
	if err != nil {
		return nil, mkError(status, err)
	}

	if status != http.StatusOK {
		return nil, mkError(status, nil)
	}

	respProto := ConvertPKReadResponse(response.(*ds.PKReadResponseGRPC))
	return respProto, nil
}

func (s *GRPCServer) Batch(c context.Context, reqProto *BatchRequestProto) (*BatchResponseProto, error) {
	req, apikey := ConvertBatchRequestProto(reqProto)

	var response ds.BatchOpResponse = (ds.BatchOpResponse)(&ds.BatchResponseGRPC{})
	response.Init()

	status, err := batchOpHandler.BathOpsHandler(&req, &apikey, response)
	if err != nil {
		return nil, mkError(status, err)
	}

	if status != http.StatusOK {
		return nil, mkError(status, nil)
	}

	respProto := ConvertBatchOpResponse(response.(*ds.BatchResponseGRPC))
	return respProto, nil
}

func mkError(status int, err error) error {
	if err != nil {
		return fmt.Errorf("Error code: %d, Error: %v ", status, err)
	} else {
		return fmt.Errorf("Error code: %d", status)
	}
}

func (s *GRPCServer) mustEmbedUnimplementedRonDBRestServerServer() {}
