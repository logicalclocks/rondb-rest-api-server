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

package dal

/*
#include <stdlib.h>
#include "./../../../data-access-rondb/src/rdrs-hopsworks-dal.h"
#include "./../../../data-access-rondb/src/rdrs-dal.h"
*/
import "C"
import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"unsafe"
)

type HopsworksAPIKey struct {
	Secret string
	Salt   string
	Name   string
	UserID int
}

func GetAPIKey(userKey string) (*HopsworksAPIKey, *DalError) {

	cUserKey := C.CString(userKey)
	defer C.free(unsafe.Pointer(cUserKey))

	apiKey := (*C.HopsworksAPIKey)(C.malloc(C.size_t(C.sizeof_HopsworksAPIKey)))
	defer C.free(unsafe.Pointer(apiKey))

	ret := C.find_api_key(cUserKey, apiKey)

	if ret.http_code != http.StatusOK {
		return nil, cToGoRet(&ret)
	}

	hopsworksAPIKey := HopsworksAPIKey{
		Secret: C.GoString(&apiKey.secret[0]),
		Salt:   C.GoString(&apiKey.salt[0]),
		Name:   C.GoString(&apiKey.name[0]),
		UserID: int(apiKey.user_id),
	}

	return &hopsworksAPIKey, nil
}

func GetUserProjects(uid int) ([]string, *DalError) {
	var dbs []string

	var count C.int
	countptr := (*C.int)(unsafe.Pointer(&count))

	var projects **C.char
	projectsPtr := (***C.char)(unsafe.Pointer(&projects))

	ret := C.find_all_projects(C.int(uid), projectsPtr, countptr)

	dstBuf := unsafe.Slice((**C.char)(projects), count)

	for _, buff := range dstBuf {
		db := C.GoString(buff)
		dbs = append(dbs, db)
		C.free(unsafe.Pointer(buff))
	}
	C.free(unsafe.Pointer(projects))

	if ret.http_code != http.StatusOK {
		return nil, cToGoRet(&ret)
	}

	return dbs, nil
}

func GetUserDatabases(apiKey string) ([]string, *DalError) {
	var dbs []string

	splits := strings.Split(apiKey, ".")
	prefix := splits[0]
	secret := splits[1]

	if len(splits) != 2 || len(splits[0]) != 16 {
		return dbs, &DalError{HttpCode: 404, Message: "Wrong API Key"}
	}

	key, err := GetAPIKey(prefix)
	if err != nil {
		return dbs, err
	}

	//sha256(client.secret + db.salt) = db.secret
	newSecret := sha256.Sum256([]byte(secret + key.Salt))
	newSecretHex := fmt.Sprintf("%x", newSecret)
	if strings.Compare(string(newSecretHex), key.Secret) != 0 {
		return dbs, &DalError{HttpCode: 404, Message: "Wrong API Key."}
	}

	dbs, err = GetUserProjects(key.UserID)
	if err != nil {
		return dbs, err
	}

	return dbs, nil
}
