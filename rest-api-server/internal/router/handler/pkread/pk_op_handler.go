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

package pkread

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"hopsworks.ai/rdrs/internal/common"
	"hopsworks.ai/rdrs/internal/config"
	"hopsworks.ai/rdrs/internal/dal"
	"hopsworks.ai/rdrs/internal/datastructs"
	ds "hopsworks.ai/rdrs/internal/datastructs"
	"hopsworks.ai/rdrs/internal/log"
	"hopsworks.ai/rdrs/internal/security/apikey"
)

func RegisterPKTestHandler(e *gin.Engine) {
	group := e.Group(ds.DB_OPS_EP_GROUP)
	group.POST(ds.PK_DB_OPERATION, PkReadHandler)
}

func PkReadHandler(c *gin.Context) {
	pkReadParams := ds.PKReadParams{}

	err := parseRequest(c, &pkReadParams)
	if err != nil {
		if log.IsDebug() {
			body, _ := ioutil.ReadAll(c.Request.Body)
			log.Debugf("Unable to parse request. Error: %v. Body: %s\n", err, body)
		}
		c.AbortWithError(http.StatusBadRequest, err)
		common.SetResponseError(c, http.StatusBadRequest, common.ErrorResponse{Error: fmt.Sprintf("%-v", err)})
		return
	}

	err = checkAPIKey(c, pkReadParams.DB)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	dalErr := processRequestNSetStatus(c, &pkReadParams)
	if dalErr != nil && log.IsDebug() {
		log.Debugf("Unable to perform pk-read request. Body: %-v. Error: %v\n", pkReadParams, err)
	}
}

func processRequestNSetStatus(c *gin.Context, pkReadParams *ds.PKReadParams) *dal.DalError {
	reqBuff, respBuff, err := CreateNativeRequest(pkReadParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"OK": false, "msg": fmt.Sprintf("%v", err)})
		return nil
	}
	defer dal.ReturnBuffer(reqBuff)
	defer dal.ReturnBuffer(respBuff)

	dalErr := dal.RonDBPKRead(reqBuff, respBuff)
	r, _, err := ProcessPKReadResponse(respBuff, true)
	response, ok := r.(*datastructs.PKReadResponseJSON)
	if !ok {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write(([]byte)(fmt.Sprintf("Wrong object type. Expecting PKReadResponseJSON. Got: %T ", *response)))
		return nil
	}

	var message string
	if dalErr != nil {

		if dalErr.HttpCode == http.StatusNotFound {
			setResponseBodyUnsafe(c, http.StatusNotFound, response)
		} else {
			if dalErr.HttpCode >= http.StatusInternalServerError {
				message = fmt.Sprintf("%v File: %v, Line: %v ", dalErr.Message, dalErr.ErrFileName, dalErr.ErrLineNo)
			} else {
				message = fmt.Sprintf("%v", dalErr.Message)
			}
			common.SetResponseError(c, dalErr.HttpCode, common.ErrorResponse{Error: message})
		}

		return dalErr
	} else {
		setResponseBodyUnsafe(c, http.StatusOK, response)
	}

	return nil
}

func setResponseBodyUnsafe(c *gin.Context, code int, response *ds.PKReadResponseJSON) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write(([]byte)(fmt.Sprintf("Unable to marshall response %v.  obj: %v", err, response)))
	} else {
		c.Writer.WriteHeader(code)
		c.Writer.Write(responseBytes)
	}
}

func parseRequest(c *gin.Context, pkReadParams *ds.PKReadParams) error {

	body := ds.PKReadBody{}
	pp := ds.PKReadPP{}

	if err := parseURI(c, &pp); err != nil {
		return err
	}

	if err := ParseBody(c.Request, &body); err != nil {
		return err
	}

	pkReadParams.DB = pp.DB
	pkReadParams.Table = pp.Table
	pkReadParams.Filters = body.Filters
	pkReadParams.ReadColumns = body.ReadColumns
	pkReadParams.OperationID = body.OperationID
	return nil
}

func ParseBody(req *http.Request, params *ds.PKReadBody) error {

	b := binding.JSON
	err := b.Bind(req, &params)
	if err != nil {
		return err
	}

	err = ValidateBody(params)
	if err != nil {
		return err
	}

	return nil
}

func ValidateBody(params *ds.PKReadBody) error {

	for _, filter := range *params.Filters {
		// make sure filter columns are valid
		if err := validateDBIdentifier(*filter.Column); err != nil {
			return err
		}
	}

	// make sure that the columns are unique.
	existingFilters := make(map[string]bool)
	for _, filter := range *params.Filters {
		if _, value := existingFilters[*filter.Column]; value {
			return fmt.Errorf("field validation for filter failed on the 'unique' tag")
		} else {
			existingFilters[*filter.Column] = true
		}
	}

	// make sure read columns are valid
	if params.ReadColumns != nil {
		for _, col := range *params.ReadColumns {
			if err := validateDBIdentifier(*col.Column); err != nil {
				return err
			}
		}
	}

	// make sure that the filter columns and read colummns do not overlap
	// and read cols are unique
	if params.ReadColumns != nil {
		existingCols := make(map[string]bool)
		for _, readCol := range *params.ReadColumns {
			if _, value := existingFilters[*readCol.Column]; value {
				return fmt.Errorf("field validation for read columns faild. '%s' already included in filter", *readCol.Column)
			}

			if _, value := existingCols[*readCol.Column]; value {
				return fmt.Errorf("field validation for 'ReadColumns' failed on the 'unique' tag.")
			} else {
				existingCols[*readCol.Column] = true
			}
		}
	}

	return nil
}

func parseURI(c *gin.Context, resource *ds.PKReadPP) error {
	err := c.ShouldBindUri(&resource)
	if err != nil {
		return err
	}

	if err = validateDBIdentifier(*resource.DB); err != nil {
		return err
	}

	if err = validateDBIdentifier(*resource.Table); err != nil {
		return err
	}

	return nil
}

func validateDBIdentifier(identifier string) error {
	if len(identifier) < 1 || len(identifier) > 64 {
		return fmt.Errorf("field length validation failed")
	}

	//https://dev.mysql.com/doc/refman/8.0/en/identifiers.html
	for _, r := range identifier {
		if !((r >= rune(0x0001) && r <= rune(0x007F)) || (r >= rune(0x0080) && r <= rune(0x0FFF))) {
			return fmt.Errorf("field validation failed. Invalid character '%U' ", r)
		}
	}
	return nil
}

func checkAPIKey(c *gin.Context, db *string) error {
	// check for Hopsworks api keys
	if config.Configuration().Security.UseHopsWorksAPIKeys {
		xapikey := c.GetHeader(ds.API_KEY_NAME)
		if xapikey == "" { // not set
			return fmt.Errorf("Unauthorized. No API key supplied")
		}
		return apikey.ValidateAPIKey(xapikey, *db)
	}
	return nil
}
