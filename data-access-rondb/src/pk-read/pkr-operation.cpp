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
#include <mysql_time.h>
#include <algorithm>
#include <utility>
#include <boost/beast/core/detail/base64.hpp>
#include <NdbDictionary.hpp>
#include "src/pk-read/pkr-request.hpp"
#include "src/pk-read/pkr-response.hpp"
#include "src/decimal_utils.hpp"
#include "src/error-strs.h"
#include "src/logger.hpp"
#include "src/rdrs-const.h"
#include "src/status.hpp"
#include "src/common/rdrs_date.hpp"
#include <boost/date_time/posix_time/posix_time.hpp>

PKROperation::PKROperation(RS_Buffer *reqBuff, RS_Buffer *respBuff, Ndb *ndbObject)
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
  if (operation != nullptr) {
    return RS_CLIENT_ERROR(ERROR_006);
  }

  operation = transaction->getNdbOperation(table_dic);
  if (operation == nullptr) {
    return RS_RONDB_SERVER_ERROR(transaction->getNdbError(), ERROR_007);
  }

  if (operation->readTuple(NdbOperation::LM_CommittedRead) != 0) {
    return RS_SERVER_ERROR(ERROR_022)
  }

  for (Uint32 i = 0; i < request.PKColumnsCount(); i++) {
    RS_Status status = SetOperationPKCols(table_dic->getColumn(request.PKName(i)), i);
    if (status.http_code != SUCCESS) {
      return status;
    }
  }

  if (request.ReadColumnsCount() > 0) {
    for (Uint32 i = 0; i < request.ReadColumnsCount(); i++) {
      NdbRecAttr *rec = operation->getValue(request.ReadColumnName(i), nullptr);
      recs.insert(recs.begin(), rec);
    }
  } else {
    std::unordered_map<std::string, const NdbDictionary::Column *>::const_iterator it =
        non_pk_cols.begin();
    while (it != non_pk_cols.end()) {
      NdbRecAttr *rec = operation->getValue(it->first.c_str(), nullptr);
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
    return RS_CLIENT_404_ERROR();
  } else {
    // iterate over all columns
    response.Append_string("{", false, false);
    if (request.OperationId() != nullptr) {
      response.Append_string("\"operationId\": ", false, false);
      response.Append_string(std::string("\"") + request.OperationId() + std::string("\""), false,
                             true);
    }
    response.Append_string("\"Data\": {", false, false);

    for (Uint32 i = 0; i < recs.size(); i++) {
      RS_Status status = response.Append_string(
          std::string("\"") + recs[i]->getColumn()->getName() + std::string("\":"), false, false);
      if (status.http_code != SUCCESS) {
        return status;
      }

      status = WriteColToRespBuff(recs[i], i == (recs.size() - 1) ? false : true);
      if (status.http_code != SUCCESS) {
        return status;
      }
    }
    response.Append_string("} } ", false, false);
    response.Append_NULL();
    return RS_OK;
  }
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
    first_byte = nullptr;
    *bytes     = 0;
    return -1;
  }
}

