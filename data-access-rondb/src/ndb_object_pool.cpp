/*
 * Copyright (C) 2022 Hopsworks AB
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301,
 * USA.
 */

#include "ndb_object_pool.hpp"
#include <iostream>
#include "src/status.hpp"
#include "src/error-strs.h"

NdbObjectPool *NdbObjectPool::__instance = nullptr;

NdbObjectPool *NdbObjectPool::GetInstance() {
  if (__instance == nullptr) {
    __instance                              = new NdbObjectPool();
    __instance->stats.ndb_objects_available = 0;
    __instance->stats.ndb_objects_count     = 0;
    __instance->stats.ndb_objects_created   = 0;
    __instance->stats.ndb_objects_deleted   = 0;
  }
  return __instance;
}

RS_Status NdbObjectPool::GetNdbObject(Ndb_cluster_connection *ndb_connection, Ndb **ndb_object) {
  __mutex.lock();
  RS_Status ret_status = RS_OK;
  if (__ndb_objects.empty()) {
    std::cout << "Creating new." << std::endl;

    *ndb_object = new Ndb(ndb_connection);
    int retCode = (*ndb_object)->init();
    if (retCode != 0) {
      ret_status = RS_SERVER_ERROR(ERROR_004 + std::string(" RetCode: ") + std::to_string(retCode));
    }
    stats.ndb_objects_created++;
    stats.ndb_objects_count++;
  } else {
    *ndb_object = __ndb_objects.front();
    __ndb_objects.pop_front();
  }
  __mutex.unlock();
  return ret_status;
}

void NdbObjectPool::ReturnResource(Ndb *object) {
  __mutex.lock();
  // reset transaction and cleanup
  __ndb_objects.push_back(object);
  __mutex.unlock();
}

RonDB_Stats NdbObjectPool::GetStats() {
  __mutex.lock();

  stats.ndb_objects_available = __ndb_objects.size();

  __mutex.unlock();
  return stats;
}

RS_Status NdbObjectPool::Close() {
  __mutex.lock();

  while (__ndb_objects.size() > 0) {
    Ndb *ndb_object = __ndb_objects.front();
    __ndb_objects.pop_front();
    delete ndb_object;
  }

  stats.ndb_objects_available = 0;
  stats.ndb_objects_count     = 0;
  stats.ndb_objects_created   = 0;
  stats.ndb_objects_deleted   = 0;
  __mutex.unlock();
  return RS_OK;
}
