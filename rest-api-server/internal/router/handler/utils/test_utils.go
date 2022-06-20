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
package utils

import (
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"hopsworks.ai/rdrs/internal/common"
	"hopsworks.ai/rdrs/internal/config"
	"hopsworks.ai/rdrs/internal/dal"
	ds "hopsworks.ai/rdrs/internal/datastructs"
	"hopsworks.ai/rdrs/internal/log"
	"hopsworks.ai/rdrs/internal/router/handler"
	"hopsworks.ai/rdrs/internal/security/tlsutils"
	"hopsworks.ai/rdrs/pkg/server/router"
	"hopsworks.ai/rdrs/version"
)

func ProcessRequest(t testing.TB, tc common.TestContext, httpVerb string,
	url string, body string, expectedStatus int, expectedMsg string) (int, string) {
	t.Helper()

	client := setupClient(tc)
	var req *http.Request
	var resp *http.Response
	var err error
	switch httpVerb {
	case "POST":
		req, err = http.NewRequest("POST", url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

	case "GET":
		req, err = http.NewRequest("GET", url, nil)

	default:
		t.Fatalf("Http verb not yet implemented. Verb %s", httpVerb)
	}

	if err != nil {
		t.Fatalf("Test failed to create request. Error: %v", err)
	}

	if config.Configuration().Security.UseHopsWorksAPIKeys {
		req.Header.Set(ds.API_KEY_NAME, common.HOPSWORKS_TEST_API_KEY)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Test failed to perform request. Error: %v", err)
	}

	respCode := resp.StatusCode
	respBodyBtyes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Test failed to read response body. Error: %v", err)
	}
	respBody := string(respBodyBtyes)

	// var prettyJSON bytes.Buffer
	// err := json.Indent(&prettyJSON, resp.Body.Bytes(), "", "\t")
	// if err != nil {
	// log.Infof("Error %v \n", err)
	// }
	// log.Infof("Response Body. %s\n", string(prettyJSON.Bytes()))

	if respCode != expectedStatus || !strings.Contains(respBody, expectedMsg) {
		if respCode != expectedStatus {
			t.Fatalf("Test failed. Expected: %d, Got: %d. Complete Response Body: %v ", expectedStatus, respCode, respBody)
		}
		if !strings.Contains(respBody, expectedMsg) {
			t.Fatalf("Test failed. Response body does not contain %s. Body: %s", expectedMsg, respBody)
		}
	}

	return respCode, respBody
}

func setupClient(tc common.TestContext) *http.Client {

	c := &http.Client{}

	if config.Configuration().Security.RootCACertFile != "" {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: tlsutils.TrustedCAs(tc.RootCACertFile),
			},
		}

		if config.Configuration().Security.RequireAndVerifyClientCert {
			clientCert, err := tls.LoadX509KeyPair(tc.ClientCertFile, tc.ClientKeyFile)
			if err != nil {
				log.Fatalf("%v\n", err)
			}
			transport.TLSClientConfig.Certificates = []tls.Certificate{clientCert}
		}
		c.Transport = transport
	}

	return c
}

func ValidateResArrayData(t testing.TB, testInfo ds.PKTestInfo, resp string, isBinaryData bool) {
	t.Helper()

	for i := 0; i < len(testInfo.RespKVs); i++ {
		key := string(testInfo.RespKVs[i].(string))

		jsonVal, found := getColumnDataFromJson(t, key, resp)
		if !found {
			t.Fatalf("Key not found in the response. Key %s", key)
		}

		dbVal, err := getColumnDataFromDB(t, testInfo.Db, testInfo.Table,
			testInfo.PkReq.Filters, key, isBinaryData)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if string(jsonVal) != string(dbVal) {
			t.Fatalf("The read value for key %s does not match. Got from REST Server: %s, Got from MYSQL Server: %s", key, jsonVal, dbVal)
		}
	}
}

func getColumnDataFromJson(t testing.TB, colName string, resp string) (string, bool) {
	t.Helper()

	if colName[0:1] != "\"" && colName[len(colName)-1:] != "\"" {
		colName = "\"" + colName + "\""
	}

	kvMap := make(map[string]string)

	var result map[string]json.RawMessage
	json.Unmarshal([]byte(resp), &result)

	dataStr := string(result["data"])
	dl := len(dataStr)
	core := dataStr[1 : dl-1] // remove the curly braces
	strs := strings.Split(core, ",")
	for _, kv := range strs {
		index := strings.Index(kv, ":")
		kvMap[kv[0:index]] = kv[index+1:]
	}

	val, ok := kvMap[colName]
	if !ok {
		return val, ok
	} else {
		var err error
		var unquote string
		unquote = val
		if val[0] == '"' {
			unquote, err = strconv.Unquote(val)
			if err != nil {
				t.Fatal(err)
			}
		}
		return unquote, ok
	}
}

