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

	ds "hopsworks.ai/rdrs/internal/datastructs"
)

func ConvertPKReadParams(req *ds.PKReadParams) *PKReadRequestProto {

	var filtersProto []*FilterProto
	if req.Filters != nil {
		for _, fillter := range *req.Filters {
			filterProto := FilterProto{}
			filterProto.Column = *fillter.Column
			filterProto.Value = *fillter.Value
			filtersProto = append(filtersProto, &filterProto)
		}
	}

	var readColumnsProto []*ReadColumnProto
	if req.ReadColumns != nil {
		for _, readColumn := range *req.ReadColumns {
			readColumnProto := ReadColumnProto{}
			readColumnProto.Column = *readColumn.Column

			if readColumn.DataReturnType != nil {
				readColumnProto.DataReturnType = *&readColumn.DataReturnType
			}
			readColumnsProto = append(readColumnsProto, &readColumnProto)
		}
	}

	pkReadRequestProto := PKReadRequestProto{}
	if req.DB != nil {
		pkReadRequestProto.DB = *req.DB
	}

	if req.Table != nil {
		pkReadRequestProto.Table = *req.Table
	}

	if req.OperationID != nil {
		pkReadRequestProto.OperationID = *&req.OperationID
	}

	pkReadRequestProto.Filters = filtersProto
	pkReadRequestProto.ReadColumns = readColumnsProto

	return &pkReadRequestProto
}

func ConvertPKReadRequestProto(reqProto *PKReadRequestProto) *ds.PKReadParams {
	pkReadParams := ds.PKReadParams{}

	db := reqProto.GetDB()
	pkReadParams.DB = &db

	table := reqProto.GetTable()
	pkReadParams.Table = &table

	if reqProto.OperationID == nil { //null check as this is optional attributed
		pkReadParams.OperationID = nil
	} else {
		opID := reqProto.GetOperationID()
		pkReadParams.OperationID = &opID
	}

	var readColumns []ds.ReadColumn
	for _, readColumnProto := range reqProto.GetReadColumns() {
		if readColumnProto != nil {
			readColumn := ds.ReadColumn{}

			col := readColumnProto.GetColumn()
			readColumn.Column = &col

			if readColumnProto.DataReturnType == nil { //null check as this is optional attributed
				readColumn.DataReturnType = nil
			} else {
				drt := readColumnProto.GetDataReturnType()
				readColumn.DataReturnType = &drt
			}

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
		filter := ds.Filter{}

		col := filterProto.GetColumn()
		filter.Column = &col

		rawVal := json.RawMessage(filterProto.GetValue())
		filter.Value = &rawVal
		filters = append(filters, filter)
	}
	if len(filters) > 0 {
		pkReadParams.Filters = &filters
	} else {
		pkReadParams.Filters = nil
	}

	return &pkReadParams
}
