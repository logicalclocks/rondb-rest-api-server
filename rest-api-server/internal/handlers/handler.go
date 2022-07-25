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

package handlers

import (
	"github.com/gin-gonic/gin"
	ds "hopsworks.ai/rdrs/internal/datastructs"
)

type RegisterTestHandler func(*gin.Engine)

type PKReader interface {
	PkReadHttpHandler(c *gin.Context)
	PkReadHandler(pkReadParams *ds.PKReadParams, apiKey *string, response ds.PKReadResponse) (int, error)
}

type BatchOp interface {
	ProcessBatchRequest(pkOperations *[]ds.PKReadParams, apiKey *string, response ds.BatchResponse) (int, error)
}
