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
		c.JSON(http.StatusBadRequest, gin.H{"OK": false, "msg": fmt.Sprintf("%-v", err)})
		return
	}

	if operations.Operations == nil {
		c.JSON(http.StatusBadRequest, gin.H{"OK": false, "msg": "No valid operations found"})
		return
	}

	pkOperations := make([]ds.PKReadParams, len(*operations.Operations))
	for i, operation := range *operations.Operations {
		err := parseOperation(&operation, &pkOperations[i])
		if err != nil {
			if log.IsDebug() {
				log.Debugf("Error: %v", err)
			}
			c.JSON(http.StatusBadRequest, gin.H{"OK": false, "msg": fmt.Sprintf("%-v", err)})
			return
		}
	}

	err = checkAPIKey(c, &pkOperations)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	dalErr := processRequestNSetStatus(c, &pkOperations)
	if dalErr != nil && log.IsDebug() {
		log.Debugf("Unable to perform batch request. Body: %-v. Error: %v\n", operations, err)
	}
}

func setResponseBodyUnsafe(c *gin.Context, code uint32, resp []*dal.NativeBuffer) {
	var response ds.BatchResponseJSON
	subResponses := []ds.PKReadResponseWithCodeJSON{}
	for _, respBuff := range resp {
		subResp, subRespCode, err := pkread.ProcessPKReadResponse(respBuff, true)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			c.Writer.Write([]byte(fmt.Sprintf("Failed to created response for batch op. Error: %v", err)))
			return
		}
		var subRespWCode ds.PKReadResponseWithCodeJSON
		subRespWCode.Code = &subRespCode

		subRespJson, ok := subResp.(*ds.PKReadResponseJSON)
		if !ok {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			c.Writer.Write(([]byte)(fmt.Sprintf("Wrong object type. Expecting PKReadResponseJSON ")))
			return
		}
		subRespWCode.Body = subRespJson
		subResponses = append(subResponses, subRespWCode)
	}
	response.Result = &subResponses

	bytes, err := json.Marshal(response)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(fmt.Sprintf("Failed to marshall batch op  response. Error: %v", err)))
		return
	} else {
		c.Writer.WriteHeader(int(code))
		c.Writer.Write(bytes)
	}
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

	err := pkread.ValidateBody(&params)
	if err != nil {
		return err
	}

	pkReadarams.DB = &splits[0]
	pkReadarams.Table = &splits[1]
	pkReadarams.Filters = params.Filters
	pkReadarams.ReadColumns = params.ReadColumns
	pkReadarams.OperationID = params.OperationID
	return nil
}

func processRequestNSetStatus(c *gin.Context, pkOperations *[]ds.PKReadParams) *dal.DalError {

	noOps := uint32(len(*pkOperations))
	reqPtrs := make([]*dal.NativeBuffer, noOps)
	respPtrs := make([]*dal.NativeBuffer, noOps)

	var err error
	for i, pkOp := range *pkOperations {
		reqPtrs[i], respPtrs[i], err = pkread.CreateNativeRequest(&pkOp)
		defer dal.ReturnBuffer(reqPtrs[i])
		defer dal.ReturnBuffer(respPtrs[i])
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"OK": false, "msg": fmt.Sprintf("%v", err)})
			return &dal.DalError{HttpCode: http.StatusInternalServerError, Message: fmt.Sprintf("%v", err)}
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
		common.SetResponseError(c, dalErr.HttpCode, common.ErrorResponse{Error: message})
		return dalErr

	} else {
		setResponseBodyUnsafe(c, http.StatusOK, respPtrs)
	}

	return nil
}

func checkAPIKey(c *gin.Context, pkOperations *[]ds.PKReadParams) error {
	// check for Hopsworks api keys
	if config.Configuration().Security.UseHopsWorksAPIKeys {
		xapikey := c.GetHeader(ds.API_KEY_NAME)
		if xapikey == "" { // not set
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

		return apikey.ValidateAPIKey(xapikey, dbArr...)
	}
	return nil
}