func getColumnDataFromDB(t testing.TB, db string, table string, filters *[]ds.Filter, col string, isBinary bool) (string, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.Configuration().MySQLServer.User,
		config.Configuration().MySQLServer.Password,
		config.Configuration().MySQLServer.IP,
		config.Configuration().MySQLServer.Port)
	dbConn, err := sql.Open("mysql", connectionString)
	defer dbConn.Close()
	if err != nil {
		t.Fatalf("failed to connect to db. %v", err)
	}

	command := "use " + db
	_, err = dbConn.Exec(command)
	if err != nil {
		t.Fatalf("failed to run command. %s. Error: %v", command, err)
	}

	if isBinary {
		command = fmt.Sprintf("select replace(replace(to_base64(%s), '\\r',''), '\\n', '') from %s where ", col, table)
	} else {
		command = fmt.Sprintf("select %s from %s where ", col, table)
	}
	where := ""
	for i := 0; i < len(*filters); i++ {
		if where != "" {
			where += " and "
		}
		if isBinary {
			where = fmt.Sprintf("%s %s = from_base64(%s)", where, *(*filters)[i].Column, string(*(*filters)[i].Value))
		} else {
			where = fmt.Sprintf("%s %s = %s", where, *(*filters)[i].Column, string(*(*filters)[i].Value))
		}
	}

	command = fmt.Sprintf(" %s %s\n ", command, where)
	rows, err := dbConn.Query(command)
	if err != nil {
		return "", err
	}

	// Get column names
	//columns, err := rows.Columns()
	//if err != nil {
	//	return "", err
	//}

	values := make([]sql.RawBytes, 1)
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return "", err
		}
		var value string
		for _, col := range values {

			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "null"
			} else {
				value = string(col)
			}
			return value, nil
		}
	}

	return "", nil
}

func RawBytes(a interface{}) json.RawMessage {
	var value json.RawMessage
	if a == nil {
		return []byte("null")
	}

	switch a.(type) {
	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, float32, float64:
		value = []byte(fmt.Sprintf("%v", a))
	case string:
		value = []byte(fmt.Sprintf("\"%v\"", a))
	default:
		panic(fmt.Errorf("Unsupported data type. Type: %v", reflect.TypeOf(a)))
	}
	return value
}

func NewReadColumns(prefix string, numReadColumns int) *[]ds.ReadColumn {
	readColumns := make([]ds.ReadColumn, numReadColumns)
	for i := 0; i < numReadColumns; i++ {
		col := prefix + fmt.Sprintf("%d", i)
		drt := ds.DRT_DEFAULT
		readColumns[i].Column = &col
		readColumns[i].DataReturnType = &drt
	}
	return &readColumns
}

func NewReadColumn(col string) *[]ds.ReadColumn {
	readColumns := make([]ds.ReadColumn, 1)
	drt := string(ds.DRT_DEFAULT)
	readColumns[0].Column = &col
	readColumns[0].DataReturnType = &drt
	return &readColumns
}

func NewPKReadURL(db string, table string) string {

	url := fmt.Sprintf("%s:%d%s%s", config.Configuration().RestServer.IP,
		config.Configuration().RestServer.Port,
		ds.DB_OPS_EP_GROUP, ds.PK_DB_OPERATION)
	url = strings.Replace(url, ":"+ds.DB_PP, db, 1)
	url = strings.Replace(url, ":"+ds.TABLE_PP, table, 1)
	appendURLProtocol(&url)
	return url
}

func NewBatchReadURL() string {
	url := fmt.Sprintf("%s:%d/%s/%s", config.Configuration().RestServer.IP,
		config.Configuration().RestServer.Port,
		version.API_VERSION, ds.BATCH_OPERATION)
	appendURLProtocol(&url)
	return url
}

func NewStatURL() string {
	url := fmt.Sprintf("%s:%d/%s/%s", config.Configuration().RestServer.IP,
		config.Configuration().RestServer.Port,
		version.API_VERSION, ds.STAT_OPERATION)
	appendURLProtocol(&url)
	return url
}

