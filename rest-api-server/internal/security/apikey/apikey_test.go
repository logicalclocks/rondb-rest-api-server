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
package apikey

import (
	"fmt"
	"testing"

	"hopsworks.ai/rdrs/internal/common"
	"hopsworks.ai/rdrs/internal/dal"
	"hopsworks.ai/rdrs/internal/router/handler"
	"hopsworks.ai/rdrs/internal/router/handler/batchops"
	tu "hopsworks.ai/rdrs/internal/router/handler/utils"
)

func TestAPIKey(t *testing.T) {
	dbs := [][][]string{}
	dbs = append(dbs, common.Database("hopsworks"))
	handlers := []handler.RegisterTestHandler{}
	handlers = append(handlers, batchops.RegisterBatchTestHandler)
	tu.WithDBs(t, dbs, handlers, func(tc common.TestContext) {

		key, err := dal.GetAPIKey("ZaCRiVfQOxuOIXZk")
		if err != nil {
			t.Fatalf("Errot: %v", err)
		}
		fmt.Printf("Secret : %s\n", key.Secret)
		fmt.Printf("UID : %d\n", key.UserID)

		dbs, err := dal.GetUserDBs(key.UserID)
		if err != nil {
		}

		fmt.Printf("Databases are %v\n", dbs)
	})
}
