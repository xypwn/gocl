package cl

/*
#cgo CFLAGS: -DCL_TARGET_OPENCL_VERSION=310
#cgo CFLAGS: -DCL_NO_NON_ICD_DISPATCH_EXTENSION_PROTOTYPES
#cgo CFLAGS: -DCL_ENABLE_LAYERS
#cgo CFLAGS: -Wno-deprecated-declarations
#cgo CFLAGS: -I../cl-3.1/inc/
#cgo windows LDFLAGS: -lcfgmgr32 -lruntimeobject -lOle32
#include "../cl-3.1/inc/CL/cl.h"
*/
import "C"

import (
	"reflect"
	"runtime"
	"unsafe"

	. "github.com/xypwn/gocl/cl-3.1"
)

// BEGIN //

//bind:clGetKernelSuggestedLocalWorkSize
func GetKernelSuggestedLocalWorkSize(command_queue CommandQueue, kernel Kernel, work_dim uint32, global_work_offset []uint64, global_work_size []uint64) (suggested_local_work_size []uint64, err error) {
	var global_work_offset_1 unsafe.Pointer
	if len(global_work_offset) != 0 {
		if len(global_work_offset) != int(work_dim) {
			panic("expected len(global_work_offset) to be 0 or equal to work_dim")
		}
		var fin func()
		global_work_offset_1, _, fin = sliceToC(global_work_offset)
		defer fin()
	}
	if len(global_work_size) != int(work_dim) {
		panic("expected len(global_work_size) to be equal to work_dim")
	}
	global_work_size_1, _, fin_1 := sliceToC(global_work_size)
	defer fin_1()
	suggested_local_work_size = make([]uint64, work_dim)
	suggested_local_work_size_1, _, fin_2 := sliceToC(suggested_local_work_size)
	defer fin_2()
	res := C.clGetKernelSuggestedLocalWorkSize(
		C.cl_command_queue(command_queue),
		C.cl_kernel(kernel),
		C.cl_uint(work_dim),
		(*C.size_t)(global_work_offset_1),
		(*C.size_t)(global_work_size_1),
		(*C.size_t)(suggested_local_work_size_1),
	)
	return suggested_local_work_size, makeError(ErrorCode(res))
}

//export go_cl_callback_clCompileProgram
//bind:clCompileProgram
func go_cl_callback_clCompileProgram(program C.cl_program, user_data *C.void) {
	program_1 := Program(program)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(program Program)))(program_1)
}

//bindc:clCompileProgram
extern void go_cl_callback_clCompileProgram(cl_program, void*);

//bind:clCompileProgram
func CompileProgram(program Program, device_list []DeviceId, options string, input_headers []Program, header_include_names []string, pfn_notify func(program Program)) (_err error) {
	if len(input_headers) != len(header_include_names) {
		panic("len(input_headers) must match len(header_include_names)")
	}

	options_1, options_1_fin := stringToC(options)
	defer options_1_fin()
	device_list_1, num_devices_1, device_list_fin := sliceToC(device_list)
	defer device_list_fin()

	input_headers_1, num_input_headers, input_headers_fin := sliceToC(input_headers)
	defer input_headers_fin()

	header_include_names_1, _, header_include_names_fin := stringsToC(header_include_names, false)
	defer header_include_names_fin()

	program_1 := C.cl_program(program)
	num_devices_2 := C.cl_uint(num_devices_1)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clCompileProgram)
	}
	res := C.clCompileProgram(
		program_1,
		num_devices_2,
		(*C.cl_device_id)(device_list_1),
		options_1,
		C.cl_uint(num_input_headers),
		(*C.cl_program)(input_headers_1),
		header_include_names_1,
		callback,
		callback_uid,
	)
	return makeError(ErrorCode(res))
}

// Sets a kernel arg using a value.
//
//bind:extra
func SetKernelArgValue[T any](kernel Kernel, arg_index uint32, value T) (_err error) {
	var pin runtime.Pinner
	defer pin.Unpin()
	pin.Pin(&value)
	return SetKernelArg(kernel, arg_index, uint64(unsafe.Sizeof(value)), unsafe.Pointer(&value))
}

