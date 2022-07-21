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
	context "context"
	"encoding/json"
	"fmt"

	ds "hopsworks.ai/rdrs/internal/datastructs"
)

type GRPCServer struct {
}

func (s *GRPCServer) PKRead(c context.Context, reqProto *PKReadRequestProto) (*PKReadResponseProto, error) {
	req := ConvertPKReadRequestProto(reqProto)
	fmt.Println("**** PKRead Called ****")
	bytes, _ := json.MarshalIndent(req, "", " ")
	fmt.Printf("Req %s \n", string(bytes))

	data := make(map[string]*string)
	name := "col_name"
	value := "123"
	data[name] = &value

	resp := ds.PKReadResponseGRPC{}
	resp.OperationID = req.OperationID
	resp.Data = &data

	respProto := ConvertPKReadResponse(&resp)

	return respProto, nil
}

func (s *GRPCServer) Batch(context.Context, *BatchRequestProto) (*BatchResponseProto, error) {
	fmt.Println("**** Batch Called ****")
	return &BatchResponseProto{}, nil
}

func (s *GRPCServer) mustEmbedUnimplementedRonDBRestServerServer() {}
