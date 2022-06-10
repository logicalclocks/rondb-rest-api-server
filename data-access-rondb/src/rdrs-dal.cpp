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

#include "src/rdrs-dal.h"
#include <mgmapi.h>
#include <cstdlib>
#include <cstring>
#include <string>
#include <iostream>
#include <iterator>
#include <sstream>
#include <NdbApi.hpp>
#include "src/error-strs.h"
#include "src/logger.hpp"
#include "src/db-operations/pk/pkr-operation.hpp"
#include "src/status.hpp"
#include "src/ndb_object_pool.hpp"
#include "src/db-operations/pk/common.hpp"

int GetAvailableAPINode(const char *connection_string);

Ndb_cluster_connection *ndb_connection;

/**
 * Initialize NDB connection
 * @param connection_string NDB connection string {url}:{port}
 * @param find_available_node_ID if set to 1 then we will first find an available node id to
 * connect to
 * @return status
 */
RS_Status init(const char *connection_string, _Bool find_available_node_id) {

  int retCode = 0;
  DEBUG(std::string("Connecting to ") + connection_string);

  retCode = ndb_init();
  if (retCode != 0) {
    return RS_SERVER_ERROR(ERROR_001 + std::string(" RetCode: ") + std::to_string(retCode));
  }

  int node_id = -1;
  if (find_available_node_id == true) {
    node_id = GetAvailableAPINode(connection_string);
    if (node_id == -1) {
      return RS_SERVER_ERROR(ERROR_024);
    }
  }

  if (node_id != -1) {
    ndb_connection = new Ndb_cluster_connection(connection_string, node_id);
  } else {
    ndb_connection = new Ndb_cluster_connection(connection_string);
  }
  retCode = ndb_connection->connect();
  if (retCode != 0) {
    return RS_SERVER_ERROR(ERROR_002 + std::string(" RetCode: ") + std::to_string(retCode));
  }

  retCode = ndb_connection->wait_until_ready(30, 0);
  if (retCode != 0) {
    return RS_SERVER_ERROR(ERROR_003 + std::string(" RetCode: ") + std::to_string(retCode));
  }

  // Initialize NDB Object Pool
  NdbObjectPool::InitPool();

  DEBUG("Connected.");
  return RS_OK;
}

RS_Status shutdown_connection() {
  try {
    // ndb_end(0); // causes seg faults when called repeated from unit tests*/
    NdbObjectPool::GetInstance()->Close();
    delete ndb_connection;
  } catch (...) {
    WARN("Exception in Shutdown");
  }
  return RS_OK;
}

/**
 * Closes a NDB Object
 *
 * @param[int] ndb_object
 *
 * @return status
 */
RS_Status closeNDBObject(Ndb *ndb_object) {
  NdbObjectPool::GetInstance()->ReturnResource(ndb_object);
  return RS_OK;
}

RS_Status pk_read(RS_Buffer *reqBuff, RS_Buffer *respBuff) {
  Ndb *ndb_object  = nullptr;
  RS_Status status = NdbObjectPool::GetInstance()->GetNdbObject(ndb_connection, &ndb_object);
  if (status.http_code != SUCCESS) {
    return status;
  }

  PKROperation pkread(reqBuff, respBuff, ndb_object);

  status = pkread.PerformOperation();
  closeNDBObject(ndb_object);
  if (status.http_code != SUCCESS) {
    return status;
  }

  return RS_OK;
}

/**
 * Batched primary key read operation
 */

RS_Status pk_batch_read(unsigned int no_req, RS_Buffer *req_buffs, RS_Buffer *resp_buffs) {
  Ndb *ndb_object  = nullptr;
  RS_Status status = NdbObjectPool::GetInstance()->GetNdbObject(ndb_connection, &ndb_object);
  if (status.http_code != SUCCESS) {
    return status;
  }

  PKROperation pkread(no_req, req_buffs, resp_buffs, ndb_object);

  status = pkread.PerformOperation();
  closeNDBObject(ndb_object);
  if (status.http_code != SUCCESS) {
    return status;
  }

  return RS_OK;
}

/**
 * Deallocate pointer array
 */
