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

#include "src/pk-read/pkr-operation.hpp"
#include "src/pk-read/pkr-request.hpp"
#include "src/pk-read/pkr-response.hpp"
#include "src/decimal_utils.hpp"
#include "src/error-strs.h"
#include "src/logger.hpp"
#include "src/rdrs-const.h"
#include "src/status.hpp"

size_t convert_to_printable(char *to, size_t to_len, const char *from, size_t from_len,
                            const CHARSET_INFO *from_cs, size_t nbytes = 0);

PKROperation::PKROperation(char *reqBuff, char *respBuff, Ndb *ndbObject)
    : request(reqBuff), response(respBuff) {
  this->ndb_object = ndbObject;
}

/**
 * start a transaction
 *
 * @param[in] ndbObject
 * @param[in] pkread
 * @param[out] table
 * @param[out] transaction
 *
 * @return status
 */

RS_Status PKROperation::SetupTransaction() {
  transaction = ndb_object->startTransaction(table_dic);
  if (transaction == nullptr) {
    return RS_RONDB_SERVER_ERROR(ndb_object->getNdbError(), ERROR_005);
  }
  return RS_OK;
}

/**
 * Set up read operation
 *
 * @param[in] ndbObject
 * @param[in] table
 * @param[in] transaction
 * @param[out] operation
 *
 * @return status
 */
RS_Status PKROperation::SetupReadOperation() {
  if (operation != NULL) {
    return RS_CLIENT_ERROR(ERROR_006);
  }

  operation = transaction->getNdbOperation(table_dic);
  if (operation == nullptr) {
    return RS_RONDB_SERVER_ERROR(transaction->getNdbError(), ERROR_007);
  }

  operation->readTuple(NdbOperation::LM_CommittedRead);

  for (Uint32 i = 0; i < request.PKColumnsCount(); i++) {
    RS_Status status = SetOperationPKCols(table_dic->getColumn(request.PKName(i)), i);
    if (status.http_code != SUCCESS) {
      return status;
    }
  }

  if (request.ReadColumnsCount() > 0) {
    for (Uint32 i = 0; i < request.ReadColumnsCount(); i++) {
      NdbRecAttr *rec = operation->getValue(request.ReadColumnName(i), NULL);
      recs.insert(recs.begin(), rec);
    }
  } else {
    std::unordered_map<std::string, const NdbDictionary::Column *>::const_iterator it =
        non_pk_cols.begin();
    while (it != non_pk_cols.end()) {
      NdbRecAttr *rec = operation->getValue(it->first.c_str(), NULL);
      it++;
      recs.insert(recs.begin(), rec);
    }
  }

  return RS_OK;
}

RS_Status PKROperation::Execute() {
  if (transaction->execute(NdbTransaction::Commit) != 0) {
    return RS_RONDB_SERVER_ERROR(transaction->getNdbError(), ERROR_009);
  }

  return RS_OK;
}

RS_Status PKROperation::CreateResponse() {
  if (transaction->getNdbError().classification == NdbError::NoDataFound) {
    char message[] = "NOT FOUND";
    memcpy(response.GetResponseBuffer(), message, sizeof(message));
    response.GetResponseBuffer()[strlen(message)] = 0x00;
  } else {
    // iterate over all columns
    response.Append_string("{", false);
    if (request.OperationId() != NULL) {
      response.Append_string("\"OperationID\": ", false);
      response.Append_string(std::string("\"") + request.OperationId() + std::string("\""), true);
    }
    response.Append_string("\"Data\": {", false);

    for (Uint32 i = 0; i < recs.size(); i++) {
      RS_Status status = response.Append_string(
          std::string("\"") + recs[i]->getColumn()->getName() + std::string("\":"), false);
      if (status.http_code != SUCCESS) {
        return status;
      }

      status = WriteColToRespBuff(recs[i], i == (recs.size() - 1) ? false : true);
      if (status.http_code != SUCCESS) {
        return status;
      }
    }
    response.Append_string("} } ", false);
    response.Append_NULL();
  }

  return RS_OK;
}

