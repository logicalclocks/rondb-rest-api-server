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

	"fmt"

	ds "hopsworks.ai/rdrs/internal/datastructs"
	"hopsworks.ai/rdrs/internal/handlers"
)

type GRPCServer struct {
}

var server GRPCServer
var pkReadHandler handlers.PKReader

func GetGRPCServer() *GRPCServer {
	return &server

}

func (s *GRPCServer) RegisterPKReadHandler(handler handlers.PKReader) {
	pkReadHandler = handler
}

func (s *GRPCServer) PKRead(c context.Context, reqProto *PKReadRequestProto) (*PKReadResponseProto, error) {
	req, apiKey := ConvertPKReadRequestProto(reqProto)

	var response ds.PKReadResponse = (ds.PKReadResponse)(&ds.PKReadResponseGRPC{})
	response.Init()

	_, err := pkReadHandler.PkReadHandler(req, &apiKey, response)
	if err != nil {
		return nil, err
	}

	respProto := ConvertPKReadResponse(response.(*ds.PKReadResponseGRPC))
	return respProto, nil
}

func (s *GRPCServer) Batch(context.Context, *BatchRequestProto) (*BatchResponseProto, error) {
	fmt.Println("**** Batch Called ****")
	return &BatchResponseProto{}, nil
}

func (s *GRPCServer) mustEmbedUnimplementedRonDBRestServerServer() {}