RS_Status get_rondb_stats(RonDB_Stats *stats) {
  RonDB_Stats ret              = NdbObjectPool::GetInstance()->GetStats();
  stats->ndb_objects_created   = ret.ndb_objects_created;
  stats->ndb_objects_deleted   = ret.ndb_objects_deleted;
  stats->ndb_objects_count     = ret.ndb_objects_count;
  stats->ndb_objects_available = ret.ndb_objects_available;

  return RS_OK;
}

static int LastConnectedInodeID = -1;
/*
 * NDB API does not support gracefull disconnection form the
 * cluster. All disconnections are treated as failures. When
 * you disconnect, the API node is not able to accept new
 * connections until the filure recovery has completed for
 * the API node. This can take upto ~5 sec, slowing down
 * unit tests which start/stop the NDB API multiple times.
 * This function returns next available API node to connect to.
 */
int GetAvailableAPINode(const char *connection_string) {
  NdbMgmHandle h;

  h = ndb_mgm_create_handle();
  if (h == 0) {
    PANIC("Failed to create mgm handle");
    return -1;
  }

  if (ndb_mgm_set_connectstring(h, connection_string) == -1) {
    PANIC("Failed set mgm connect string");
    return -1;
  }

  if (ndb_mgm_connect(h, 0, 0, 0)) {
    PANIC("Failed to connect to mgm node");
    return -1;
  }

  // look for api nodes only
  ndb_mgm_node_type node_types[2]   = {NDB_MGM_NODE_TYPE_API, NDB_MGM_NODE_TYPE_UNKNOWN};
  struct ndb_mgm_cluster_state *ret = ndb_mgm_get_status2(h, node_types);

  if (ret->no_of_nodes > 1) {
    int max_node_id = ret->node_states[0].node_id;
    for (int i = 1; i < ret->no_of_nodes; i++) {
      if (ret->node_states[i].node_id > max_node_id) {
        max_node_id = ret->node_states[i].node_id;
      }
    }

    if (LastConnectedInodeID == max_node_id) {
      LastConnectedInodeID = -1;
    }

    for (int i = 0; i < ret->no_of_nodes; i++) {
      if (ret->node_states[i].node_id > LastConnectedInodeID &&
          ret->node_states[i].node_status == NDB_MGM_NODE_STATUS_NO_CONTACT) {
        LastConnectedInodeID = ret->node_states[i].node_id;
        free(ret);
        return LastConnectedInodeID;
      }
    }
  }

  free(ret);
  return -1;
}

/**
 * Register callbacks
 */
void register_callbacks(Callbacks cbs) {
  setLogCallBackFns(cbs);
}

RS_Status select_table(Ndb *ndb_object, const char *database_str, const char *table_str,
                       const NdbDictionary::Table **table_dict) {
  if (ndb_object->setCatalogName(database_str) != 0) {
    return RS_CLIENT_ERROR(ERROR_011 + std::string(" Database: ") + std::string(database_str) +
                           std::string(". Table: ") + std::string(table_str));
  }

  const NdbDictionary::Dictionary *dict = ndb_object->getDictionary();
  *table_dict                           = dict->getTable(table_str);

  if (table_dict == nullptr) {
    return RS_CLIENT_ERROR(ERROR_011 + std::string(" Database: ") + std::string(database_str) +
                           std::string(". Table: ") + std::string(table_str));
  }
  return RS_OK;
}

RS_Status start_transaction(Ndb *ndb_object, NdbTransaction **tx) {
  NdbError err;
  *tx = ndb_object->startTransaction();
  if (tx == nullptr) {
    err = ndb_object->getNdbError();
    return RS_RONDB_SERVER_ERROR(err, ERROR_005);
  }
  return RS_OK;
}

RS_Status get_scan_op(Ndb *ndb_object, NdbTransaction *tx, const NdbDictionary::Table *table_dict,
                      NdbScanOperation **scanOp) {
  NdbError err;
  *scanOp = tx->getNdbScanOperation(table_dict);
  if (scanOp == nullptr) {
    err = ndb_object->getNdbError();
    return RS_RONDB_SERVER_ERROR(err, ERROR_029);
  }
  return RS_OK;
}

