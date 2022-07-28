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

	"hopsworks.ai/rdrs/internal/handlers"
	"hopsworks.ai/rdrs/pkg/api"
)

type GRPCServer struct {
	api.UnimplementedRonDBRESTServer
}

var _ api.RonDBRESTServer = (*GRPCServer)(nil)

var server GRPCServer

var pkReadHandler handlers.PKReader
var batchOpHandler handlers.Batcher
var statOpHandler handlers.Stater

func GetGRPCServer() *GRPCServer {
	return &server

}

func (s *GRPCServer) RegisterPKReadHandler(handler handlers.PKReader) {
	pkReadHandler = handler
}

func (s *GRPCServer) RegisterBatchOpHandler(handler handlers.Batcher) {
	batchOpHandler = handler
}

func (s *GRPCServer) RegisterStatOpHandler(handler handlers.Stater) {
	statOpHandler = handler
}

func (s *GRPCServer) PKRead(c context.Context, reqProto *api.PKReadRequestProto) (*api.PKReadResponseProto, error) {
	req, apiKey := api.ConvertPKReadRequestProto(reqProto)

	var response api.PKReadResponse = (api.PKReadResponse)(&api.PKReadResponseGRPC{})
	response.Init()

	status, err := pkReadHandler.PkReadHandler(req, &apiKey, response)
	if err != nil {
		return nil, mkError(status, err)
	}

	if status != http.StatusOK {
		return nil, mkError(status, nil)
	}

	respProto := api.ConvertPKReadResponse(response.(*api.PKReadResponseGRPC))
	return respProto, nil
}

func (s *GRPCServer) Batch(c context.Context, reqProto *api.BatchRequestProto) (*api.BatchResponseProto, error) {
	req, apikey := api.ConvertBatchRequestProto(reqProto)

	var response api.BatchOpResponse = (api.BatchOpResponse)(&api.BatchResponseGRPC{})
	response.Init()

	status, err := batchOpHandler.BatchOpsHandler(req, &apikey, response)
	if err != nil {
		return nil, mkError(status, err)
	}

	if status != http.StatusOK {
		return nil, mkError(status, nil)
	}

	respProto := api.ConvertBatchOpResponse(response.(*api.BatchResponseGRPC))
	return respProto, nil
}

func (s *GRPCServer) Stat(ctx context.Context, reqProto *api.StatRequestProto) (*api.StatResponseProto, error) {

	response := &api.StatResponse{}
	status, err := statOpHandler.StatOpsHandler(response)
	if err != nil {
		return nil, mkError(status, err)
	}

	if status != http.StatusOK {
		return nil, mkError(status, nil)
	}

	respProto := api.ConvertStatResponse(response)
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