// Set multiple kernel args. arg_offset specifies the arg index of the first value.
//
// Slower than [SetKernelArgValue] due to its use of reflection and the need
// to make (temporary) allocations.
//
// Returns the first error it encounters (if any).
//
//bind:extra
func SetKernelArgValues(kernel Kernel, arg_offset uint32, values ...any) (_err error) {
	var pin runtime.Pinner
	defer pin.Unpin()
	for i := range values {
		idx := arg_offset + uint32(i)
		value := reflect.ValueOf(values[i])
		nv := reflect.New(value.Type())
		nv.Elem().Set(value)
		ptr, size := nv.UnsafePointer(), nv.Elem().Type().Size()
		pin.Pin(ptr)
		if err := SetKernelArg(kernel, idx, uint64(size), ptr); err != nil {
			return err
		}
	}
	return nil
}

// The [CreateBufferSlice] of [CreateBufferWithProperties].
//
//bind:extra
func CreateBufferSliceWithProperties[E any, S ~[]E](context Context, properties []MemProperties, flags MemFlags, items S) (_res Mem, _errcode_ret error) {
	if len(items) == 0 {
		panic("items must be non-empty")
	}
	if flags & MEM_ALLOC_HOST_PTR != 0 {
		panic("MEM_ALLOC_HOST_PTR not allowed")
	}
	size := uint64(len(items))*uint64(unsafe.Sizeof(items[0]))
	ptr := unsafe.Pointer(&items[0])
	if flags&MEM_COPY_HOST_PTR == 0 && flags&MEM_USE_HOST_PTR == 0 {
		// Prevent CL_INVALID_HOST_PTR
		ptr = nil
	} else {
		var pin runtime.Pinner
		defer pin.Unpin()
		pin.Pin(ptr)
	}
	if len(properties) == 0 {
		// BUG: CreateBufferWithProperties seems to cause a segfault on some system, so I'll just
		// use regular CreateBuffer if possible until I have figured this out.
		return CreateBuffer(context, flags, size, ptr)
	} else {
		return CreateBufferWithProperties(context, properties, flags, size, ptr)
	}
}

// Like [CreateBuffer], but accepts a slice and automatically determines the memory region size.
//
// Does NOT properly handle Go memory pinning in the case of async operations or MEM_USE_PTR. You
// either have to handle pinning yourself, or use [BackedBuffer].
//
// items must be non-empty.
//
// Unlike [CreateBuffer], [CreateBufferSlice] accepts non-nil items even if neither MEM_COPY_HOST_PTR nor MEM_USE_HOST_PTR is set.
// However, note that the data in items will only be copied if MEM_COPY_HOST_PTR or MEM_USE_HOST_PTR is set.
//
// To allocate an empty buffer from a data type and item count, see [CreateBufferEmpty].
//
//bind:extra
func CreateBufferSlice[E any, S ~[]E](context Context, flags MemFlags, items S) (_res Mem, _errcode_ret error) {
	return CreateBufferSliceWithProperties(context, nil, flags, items)
}

// The [CreateBufferEmpty] of [CreateBufferWithProperties].
//
//bind:extra
func CreateBufferEmptyWithProperties[E any](context Context, properties []MemProperties, flags MemFlags, num_items int) (_res Mem, _errcode_ret error) {
	if flags&MEM_COPY_HOST_PTR != 0 || flags&MEM_USE_HOST_PTR != 0 {
		panic("CreateBufferEmpty forbids flags MEM_COPY_HOST_PTR and MEM_USE_HOST_PTR (use CreateBufferSlice to create a buffer with initial contents)")
	}
	var zero E
	itemSize := uint64(unsafe.Sizeof(zero))
	if len(properties) == 0 {
		// BUG: CreateBufferWithProperties seems to cause a segfault on some system, so I'll just
		// use regular CreateBuffer if possible until I have figured this out.
		return CreateBuffer(context, flags, uint64(num_items)*itemSize, nil)
	} else {
		return CreateBufferWithProperties(context, properties, flags, uint64(num_items)*itemSize, nil)
	}
}

// Like [CreateBuffer], but determines memory requirements from data type and item count.
//
// Flags MUST NOT CONTAIN MEM_COPY_HOST_PTR or MEM_USE_HOST_PTR (will panic otherwise).
// 
// To allocate a buffer from a slice, see [CreateBufferSlice].
//
//bind:extra
func CreateBufferEmpty[E any](context Context, flags MemFlags, num_items int) (_res Mem, _errcode_ret error) {
	return CreateBufferEmptyWithProperties[E](context, nil, flags, num_items)
}