int PKROperation::GetByteArray(const NdbRecAttr *attr, const char **first_byte, int *bytes) {
  const NdbDictionary::Column::ArrayType array_type = attr->getColumn()->getArrayType();
  const size_t attr_bytes                           = attr->get_size_in_bytes();
  const char *aRef                                  = attr->aRef();
  std::string result;

  switch (array_type) {
  case NdbDictionary::Column::ArrayTypeFixed:
    /*
     No prefix length is stored in aRef. Data starts from aRef's first byte
     data might be padded with blank or null bytes to fill the whole column
     */
    *first_byte = aRef;
    *bytes      = attr_bytes;
    return 0;
  case NdbDictionary::Column::ArrayTypeShortVar:
    /*
     First byte of aRef has the length of data stored
     Data starts from second byte of aRef
     */
    *first_byte = aRef + 1;
    *bytes      = static_cast<size_t>(aRef[0]);
    return 0;
  case NdbDictionary::Column::ArrayTypeMediumVar:
    /*
     First two bytes of aRef has the length of data stored
     Data starts from third byte of aRef
     */
    *first_byte = aRef + 2;
    *bytes      = static_cast<size_t>(aRef[1]) * 256 + static_cast<size_t>(aRef[0]);
    return 0;
  default:
    first_byte = NULL;
    *bytes     = 0;
    return -1;
  }
}

// https://docs.oracle.com/cd/E17952_01/ndbapi-en/ndbapi-examples-array-simple.html
int PKROperation::CopyString(const NdbRecAttr *attr, int start) {
  int attr_bytes;
  const char *data_start_ptr = NULL;

  /* get stored length and data using get_byte_array */
  if (GetByteArray(attr, &data_start_ptr, &attr_bytes) == 0) {
    memcpy(response.GetResponseBuffer() + start, data_start_ptr, attr_bytes);

    std::string str = std::string(data_start_ptr, attr_bytes);
    /* sprintf(NULL,"PTR: %p\n", data_start_ptr); */
    /* we have length of the string and start location */
    //    str = string(data_start_ptr, attr_bytes);
    //    if (attr->getType() == NdbDictionary::Column::Char) {
    //      /* Fixed Char : remove blank spaces at the end */
    //      size_t endpos = str.find_last_not_of(" ");
    //      if (string::npos != endpos) {
    //        str = str.substr(0, endpos + 1);
    //      }
    //    }
    response.GetResponseBuffer()[start + attr_bytes] = 0x00;
    return start + attr_bytes + 1;
  }
  return -1;
}

RS_Status PKROperation::Init() {
  if (table_dic == NULL) {
    if (ndb_object->setCatalogName(request.DB()) != 0) {
      return RS_CLIENT_ERROR(ERROR_011 + std::string(" Database: ") + std::string(request.DB()) +
                             " Table: " + request.Table());
    }
    const NdbDictionary::Dictionary *dict = ndb_object->getDictionary();
    table_dic                             = dict->getTable(request.Table());

    if (table_dic == nullptr) {
      return RS_CLIENT_ERROR(ERROR_011 + std::string(" Database: ") + std::string(request.DB()) +
                             " Table: " + request.Table());
    }
  }

  // get all primary key columnns
  for (int i = 0; i < table_dic->getNoOfPrimaryKeys(); i++) {
    const char *priName           = table_dic->getPrimaryKey(i);
    pk_cols[std::string(priName)] = table_dic->getColumn(priName);
  }

  // get all non primary key columnns
  for (int i = 0; i < table_dic->getNoOfColumns(); i++) {
    const NdbDictionary::Column *col = table_dic->getColumn(i);
    std::string colNameStr(col->getName());
    std::unordered_map<std::string, const NdbDictionary::Column *>::const_iterator got =
        pk_cols.find(colNameStr);
    if (got == pk_cols.end()) {  // not found
      non_pk_cols[std::string(col->getName())] = table_dic->getColumn(col->getName());
    }
  }

  return RS_OK;
}

