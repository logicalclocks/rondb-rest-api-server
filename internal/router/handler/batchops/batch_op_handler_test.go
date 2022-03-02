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
package batchops

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"hopsworks.ai/rondb-rest-api-server/internal/router/handler/pkread"
	tu "hopsworks.ai/rondb-rest-api-server/internal/router/handler/utils"
)

func TestBatchOps(t *testing.T) {
	router := gin.Default()
	group := router.Group(DB_OPS_EP_GROUP)
	group.POST(DB_OPERATION, BatchOpsHandler)
	url := URL()

	operations := NewOperations(t, 3)
	operationsWrapper := Operations{Operations: &operations}
	body, _ := json.Marshal(operationsWrapper)

	tu.ProcessRequest(t, router, HTTP_VERB, url, string(body), http.StatusOK, "")
}

func TestBatchMissingReqField(t *testing.T) {
	router := gin.Default()
	group := router.Group(DB_OPS_EP_GROUP)
	group.POST(DB_OPERATION, BatchOpsHandler)
	url := URL()

	// Test missing method
	operations := NewOperations(t, 3)
	operations[1].Method = nil
	operationsWrapper := Operations{Operations: &operations}
	body, _ := json.Marshal(operationsWrapper)
	tu.ProcessRequest(t, router, HTTP_VERB, url, string(body), http.StatusBadRequest,
		"Error:Field validation for 'Method' failed ")

	// Test missing relative URL
	operations = NewOperations(t, 3)
	operations[1].RelativeURL = nil
	operationsWrapper = Operations{Operations: &operations}
	body, _ = json.Marshal(operationsWrapper)
	tu.ProcessRequest(t, router, HTTP_VERB, url, string(body), http.StatusBadRequest,
		"Error:Field validation for 'RelativeURL' failed ")

	// Test missing body
	operations = NewOperations(t, 3)
	operations[1].Body = nil
	operationsWrapper = Operations{Operations: &operations}
	body, _ = json.Marshal(operationsWrapper)
	tu.ProcessRequest(t, router, HTTP_VERB, url, string(body), http.StatusBadRequest,
		"Error:Field validation for 'Body' failed ")

	// Test missing filter in an operation
	operations = NewOperations(t, 3)
	*operations[1].Body = strings.Replace(*operations[1].Body, pkread.FILTER_PARAM_NAME, "XXX", -1)
	operationsWrapper = Operations{Operations: &operations}
	body, _ = json.Marshal(operationsWrapper)
	tu.ProcessRequest(t, router, HTTP_VERB, url, string(body), http.StatusBadRequest,
		"Error:Field validation for 'Filters' failed")

	// Test missing non-required fields
	operations = NewOperations(t, 1)
	*operations[0].Body = strings.Replace(*operations[0].Body, pkread.READ_COL_PARAM_NAME, "XXX", -1)
	*operations[0].Body = strings.Replace(*operations[0].Body, pkread.OPERATION_ID_PARAM_NAME, "XXX", -1)
	operationsWrapper = Operations{Operations: &operations}
	body, _ = json.Marshal(operationsWrapper)
	tu.ProcessRequest(t, router, HTTP_VERB, url, string(body), http.StatusOK, "")
}

func NewOperations(t *testing.T, numOps int) []Operation {
	operations := make([]Operation, numOps)
	for i := 0; i < numOps; i++ {
		operations[i] = NewOperation(t)
	}
	return operations
}

func NewOperation(t *testing.T) Operation {
	opBody, _ := json.MarshalIndent(pkread.NewPKReadReqBody(t), "", "\t")
	opBodyStr := string(opBody)
	method := "POST"
	relativeURL := pkread.NewPKReadURL("db", "table")

	op := Operation{
		Method:      &method,
		RelativeURL: &relativeURL,
		Body:        &opBodyStr,
	}

	return op
}

func URL() string {
	return fmt.Sprintf("%s%s", DB_OPS_EP_GROUP, DB_OPERATION)
}
