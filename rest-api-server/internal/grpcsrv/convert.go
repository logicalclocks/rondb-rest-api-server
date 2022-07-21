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
	"encoding/json"
	"strconv"

	ds "hopsworks.ai/rdrs/internal/datastructs"
)

// Converters for PK Read Request
func ConvertPKReadParams(req *ds.PKReadParams) (*PKReadRequestProto, error) {

	pkReadRequestProto := PKReadRequestProto{}

	var filtersProto []*FilterProto
	if req.Filters != nil {
		for _, fillter := range *req.Filters {
			filterProto := FilterProto{}
			filterProto.Column = fillter.Column

			// remove quotes if any
			if *fillter.Value != nil {
				valueStr := string([]byte(*fillter.Value))
				var err error
				if valueStr[0] == '"' {
					valueStr, err = strconv.Unquote(valueStr)
					if err != nil {
						return nil, err
					}
				}
				filterProto.Value = &valueStr
			}

			filtersProto = append(filtersProto, &filterProto)
		}
	}
	pkReadRequestProto.Filters = filtersProto

	var readColumnsProto []*ReadColumnProto
	if req.ReadColumns != nil {
		for _, readColumn := range *req.ReadColumns {
			readColumnProto := ReadColumnProto{}
			readColumnProto.Column = readColumn.Column

			if readColumn.DataReturnType != nil {
				readColumnProto.DataReturnType = readColumn.DataReturnType
			}

			readColumnsProto = append(readColumnsProto, &readColumnProto)
		}
	}
	pkReadRequestProto.ReadColumns = readColumnsProto

	pkReadRequestProto.DB = req.DB
	pkReadRequestProto.Table = req.Table
	pkReadRequestProto.OperationID = req.OperationID

	return &pkReadRequestProto, nil
}

func ConvertPKReadRequestProto(reqProto *PKReadRequestProto) *ds.PKReadParams {
	pkReadParams := ds.PKReadParams{}

	pkReadParams.DB = reqProto.DB
	pkReadParams.Table = reqProto.Table
	pkReadParams.OperationID = reqProto.OperationID

	var readColumns []ds.ReadColumn
	for _, readColumnProto := range reqProto.GetReadColumns() {
		if readColumnProto != nil {
			readColumn := ds.ReadColumn{}

			readColumn.Column = readColumnProto.Column
			readColumn.DataReturnType = readColumnProto.DataReturnType

			readColumns = append(readColumns, readColumn)
		}
	}
	if len(readColumns) > 0 {
		pkReadParams.ReadColumns = &readColumns
	} else {
		pkReadParams.ReadColumns = nil
	}

	var filters []ds.Filter
	for _, filterProto := range reqProto.Filters {
		if filterProto != nil {
			filter := ds.Filter{}

			filter.Column = filterProto.Column
			rawMsg := json.RawMessage([]byte(*filterProto.Value))
			filter.Value = &rawMsg

			filters = append(filters, filter)
		}
	}
	if len(filters) > 0 {
		pkReadParams.Filters = &filters
	} else {
		pkReadParams.Filters = nil
	}

	return &pkReadParams
}

// Converters for PK Read Response
func ConvertPKReadResponseProto(respProto *PKReadResponseProto) *ds.PKReadResponseGRPC {
	resp := ds.PKReadResponseGRPC{}

	data := make(map[string]*string)
	if respProto.Data != nil {
		for colName, colVal := range respProto.Data {
			if colVal != nil {
				data[colName] = colVal.Name
			} else {
				data[colName] = nil
			}
		}
	}
	if len(data) > 0 {
		resp.Data = &data
	} else {
		resp.Data = nil
	}

	resp.OperationID = respProto.OperationID
	return &resp
}

func ConvertPKReadResponse(resp *ds.PKReadResponseGRPC) *PKReadResponseProto {
	respProto := PKReadResponseProto{}
	respProto.Data = make(map[string]*ColumnValueProto)
	if resp.Data != nil {
		for colName, colVal := range *resp.Data {
			if colVal != nil {
				respProto.Data[colName] = &ColumnValueProto{Name: colVal}
			} else {
				respProto.Data[colName] = nil
			}
		}
	}

	respProto.OperationID = resp.OperationID
	return &respProto
}