RS_Status read_tuples(Ndb *ndb_object, NdbScanOperation *scanOp) {
  NdbError err;
  if (scanOp->readTuples(NdbOperation::LM_Exclusive) != 0) {
    err = ndb_object->getNdbError();
    return RS_RONDB_SERVER_ERROR(err, ERROR_030);
  }
  return RS_OK;
}

RS_Status find_api_key(Ndb *ndb_object, const char *prefix, HopsworksAPIKey *api_key) {

  NdbError err;
  const NdbDictionary::Table *table_dict;
  NdbTransaction *tx;
  NdbScanOperation *scanOp;

  RS_Status status = select_table(ndb_object, "hopsworks", "api_key", &table_dict);
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = start_transaction(ndb_object, &tx);
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = get_scan_op(ndb_object, tx, table_dict, &scanOp);
  if (status.http_code != SUCCESS) {
    ndb_object->closeTransaction(tx);
    return status;
  }

  status = read_tuples(ndb_object, scanOp);
  if (status.http_code != SUCCESS) {
    ndb_object->closeTransaction(tx);
    return status;
  }

  int col_id      = table_dict->getColumn("prefix")->getColumnNo();
  Uint32 col_size = (Uint32)table_dict->getColumn("prefix")->getSizeInBytes();
  if (strlen(prefix) > col_size) {
    return RS_CLIENT_ERROR("Wrong length of the search key");
  }

  char cmp_str[col_size];
  memcpy(cmp_str + 1, prefix, col_size - 1);
  cmp_str[0] = (char)strlen(prefix);

  NdbScanFilter filter(scanOp);
  if (filter.begin(NdbScanFilter::AND) < 0 ||
      filter.cmp(NdbScanFilter::COND_EQ, col_id, cmp_str, col_size) < 0 || filter.end() < 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_031);
  }

  bool check;
  NdbRecAttr *user_id = scanOp->getValue("user_id");
  NdbRecAttr *secret  = scanOp->getValue("secret");
  NdbRecAttr *salt    = scanOp->getValue("salt");
  NdbRecAttr *name    = scanOp->getValue("name");

  if (user_id == nullptr || secret == nullptr || salt == nullptr || name == nullptr) {
    return RS_RONDB_SERVER_ERROR(err, ERROR_019);
  }

  if (tx->execute(NdbTransaction::NoCommit) != 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_009);
  }

  while ((check = scanOp->nextResult(true)) == 0) {
    do {

      Uint32 name_attr_bytes;
      const char *name_data_start = nullptr;
      if (GetByteArray(name, &name_data_start, &name_attr_bytes) != 0) {
        return RS_CLIENT_ERROR(ERROR_019);
      }

      Uint32 salt_attr_bytes;
      const char *salt_data_start = nullptr;
      if (GetByteArray(salt, &salt_data_start, &salt_attr_bytes) != 0) {
        return RS_CLIENT_ERROR(ERROR_019);
      }

      Uint32 secret_attr_bytes;
      const char *secret_data_start = nullptr;
      if (GetByteArray(secret, &secret_data_start, &secret_attr_bytes) != 0) {
        return RS_CLIENT_ERROR(ERROR_019);
      }

      if (sizeof(api_key->secret) < secret_attr_bytes || sizeof(api_key->name) < name_attr_bytes ||
          sizeof(api_key->salt) < salt_attr_bytes) {
        return RS_CLIENT_ERROR(ERROR_021);
      }

      memcpy(api_key->name, name_data_start, name_attr_bytes);
      api_key->name[name_attr_bytes] = 0;

      memcpy(api_key->secret, secret_data_start, secret_attr_bytes);
      api_key->secret[secret_attr_bytes] = 0;

      memcpy(api_key->salt, salt_data_start, salt_attr_bytes);
      api_key->salt[salt_attr_bytes] = 0;

      api_key->user_id = user_id->int32_value();
    } while ((check = scanOp->nextResult(false)) == 0);
  }

  ndb_object->closeTransaction(tx);

  return RS_OK;
}