// Like [EnqueueReadBuffer], but accepts a slice and automatically determines the memory region size.
//
// offset is now the number of ITEMS, not bytes (unlike [EnqueueReadBuffer]).
//
//bind:extra
func EnqueueReadBufferSlice[E any](command_queue CommandQueue, buffer Mem, blocking_read bool, offset int, items []E, event_wait_list []Event, event *Event) (_err error) {
	if len(items) == 0 {
		return makeError(INVALID_VALUE)
	}
	var pin runtime.Pinner
	defer pin.Unpin()
	itemSize := uint64(unsafe.Sizeof(items[0]))
	pin.Pin(&items[0])
	return EnqueueReadBuffer(command_queue, buffer, blocking_read, uint64(offset)*itemSize, uint64(len(items))*itemSize, unsafe.Pointer(&items[0]), event_wait_list, event)
}

// Like [EnqueueWriteBuffer], but accepts a slice and automatically determines the memory region size.
//
// offset is now the number of ITEMS, not bytes (unlike [EnqueueWriteBuffer]).
//
//bind:extra
func EnqueueWriteBufferSlice[E any](command_queue CommandQueue, buffer Mem, blocking_write bool, offset int, items []E, event_wait_list []Event, event *Event) (_err error) {
	if len(items) == 0 {
		return makeError(INVALID_VALUE)
	}
	var pin runtime.Pinner
	defer pin.Unpin()
	itemSize := uint64(unsafe.Sizeof(items[0]))
	pin.Pin(&items[0])
	return EnqueueWriteBuffer(command_queue, buffer, blocking_write, uint64(offset)*itemSize, uint64(len(items))*itemSize, unsafe.Pointer(&items[0]), event_wait_list, event)
}

// Like [EnqueueFillBuffer].
//
// offset and size is now specified as a number of items (unlike [EnqueueFillBuffer]).
//
// Pattern must be a slice with a power of two size in bytes, which will repeatedly
// be filled into the device buffer starting at item offset, spanning size items.
//
// size must be a multiple of len(pattern).
//
// Returns [INVALID_VALUE] if len(pattern) == 0, size < 0, offset < 0, or the operation would write out of bounds.
//
//bind:extra
func EnqueueFillBufferSlice[E any](command_queue CommandQueue, buffer Mem, pattern []E, offset int, size int, event_wait_list []Event, event *Event) (_err error) {
	if len(pattern) == 0 || size < 0 || offset < 0 {
		return makeError(INVALID_VALUE)
	}
	var pin runtime.Pinner
	defer pin.Unpin()
	itemSize := uint64(unsafe.Sizeof(pattern[0]))
	pin.Pin(&pattern[0])
	return EnqueueFillBuffer(command_queue, buffer, unsafe.Pointer(&pattern[0]), uint64(len(pattern))*itemSize, uint64(offset)*itemSize, uint64(size)*itemSize, event_wait_list, event)
}

// BackedBuffer is a typed [Mem] ory object that
// holds the host buffer in itself. It is
// meant to simplify common buffer use cases
// and is not a full replacement for all OpenCL
// buffer features.
//
// It automatically pins the item memory and
// unpins it when [BackedBuffer.Release] is called.
// (Long-term pinning is necessary to guarantee
// valid memory accesses from C if memory is
// accessed outside of the function itself, e.g.
// through asynchronous use, or MEM_USE_HOST_PTR).
//
// Data synchronization between host and device
// memory is still left up to the user.
//
// When passing a BackedBuffer to [SetKernelArgValue]
// or similar, the BackedBuffer.Mem field must be
// passed instead of the BackedBuffer itself.
//
//bind:extra
type BackedBuffer[E any] struct {
	// OpenCL memory handle.
	Mem Mem
	// Internal Go buffer slice.
	Items []E
	// Pinner used to pin the buffer
	// slice memory. DO NOT TOUCH THIS
	// UNLESS YOU KNOW WHAT YOU'RE DOING.
	Pinner runtime.Pinner
}