RS_Status PKROperation::Init() {
  if (table_dic == nullptr) {
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
  // check that data return type is supported
  // check for reading blob columns
  if (request.ReadColumnsCount() > 0) {
    for (Uint32 i = 0; i < request.ReadColumnsCount(); i++) {
      std::unordered_map<std::string, const NdbDictionary::Column *>::const_iterator got =
          non_pk_cols.find(std::string(request.ReadColumnName(i)));
      if (got == non_pk_cols.end()) {  // not found
        return RS_CLIENT_ERROR(ERROR_012 + std::string(" Column: ") +
                               std::string(request.ReadColumnName(i)));
      }

      // check that the data return type is supported
      // for now we only support DataReturnType.DEFAULT
      if (request.ReadColumnReturnType(i) > __MAX_TYPE_NOT_A_DRT ||
          DEFAULT_DRT != request.ReadColumnReturnType(i)) {
        return RS_SERVER_ERROR(ERROR_025 + std::string(" Column: ") +
                               std::string(request.ReadColumnName(i)));
      }

      if (table_dic->getColumn(request.ReadColumnName(i))->getType() ==
              NdbDictionary::Column::Blob ||
          table_dic->getColumn(request.ReadColumnName(i))->getType() ==
              NdbDictionary::Column::Text) {
        return RS_SERVER_ERROR(ERROR_026 + std::string(" Column: ") +
                               std::string(request.ReadColumnName(i)));
      }
    }
  } else {
    // user wants to read all columns. make sure that we are not reading Blobs
    std::unordered_map<std::string, const NdbDictionary::Column *>::const_iterator it =
        non_pk_cols.begin();
    while (it != non_pk_cols.end()) {
      NdbDictionary::Column::Type type = it->second->getType();
      std::cout << "here 2 --" << std::endl;
      if (type == NdbDictionary::Column::Blob || type == NdbDictionary::Column::Text) {
        return RS_SERVER_ERROR(ERROR_026 + std::string(" Column: ") + it->first);
      }
      it++;
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
  if (transaction != nullptr) {
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
  if (attr->isNULL()) {
    return response.Append_string("null", false, appendComma);
  }

  switch (col->getType()) {
  case NdbDictionary::Column::Undefined: {
    ///< 4 bytes + 0-3 fraction
    return RS_CLIENT_ERROR(ERROR_018 + std::string(" Column: ") + std::string(col->getName()));
  }
  case NdbDictionary::Column::Tinyint: {
    ///< 8 bit. 1 byte signed integer, can be used in array
    return response.Append_i8(attr->int8_value(), appendComma);
  }
  case NdbDictionary::Column::Tinyunsigned: {
    ///< 8 bit. 1 byte unsigned integer, can be used in array
    return response.Append_iu8(attr->u_8_value(), appendComma);
  }
  case NdbDictionary::Column::Smallint: {
    ///< 16 bit. 2 byte signed integer, can be used in array
    return response.Append_i16(attr->short_value(), appendComma);
  }
  case NdbDictionary::Column::Smallunsigned: {
    ///< 16 bit. 2 byte unsigned integer, can be used in array
    return response.Append_iu16(attr->u_short_value(), appendComma);
  }
  case NdbDictionary::Column::Mediumint: {
    ///< 24 bit. 3 byte signed integer, can be used in array
    return response.Append_i24(attr->medium_value(), appendComma);
  }
  case NdbDictionary::Column::Mediumunsigned: {
    ///< 24 bit. 3 byte unsigned integer, can be used in array
    return response.Append_iu24(attr->u_medium_value(), appendComma);
  }
  case NdbDictionary::Column::Int: {
    ///< 32 bit. 4 byte signed integer, can be used in array
    return response.Append_i32(attr->int32_value(), appendComma);
  }
  case NdbDictionary::Column::Unsigned: {
    ///< 32 bit. 4 byte unsigned integer, can be used in array
    return response.Append_iu32(attr->u_32_value(), appendComma);
  }
  case NdbDictionary::Column::Bigint: {
    ///< 64 bit. 8 byte signed integer, can be used in array
    return response.Append_i64(attr->int64_value(), appendComma);
  }
  case NdbDictionary::Column::Bigunsigned: {
    ///< 64 Bit. 8 byte signed integer, can be used in array
    return response.Append_iu64(attr->u_64_value(), appendComma);
  }
  case NdbDictionary::Column::Float: {
    ///< 32-bit float. 4 bytes float, can be used in array
    return response.Append_f32(attr->float_value(), appendComma);
  }
  case NdbDictionary::Column::Double: {
    ///< 64-bit float. 8 byte float, can be used in array
    return response.Append_d64(attr->double_value(), appendComma);
  }
  case NdbDictionary::Column::Olddecimal: {
    ///< MySQL < 5.0 signed decimal,  Precision, Scale
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Olddecimalunsigned: {
    ///< MySQL < 5.0 signed decimal,  Precision, Scale
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
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
    return response.Append_string(decStr, false, appendComma);
  }
  case NdbDictionary::Column::Char:
    ///< Len. A fixed array of 1-byte chars
    [[fallthrough]];
  case NdbDictionary::Column::Varchar:
    ///< Length bytes: 1, Max: 255
    [[fallthrough]];
  case NdbDictionary::Column::Longvarchar: {
    ///< Length bytes: 2, little-endian
    int attr_bytes;
    const char *data_start = nullptr;
    if (GetByteArray(attr, &data_start, &attr_bytes) != 0) {
      return RS_CLIENT_ERROR(ERROR_019);
    } else {
      return response.Append_char(data_start, attr_bytes, attr->getColumn()->getCharset(),
                                  appendComma);
    }
  }
  case NdbDictionary::Column::Binary:
    [[fallthrough]];
  case NdbDictionary::Column::Varbinary:
    ///< Length bytes: 1, Max: 255
    [[fallthrough]];
  case NdbDictionary::Column::Longvarbinary: {
    ///< Length bytes: 2, little-endian
    int attr_bytes;
    const char *data_start = nullptr;
    if (GetByteArray(attr, &data_start, &attr_bytes) != 0) {
      return RS_CLIENT_ERROR(ERROR_019);
    } else {
      size_t encoded_str_size = boost::beast::detail::base64::encoded_size(attr_bytes);
      char buffer[encoded_str_size];
      size_t ret = boost::beast::detail::base64::encode(reinterpret_cast<void *>(buffer),
                                                        data_start, attr_bytes);
      return response.Append_string(std::string(buffer, ret), true, appendComma);
    }
  }
  case NdbDictionary::Column::Datetime: {
    ///< Precision down to 1 sec (sizeof(Datetime) == 8 bytes )
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Date: {
    ///< Precision down to 1 day(sizeof(Date) == 4 bytes )
    MYSQL_TIME l_time;
    my_unpack_date(&l_time, attr->aRef());
    char to[MAX_DATE_STRING_REP_LENGTH];
    my_date_to_str(l_time, to);
    return response.Append_string(std::string(to), true, appendComma);
  }
  case NdbDictionary::Column::Blob: {
    ///< Binary large object (see NdbBlob)
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Text: {
    ///< Text blob
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Bit: {
    //< Bit, length specifies no of bits
    int32 attr_bytes = col->getSizeInBytes();
    Uint32 words     = attr->getColumn()->getLength() / 8;
    if (attr->getColumn()->getLength() % 8 != 0) {
      words += 1;
    }

    // change endieness
    int i = 0;
    char reversed[words];
    for (int j = words - 1; j >= 0; j--) {
      reversed[i++] = attr->aRef()[j];
    }

    size_t encoded_str_size = boost::beast::detail::base64::encoded_size(words);
    char buffer[encoded_str_size];
    size_t ret =
        boost::beast::detail::base64::encode(reinterpret_cast<void *>(buffer), reversed, words);
    return response.Append_string(std::string(buffer, ret), true, appendComma);
  }
  case NdbDictionary::Column::Time: {
    ///< Time without date
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Year: {
    ///< Year 1901-2155 (1 byte)
    Int32 year = (uint)(1900 + attr->aRef()[0]);
    return response.Append_i32(year, appendComma);
  }
  case NdbDictionary::Column::Timestamp: {
    ///< Unix time
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
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
    uint precision = col->getPrecision();

    longlong numeric_time =
        my_time_packed_from_binary((const unsigned char *)attr->aRef(), precision);

    MYSQL_TIME l_time;
    TIME_from_longlong_time_packed(&l_time, numeric_time);

    char to[MAX_DATE_STRING_REP_LENGTH];
    my_TIME_to_str(l_time, to, precision);

    return response.Append_string(std::string(to), true, appendComma);
  }
  case NdbDictionary::Column::Datetime2: {
    ///< 5 bytes plus 0-3 fraction
    uint precision = col->getPrecision();

    longlong numeric_date =
        my_datetime_packed_from_binary((const unsigned char *)attr->aRef(), precision);

    MYSQL_TIME l_time;
    TIME_from_longlong_datetime_packed(&l_time, numeric_date);

    char to[MAX_DATE_STRING_REP_LENGTH];
    my_TIME_to_str(l_time, to, precision);

    return response.Append_string(std::string(to), true, appendComma);
  }
  case NdbDictionary::Column::Timestamp2: {
    ///< 4 bytes + 0-3 fraction
    uint precision = col->getPrecision();

    my_timeval my_tv{};
    my_timestamp_from_binary(&my_tv, (const unsigned char *)attr->aRef(), precision);

    Int64 epoch_in = my_tv.m_tv_sec;
    std::time_t stdtime(epoch_in);
    boost::posix_time::ptime ts = boost::posix_time::from_time_t(stdtime);

    MYSQL_TIME l_time  = {};
    l_time.year        = ts.date().year();
    l_time.month       = ts.date().month();
    l_time.day         = ts.date().day();
    l_time.hour        = ts.time_of_day().hours();
    l_time.minute      = ts.time_of_day().minutes();
    l_time.second      = ts.time_of_day().seconds();
    l_time.second_part = my_tv.m_tv_usec;
    l_time.time_type   = MYSQL_TIMESTAMP_DATETIME;

    char to[MAX_DATE_STRING_REP_LENGTH];
    my_TIME_to_str(l_time, to, precision);

    return response.Append_string(std::string(to), true, appendComma);
  }
  }

  return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                         " Type: " + std::to_string(col->getType()));
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
        if (operation->equal(request.PKName(colIdx), static_cast<char>(num)) != 0) {
          return RS_SERVER_ERROR(ERROR_023);
        }
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
        if (operation->equal(request.PKName(colIdx), static_cast<char>(num)) != 0) {
          return RS_SERVER_ERROR(ERROR_023);
        }
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
        if (operation->equal(request.PKName(colIdx), (Int16)num) != 0) {
          return RS_SERVER_ERROR(ERROR_023);
        }
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
        if (operation->equal(request.PKName(colIdx), (Uint16)num) != 0) {
          return RS_SERVER_ERROR(ERROR_023);
        }
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
        if (operation->equal(request.PKName(colIdx), static_cast<int>(num)) != 0) {
          return RS_SERVER_ERROR(ERROR_023);
        }
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
        if (operation->equal(request.PKName(colIdx), (unsigned int)num)) {
          return RS_SERVER_ERROR(ERROR_023);
        }
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
      if (operation->equal(request.PKName(colIdx), num) != 0) {
        return RS_SERVER_ERROR(ERROR_023);
      }
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
        if (operation->equal(request.PKName(colIdx), result) != 0) {
          return RS_SERVER_ERROR(ERROR_023);
        }
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
      if (operation->equal(request.PKName(colIdx), num) != 0) {
        return RS_SERVER_ERROR(ERROR_023);
      }
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
        if (operation->equal(request.PKName(colIdx), num) != 0) {
          return RS_SERVER_ERROR(ERROR_023);
        }
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
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Olddecimalunsigned: {
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
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

    if (operation->equal(request.PKName(colIdx), decBin, bytesNeeded) != 0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Char: {
    ///< Len. A fixed array of 1-byte chars

    const int len = request.PKValueLen(colIdx);
    if (len > col->getLength()) {
      // the user is searching a key greater than all the possible keys so return 404
      // additionally using a pk greater in size than the table definition
      // causes seg fault https://github.com/logicalclocks/rondb/issues/122
      return RS_CLIENT_404_ERROR();
    }

    const char *charStr = request.PKValueCStr(colIdx);
    char pk[col->getLength()];
    for (int i = 0; i < col->getLength(); i++) {
      pk[i] = 0;
    }
    memcpy(pk, charStr, len);

    if (operation->equal(request.PKName(colIdx), pk, col->getLength()) != 0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Varchar:
    ///< Length bytes: 1, Max: 255
    [[fallthrough]];
  case NdbDictionary::Column::Longvarchar: {
    ///< Length bytes: 2, little-endian
    const int len = request.PKValueLen(colIdx);
    if (len > col->getLength()) {
      // the user is searching a key greater than all the possible keys so return 404
      // additionally using a pk greater in size than the table definition
      // causes seg fault https://github.com/logicalclocks/rondb/issues/122
      return RS_CLIENT_404_ERROR();
    }
    char *charStr;
    if (request.PKValueNDBStr(colIdx, col, &charStr) != 0) {
      return RS_SERVER_ERROR(ERROR_019);
    }
    if (operation->equal(request.PKName(colIdx), charStr, len) != 0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Binary: {
    ///< Len
    // we get the data in base64
    const char *encodedStr = request.PKValueCStr(colIdx);
    size_t decoded_size    = boost::beast::detail::base64::decoded_size(request.PKValueLen(colIdx));
    int maxlen             = std::max(col->getLength(), static_cast<int>(decoded_size));

    char pk[maxlen];
    for (int i = 0; i < col->getLength(); i++) {
      pk[i] = 0;
    }

    std::pair<std::size_t, std::size_t> ret =
        boost::beast::detail::base64::decode(pk, encodedStr, request.PKValueLen(colIdx));

    if (static_cast<int>(ret.first) > col->getLength()) {
      // the user is searching a key greater than all the possible keys so return 404
      // additionally using a pk greater in size than the table definition
      // causes seg fault https://github.com/logicalclocks/rondb/issues/122
      return RS_CLIENT_404_ERROR();
    }

    if (operation->equal(request.PKName(colIdx), pk, col->getLength()) != 0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Varbinary:
    ///< Length bytes: 1, Max: 255
    [[fallthrough]];
  case NdbDictionary::Column::Longvarbinary: {
    ///< Length bytes: 2, little-endian

    const char *encodedStr = request.PKValueCStr(colIdx);
    size_t decoded_size    = boost::beast::detail::base64::decoded_size(request.PKValueLen(colIdx));
    int additional_len     = 1;
    if (col->getType() == NdbDictionary::Column::Longvarbinary) {
      additional_len = 2;
    }

    int maxlen = std::max(col->getLength(), static_cast<int>(decoded_size) + additional_len);
    char pk[maxlen];
    for (int i = 0; i < maxlen; i++) {
      pk[i] = 0;
    }

    std::pair<std::size_t, std::size_t> ret = boost::beast::detail::base64::decode(
        pk + additional_len, encodedStr, request.PKValueLen(colIdx));

    if (static_cast<int>(ret.first) > col->getLength()) {
      // the user is searching a key greater than all the possible keys so return 404
      // additionally using a pk greater in size than the table definition
      // causes seg fault https://github.com/logicalclocks/rondb/issues/122
      return RS_CLIENT_404_ERROR();
    }

    // insert the length at the begenning of the array
    if (col->getType() == NdbDictionary::Column::Varbinary) {
      pk[0] = (Uint8)ret.first;
    } else if (col->getType() == NdbDictionary::Column::Longvarbinary) {
      pk[0] = (Uint8)(ret.first % 256);
      pk[1] = (Uint8)(ret.first / 256);
    } else {
      return RS_SERVER_ERROR(ERROR_015);
    }

    if (operation->equal(request.PKName(colIdx), pk, ret.first + additional_len) != 0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Datetime: {
    ///< Precision down to 1 sec (sizeof(Datetime) == 8 bytes )
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Date: {
    ///< Precision down to 1 day(sizeof(Date) == 4 bytes )
    const char *date_str = request.PKValueCStr(colIdx);
    size_t date_str_len  = request.PKValueLen(colIdx);

    MYSQL_TIME l_time;
    MYSQL_TIME_STATUS status;
    bool ret = str_to_datetime(date_str, date_str_len, &l_time, 0, &status);
    if (ret != 0) {
      return RS_CLIENT_ERROR(std::string(ERROR_027) + std::string(" Column: ") +
                             std::string(col->getName()))
    }

    if (l_time.hour != 0 || l_time.minute != 0 || l_time.second != 0 || l_time.second_part != 0) {
      // the user has also specified time. as we only store YYMMDD therefore return 404 error
      return RS_CLIENT_404_ERROR();
    }

    unsigned char packed[col->getSizeInBytes()];
    my_date_to_binary(&l_time, packed);

    if (operation->equal(request.PKName(colIdx), reinterpret_cast<char *>(packed),
                         col->getSizeInBytes()) != 0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Blob: {
    ///< Binary large object (see NdbBlob)
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Text: {
    ///< Text blob
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Bit: {
    ///< Bit, length specifies no of bits
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Time: {
    ///< Time without date
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
  }
  case NdbDictionary::Column::Year: {
    ///< Year 1901-2155 (1 byte)
    bool success = false;
    try {
      Int32 year = std::stoi(request.PKValueCStr(colIdx));
      if (year >= 1901 && year <= 2155) {
        Uint8 year_char = (year - 1900);
        if (operation->equal(request.PKName(colIdx), year_char) != 0) {
          return RS_SERVER_ERROR(ERROR_023);
        }
        success = true;
      }
    } catch (...) {
    }
    if (!success) {
      return RS_CLIENT_ERROR(
          ERROR_015 + std::string(" Expecting YEAR column. Possible values [1901-2155]. Column: ") +
          std::string(request.PKName(colIdx)));
    } else {
      return RS_OK;
    }
  }
  case NdbDictionary::Column::Timestamp: {
    ///< Unix time
    return RS_SERVER_ERROR(ERROR_028 + std::string(" Column: ") + std::string(col->getName()) +
                           " Type: " + std::to_string(col->getType()));
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
    const char *time_str = request.PKValueCStr(colIdx);
    size_t time_str_len  = request.PKValueLen(colIdx);

    MYSQL_TIME l_time;
    MYSQL_TIME_STATUS status;
    bool ret = str_to_time(time_str, time_str_len, &l_time, &status, 0);
    if (ret != 0) {
      return RS_CLIENT_ERROR(std::string(ERROR_027) + std::string(" Column: ") +
                             std::string(col->getName()))
    }

    size_t packed_len = col->getSizeInBytes();
    int precision     = col->getPrecision();
    unsigned char packed[packed_len];

    longlong numaric_date_time = TIME_to_longlong_time_packed(l_time);
    my_time_packed_to_binary(numaric_date_time, packed, precision);

    if (operation->equal(request.PKName(colIdx), reinterpret_cast<char *>(packed), packed_len) !=
        0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Datetime2: {
    ///< 5 bytes plus 0-3 fraction
    const char *date_str = request.PKValueCStr(colIdx);
    size_t date_str_len  = request.PKValueLen(colIdx);

    MYSQL_TIME l_time;
    MYSQL_TIME_STATUS status;
    bool ret = str_to_datetime(date_str, date_str_len, &l_time, 0, &status);
    if (ret != 0) {
      return RS_CLIENT_ERROR(std::string(ERROR_027) + std::string(" Column: ") +
                             std::string(col->getName()))
    }

    size_t packed_len = col->getSizeInBytes();
    int precision     = col->getPrecision();
    unsigned char packed[packed_len];

    longlong numaric_date_time = TIME_to_longlong_datetime_packed(l_time);

    my_datetime_packed_to_binary(numaric_date_time, packed, precision);

    if (operation->equal(request.PKName(colIdx), reinterpret_cast<char *>(packed), packed_len) !=
        0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  case NdbDictionary::Column::Timestamp2: {
    // epoch range 0 , 2147483647
    /// < 4 bytes + 0-3 fraction
    const char *ts_str = request.PKValueCStr(colIdx);
    size_t ts_str_len  = request.PKValueLen(colIdx);
    size_t packed_len  = col->getSizeInBytes();
    unsigned char packed[packed_len];
    uint precision = col->getPrecision();

    MYSQL_TIME l_time;
    MYSQL_TIME_STATUS status;
    bool ret = str_to_datetime(ts_str, ts_str_len, &l_time, 0, &status);
    if (ret != 0) {
      return RS_CLIENT_ERROR(std::string(ERROR_027) + std::string(" Column: ") +
                             std::string(col->getName()))
    }

    time_t epoch = 0;
    try {
      char bts_str[MAX_DATE_STRING_REP_LENGTH];
      snprintf(bts_str, MAX_DATE_STRING_REP_LENGTH, "%d-%d-%d %d:%d:%d", l_time.year, l_time.month,
               l_time.day, l_time.hour, l_time.minute, l_time.second);
      boost::posix_time::ptime bt(boost::posix_time::time_from_string(std::string(bts_str)));
      boost::posix_time::ptime start(boost::gregorian::date(1970, 1, 1));
      boost::posix_time::time_duration dur = bt - start;
      epoch                                = dur.total_seconds();
    } catch (...) {
      return RS_CLIENT_ERROR(std::string(ERROR_027) + std::string(" Column: ") +
                             std::string(col->getName()))
    }

    // 1970-01-01 00:00:01' UTC to '2038-01-19 03:14:07' UTC.
    if (epoch <= 0 || epoch > 2147483647) {
      return RS_CLIENT_ERROR(std::string(ERROR_027) + std::string(" Column: ") +
                             std::string(col->getName()))
    }

    std::cout << "Boost secs : " << epoch << std::endl;
    // TODO(salman) 1 apply timezone changes
    // https://dev.mysql.com/doc/refman/8.0/en/datetime.html
    // iMySQL converts TIMESTAMP values from the current time zone to UTC for storage, and back from
    // UTC to the current time zone for retrieval. (This does not occur for other types such as
    // DATETIME.) By default, the current time zone for each connection is the server's time. The
    // time zone can be set on a per-connection basis. As long as the time zone setting remains
    // constant, you get back the same value you store. If you store a TIMESTAMP value, and then
    // change the time zone and retrieve the value, the retrieved value is different from the value
    // you stored. This occurs because the same time zone was not used for conversion in both
    // directions. The current time zone is available as the value of the time_zone system variable.
    // For more information, see Section 5.1.15, “MySQL Server Time Zone Support”.
    // TODO(salman) 2 Investigate how clusterj inserts time stamps. Does it apply time zone changes
    // TODO(salman) how to deal with time zone setting in mysql server
    //

    my_timeval my_tv{epoch, (Int64)l_time.second_part};
    my_timestamp_to_binary(&my_tv, packed, precision);

    if (operation->equal(request.PKName(colIdx), reinterpret_cast<char *>(packed), packed_len) !=
        0) {
      return RS_SERVER_ERROR(ERROR_023);
    }
    return RS_OK;
  }
  }
  return RS_OK;
}
