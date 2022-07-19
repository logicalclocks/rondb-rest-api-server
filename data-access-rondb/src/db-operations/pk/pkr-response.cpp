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

#include "src/db-operations/pk/pkr-response.hpp"

#include <mysql.h>
#include <iostream>
#include <sstream>
#include <cassert>
#include "src/rondb-lib/rdrs_string.hpp"
#include "src/mystring.hpp"
#include "src/rdrs-const.h"

PKRResponse::PKRResponse(const RS_Buffer *respBuff) {
  this->resp        = respBuff;
  this->writeHeader = PK_RESP_HEADER_END;
  this->WriteHeaderField(PK_RESP_OP_TYPE_IDX, RDRS_PK_RESP_ID);
  this->WriteHeaderField(PK_RESP_CAPACITY_IDX, resp->size);
}

RS_Status PKRResponse::WriteHeaderField(Uint32 index, Uint32 value) {
  Uint32 *b = reinterpret_cast<Uint32 *>(this->resp->buffer);
  b[index]  = value;
  return RS_OK;
}

RS_Status PKRResponse::SetStatus(Uint32 value) {
  this->WriteHeaderField(PK_RESP_OP_STATUS_IDX, value);
  return RS_OK;
}

RS_Status PKRResponse::Close() {
  this->WriteHeaderField(PK_RESP_LENGTH_IDX, writeHeader);
  return RS_OK;
}

RS_Status PKRResponse::SetDB(const char *db) {
  Uint32 dbAddr    = this->writeHeader;
  RS_Status status = this->Append_cstring(db);
  if (status.http_code != SUCCESS) {
    return status;
  }
  this->WriteHeaderField(PK_RESP_DB_IDX, dbAddr);
  return RS_OK;
}

RS_Status PKRResponse::SetTable(const char *table) {
  Uint32 tableAddr = this->writeHeader;
  RS_Status status = this->Append_cstring(table);
  if (status.http_code != SUCCESS) {
    return status;
  }
  this->WriteHeaderField(PK_RESP_TABLE_IDX, tableAddr);
  return RS_OK;
}

RS_Status PKRResponse::SetOperationID(const char *opID) {
  Uint32 opIDAddr    = this->writeHeader;
  RS_Status status = this->Append_cstring(opID);
  if (status.http_code != SUCCESS) {
    return status;
  }
  this->WriteHeaderField(PK_RESP_OP_ID_IDX, opIDAddr);
  return RS_OK;
}

bool PKRResponse::HasCapacity(char *str) {
  Uint32 strl = strlen(str) + 1;  // +1 for null terminator
  if (strl > GetRemainingCapacity()) {
    return false;
  }
  return true;
}

RS_Status PKRResponse::Append_cstring(const char *str) {
  Uint32 strl = strlen(str) + 1;  // for null terminator
  if (strl > GetRemainingCapacity()) {
    return RS_SERVER_ERROR(ERROR_016);
  }

  std::memcpy(resp->buffer + writeHeader, str, strl);
  writeHeader += strl;
  return RS_OK;
}

RS_Status PKRResponse::SetNoOfColumns(Uint32 cols) {
  // 3 because +1 for Col Name Address
  // +1 for Value Address
  // +1 for isNULL
  Uint32 spaceNeeded = cols * ADDRESS_SIZE * 3;
  if (spaceNeeded > GetRemainingCapacity()) {
    return RS_SERVER_ERROR(ERROR_016);
  }

  this->writeHeader = (this->writeHeader + spaceNeeded);
  this->colsToWrite = cols;
  return RS_OK;
}

RS_Status PKRResponse::SetColumnDataNull(const char *colName) {
  return SetColumnDataInt(colName, nullptr, RDRS_UNKNOWN_DATATYPE);
}

RS_Status PKRResponse::SetColumnData(const char *colName, const char *value, Uint32 type) {
  return this->SetColumnDataInt(colName, value, type);
}

RS_Status PKRResponse::SetColumnDataInt(const char *colName, const char *value, Uint32 type) {
  // first index is for column name
  // second index is for column value
  // thrid index is for isNULL
  // forth index is for data type, e.g., string or non-string data
  Uint32 *b           = reinterpret_cast<Uint32 *>(this->resp->buffer);
  Uint32 indexWritten = (PK_RESP_HEADER_END / ADDRESS_SIZE);
  indexWritten += (colsWritten * 4);

  Uint32 nameAddress = this->writeHeader;
  RS_Status status   = Append_cstring(colName);
  if (status.http_code != SUCCESS) {
    return status;
  }
  b[indexWritten + 0] = nameAddress;

  if (value == nullptr) {
    b[indexWritten + 1] = 0;                      // value address not set
    b[indexWritten + 2] = 1;                      // isNULL
    b[indexWritten + 3] = RDRS_UNKNOWN_DATATYPE;  // data type
  } else {
    Uint32 valueAddress = this->writeHeader;
    RS_Status status    = Append_cstring(value);
    if (status.http_code != SUCCESS) {
      return status;
    }
    b[indexWritten + 1] = valueAddress;  // value address
    b[indexWritten + 2] = 0;             // isNULL
    b[indexWritten + 3] = type;          // data type
  }

  colsWritten++;
  return RS_OK;
}

char *PKRResponse::GetResponseBuffer() {
  return resp->buffer;
}

Uint32 PKRResponse::GetMaxCapacity() {
  return this->resp->size;
}