func appendURLProtocol(url *string) {
	if config.Configuration().Security.EnableTLS {
		*url = fmt.Sprintf("https://%s", *url)
	} else {
		*url = fmt.Sprintf("http://%s", *url)
	}
}

func NewOperationID(size int) *string {
	opID := RandString(size)
	return &opID
}

func NewPKReadReqBodyTBD() ds.PKReadBody {
	param := ds.PKReadBody{
		Filters:     NewFilters("filter_col_", 3),
		ReadColumns: NewReadColumns("read_col_", 5),
		OperationID: NewOperationID(64),
	}
	return param
}

// creates dummy filter columns of type string
func NewFilters(prefix string, numFilters int) *[]ds.Filter {
	filters := make([]ds.Filter, numFilters)
	for i := 0; i < numFilters; i++ {
		col := prefix + fmt.Sprintf("%d", i)
		val := col + "_data"
		v := RawBytes(val)
		filters[i] = ds.Filter{Column: &col, Value: &v}
	}
	return &filters
}

func NewFilter(column *string, a interface{}) *[]ds.Filter {
	filter := make([]ds.Filter, 1)

	filter[0] = ds.Filter{Column: column}
	v := RawBytes(a)
	filter[0].Value = &v
	return &filter
}

func NewFiltersKVs(vals ...interface{}) *[]ds.Filter {
	if len(vals)%2 != 0 {
		log.Panic("Expecting key value pairs")
	}

	filters := make([]ds.Filter, len(vals)/2)
	fidx := 0
	for i := 0; i < len(vals); {
		c := fmt.Sprintf("%v", vals[i])
		v := RawBytes(vals[i+1])
		filters[fidx] = ds.Filter{Column: &c, Value: &v}
		fidx++
		i += 2
	}
	return &filters
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func WithDBs(t testing.TB, dbs []string, registerHandlers []handler.RegisterTestHandler,
	fn func(tc common.TestContext)) {
	t.Helper()

	tc := common.TestContext{}

	// set log level to warn for testing
	log.SetLevel("WARN")

	if config.Configuration().Security.EnableTLS {
		tlsutils.SetupCerts(&tc)
	}

	rand.Seed(int64(time.Now().Nanosecond()))

	common.CreateDatabases(t, dbs...)
	defer common.DropDatabases(t, dbs...)

	routerCtx := router.CreateRouterContext()
	routerCtx.SetupRouter(registerHandlers)
	err := routerCtx.StartRouter()
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer shutDownRouter(t, routerCtx)

	time.Sleep(250 * time.Millisecond)

	fn(tc)

	stats := dal.GetNativeBuffersStats()
	if stats.BuffersCount != stats.FreeBuffers {
		t.Fatalf("Number of free buffers do not match. Expecting: %d, Got: %d",
			stats.BuffersCount, stats.FreeBuffers)
	}

	if config.Configuration().Security.EnableTLS {
		tlsutils.DeleteCerts(&tc)
	}
}

func shutDownRouter(t testing.TB, router router.Router) error {
	t.Helper()
	return router.StopRouter()
}

func PkTest(t *testing.T, tests map[string]ds.PKTestInfo, isBinaryData bool, registerHandler ...handler.RegisterTestHandler) {
	for name, testInfo := range tests {
		t.Run(name, func(t *testing.T) {
			dbs := []string{}
			dbs = append(dbs, testInfo.Db)

			WithDBs(t, dbs, registerHandler, func(tc common.TestContext) {
				url := NewPKReadURL(testInfo.Db, testInfo.Table)
				body, _ := json.MarshalIndent(testInfo.PkReq, "", "\t")
				httpCode, res := ProcessRequest(t, tc, ds.PK_HTTP_VERB, url,
					string(body), testInfo.HttpCode, testInfo.BodyContains)
				if httpCode == http.StatusOK {
					ValidateResArrayData(t, testInfo, res, isBinaryData)
				}
			})
		})
	}
}

func BatchTest(t *testing.T, tests map[string]ds.BatchOperationTestInfo, isBinaryData bool,
	registerHandlers ...handler.RegisterTestHandler) {
	for name, testInfo := range tests {
		t.Run(name, func(t *testing.T) {

			// all databases used in this test
			dbNamesMap := map[string]bool{}
			dbNamesArr := []string{}
			for _, op := range testInfo.Operations {
				if _, ok := dbNamesMap[op.DB]; !ok {
					dbNamesMap[op.DB] = true
				}
			}

			for k := range dbNamesMap {
				dbNamesArr = append(dbNamesArr, k)
			}

			//batch operation
			subOps := []ds.BatchSubOperation{}
			for _, op := range testInfo.Operations {
				subOps = append(subOps, op.SubOperation)
			}
			batch := ds.BatchOperation{Operations: &subOps}

			WithDBs(t, dbNamesArr, registerHandlers, func(tc common.TestContext) {
				url := NewBatchReadURL()
				body, _ := json.MarshalIndent(batch, "", "\t")
				httpCode, res := ProcessRequest(t, tc, ds.BATCH_HTTP_VERB, url,
					string(body), testInfo.HttpCode, "")
				if httpCode == http.StatusOK {
					validateBatchResponse(t, testInfo, res, isBinaryData)
				}
			})
		})
	}
}

func validateBatchResponse(t testing.TB, testInfo ds.BatchOperationTestInfo, resp string, isBinaryData bool) {
	t.Helper()
	validateBatchResponseOpIdsNCode(t, testInfo, resp)
	validateBatchResponseMsg(t, testInfo, resp)
	validateBatchResponseValues(t, testInfo, resp, isBinaryData)

}

func validateBatchResponseOpIdsNCode(t testing.TB, testInfo ds.BatchOperationTestInfo, resp string) {
	var res []struct {
		Code int
		Body struct {
			OperationId string
		}
	}
	json.Unmarshal([]byte(resp), &res)

	if len(res) != len(testInfo.Operations) {
		t.Fatal("Wrong number of operation responses received")
	}

	for i := 0; i < len(res); i++ {
		expectingId := testInfo.Operations[i].SubOperation.Body.OperationID
		if expectingId != nil {
			idGot := res[i].Body.OperationId
			if *expectingId != idGot {
				t.Fatalf("Operation ID does not match. Expecting: %s, Got: %s", *expectingId, idGot)
			}
		}

		expectingCode := testInfo.Operations[i].HttpCode
		codeGot := res[i].Code
		if expectingCode != codeGot {
			t.Fatalf("Return code does not match. Expecting: %d, Got: %d", expectingCode, codeGot)
		}
	}
}

func validateBatchResponseMsg(t testing.TB, testInfo ds.BatchOperationTestInfo, resp string) {

	var res []json.RawMessage
	json.Unmarshal([]byte(resp), &res)

	for i := 0; i < len(testInfo.Operations); i++ {
		if !strings.Contains(string(res[i]), testInfo.Operations[i].BodyContains) {
			t.Fatalf("Test failed. Response body does not contain %s. Body: %s",
				testInfo.Operations[i].BodyContains, string(res[i]))
		}
	}
}

func validateBatchResponseValues(t testing.TB, testInfo ds.BatchOperationTestInfo, resp string, isBinaryData bool) {
	var res []struct {
		Code int
		Body json.RawMessage
	}
	json.Unmarshal([]byte(resp), &res)

	for o := 0; o < len(testInfo.Operations); o++ {
		if res[o].Code != http.StatusOK {
			continue // data is null if the status is not OK
		}

		operation := testInfo.Operations[o]
		for i := 0; i < len(operation.RespKVs); i++ {
			key := string(operation.RespKVs[i].(string))
			bodyGot := string(res[o].Body)
			jsonVal, found := getColumnDataFromJson(t, key, bodyGot)
			if !found {
				t.Fatalf("Key not found in the response. Key %s", key)
			}
			dbVal, err := getColumnDataFromDB(t, operation.DB, operation.Table,
				operation.SubOperation.Body.Filters, key, isBinaryData)
			if err != nil {
				t.Fatalf("%v", err)
			}

			if string(jsonVal) != string(dbVal) {

				t.Fatalf("The read value for key %s does not match. Got from REST Server: %s, Got from MYSQL Server: %s", key, jsonVal, dbVal)
			}
		}
	}
}

func Encode(data string, binary bool, colWidth int, padding bool) string {

	if binary {

		newData := []byte(data)
		if padding {
			length := colWidth
			if length < len(data) {
				length = len(data)
			}

			newData = make([]byte, length)
			for i := 0; i < length; i++ {
				newData[i] = 0x00
			}
			for i := 0; i < len(data); i++ {
				newData[i] = data[i]
			}
		}
		return base64.StdEncoding.EncodeToString(newData)
	} else {
		return data
	}
}
