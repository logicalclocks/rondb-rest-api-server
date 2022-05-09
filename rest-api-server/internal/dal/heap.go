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
#include "./../../../data-access-rondb/src/rdrs-const.h"
#include "./../../../data-access-rondb/src/rdrs-dal.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"hopsworks.ai/rdrs/internal/config"
)

type NativeBuffer struct {
	Size   uint32
	Buffer unsafe.Pointer
}

type NativeBufferStats struct {
	AllocationsCount   uint64
	DeallocationsCount uint64
	BuffersCount       uint64
	FreeBuffers        uint64
}

var buffers []*NativeBuffer
var buffersStats NativeBufferStats
var initialized bool
var mutex sync.Mutex

func InitializeBuffers() {
	mutex.Lock()
	defer mutex.Unlock()

	if initialized {
		panic(fmt.Sprintf("Native buffers are already initialized"))
	}

	if C.ADDRESS_SIZE != 4 {
		panic(fmt.Sprintf("Only 4 byte address are supported"))
	}

	if config.BufferSize()%C.ADDRESS_SIZE != 0 {
		panic(fmt.Sprintf("Buffer size must be multiple of %d", C.ADDRESS_SIZE))
	}

	for i := uint32(0); i < config.PreAllocBuffers(); i++ {
		buffers = append(buffers, __allocateBuffer())
	}

	buffersStats.AllocationsCount = uint64(config.PreAllocBuffers())
	buffersStats.BuffersCount = buffersStats.AllocationsCount
	buffersStats.DeallocationsCount = 0

	initialized = true
}

func __allocateBuffer() *NativeBuffer {
	buff := NativeBuffer{Buffer: C.malloc(C.size_t(config.BufferSize())), Size: config.BufferSize()}
	dstBuf := unsafe.Slice((*byte)(buff.Buffer), config.BufferSize())
	dstBuf[0] = 0x00 // reset buffer by putting null terminator in the begenning
	return &buff
}

func GetBuffer() *NativeBuffer {
	if !initialized {
		panic(fmt.Sprintf("Native buffers are not initialized"))
	}

	mutex.Lock()
	defer mutex.Unlock()

	var buff *NativeBuffer
	if len(buffers) > 0 {
		buff = buffers[len(buffers)-1]
		buffers = buffers[:len(buffers)-1]
	} else {
		buff = __allocateBuffer()
		buffersStats.BuffersCount++
		buffersStats.AllocationsCount++
	}

	return buff
}

func ReturnBuffer(buffer *NativeBuffer) {
	if !initialized {
		panic(fmt.Sprintf("Native buffers are not initialized"))
	}
	mutex.Lock()
	defer mutex.Unlock()

	buffers = append(buffers, buffer)
}

func GetNativeBuffersStats() NativeBufferStats {
	if !initialized {
		panic(fmt.Sprintf("Native buffers are not initialized"))
	}
	//update the free buffers cound
	mutex.Lock()
	defer mutex.Unlock()
	buffersStats.FreeBuffers = uint64(len(buffers))
	return buffersStats
}

func BuffersInitialized() bool {
	return initialized
}
