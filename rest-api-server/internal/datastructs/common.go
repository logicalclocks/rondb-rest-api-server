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
package datastructs

import "hopsworks.ai/rdrs/version"

const DB_PP = "db"
const TABLE_PP = "table"
const DB_OPS_EP_GROUP = "/" + version.API_VERSION + "/:" + DB_PP + "/:" + TABLE_PP + "/"

const API_KEY_NAME = "X-API-KEY"
