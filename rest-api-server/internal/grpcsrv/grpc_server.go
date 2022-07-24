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
	// ds "hopsworks.ai/rdrs/internal/datastructs"
	// "hopsworks.ai/rdrs/internal/handlers/pkread"
)

type GRPCServer struct {
}

func (s *GRPCServer) PKRead(c context.Context, reqProto *PKReadRequestProto) (*PKReadResponseProto, error) {
	//	req, apiKey := ConvertPKReadRequestProto(reqProto)
	//	fmt.Println("**** GRPC Server ****")
	//
	//	var response ds.PKReadResponse = (ds.PKReadResponse)(&ds.PKReadResponseGRPC{})
	//	response.Init()
	//
	//	_, err := pkread.ProcessPKReadRequest(req, &apiKey, response)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	respProto := ConvertPKReadResponse(response.(*ds.PKReadResponseGRPC))
	//	return respProto, nil
	return nil, nil
}

func (s *GRPCServer) Batch(context.Context, *BatchRequestProto) (*BatchResponseProto, error) {
	fmt.Println("**** Batch Called ****")
	return &BatchResponseProto{}, nil
}

func (s *GRPCServer) mustEmbedUnimplementedRonDBRestServerServer() {}
