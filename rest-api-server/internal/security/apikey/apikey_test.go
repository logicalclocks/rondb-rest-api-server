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
	"time"

	"hopsworks.ai/rdrs/internal/common"
	"hopsworks.ai/rdrs/internal/config"
	"hopsworks.ai/rdrs/internal/dal"
)

func TestAPIKey(t *testing.T) {

	conString := fmt.Sprintf("%s:%d", config.Configuration().RonDBConfig.IP,
		config.Configuration().RonDBConfig.Port)

	dal.InitRonDBConnection(conString, true)
	defer dal.ShutdownConnection()

	common.CreateDatabases(t, []string{"DB001", "DB002"}...)
	defer common.DropDatabases(t, []string{"DB001", "DB002"}...)

	err := ValidateAPIKey("bkYjEz6OTZyevbqT.ocHajJhnE0ytBh8zbYj3IXupyMqeMZp8PW464eTxzxqP5afBjodEQUgY0lmL33ub", "")
	if err == nil {
		t.Fatalf("Supplied wrong prefix. This should have failed. ")
	}

	err = ValidateAPIKey("bkYjEz6OTZyevbqT.")
	if err == nil {
		t.Fatalf("No secret. This should have failed")
	}

	err = ValidateAPIKey("bkYjEz6OTZyevbq.ocHajJhnE0ytBh8zbYj3IXupyMqeMZp8PW464eTxzxqP5afBjodEQUgY0lmL33ub")
	if err == nil {
		t.Fatalf("Wrong length prefix. This should have failed")
	}

	// correct api key but wrong db. this api key can not access test3 db
	err = ValidateAPIKey(common.HOPSWORKS_TEST_API_KEY, "test3")
	if err == nil {
		t.Fatalf("This should have failed")
	}

	// correct api key
	err = ValidateAPIKey(common.HOPSWORKS_TEST_API_KEY, "DB001")
	if err != nil {
		t.Fatalf("No error expected")
	}

	err = ValidateAPIKey(common.HOPSWORKS_TEST_API_KEY, "DB001", "DB002")
	if err != nil {
		t.Fatalf("No error expected")
	}

}

// check that cache is updated every N secs
func TestAPIKeyCache1(t *testing.T) {

	conString := fmt.Sprintf("%s:%d", config.Configuration().RonDBConfig.IP,
		config.Configuration().RonDBConfig.Port)

	dal.InitRonDBConnection(conString, true)
	defer dal.ShutdownConnection()

	common.CreateDatabases(t, []string{"DB001", "DB002"}...)
	defer common.DropDatabases(t, []string{"DB001", "DB002"}...)

	err := ValidateAPIKey(common.HOPSWORKS_TEST_API_KEY, "DB001")
	if err != nil {
		t.Fatalf("No error expected")
	}

	lastUpdated1 := cacheUpdateTime(common.HOPSWORKS_TEST_API_KEY)

	err = ValidateAPIKey(common.HOPSWORKS_TEST_API_KEY, "DB001")
	if err != nil {
		t.Fatalf("No error expected")
	}

	lastUpdated2 := cacheUpdateTime(common.HOPSWORKS_TEST_API_KEY)

	if lastUpdated1 != lastUpdated2 {
		t.Fatalf("Cache update time is expected to be the same")
	}

	time.Sleep(time.Duration(config.Configuration().Security.HopsWorksAPIKeysCacheValiditySec))
	err = ValidateAPIKey(common.HOPSWORKS_TEST_API_KEY, "DB001")
	if err != nil {
		t.Fatalf("No error expected")
	}

	lastUpdated3 := cacheUpdateTime(common.HOPSWORKS_TEST_API_KEY)

	lastUpdated2p := lastUpdated2.Add(time.Duration(config.Configuration().Security.HopsWorksAPIKeysCacheValiditySec))
	if lastUpdated2 != lastUpdated3 && lastUpdated2p.Before(lastUpdated3) {
		t.Fatalf("Cache time is not updated properly")
	}

}

// check that cache is updated every N secs even if the user is not authorized to access a DB
func TestAPIKeyCache2(t *testing.T) {

	conString := fmt.Sprintf("%s:%d", config.Configuration().RonDBConfig.IP,
		config.Configuration().RonDBConfig.Port)

	dal.InitRonDBConnection(conString, true)
	defer dal.ShutdownConnection()

	common.CreateDatabases(t, []string{"DB001", "DB002"}...)
	defer common.DropDatabases(t, []string{"DB001", "DB002"}...)

	err := ValidateAPIKey(common.HOPSWORKS_TEST_API_KEY, "DB003")
	if err == nil {
		t.Fatalf("Expected it to fail")
	}

	lastUpdated1 := cacheUpdateTime(common.HOPSWORKS_TEST_API_KEY)

	err = ValidateAPIKey(common.HOPSWORKS_TEST_API_KEY, "DB001")
	if err != nil {
		t.Fatalf("No error expected")
	}

	lastUpdated2 := cacheUpdateTime(common.HOPSWORKS_TEST_API_KEY)

	if lastUpdated1 != lastUpdated2 {
		t.Fatalf("Cache update time is expected to be the same")
	}
}
