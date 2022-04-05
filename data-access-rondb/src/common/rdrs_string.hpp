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

#ifndef RDRS_STRING_H
#define RDRS_STRING_H

#include <NdbApi.hpp>
#include <cstring>
#include <stdint.h>
#include <string>

using namespace std;
// function defined in RonDB lib
size_t convert_to_printable(char *to, size_t to_len, const char *from, size_t from_len,
                            const CHARSET_INFO *from_cs, size_t nbytes = 0);

size_t well_formed_copy_nchars(const CHARSET_INFO *to_cs, char *to,
                               size_t to_length, const CHARSET_INFO *from_cs,
                               const char *from, size_t from_length,
                               size_t nchars,
                               const char **well_formed_error_pos,
                               const char **cannot_convert_error_pos,
                               const char **from_end_pos);

/*!
    @brief calculates the extra space to escape a JSON string
    @param[in] s  the string to escape
    @return the number of characters required to escape string @a s
    @complexity Linear in the length of string @a s.
    */
std::size_t extra_space(const string &s) noexcept {
  std::size_t result = 0;

  for (const auto &c : s) {
    switch (c) {
    case '"':
    case '\\':
    case '\b':
    case '\f':
    case '\n':
    case '\r':
    case '\t': {
      // from c (1 byte) to \x (2 bytes)
      result += 1;
      break;
    }

    default: {
      if (c >= 0x00 and c <= 0x1f) {
        // from c (1 byte) to \uxxxx (6 bytes)
        result += 5;
      }
      break;
    }
    }
  }

  return result;
}

//https://github.com/nlohmann/json/blob/ec7a1d834773f9fee90d8ae908a0c9933c5646fc/src/json.hpp#L4604-L4697
/*!
    @brief escape a string
    Escape a string by replacing certain special characters by a sequence of an
    escape character (backslash) and another character and other control
    characters by a sequence of "\u" followed by a four-digit hex
    representation.
    @param[in] s  the string to escape
    @return  the escaped string
    @complexity Linear in the length of string @a s.
    */
string escape_string(const string &s) noexcept {
  const auto space = extra_space(s);
  if (space == 0) {
    return s;
  }

  // create a result string of necessary size
  string result(s.size() + space, '\\');
  std::size_t pos = 0;

  for (const auto &c : s) {
    switch (c) {
    // quotation mark (0x22)
    case '"': {
      result[pos + 1] = '"';
      pos += 2;
      break;
    }

    // reverse solidus (0x5c)
    case '\\': {
      // nothing to change
      pos += 2;
      break;
    }

    // backspace (0x08)
    case '\b': {
      result[pos + 1] = 'b';
      pos += 2;
      break;
    }

    // formfeed (0x0c)
    case '\f': {
      result[pos + 1] = 'f';
      pos += 2;
      break;
    }

    // newline (0x0a)
    case '\n': {
      result[pos + 1] = 'n';
      pos += 2;
      break;
    }

    // carriage return (0x0d)
    case '\r': {
      result[pos + 1] = 'r';
      pos += 2;
      break;
    }

    // horizontal tab (0x09)
    case '\t': {
      result[pos + 1] = 't';
      pos += 2;
      break;
    }

    default: {
      if (c >= 0x00 and c <= 0x1f) {
        // print character c as \uxxxx
        sprintf(&result[pos + 1], "u%04x", int(c));
        pos += 6;
        // overwrite trailing null character
        result[pos] = '\\';
      } else {
        // all other characters are added as-is
        result[pos++] = c;
      }
      break;
    }
    }
  }

  return result;
}


#endif