RS_Status find_user(Ndb *ndb_object, HopsworksAPIKey api_key, HopsworksUsers *users) {

  NdbError err;
  const NdbDictionary::Table *table_dict;
  NdbTransaction *tx;
  NdbScanOperation *scanOp;

  RS_Status status = select_table(ndb_object, "hopsworks", "users", &table_dict);
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = start_transaction(ndb_object, &tx);
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = get_scan_op(ndb_object, tx, table_dict, &scanOp);
  if (status.http_code != SUCCESS) {
    ndb_object->closeTransaction(tx);
    return status;
  }

  status = read_tuples(ndb_object, scanOp);
  if (status.http_code != SUCCESS) {
    ndb_object->closeTransaction(tx);
    return status;
  }

  int col_id = table_dict->getColumn("uid")->getColumnNo();
  NdbScanFilter filter(scanOp);
  if (filter.begin(NdbScanFilter::AND) < 0 || filter.eq(col_id, (Uint32)api_key.user_id) < 0 ||
      filter.end() < 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_031);
  }

  bool check;
  NdbRecAttr *email = scanOp->getValue("email");

  if (email == nullptr) {
    return RS_RONDB_SERVER_ERROR(err, ERROR_019);
  }

  if (tx->execute(NdbTransaction::NoCommit) != 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_009);
  }

  while ((check = scanOp->nextResult(true)) == 0) {
    do {

      Uint32 email_attr_bytes;
      const char *email_data_start = nullptr;
      if (GetByteArray(email, &email_data_start, &email_attr_bytes) != 0) {
        return RS_CLIENT_ERROR(ERROR_019);
      }

      if (sizeof(users->email) < email_attr_bytes) {
        return RS_CLIENT_ERROR(ERROR_021);
      }

      memcpy(users->email, email_data_start, email_attr_bytes);
      users->email[email_attr_bytes] = 0;

    } while ((check = scanOp->nextResult(false)) == 0);
  }

  ndb_object->closeTransaction(tx);

  return RS_OK;
}

RS_Status find_project_team(Ndb *ndb_object, HopsworksUsers users,
                            std::vector<HopsworksProjectTeam> *project_team_vec) {

  NdbError err;
  const NdbDictionary::Table *table_dict;
  NdbTransaction *tx;
  NdbScanOperation *scanOp;

  RS_Status status = select_table(ndb_object, "hopsworks", "project_team", &table_dict);
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = start_transaction(ndb_object, &tx);
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = get_scan_op(ndb_object, tx, table_dict, &scanOp);
  if (status.http_code != SUCCESS) {
    ndb_object->closeTransaction(tx);
    return status;
  }

  status = read_tuples(ndb_object, scanOp);
  if (status.http_code != SUCCESS) {
    ndb_object->closeTransaction(tx);
    return status;
  }

  int col_id      = table_dict->getColumn("team_member")->getColumnNo();
  Uint32 col_size = (Uint32)table_dict->getColumn("team_member")->getSizeInBytes();
  if (strlen(users.email) > col_size) {
    return RS_CLIENT_ERROR("Wrong length of the search key");
  }

  char cmp_str[col_size];
  memcpy(cmp_str + 1, users.email, strlen(users.email));
  cmp_str[0] = (char)strlen(users.email);

  NdbScanFilter filter(scanOp);
  if (filter.begin(NdbScanFilter::AND) < 0 ||
      filter.cmp(NdbScanFilter::COND_EQ, col_id, cmp_str, col_size) < 0 || filter.end() < 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_031);
  }

  bool check;
  NdbRecAttr *project_id = scanOp->getValue("project_id");

  if (project_id == nullptr) {
    return RS_RONDB_SERVER_ERROR(err, ERROR_019);
  }

  if (tx->execute(NdbTransaction::NoCommit) != 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_009);
  }

  while ((check = scanOp->nextResult(true)) == 0) {
    do {
      HopsworksProjectTeam project_team;
      project_team.porject_id = project_id->int32_value();
      project_team_vec->push_back(project_team);
    } while ((check = scanOp->nextResult(false)) == 0);
  }

  ndb_object->closeTransaction(tx);

  return RS_OK;
}

