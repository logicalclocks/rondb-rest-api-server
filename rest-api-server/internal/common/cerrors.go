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

package common

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -L./../../../data-access-rondb/build/ -lrdrclient
#cgo LDFLAGS: -L/usr/local/mysql/lib -lndbclient
#include "./../../../data-access-rondb/src/error-strs.h"
*/
import "C"

func ERROR_011() string {
	return C.ERROR_011
}

func ERROR_012() string {
	return C.ERROR_012
}

func ERROR_013() string {
	return C.ERROR_013
}

func ERROR_014() string {
	return C.ERROR_014
}

func ERROR_001() string {
	return C.ERROR_001
}