// The [CreateBackedBuffer] of [CreateBufferWithProperties].
//
//bind:extra
func CreateBackedBufferWithProperties[E any](context Context, properties []MemProperties, flags MemFlags, items []E) (res BackedBuffer[E], errcode_ret error) {
	if flags & MEM_ALLOC_HOST_PTR != 0 {
		panic("MEM_ALLOC_HOST_PTR not allowed")
	}
	res.Items = items
	if len(res.Items) <= 0 {
		return BackedBuffer[E]{}, makeError(INVALID_VALUE)
	}
	itemSize := uint64(unsafe.Sizeof(res.Items[0]))
	res.Pinner.Pin(&res.Items[0])
	ptr := unsafe.Pointer(&res.Items[0])
	if flags&MEM_COPY_HOST_PTR == 0 && flags&MEM_USE_HOST_PTR == 0 {
		// Prevent CL_INVALID_HOST_PTR
		ptr = nil
	}
	if len(properties) == 0 {
		// BUG: CreateBufferWithProperties seems to cause a segfault on some system, so I'll just
		// use regular CreateBuffer if possible until I have figured this out.
		res.Mem, errcode_ret = CreateBuffer(context, flags, uint64(len(res.Items))*itemSize, ptr)
	} else {
		res.Mem, errcode_ret = CreateBufferWithProperties(context, properties, flags, uint64(len(res.Items))*itemSize, ptr)
	}
	if errcode_ret != nil {
		return BackedBuffer[E]{}, errcode_ret
	}
	return
}

// Creates a new buffer backed by a Go slice.
//
// Automatically pins the slice until released to ensure
// async usage works correctly.
//
// Note that the given items are only copied to the device if
// MEM_COPY_HOST_PTR or MEM_USE_HOST_PTR is set.
//
// The user MUST call [BackedBuffer.Release] when done to ensure
// the slice is deallocated.
//
//bind:extra
func CreateBackedBuffer[E any](context Context, flags MemFlags, items []E) (res BackedBuffer[E], errcode_ret error) {
	return CreateBackedBufferWithProperties(context, nil, flags, items)
}

// Calls [ReleaseMemObject] and unpins the memory.
//
//bind:extra
func (m *BackedBuffer[E]) Release() (err error) {
	err = ReleaseMemObject(m.Mem)
	m.Pinner.Unpin()
	return
}

// Like [EnqueueReadBuffer].
//
// items from indexing from start_offset to end_offset (exclusive) will be
// read back into the host buffer.
//
// end_offset may be set to -1 to indicate the range over all items.
//
// Returns [INVALID_VALUE] if start_offset >= end_offset, or any start_offset is negative.
//
//bind:extra
func (m BackedBuffer[E]) EnqueueRead(command_queue CommandQueue, blocking_read bool, start_offset int, end_offset int, event_wait_list []Event, event *Event) (_err error) {
	if end_offset == -1 {
		end_offset = len(m.Items)
	}
	if start_offset < 0 || end_offset < 0 || start_offset >= end_offset {
		return makeError(INVALID_VALUE)
	}
	itemSize := uint64(unsafe.Sizeof(m.Items[0]))
	return EnqueueReadBuffer(command_queue, m.Mem, blocking_read, uint64(start_offset)*itemSize, uint64(end_offset-start_offset)*itemSize, unsafe.Pointer(&m.Items[0]), event_wait_list, event)
}

// Like [EnqueueWriteBuffer].
//
// items from indexing from start_offset to end_offset (exclusive) will be
// written into the device buffer.
//
//
// end_offset may be set to -1 to indicate the range over all items.
//
// Returns [INVALID_VALUE] if start_offset >= end_offset, or any start_offset is negative.
//
//bind:extra
func (m BackedBuffer[E]) EnqueueWrite(command_queue CommandQueue, blocking_write bool, start_offset int, end_offset int, event_wait_list []Event, event *Event) (_err error) {
	if end_offset == -1 {
		end_offset = len(m.Items)
	}
	if start_offset < 0 || end_offset < 0 || start_offset >= end_offset {
		return makeError(INVALID_VALUE)
	}
	itemSize := uint64(unsafe.Sizeof(m.Items[0]))
	return EnqueueWriteBuffer(command_queue, m.Mem, blocking_write, uint64(start_offset)*itemSize, uint64(end_offset-start_offset)*itemSize, unsafe.Pointer(&m.Items[0]), event_wait_list, event)
}

// Calls [EnqueueFillBufferSlice] (see for documentation).
//
// Additionally, size may be -1 to indicate len(Items).
//
//bind:extra
func (m BackedBuffer[E]) EnqueueFill(command_queue CommandQueue, pattern []E, offset int, size int, event_wait_list []Event, event *Event) (_err error) {
	if size == -1 {
		size = len(m.Items)
	}
	return EnqueueFillBufferSlice(command_queue, m.Mem, pattern, offset, size, event_wait_list, event)
}