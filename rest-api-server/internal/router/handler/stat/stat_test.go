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

package stat

import (
	"encoding/json"
	"net/http"
	"testing"

	"hopsworks.ai/rdrs/internal/common"
	"hopsworks.ai/rdrs/internal/config"
	ds "hopsworks.ai/rdrs/internal/datastructs"
	"hopsworks.ai/rdrs/internal/router/handler"
	"hopsworks.ai/rdrs/internal/router/handler/pkread"
	tu "hopsworks.ai/rdrs/internal/router/handler/utils"
)

func TestStat(t *testing.T) {

	db := "DB004"
	table := "int_table"

	ch := make(chan int)

	numOps := uint32(5)
	expectedAllocations := numOps * 2

	if config.Configuration().RestServer.PreAllocatedBuffers > numOps {
		expectedAllocations = config.Configuration().RestServer.PreAllocatedBuffers
	}

	tu.WithDBs(t, [][][]string{common.Database(db)},
		[]handler.RegisterTestHandler{pkread.RegisterPKTestHandler, RegisterStatTestHandler}, func(tc common.TestContext) {
			for i := uint32(0); i < numOps; i++ {
				go performPkOp(t, tc, db, table, ch)
			}
			for i := uint32(0); i < numOps; i++ {
				<-ch
			}

			// get stats
			stats := getStats(t, tc)
			if stats.NativeBufferStats.AllocationsCount != uint64(expectedAllocations) ||
				stats.NativeBufferStats.BuffersCount != uint64(expectedAllocations) ||
				stats.NativeBufferStats.FreeBuffers != uint64(expectedAllocations) {
				t.Fatalf("Native buffer stats do not match Got: %v", stats)
			}

			if stats.RonDBStats.NdbObjectsCreationCount != uint64(numOps) ||
				stats.RonDBStats.NdbObjectsTotalCount != uint64(numOps) ||
				stats.RonDBStats.NdbObjectsFreeCount != uint64(numOps) {
				t.Fatalf("RonDB stats do not match. %#v", stats.RonDBStats)
			}
		})
}

func performPkOp(t *testing.T, tc common.TestContext, db string, table string, ch chan int) {
	param := ds.PKReadBody{
		Filters:     tu.NewFiltersKVs("id0", 0, "id1", 0),
		ReadColumns: tu.NewReadColumn("col0"),
	}
	body, _ := json.MarshalIndent(param, "", "\t")

	url := tu.NewPKReadURL(db, table)
	tu.ProcessRequest(t, tc, ds.PK_HTTP_VERB, url, string(body), http.StatusOK, "")

	ch <- 0
}

func getStats(t *testing.T, tc common.TestContext) ds.StatInfo {
	body := ""
	url := tu.NewStatURL()
	_, respBody := tu.ProcessRequest(t, tc, ds.STAT_HTTP_VERB, url, string(body), http.StatusOK, "")

	var stats ds.StatInfo
	err := json.Unmarshal([]byte(respBody), &stats)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return stats
}
