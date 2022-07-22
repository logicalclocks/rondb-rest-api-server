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
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"hopsworks.ai/rdrs/internal/common"
	"hopsworks.ai/rdrs/internal/config"
	"hopsworks.ai/rdrs/internal/dal"
	ds "hopsworks.ai/rdrs/internal/datastructs"
	"hopsworks.ai/rdrs/internal/log"
	"hopsworks.ai/rdrs/internal/router/handler/pkread"
	"hopsworks.ai/rdrs/internal/security/apikey"
	"hopsworks.ai/rdrs/version"
)

func RegisterBatchTestHandler(engine *gin.Engine) {
	engine.POST("/"+version.API_VERSION+"/"+ds.BATCH_OPERATION, BatchOpsHandler)
}

func BatchOpsHandler(c *gin.Context) {
	operations := ds.BatchOperation{}
	err := c.ShouldBindJSON(&operations)
	if err != nil {
		if log.IsDebug() {
			body, _ := ioutil.ReadAll(c.Request.Body)
			log.Debugf("Unable to parse request. Error: %v. Body: %s\n", err, body)
		}
		common.SetResponseBodyError(c, http.StatusBadRequest, err)
		return
	}

	if operations.Operations == nil {
		common.SetResponseBodyError(c, http.StatusBadRequest, fmt.Errorf("No valid operations found"))
		return
	}

	pkOperations := make([]ds.PKReadParams, len(*operations.Operations))
	for i, operation := range *operations.Operations {
		err := parseOperation(&operation, &pkOperations[i])
		if err != nil {
			if log.IsDebug() {
				log.Debugf("Error: %v", err)
			}
			common.SetResponseBodyError(c, http.StatusBadRequest, err)
			return
		}
	}

	var response ds.BatchResponse = (ds.BatchResponse)(&ds.BatchResponseJSON{})
	response.Init()
	status, err := ProcessBatchRequest(&pkOperations, getAPIKey(c), response)
	if err != nil {
		common.SetResponseBodyError(c, status, err)
	}

	batchResponseJSON, ok := response.(*ds.BatchResponseJSON)
	if !ok {
		common.SetResponseBodyError(c, http.StatusInternalServerError, fmt.Errorf("Wrong object type. Expecting BatchResponseJSON "))
		return
	}

	common.SetResponseBody(c, status, batchResponseJSON)
}

func ProcessBatchRequest(pkOperations *[]ds.PKReadParams, apiKey string, response ds.BatchResponse) (int, error) {

	err := checkAPIKey(pkOperations, apiKey)
	if err != nil {
		return http.StatusUnauthorized, err
	}

	noOps := uint32(len(*pkOperations))
	reqPtrs := make([]*dal.NativeBuffer, noOps)
	respPtrs := make([]*dal.NativeBuffer, noOps)

	for i, pkOp := range *pkOperations {
		reqPtrs[i], respPtrs[i], err = pkread.CreateNativeRequest(&pkOp)
		defer dal.ReturnBuffer(reqPtrs[i])
		defer dal.ReturnBuffer(respPtrs[i])
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	dalErr := dal.RonDBBatchedPKRead(noOps, reqPtrs, respPtrs)
	var message string
	if dalErr != nil {
		if dalErr.HttpCode >= http.StatusInternalServerError {
			message = fmt.Sprintf("%v File: %v, Line: %v ", dalErr.Message, dalErr.ErrFileName, dalErr.ErrLineNo)
		} else {
			message = fmt.Sprintf("%v", dalErr.Message)
		}
		return dalErr.HttpCode, fmt.Errorf("%s", message)
	}

	status, err := processResponses(&respPtrs, response)
	if err != nil {
		return status, err
	}

	return http.StatusOK, nil
}

func processResponses(respBuffs *[]*dal.NativeBuffer, response ds.BatchResponse) (int, error) {
	for _, respBuff := range *respBuffs {

		subResponse := response.CreateNewSubResponse()
		pkReadResponseWithCode, ok := subResponse.(ds.PKReadResponseWithCode)
		if !ok {
			return http.StatusInternalServerError, fmt.Errorf("Wrong object type. Expecting PKReadResponseWithCode")
		}

		pkReadResponse, ok := (pkReadResponseWithCode.GetPKReadResponse()).(ds.PKReadResponse)
		if !ok {
			return http.StatusInternalServerError, fmt.Errorf("Wrong object type. Expecting PKReadResponse")
		}

		subRespCode, err := pkread.ProcessPKReadResponse(respBuff, pkReadResponse)
		if err != nil {
			return int(subRespCode), err
		}

		pkReadResponseWithCode.SetCode(&subRespCode)
		err = response.AppendSubResponse(pkReadResponseWithCode)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, nil
}

func parseOperation(operation *ds.BatchSubOperation, pkReadarams *ds.PKReadParams) error {

	//remove leading / character
	if strings.HasPrefix(*operation.RelativeURL, "/") {
		trimmed := strings.Trim(*operation.RelativeURL, "/")
		operation.RelativeURL = &trimmed
	}

	match, err := regexp.MatchString("^[a-zA-Z0-9$_]+/[a-zA-Z0-9$_]+/pk-read",
		*operation.RelativeURL)
	if !match || err != nil {
		return fmt.Errorf("Invalid Relative URL: %s", *operation.RelativeURL)
	} else {
		err := parsePKRead(operation, pkReadarams)
		if err != nil {
			return err
		}
	}
	return nil
}

func parsePKRead(operation *ds.BatchSubOperation, pkReadarams *ds.PKReadParams) error {
	params := *operation.Body

	//split the relative url to extract path parameters
	splits := strings.Split(*operation.RelativeURL, "/")
	if len(splits) != 3 {
		return fmt.Errorf("Failed to extract database and table information from relative url")
	}

	pkReadarams.DB = &splits[0]
	pkReadarams.Table = &splits[1]
	pkReadarams.Filters = params.Filters
	pkReadarams.ReadColumns = params.ReadColumns
	pkReadarams.OperationID = params.OperationID

	return nil
}

func getAPIKey(c *gin.Context) string {
	return c.GetHeader(ds.API_KEY_NAME)
}

func checkAPIKey(pkOperations *[]ds.PKReadParams, apiKey string) error {
	// check for Hopsworks api keys
	if config.Configuration().Security.UseHopsWorksAPIKeys {
		if apiKey == "" { // not set
			return fmt.Errorf("Unauthorized. No API key supplied")
		}

		dbMap := make(map[string]bool)
		dbArr := []string{}

		for _, op := range *pkOperations {
			dbMap[*op.DB] = true
		}

		for dbKey := range dbMap {
			dbArr = append(dbArr, dbKey)
		}

		return apikey.ValidateAPIKey(apiKey, dbArr...)
	}
	return nil
}