Uint32 PKRResponse::GetRemainingCapacity() {
  return GetMaxCapacity() - GetWriteHeader();
}

Uint32 PKRResponse::GetWriteHeader() {
  return this->writeHeader;
}

void *PKRResponse::GetWritePointer() {
  return resp->buffer + writeHeader;
}

RS_Status PKRResponse::Append_string(const char *colName, std::string value, Uint32 type) {
  if ((value.length() + 1) > GetRemainingCapacity()) {  // +1 null terminator
    return RS_SERVER_ERROR(ERROR_016);
  }

  return SetColumnData(colName, value.c_str(), type);
}

RS_Status PKRResponse::Append_i8(const char *colName, char num) {
  return Append_i64(colName, num);
}

RS_Status PKRResponse::Append_iu8(const char *colName, unsigned char num) {
  return Append_iu64(colName, num);
}

RS_Status PKRResponse::Append_i16(const char *colName, Int16 num) {
  return Append_i64(colName, num);
}

RS_Status PKRResponse::Append_iu16(const char *colName, Uint16 num) {
  return Append_iu64(colName, num);
}

RS_Status PKRResponse::Append_i24(const char *colName, int num) {
  return Append_i64(colName, num);
}

RS_Status PKRResponse::Append_iu24(const char *colName, Uint32 num) {
  return Append_iu64(colName, num);
}

RS_Status PKRResponse::Append_iu32(const char *colName, Uint32 num) {
  return Append_iu64(colName, num);
}

RS_Status PKRResponse::Append_i32(const char *colName, Int32 num) {
  return Append_i64(colName, num);
}

RS_Status PKRResponse::Append_f32(const char *colName, float num) {
  return Append_d64(colName, num);
}

RS_Status PKRResponse::Append_d64(const char *colName, double num) {
  try {
    std::stringstream ss;
    ss << num;
    return this->SetColumnData(colName, ss.str().c_str(), RDRS_FLOAT_DATATYPE);
  } catch (...) {
    return RS_SERVER_ERROR(ERROR_015);
  }
}

RS_Status PKRResponse::Append_iu64(const char *colName, Uint64 num) {
  try {
    std::string numStr = std::to_string(num);
    return this->SetColumnData(colName, numStr.c_str(), RDRS_INTEGER_DATATYPE);
  } catch (...) {
    return RS_SERVER_ERROR(ERROR_015);
  }
}

RS_Status PKRResponse::Append_i64(const char *colName, Int64 num) {
  try {
    std::string numStr = std::to_string(num);
    return this->SetColumnData(colName, numStr.c_str(), RDRS_INTEGER_DATATYPE);
  } catch (...) {
    return RS_SERVER_ERROR(ERROR_015);
  }
}

RS_Status PKRResponse::Append_char(const char *colName, const char *fromBuff, Uint32 fromBuffLen,
                                   CHARSET_INFO *fromCS) {

  Uint32 extraSpace     = 1;  // +1 for null terminator
  Uint32 estimatedBytes = fromBuffLen + extraSpace;

  if (estimatedBytes > GetRemainingCapacity()) {
    return RS_SERVER_ERROR(ERROR_010 + std::string(" Response buffer remaining capacity: ") +
                           std::to_string(GetRemainingCapacity()) + std::string(" Required: ") +
                           std::to_string(estimatedBytes));
  }

  // from_buffer -> printable string  -> escaped string
  char tempBuff[estimatedBytes];
  const char *well_formed_error_pos;
  const char *cannot_convert_error_pos;
  const char *from_end_pos;
  const char *error_pos;

  /* convert_to_printable(tempBuff, tempBuffLen, fromBuffer, fromLength, fromCS, 0); */
  int bytesFormed = well_formed_copy_nchars(fromCS, tempBuff, estimatedBytes, fromCS, fromBuff,
                                            fromBuffLen, UINT32_MAX, &well_formed_error_pos,
                                            &cannot_convert_error_pos, &from_end_pos);

  error_pos = well_formed_error_pos ? well_formed_error_pos : cannot_convert_error_pos;
  if (error_pos) {
    char printable_buff[32];
    convert_to_printable(printable_buff, sizeof(printable_buff), error_pos,
                         fromBuff + fromBuffLen - error_pos, fromCS, 6);
    return RS_SERVER_ERROR(ERROR_008 + std::string(" Invalid string: ") +
                           std::string(printable_buff));
  } else if (from_end_pos < fromBuff + fromBuffLen) {
    /*
      result is longer than UINT_MAX32 and doesn't fit into String
    */
    return RS_SERVER_ERROR(ERROR_021 + std::string(" Buffer size: ") +
                           std::to_string(estimatedBytes) + std::string(". Bytes left to copy: ") +
                           std::to_string((fromBuff + fromBuffLen) - from_end_pos));
  }
  std::string wellFormedString = std::string(tempBuff, bytesFormed);
  // remove blank spaces that are padded to the string
  size_t endpos = wellFormedString.find_last_not_of(" ");
  if (std::string::npos != endpos) {
    wellFormedString = wellFormedString.substr(0, endpos + 1);
  }

  std::string escapedstr = escape_string(wellFormedString);
  if ((escapedstr.length() + extraSpace) >= GetRemainingCapacity()) {  // +2 for quotation marks
    return RS_SERVER_ERROR(ERROR_010);
  }
  return this->SetColumnData(colName, escapedstr.c_str(), RDRS_STRING_DATATYPE);
}
