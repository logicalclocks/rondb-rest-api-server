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

//import (
//	"encoding/json"
//
//	ds "hopsworks.ai/rdrs/internal/datastructs"
//)
//
//// Converters for PK Read Request
//func ConvertPKReadParams(req *ds.PKReadParams) *PKReadRequestProto {
//
//	pkReadRequestProto := PKReadRequestProto{}
//
//	var filtersProto []*FilterProto
//	if req.Filters != nil {
//		for _, fillter := range *req.Filters {
//			filterProto := FilterProto{}
//			filterProto.Column = fillter.Column
//			filterProto.Value = *fillter.Value
//			filtersProto = append(filtersProto, &filterProto)
//		}
//	}
//	pkReadRequestProto.Filters = filtersProto
//
//	var readColumnsProto []*ReadColumnProto
//	if req.ReadColumns != nil {
//		for _, readColumn := range *req.ReadColumns {
//			readColumnProto := ReadColumnProto{}
//			readColumnProto.Column = readColumn.Column
//
//			if readColumn.DataReturnType != nil {
//				readColumnProto.DataReturnType = readColumn.DataReturnType
//			}
//
//			readColumnsProto = append(readColumnsProto, &readColumnProto)
//		}
//	}
//	pkReadRequestProto.ReadColumns = readColumnsProto
//
//	pkReadRequestProto.DB = req.DB
//	pkReadRequestProto.Table = req.Table
//	pkReadRequestProto.OperationID = req.OperationID
//
//	return &pkReadRequestProto
//}
//
//func ConvertPKReadRequestProto(reqProto *PKReadRequestProto) *ds.PKReadParams {
//	pkReadParams := ds.PKReadParams{}
//
//	pkReadParams.DB = reqProto.DB
//	pkReadParams.Table = reqProto.Table
//	pkReadParams.OperationID = reqProto.OperationID
//
//	var readColumns []ds.ReadColumn
//	for _, readColumnProto := range reqProto.GetReadColumns() {
//		if readColumnProto != nil {
//			readColumn := ds.ReadColumn{}
//
//			readColumn.Column = readColumnProto.Column
//			readColumn.DataReturnType = readColumnProto.DataReturnType
//
//			readColumns = append(readColumns, readColumn)
//		}
//	}
//	if len(readColumns) > 0 {
//		pkReadParams.ReadColumns = &readColumns
//	} else {
//		pkReadParams.ReadColumns = nil
//	}
//
//	var filters []ds.Filter
//	for _, filterProto := range reqProto.Filters {
//		if filterProto != nil {
//			filter := ds.Filter{}
//
//			filter.Column = filterProto.Column
//			rawMsg := json.RawMessage(filterProto.Value)
//			filter.Value = &rawMsg
//
//			filters = append(filters, filter)
//		}
//	}
//	if len(filters) > 0 {
//		pkReadParams.Filters = &filters
//	} else {
//		pkReadParams.Filters = nil
//	}
//
//	return &pkReadParams
//}
//
//// Converters for PK Read Response
//func ConvertPKReadResponseProto(respProto *PKReadResponseProto) *ds.PKReadResponse {
//	resp := ds.PKReadResponse{}
//
//	data := []ds.Column{}
//	if respProto.Data != nil {
//		for _, columnProto := range respProto.Data {
//			if columnProto != nil {
//				column := ds.Column{}
//				column.Name = columnProto.Name
//				rawMsg := json.RawMessage(columnProto.Value)
//				column.Value = &rawMsg
//				data = append(data, column)
//			}
//		}
//	}
//	if len(data) > 0 {
//		resp.Data = &data
//	} else {
//		resp.Data = nil
//	}
//
//	resp.OperationID = respProto.OperationID
//	return &resp
//}
//
//func ConvertPKReadResponse(resp *ds.PKReadResponseGRPC) *PKReadResponseProto {
//	respProto := PKReadResponseProto{}
//
//	dataProto := []*ColumnProto{}
//	if resp.Data != nil {
//
//		for _, column := range *resp.Data {
//			columnProto := ColumnProto{}
//			columnProto.Name = column.Name
//			columnProto.Value = *column.Value
//			dataProto = append(dataProto, &columnProto)
//		}
//	}
//
//	respProto.Data = dataProto
//	respProto.OperationID = resp.OperationID
//
//	return &respProto
//}