RS_Status PKROperation::ValidateRequest() {
  // Check primary key columns
  if (request.PKColumnsCount() != pk_cols.size()) {
    return RS_CLIENT_ERROR(ERROR_013 + std::string(" Expecting: ") +
                           std::to_string(pk_cols.size()) +
                           " Got: " + std::to_string(request.PKColumnsCount()));
  }

  for (Uint32 i = 0; i < request.PKColumnsCount(); i++) {
    std::unordered_map<std::string, const NdbDictionary::Column *>::const_iterator got =
        pk_cols.find(std::string(request.PKName(i)));
    if (got == pk_cols.end()) {  // not found
      return RS_CLIENT_ERROR(ERROR_014 + std::string(" Column: ") + std::string(request.PKName(i)));
    }
  }

  // Check non primary key columns
  // check that all columns exist
  if (request.ReadColumnsCount() > 0) {
    for (Uint32 i = 0; i < request.ReadColumnsCount(); i++) {
      std::unordered_map<std::string, const NdbDictionary::Column *>::const_iterator got =
          non_pk_cols.find(std::string(request.ReadColumnName(i)));
      if (got == non_pk_cols.end()) {  // not found
        return RS_CLIENT_ERROR(ERROR_012 + std::string(" Column: ") +
                               std::string(request.ReadColumnName(i)));
      }
    }
  }

  return RS_OK;
}

void PKROperation::CloseTransaction() {
  ndb_object->closeTransaction(transaction);
}

RS_Status PKROperation::PerformOperation() {
  RS_Status status = Init();
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = ValidateRequest();
  if (status.http_code != SUCCESS) {
    return status;
  }

  status = SetupTransaction();
  if (status.http_code != SUCCESS) {
    this->Abort();
    return status;
  }

  status = SetupReadOperation();
  if (status.http_code != SUCCESS) {
    this->Abort();
    return status;
  }

  status = Execute();
  if (status.http_code != SUCCESS) {
    this->Abort();
    return status;
  }

  status = CreateResponse();
  if (status.http_code != SUCCESS) {
    this->Abort();
    return status;
  }

  CloseTransaction();
  return RS_OK;
}

RS_Status PKROperation::Abort() {
  if (transaction != NULL) {
    NdbTransaction::CommitStatusType status = transaction->commitStatus();
    if (status == NdbTransaction::CommitStatusType::Started) {
      transaction->execute(NdbTransaction::Rollback);
    }
    ndb_object->closeTransaction(transaction);
  }

  return RS_OK;
}