RS_Status find_projects(Ndb *ndb_object, std::vector<HopsworksProjectTeam> *project_team_vec,
                        std::vector<HopsworksProject> *project_vec) {

  NdbError err;
  const NdbDictionary::Table *table_dict;
  NdbTransaction *tx;
  NdbScanOperation *scanOp;

  RS_Status status = select_table(ndb_object, "hopsworks", "project", &table_dict);
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = start_transaction(ndb_object, &tx);
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = get_scan_op(ndb_object, tx, table_dict, &scanOp);
  if (status.http_code != SUCCESS) {
    ndb_object->closeTransaction(tx);
    return status;
  }

  status = read_tuples(ndb_object, scanOp);
  if (status.http_code != SUCCESS) {
    ndb_object->closeTransaction(tx);
    return status;
  }

  int col_id = table_dict->getColumn("id")->getColumnNo();

  NdbScanFilter filter(scanOp);
  if (filter.begin(NdbScanFilter::OR) < 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_031);
  }

  for (Uint32 i = 0; i < project_team_vec->size(); i++) {
    if (filter.eq(col_id, (Uint32)(*project_team_vec)[i].porject_id) < 0) {
      err = ndb_object->getNdbError();
      ndb_object->closeTransaction(tx);
      return RS_RONDB_SERVER_ERROR(err, ERROR_031);
    }
  }

  if (filter.end() < 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_031);
  }

  bool check;
  NdbRecAttr *projectname = scanOp->getValue("projectname");

  if (projectname == nullptr) {
    return RS_RONDB_SERVER_ERROR(err, ERROR_019);
  }

  if (tx->execute(NdbTransaction::NoCommit) != 0) {
    err = ndb_object->getNdbError();
    ndb_object->closeTransaction(tx);
    return RS_RONDB_SERVER_ERROR(err, ERROR_009);
  }

  while ((check = scanOp->nextResult(true)) == 0) {
    do {
      HopsworksProject project;
      Uint32 projectname_attr_bytes;
      const char *projectname_data_start = nullptr;
      if (GetByteArray(projectname, &projectname_data_start, &projectname_attr_bytes) != 0) {
        return RS_CLIENT_ERROR(ERROR_019);
      }

      if (sizeof(project.porjectname) < projectname_attr_bytes) {
        return RS_CLIENT_ERROR(ERROR_021);
      }

      memcpy(project.porjectname, projectname_data_start, projectname_attr_bytes);
      project.porjectname[projectname_attr_bytes] = 0;
      project_vec->push_back(project);

    } while ((check = scanOp->nextResult(false)) == 0);
  }

  ndb_object->closeTransaction(tx);

  return RS_OK;
}

/**
 * only for testing
 */
int main(int argc, char **argv) {
  char connection_string[] = "localhost:1186";
  init(connection_string, true);

  Ndb *ndb_object  = nullptr;
  RS_Status status = NdbObjectPool::GetInstance()->GetNdbObject(ndb_connection, &ndb_object);
  if (status.http_code != SUCCESS) {
    INFO("Failed to get the NDB object");
    return 1;
  }

  HopsworksAPIKey api_key;
  status = find_api_key(ndb_object, "ZaCRiVfQOxuOIXZk", &api_key);
  if (status.http_code != SUCCESS) {
    ERROR(status.message);
  }

  std::cout << "User ID: " << api_key.user_id << std::endl;
  std::cout << "name: " << api_key.name << std::endl;
  std::cout << "secret: " << api_key.secret << std::endl;
  std::cout << "salt: " << api_key.salt << std::endl;

  HopsworksUsers user;
  status = find_user(ndb_object, api_key, &user);
  if (status.http_code != SUCCESS) {
    ERROR(status.message);
  }
  std::cout << "Email: " << user.email << std::endl;

  std::vector<HopsworksProjectTeam> project_team_vec;
  status = find_project_team(ndb_object, user, &project_team_vec);
  if (status.http_code != SUCCESS) {
    ERROR(status.message);
  }

  for (Uint32 i = 0; i < project_team_vec.size(); i++) {
    std::cout << "Proj ID : " << project_team_vec[i].porject_id << std::endl;
  }

  std::vector<HopsworksProject> project_vec;
  status = find_projects(ndb_object, &project_team_vec, &project_vec);
  if (status.http_code != SUCCESS) {
    ERROR(status.message);
  }

  for (Uint32 i = 0; i < project_vec.size(); i++) {
    std::cout << "Proj Name : " << project_vec[i].porjectname << std::endl;
  }

  ndb_end(0);
}