RS_Status PKROperation::WriteColToRespBuff(const NdbRecAttr *attr, bool appendComma) {
  const NdbDictionary::Column *col = attr->getColumn();
  RS_Status status;

  if (attr->isNULL()) {
    return response.Append_string("null", appendComma);
  }

  switch (col->getType()) {
  case NdbDictionary::Column::Undefined: {
    ///< 4 bytes + 0-3 fraction
    return RS_CLIENT_ERROR(ERROR_018 + std::string(" Column: ") + std::string(col->getName()));
  }
  case NdbDictionary::Column::Tinyint: {
    ///< 8 bit. 1 byte signed integer, can be used in array
    status = response.Append_i8(attr->int8_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Tinyunsigned: {
    ///< 8 bit. 1 byte unsigned integer, can be used in array
    status = response.Append_iu8(attr->u_8_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Smallint: {
    ///< 16 bit. 2 byte signed integer, can be used in array
    status = response.Append_i16(attr->short_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Smallunsigned: {
    ///< 16 bit. 2 byte unsigned integer, can be used in array
    status = response.Append_iu16(attr->u_short_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Mediumint: {
    ///< 24 bit. 3 byte signed integer, can be used in array
    status = response.Append_i24(attr->medium_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Mediumunsigned: {
    ///< 24 bit. 3 byte unsigned integer, can be used in array
    status = response.Append_iu24(attr->u_medium_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Int: {
    ///< 32 bit. 4 byte signed integer, can be used in array
    status = response.Append_i32(attr->int32_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Unsigned: {
    ///< 32 bit. 4 byte unsigned integer, can be used in array
    status = response.Append_iu32(attr->u_32_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Bigint: {
    ///< 64 bit. 8 byte signed integer, can be used in array
    status = response.Append_i64(attr->int64_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Bigunsigned: {
    ///< 64 Bit. 8 byte signed integer, can be used in array
    status = response.Append_iu64(attr->u_64_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Float: {
    ///< 32-bit float. 4 bytes float, can be used in array
    status = response.Append_f32(attr->float_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Double: {
    ///< 64-bit float. 8 byte float, can be used in array
    status = response.Append_d64(attr->double_value(), appendComma);
    break;
  }
  case NdbDictionary::Column::Olddecimal: {
    ///< MySQL < 5.0 signed decimal,  Precision, Scale
    return RS_CLIENT_ERROR("Not Implemented. MySQL < 5.0 Olddecimal.");
  }
  case NdbDictionary::Column::Olddecimalunsigned: {
    ///< MySQL < 5.0 signed decimal,  Precision, Scale
    return RS_CLIENT_ERROR("Not Implemented. MySQL < 5.0 Olddecimalunsigned.");
  }
  case NdbDictionary::Column::Decimal:
    ///< MySQL >= 5.0 signed decimal,  Precision, Scale
    [[fallthrough]];
  case NdbDictionary::Column::Decimalunsigned: {
    char decStr[MaxDecimalStrLen];
    int precision = attr->getColumn()->getPrecision();
    int scale     = attr->getColumn()->getScale();
    void *bin     = attr->aRef();
    int bin_len   = attr->get_size_in_bytes();
    decimal_bin2str(bin, bin_len, precision, scale, decStr, MaxDecimalStrLen);
    return response.Append_string(decStr, appendComma);
  }
  case NdbDictionary::Column::Char: {
    ///< Len. A fixed array of 1-byte chars
    int attr_bytes;
    const char *data_start = NULL;
    if (GetByteArray(attr, &data_start, &attr_bytes) != 0) {
      return RS_CLIENT_ERROR(std::string(ERROR_019) + std::string(" Char column. Column: ") +
                             attr->getColumn()->getName());
    } else {
      return response.Append_char(data_start, attr_bytes, attr->getColumn()->getCharset(),
                                  appendComma);
    }
  }
  case NdbDictionary::Column::Varchar: {
    ///< Length bytes: 1, Max: 255
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Varchar")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Binary: {
    ///< Len
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Binary")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Varbinary: {
    ///< Length bytes: 1, Max: 255
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Varbinary")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Datetime: {
    ///< Precision down to 1 sec (sizeof(Datetime) == 8 bytes )
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Datetime")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Date: {
    ///< Precision down to 1 day(sizeof(Date) == 4 bytes )
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Date")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Blob: {
    ///< Binary large object (see NdbBlob)
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Blob")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Text: {
    ///< Text blob
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Text")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Bit: {
    ///< Bit, length specifies no of bits
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Bit")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Longvarchar: {
    ///< Length bytes: 2, little-endian
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Longvarchar")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Longvarbinary: {
    ///< Length bytes: 2, little-endian
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Longvarbinary")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Time: {
    ///< Time without date
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Time")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Year: {
    ///< Year 1901-2155 (1 byte)
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Year")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Timestamp: {
    ///< Unix time
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Timestamp")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  ///**
  // * Time types in MySQL 5.6 add microsecond fraction.
  // * One should use setPrecision(x) to set number of fractional
  // * digits (x = 0-6, default 0).  Data formats are as in MySQL
  // * and must use correct byte length.  NDB does not check data
  // * itself since any values can be compared as binary strings.
  // */
  case NdbDictionary::Column::Time2: {
    ///< 3 bytes + 0-3 fraction
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Time2")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Datetime2: {
    ///< 5 bytes plus 0-3 fraction
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Datetime2")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Timestamp2: {
    ///< 4 bytes + 0-3 fraction
    TRACE(std::string("Getting PK Column: ") + std::string(col->getName()) + " Type: Timestamp2");
    return RS_CLIENT_ERROR("Not Implemented");
  }
  }

  if (status.http_code != SUCCESS) {
    return status;
  } else {
    return RS_OK;
  }
}

RS_Status PKROperation::SetOperationPKCols(const NdbDictionary::Column *col, Uint32 colIdx) {
  // validate the data and set data according to column type
  char *data;

  switch (col->getType()) {
  case NdbDictionary::Column::Undefined: {
    ///< 4 bytes + 0-3 fraction
    return RS_CLIENT_ERROR(ERROR_018 + std::string(" Column: ") +
                           std::string(request.PKName(colIdx)));
  }
  case NdbDictionary::Column::Tinyint: {
    ///< 8 bit. 1 byte signed integer, can be used in array
    bool success = false;
    try {
      int num = std::stoi(request.PKValueCStr(colIdx));
      if (num >= -128 && num <= 127) {
        operation->equal(request.PKName(colIdx), static_cast<char>(num));
        success = true;
      }
    } catch (...) {
    }
    if (!success) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting TINYINT. Column: ") +
                             std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Tinyunsigned: {
    ///< 8 bit. 1 byte unsigned integer, can be used in array
    bool success = false;
    try {
      int num = std::stoi(request.PKValueCStr(colIdx));
      if (num >= 0 && num <= 255) {
        operation->equal(request.PKName(colIdx), static_cast<char>(num));
        success = true;
      }
    } catch (...) {
    }
    if (!success) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting TINYINT. Column: ") +
                             std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Smallint: {
    ///< 16 bit. 2 byte signed integer, can be used in array
    bool success = false;
    try {
      int num = std::stoi(request.PKValueCStr(colIdx));
      if (num >= -32768 && num <= 32767) {
        operation->equal(request.PKName(colIdx), (Int16)num);
        success = true;
      }
    } catch (...) {
    }
    if (!success) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting SMALLINT. Column: ") +
                             std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Smallunsigned: {
    ///< 16 bit. 2 byte unsigned integer, can be used in array
    bool success = false;
    try {
      int num = std::stoi(request.PKValueCStr(colIdx));
      if (num >= 0 && num <= 65535) {
        operation->equal(request.PKName(colIdx), (Uint16)num);
        success = true;
      }
    } catch (...) {
    }
    if (!success) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting TINYINT UNSIGNED. Column: ") +
                             std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Mediumint: {
    ///< 24 bit. 3 byte signed integer, can be used in array
    bool success = false;
    try {
      int num = std::stoi(request.PKValueCStr(colIdx));
      if (num >= -8388608 && num <= 8388607) {
        operation->equal(request.PKName(colIdx), static_cast<int>(num));
        success = true;
      }
    } catch (...) {
    }
    if (!success) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting MEDIUMINT. Column: ") +
                             std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Mediumunsigned: {
    ///< 24 bit. 3 byte unsigned integer, can be used in array
    bool success = false;
    try {
      int num = std::stoi(request.PKValueCStr(colIdx));
      if (num >= 0 && num <= 16777215) {
        operation->equal(request.PKName(colIdx), (unsigned int)num);
        success = true;
      }
    } catch (...) {
    }
    if (!success) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting MEDIUMINT UNSIGNED. Column: ") +
                             std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Int: {
    ///< 32 bit. 4 byte signed integer, can be used in array
    try {
      Int32 num = std::stoi(request.PKValueCStr(colIdx));
      operation->equal(request.PKName(colIdx), num);
    } catch (...) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting Int. Column: ") +
                             std::string(request.PKName(colIdx)));
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Unsigned: {
    ///< 32 bit. 4 byte unsigned integer, can be used in array
    bool success = false;
    try {
      Int64 lresult = std::stoll(request.PKValueCStr(colIdx));
      Uint32 result = lresult;
      if (result == lresult) {
        operation->equal(request.PKName(colIdx), result);
        success = true;
      }
    } catch (...) {
    }

    if (!success) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting Unsigned Int. Column: ") +
                             std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Bigint: {
    ///< 64 bit. 8 byte signed integer, can be used in array
    try {
      Int64 num = std::stoll(request.PKValueCStr(colIdx));
      operation->equal(request.PKName(colIdx), num);
    } catch (...) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting BIGINT. Column: ") +
                             std::string(request.PKName(colIdx)));
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Bigunsigned: {
    ///< 64 Bit. 8 byte signed integer, can be used in array
    bool success = false;
    try {
      const char *numCStr      = request.PKValueCStr(colIdx);
      const std::string numStr = std::string(numCStr);
      if (numStr.find('-') == std::string::npos) {
        Uint64 num = std::stoul(numCStr);
        operation->equal(request.PKName(colIdx), num);
        success = true;
      }
    } catch (...) {
    }
    if (!success) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting BIGINT UNSIGNED. Column: ") +
                             std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Float: {
    ///< 32-bit float. 4 bytes float, can be used in array
    return RS_CLIENT_ERROR(ERROR_017 + std::string(" Column: ") +
                           std::string(request.PKName(colIdx)));
  }
  case NdbDictionary::Column::Double: {
    ///< 64-bit float. 8 byte float, can be used in array
    return RS_CLIENT_ERROR(ERROR_017 + std::string(" Column: ") +
                           std::string(request.PKName(colIdx)));
  }
  case NdbDictionary::Column::Olddecimal: {
    ///< MySQL < 5.0 signed decimal,  Precision, Scale
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Olddecimal")
    return RS_CLIENT_ERROR("Not Implemented. Type: Olddecimal");
  }
  case NdbDictionary::Column::Olddecimalunsigned: {
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) +
          " Type: Olddecimalunsigned")
    return RS_CLIENT_ERROR("Not Implemented. Type: Olddecimalunsigned");
  }
  case NdbDictionary::Column::Decimalunsigned: {
    ///< MySQL >= 5.0 signed decimal,  Precision, Scale
    const std::string decStr = std::string(request.PKValueCStr(colIdx));
    if (decStr.find('-') != std::string::npos) {
      return RS_CLIENT_ERROR(ERROR_015 +
                             std::string(" Expecting Decimalunsigned UNSIGNED. Column: ") +
                             std::string(request.PKName(colIdx)));
    }
    [[fallthrough]];
  }
  case NdbDictionary::Column::Decimal: {
    int precision      = col->getPrecision();
    int scale          = col->getScale();
    int bytesNeeded    = getDecimalColumnSpace(precision, scale);
    const char *decStr = request.PKValueCStr(colIdx);
    char decBin[bytesNeeded];
    if (decimal_str2bin(decStr, strlen(decStr), precision, scale, decBin, bytesNeeded) != 0) {
      return RS_CLIENT_ERROR(ERROR_015 + std::string(" Expecting Decimal with Precision: ") +
                             std::to_string(precision) + std::string(" and Scale: ") +
                             std::to_string(scale));
    }
    operation->equal(request.PKName(colIdx), decBin, bytesNeeded);
    return RS_OK;
  }
  case NdbDictionary::Column::Char: {
    ///< Len. A fixed array of 1-byte chars
    const char *charStr = request.PKValueCStr(colIdx);
    const int len       = strlen(charStr);
    if (len > col->getLength()) {
      return RS_CLIENT_ERROR(ERROR_020 + std::string(" Column: ") + request.PKName(colIdx));
    }
    char pk[col->getLength()];
    for (int i = 0; i < col->getLength(); i++) {
      pk[i] = 0;
    }
    memcpy(pk, charStr, len);
    operation->equal(request.PKName(colIdx), pk, col->getLength());
    return RS_OK;
  }
  case NdbDictionary::Column::Varchar: {
    ///< Length bytes: 1, Max: 255
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Varchar")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Binary: {
    ///< Len
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Binary")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Varbinary: {
    ///< Length bytes: 1, Max: 255
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Varbinary")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Datetime: {
    ///< Precision down to 1 sec (sizeof(Datetime) == 8 bytes )
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Datetime")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Date: {
    ///< Precision down to 1 day(sizeof(Date) == 4 bytes )
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Date")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Blob: {
    ///< Binary large object (see NdbBlob)
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Blob")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Text: {
    ///< Text blob
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Text")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Bit: {
    ///< Bit, length specifies no of bits
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Bit")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Longvarchar: {
    ///< Length bytes: 2, little-endian
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Longvarchar")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Longvarbinary: {
    ///< Length bytes: 2, little-endian
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Longvarbinary")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Time: {
    ///< Time without date
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Time")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Year: {
    ///< Year 1901-2155 (1 byte)
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Year")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Timestamp: {
    ///< Unix time
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Timestamp")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  ///**
  // * Time types in MySQL 5.6 add microsecond fraction.
  // * One should use setPrecision(x) to set number of fractional
  // * digits (x = 0-6, default 0).  Data formats are as in MySQL
  // * and must use correct byte length.  NDB does not check data
  // * itself since any values can be compared as binary strings.
  // */
  case NdbDictionary::Column::Time2: {
    ///< 3 bytes + 0-3 fraction
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Time2")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Datetime2: {
    ///< 5 bytes plus 0-3 fraction
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Datetime2")
    return RS_CLIENT_ERROR("Not Implemented");
  }
  case NdbDictionary::Column::Timestamp2: {
    ///< 4 bytes + 0-3 fraction
    TRACE(std::string("Setting PK Column: ") + std::string(col->getName()) + " Type: Timestamp2");
    return RS_CLIENT_ERROR("Not Implemented");
  }
  }

  /* if (request.pkValueNDBStr(i, tableDic->getColumn(request.pkName(i)), &data) != 0)  */
  /* operation->equal(request.pkName(i), data); */
  return RS_OK;
}

