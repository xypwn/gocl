package cl

/*
#cgo CFLAGS: -DCL_TARGET_OPENCL_VERSION=310
#cgo CFLAGS: -DCL_NO_NON_ICD_DISPATCH_EXTENSION_PROTOTYPES
#cgo CFLAGS: -DCL_ENABLE_LAYERS
#cgo CFLAGS: -Wno-deprecated-declarations
#cgo CFLAGS: -Iinc/
#cgo windows LDFLAGS: -lcfgmgr32 -lruntimeobject -lOle32
#include "inc/CL/cl.h"
#include "cl_go_defs.h"
*/
import "C"

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"unsafe"
)

// Custom utilities
func boolToClBool(b bool) C.cl_bool {
	if b {
		return 1
	} else {
		return 0
	}
}
func cstringToString(cstr *C.char) string {
	s := (*[1 << 30]C.char)(unsafe.Pointer(cstr))
	i := 0
	for ; s[i] != 0; i++ {
	}
	str := make([]byte, i)
	for j := 0; j < i; j++ {
		str[j] = byte(s[j])
	}
	return string(str)
}
func sliceToC[E any, S ~[]E](s S) (data unsafe.Pointer, length C.size_t, fin func()) {
	if len(s) == 0 {
		return nil, 0, func() {}
	}
	var pin runtime.Pinner
	pin.Pin(&s[0])
	return unsafe.Pointer(&s[0]), C.size_t(len(s)), pin.Unpin
}
func sliceToCZeroTerm[E any, S ~[]E](s S) (data unsafe.Pointer, fin func()) {
	if len(s) == 0 {
		return nil, func() {}
	}
	var zero E
	s = append(slices.Clone(s), zero)
	var pin runtime.Pinner
	pin.Pin(&s[0])
	return unsafe.Pointer(&s[0]), pin.Unpin
}
func stringToC(s string) (_ *C.char, fin func()) {
	var pin runtime.Pinner
	cstr := append([]byte(s), 0)
	pin.Pin(&cstr[0])
	return (*C.char)(unsafe.Pointer(&cstr[0])), pin.Unpin
}
func stringsToC(ss []string, needLengths bool) (data **C.char, lengths *C.size_t, fin func()) {
	if len(ss) == 0 {
		return nil, nil, func() {}
	}
	var pin runtime.Pinner
	cstrs := make([][]byte, len(ss))
	for i, s := range ss {
		cstrs[i] = append([]byte(s), 0)
		pin.Pin(&cstrs[i][0])
	}
	pin.Pin(&cstrs[0])
	if needLengths {
		lens := make([]C.size_t, len(ss))
		for i := range ss {
			lens[i] = C.size_t(len(ss[i]))
		}
		pin.Pin(&lens[0])
		lengths = (*C.size_t)(unsafe.Pointer(&lens[0]))
	}
	ptrs := make([]*C.char, len(ss))
	for i := range ss {
		ptrs[i] = (*C.char)(unsafe.Pointer(&cstrs[i][0]))
	}
	pin.Pin(&ptrs[0])
	data = (**C.char)(unsafe.Pointer(&ptrs[0]))
	fin = pin.Unpin
	return
}
func byteSlicesToC(bs [][]byte) (data **C.uchar, lengths *C.size_t, fin func()) {
	if len(bs) == 0 {
		return nil, nil, func() {}
	}
	var pin runtime.Pinner
	pin.Pin(&bs[0])
	lens := make([]C.size_t, len(bs))
	for i := range bs {
		lens[i] = C.size_t(len(bs[i]))
		if len(bs[i]) > 0 {
			pin.Pin(&bs[i][0])
		}
	}
	pin.Pin(&lens[0])
	lengths = (*C.size_t)(unsafe.Pointer(&lens[0]))
	ptrs := make([]*C.uchar, len(bs))
	for i := range bs {
		ptrs[i] = (*C.uchar)(unsafe.Pointer(&bs[i][0]))
	}
	pin.Pin(&ptrs[0])
	data = (**C.uchar)(unsafe.Pointer(&ptrs[0]))
	fin = pin.Unpin
	return
}
func makeError(ec ErrorCode) error {
	if ec != SUCCESS {
		return &Error{Code: ec}
	} else {
		return nil
	}
}

// Callback handling
var (
	callbackMu  sync.Mutex
	callbackUid int = 1
	callbackFns     = make(map[int]any)
)

func callbackRegister(fn any) (uid int) {
	if fn == nil {
		panic("attempt to register nil callback")
	}
	callbackMu.Lock()
	defer callbackMu.Unlock()
	uid = callbackUid
	callbackFns[uid] = fn
	callbackUid++
	return
}
func callbackUnregister(uid int) {
	callbackMu.Lock()
	defer callbackMu.Unlock()
	delete(callbackFns, uid)
}
func callbackFn(uid int) (fn any) {
	callbackMu.Lock()
	defer callbackMu.Unlock()
	fn = callbackFns[uid]
	return
}

// Error type
type Error struct {
	Code ErrorCode
}

func (err *Error) Error() string {
	return err.Code.String()
}

// ErrorCode enum type
type ErrorCode int
// BEGIN cl-3.1/inc/CL/cl_platform.h //

// Typedefs

// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_char3.html
type Char3 C.cl_char3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_uchar3.html
type Uchar3 C.cl_uchar3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_short3.html
type Short3 C.cl_short3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_ushort3.html
type Ushort3 C.cl_ushort3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_half3.html
type Half3 C.cl_half3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_int3.html
type Int3 C.cl_int3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_uint3.html
type Uint3 C.cl_uint3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_long3.html
type Long3 C.cl_long3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_ulong3.html
type Ulong3 C.cl_ulong3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_float3.html
type Float3 C.cl_float3
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_double3.html
type Double3 C.cl_double3

// Structs


// END cl-3.1/inc/CL/cl_platform.h //

// BEGIN cl-3.1/inc/CL/cl.h //

// Typedefs

// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_platform_id.html
type PlatformId C.cl_platform_id
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_id.html
type DeviceId C.cl_device_id
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_context.html
type Context C.cl_context
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_command_queue.html
type CommandQueue C.cl_command_queue
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_mem.html
type Mem C.cl_mem
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_program.html
type Program C.cl_program
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel.html
type Kernel C.cl_kernel
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_event.html
type Event C.cl_event
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_sampler.html
type Sampler C.cl_sampler
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_bitfield.html
type Bitfield C.cl_bitfield
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_properties.html
type Properties C.cl_properties
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_type.html
type DeviceType C.cl_device_type
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_platform_info.html
type PlatformInfo C.cl_platform_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_info.html
type DeviceInfo C.cl_device_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_fp_config.html
type DeviceFpConfig C.cl_device_fp_config
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_mem_cache_type.html
type DeviceMemCacheType C.cl_device_mem_cache_type
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_local_mem_type.html
type DeviceLocalMemType C.cl_device_local_mem_type
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_exec_capabilities.html
type DeviceExecCapabilities C.cl_device_exec_capabilities
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_svm_capabilities.html
type DeviceSvmCapabilities C.cl_device_svm_capabilities
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_command_queue_properties.html
type CommandQueueProperties C.cl_command_queue_properties
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_partition_property.html
type DevicePartitionProperty C.cl_device_partition_property
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_affinity_domain.html
type DeviceAffinityDomain C.cl_device_affinity_domain
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_context_properties.html
type ContextProperties C.cl_context_properties
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_context_info.html
type ContextInfo C.cl_context_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_queue_properties.html
type QueueProperties C.cl_queue_properties
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_command_queue_info.html
type CommandQueueInfo C.cl_command_queue_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_channel_order.html
type ChannelOrder C.cl_channel_order
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_channel_type.html
type ChannelType C.cl_channel_type
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_mem_flags.html
type MemFlags C.cl_mem_flags
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_svm_mem_flags.html
type SvmMemFlags C.cl_svm_mem_flags
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_mem_object_type.html
type MemObjectType C.cl_mem_object_type
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_mem_info.html
type MemInfo C.cl_mem_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_mem_migration_flags.html
type MemMigrationFlags C.cl_mem_migration_flags
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_image_info.html
type ImageInfo C.cl_image_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_buffer_create_type.html
type BufferCreateType C.cl_buffer_create_type
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_addressing_mode.html
type AddressingMode C.cl_addressing_mode
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_filter_mode.html
type FilterMode C.cl_filter_mode
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_sampler_info.html
type SamplerInfo C.cl_sampler_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_map_flags.html
type MapFlags C.cl_map_flags
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_pipe_properties.html
type PipeProperties C.cl_pipe_properties
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_pipe_info.html
type PipeInfo C.cl_pipe_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_program_info.html
type ProgramInfo C.cl_program_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_program_build_info.html
type ProgramBuildInfo C.cl_program_build_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_program_binary_type.html
type ProgramBinaryType C.cl_program_binary_type
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_build_status.html
type BuildStatus C.cl_build_status
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel_info.html
type KernelInfo C.cl_kernel_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel_arg_info.html
type KernelArgInfo C.cl_kernel_arg_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel_arg_address_qualifier.html
type KernelArgAddressQualifier C.cl_kernel_arg_address_qualifier
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel_arg_access_qualifier.html
type KernelArgAccessQualifier C.cl_kernel_arg_access_qualifier
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel_arg_type_qualifier.html
type KernelArgTypeQualifier C.cl_kernel_arg_type_qualifier
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel_work_group_info.html
type KernelWorkGroupInfo C.cl_kernel_work_group_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel_sub_group_info.html
type KernelSubGroupInfo C.cl_kernel_sub_group_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_event_info.html
type EventInfo C.cl_event_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_command_type.html
type CommandType C.cl_command_type
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_profiling_info.html
type ProfilingInfo C.cl_profiling_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_sampler_properties.html
type SamplerProperties C.cl_sampler_properties
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_kernel_exec_info.html
type KernelExecInfo C.cl_kernel_exec_info
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_atomic_capabilities.html
type DeviceAtomicCapabilities C.cl_device_atomic_capabilities
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_device_enqueue_capabilities.html
type DeviceDeviceEnqueueCapabilities C.cl_device_device_enqueue_capabilities
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_khronos_vendor_id.html
type KhronosVendorId C.cl_khronos_vendor_id
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_mem_properties.html
type MemProperties C.cl_mem_properties
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_version.html
type Version C.cl_version
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_integer_dot_product_capabilities.html
type DeviceIntegerDotProductCapabilities C.cl_device_integer_dot_product_capabilities

// Structs

// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_image_format.html
type ImageFormat C.cl_image_format
func (s *ImageFormat) ImageChannelOrder() ChannelOrder { return ChannelOrder(s.image_channel_order) }
func (s *ImageFormat) SetImageChannelOrder(v ChannelOrder) { s.image_channel_order = C.cl_channel_order(v) }
func (s *ImageFormat) ImageChannelDataType() ChannelType { return ChannelType(s.image_channel_data_type) }
func (s *ImageFormat) SetImageChannelDataType(v ChannelType) { s.image_channel_data_type = C.cl_channel_type(v) }
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_image_desc.html
type ImageDesc C.cl_image_desc
func (s *ImageDesc) ImageType() MemObjectType { return MemObjectType(s.image_type) }
func (s *ImageDesc) SetImageType(v MemObjectType) { s.image_type = C.cl_mem_object_type(v) }
func (s *ImageDesc) ImageWidth() uint64 { return uint64(s.image_width) }
func (s *ImageDesc) SetImageWidth(v uint64) { s.image_width = C.size_t(v) }
func (s *ImageDesc) ImageHeight() uint64 { return uint64(s.image_height) }
func (s *ImageDesc) SetImageHeight(v uint64) { s.image_height = C.size_t(v) }
func (s *ImageDesc) ImageDepth() uint64 { return uint64(s.image_depth) }
func (s *ImageDesc) SetImageDepth(v uint64) { s.image_depth = C.size_t(v) }
func (s *ImageDesc) ImageArraySize() uint64 { return uint64(s.image_array_size) }
func (s *ImageDesc) SetImageArraySize(v uint64) { s.image_array_size = C.size_t(v) }
func (s *ImageDesc) ImageRowPitch() uint64 { return uint64(s.image_row_pitch) }
func (s *ImageDesc) SetImageRowPitch(v uint64) { s.image_row_pitch = C.size_t(v) }
func (s *ImageDesc) ImageSlicePitch() uint64 { return uint64(s.image_slice_pitch) }
func (s *ImageDesc) SetImageSlicePitch(v uint64) { s.image_slice_pitch = C.size_t(v) }
func (s *ImageDesc) NumMipLevels() uint32 { return uint32(s.num_mip_levels) }
func (s *ImageDesc) SetNumMipLevels(v uint32) { s.num_mip_levels = C.cl_uint(v) }
func (s *ImageDesc) NumSamples() uint32 { return uint32(s.num_samples) }
func (s *ImageDesc) SetNumSamples(v uint32) { s.num_samples = C.cl_uint(v) }
func (s *ImageDesc) Buffer() Mem { return Mem(s.buffer) }
func (s *ImageDesc) SetBuffer(v Mem) { s.buffer = C.cl_mem(v) }
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_buffer_region.html
type BufferRegion C.cl_buffer_region
func (s *BufferRegion) Origin() uint64 { return uint64(s.origin) }
func (s *BufferRegion) SetOrigin(v uint64) { s.origin = C.size_t(v) }
func (s *BufferRegion) Size() uint64 { return uint64(s.size) }
func (s *BufferRegion) SetSize(v uint64) { s.size = C.size_t(v) }
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_name_version.html
type NameVersion C.cl_name_version
func (s *NameVersion) Version() Version { return Version(s.version) }
func (s *NameVersion) SetVersion(v Version) { s.version = C.cl_version(v) }
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/cl_device_integer_dot_product_acceleration_properties.html
type DeviceIntegerDotProductAccelerationProperties C.cl_device_integer_dot_product_acceleration_properties
func (s *DeviceIntegerDotProductAccelerationProperties) SignedAccelerated() bool { return (s.signed_accelerated != 0) }
func (s *DeviceIntegerDotProductAccelerationProperties) SetSignedAccelerated(v bool) { s.signed_accelerated = boolToClBool(v) }
func (s *DeviceIntegerDotProductAccelerationProperties) UnsignedAccelerated() bool { return (s.unsigned_accelerated != 0) }
func (s *DeviceIntegerDotProductAccelerationProperties) SetUnsignedAccelerated(v bool) { s.unsigned_accelerated = boolToClBool(v) }
func (s *DeviceIntegerDotProductAccelerationProperties) MixedSignednessAccelerated() bool { return (s.mixed_signedness_accelerated != 0) }
func (s *DeviceIntegerDotProductAccelerationProperties) SetMixedSignednessAccelerated(v bool) { s.mixed_signedness_accelerated = boolToClBool(v) }
func (s *DeviceIntegerDotProductAccelerationProperties) AccumulatingSaturatingSignedAccelerated() bool { return (s.accumulating_saturating_signed_accelerated != 0) }
func (s *DeviceIntegerDotProductAccelerationProperties) SetAccumulatingSaturatingSignedAccelerated(v bool) { s.accumulating_saturating_signed_accelerated = boolToClBool(v) }
func (s *DeviceIntegerDotProductAccelerationProperties) AccumulatingSaturatingUnsignedAccelerated() bool { return (s.accumulating_saturating_unsigned_accelerated != 0) }
func (s *DeviceIntegerDotProductAccelerationProperties) SetAccumulatingSaturatingUnsignedAccelerated(v bool) { s.accumulating_saturating_unsigned_accelerated = boolToClBool(v) }
func (s *DeviceIntegerDotProductAccelerationProperties) AccumulatingSaturatingMixedSignednessAccelerated() bool { return (s.accumulating_saturating_mixed_signedness_accelerated != 0) }
func (s *DeviceIntegerDotProductAccelerationProperties) SetAccumulatingSaturatingMixedSignednessAccelerated(v bool) { s.accumulating_saturating_mixed_signedness_accelerated = boolToClBool(v) }

// Enums

// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ErrorCode
	SUCCESS ErrorCode = 0
	DEVICE_NOT_FOUND ErrorCode = -1
	DEVICE_NOT_AVAILABLE ErrorCode = -2
	COMPILER_NOT_AVAILABLE ErrorCode = -3
	MEM_OBJECT_ALLOCATION_FAILURE ErrorCode = -4
	OUT_OF_RESOURCES ErrorCode = -5
	OUT_OF_HOST_MEMORY ErrorCode = -6
	PROFILING_INFO_NOT_AVAILABLE ErrorCode = -7
	MEM_COPY_OVERLAP ErrorCode = -8
	IMAGE_FORMAT_MISMATCH ErrorCode = -9
	IMAGE_FORMAT_NOT_SUPPORTED ErrorCode = -10
	BUILD_PROGRAM_FAILURE ErrorCode = -11
	MAP_FAILURE ErrorCode = -12
	MISALIGNED_SUB_BUFFER_OFFSET ErrorCode = -13
	EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST ErrorCode = -14
	COMPILE_PROGRAM_FAILURE ErrorCode = -15
	LINKER_NOT_AVAILABLE ErrorCode = -16
	LINK_PROGRAM_FAILURE ErrorCode = -17
	DEVICE_PARTITION_FAILED ErrorCode = -18
	KERNEL_ARG_INFO_NOT_AVAILABLE ErrorCode = -19
	INVALID_VALUE ErrorCode = -30
	INVALID_DEVICE_TYPE ErrorCode = -31
	INVALID_PLATFORM ErrorCode = -32
	INVALID_DEVICE ErrorCode = -33
	INVALID_CONTEXT ErrorCode = -34
	INVALID_QUEUE_PROPERTIES ErrorCode = -35
	INVALID_COMMAND_QUEUE ErrorCode = -36
	INVALID_HOST_PTR ErrorCode = -37
	INVALID_MEM_OBJECT ErrorCode = -38
	INVALID_IMAGE_FORMAT_DESCRIPTOR ErrorCode = -39
	INVALID_IMAGE_SIZE ErrorCode = -40
	INVALID_SAMPLER ErrorCode = -41
	INVALID_BINARY ErrorCode = -42
	INVALID_BUILD_OPTIONS ErrorCode = -43
	INVALID_PROGRAM ErrorCode = -44
	INVALID_PROGRAM_EXECUTABLE ErrorCode = -45
	INVALID_KERNEL_NAME ErrorCode = -46
	INVALID_KERNEL_DEFINITION ErrorCode = -47
	INVALID_KERNEL ErrorCode = -48
	INVALID_ARG_INDEX ErrorCode = -49
	INVALID_ARG_VALUE ErrorCode = -50
	INVALID_ARG_SIZE ErrorCode = -51
	INVALID_KERNEL_ARGS ErrorCode = -52
	INVALID_WORK_DIMENSION ErrorCode = -53
	INVALID_WORK_GROUP_SIZE ErrorCode = -54
	INVALID_WORK_ITEM_SIZE ErrorCode = -55
	INVALID_GLOBAL_OFFSET ErrorCode = -56
	INVALID_EVENT_WAIT_LIST ErrorCode = -57
	INVALID_EVENT ErrorCode = -58
	INVALID_OPERATION ErrorCode = -59
	INVALID_GL_OBJECT ErrorCode = -60
	INVALID_BUFFER_SIZE ErrorCode = -61
	INVALID_MIP_LEVEL ErrorCode = -62
	INVALID_GLOBAL_WORK_SIZE ErrorCode = -63
	INVALID_PROPERTY ErrorCode = -64
	INVALID_IMAGE_DESCRIPTOR ErrorCode = -65
	INVALID_COMPILER_OPTIONS ErrorCode = -66
	INVALID_LINKER_OPTIONS ErrorCode = -67
	INVALID_DEVICE_PARTITION_COUNT ErrorCode = -68
	INVALID_PIPE_SIZE ErrorCode = -69
	INVALID_DEVICE_QUEUE ErrorCode = -70
	INVALID_SPEC_ID ErrorCode = -71
	MAX_SIZE_RESTRICTION_EXCEEDED ErrorCode = -72
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // bool
	FALSE bool = false
	TRUE bool = true
	BLOCKING bool = true
	NON_BLOCKING bool = false
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // PlatformInfo
	PLATFORM_PROFILE PlatformInfo = 0x0900
	PLATFORM_VERSION PlatformInfo = 0x0901
	PLATFORM_NAME PlatformInfo = 0x0902
	PLATFORM_VENDOR PlatformInfo = 0x0903
	PLATFORM_EXTENSIONS PlatformInfo = 0x0904
	PLATFORM_HOST_TIMER_RESOLUTION PlatformInfo = 0x0905
	PLATFORM_NUMERIC_VERSION PlatformInfo = 0x0906
	PLATFORM_EXTENSIONS_WITH_VERSION PlatformInfo = 0x0907
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceType
	DEVICE_TYPE_DEFAULT DeviceType = (1 << 0)
	DEVICE_TYPE_CPU DeviceType = (1 << 1)
	DEVICE_TYPE_GPU DeviceType = (1 << 2)
	DEVICE_TYPE_ACCELERATOR DeviceType = (1 << 3)
	DEVICE_TYPE_CUSTOM DeviceType = (1 << 4)
	DEVICE_TYPE_ALL DeviceType = 0xFFFFFFFF
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceInfo
	DEVICE_TYPE DeviceInfo = 0x1000
	DEVICE_VENDOR_ID DeviceInfo = 0x1001
	DEVICE_MAX_COMPUTE_UNITS DeviceInfo = 0x1002
	DEVICE_MAX_WORK_ITEM_DIMENSIONS DeviceInfo = 0x1003
	DEVICE_MAX_WORK_GROUP_SIZE DeviceInfo = 0x1004
	DEVICE_MAX_WORK_ITEM_SIZES DeviceInfo = 0x1005    /* deprecated */
	DEVICE_MAX_WORK_GROUP_SIZES DeviceInfo = 0x1005
	DEVICE_PREFERRED_VECTOR_WIDTH_CHAR DeviceInfo = 0x1006
	DEVICE_PREFERRED_VECTOR_WIDTH_SHORT DeviceInfo = 0x1007
	DEVICE_PREFERRED_VECTOR_WIDTH_INT DeviceInfo = 0x1008
	DEVICE_PREFERRED_VECTOR_WIDTH_LONG DeviceInfo = 0x1009
	DEVICE_PREFERRED_VECTOR_WIDTH_FLOAT DeviceInfo = 0x100A
	DEVICE_PREFERRED_VECTOR_WIDTH_DOUBLE DeviceInfo = 0x100B
	DEVICE_MAX_CLOCK_FREQUENCY DeviceInfo = 0x100C
	DEVICE_ADDRESS_BITS DeviceInfo = 0x100D
	DEVICE_MAX_READ_IMAGE_ARGS DeviceInfo = 0x100E
	DEVICE_MAX_WRITE_IMAGE_ARGS DeviceInfo = 0x100F
	DEVICE_MAX_MEM_ALLOC_SIZE DeviceInfo = 0x1010
	DEVICE_IMAGE2D_MAX_WIDTH DeviceInfo = 0x1011
	DEVICE_IMAGE2D_MAX_HEIGHT DeviceInfo = 0x1012
	DEVICE_IMAGE3D_MAX_WIDTH DeviceInfo = 0x1013
	DEVICE_IMAGE3D_MAX_HEIGHT DeviceInfo = 0x1014
	DEVICE_IMAGE3D_MAX_DEPTH DeviceInfo = 0x1015
	DEVICE_IMAGE_SUPPORT DeviceInfo = 0x1016
	DEVICE_MAX_PARAMETER_SIZE DeviceInfo = 0x1017
	DEVICE_MAX_SAMPLERS DeviceInfo = 0x1018
	DEVICE_MEM_BASE_ADDR_ALIGN DeviceInfo = 0x1019
	DEVICE_MIN_DATA_TYPE_ALIGN_SIZE DeviceInfo = 0x101A
	DEVICE_SINGLE_FP_CONFIG DeviceInfo = 0x101B
	DEVICE_GLOBAL_MEM_CACHE_TYPE DeviceInfo = 0x101C
	DEVICE_GLOBAL_MEM_CACHELINE_SIZE DeviceInfo = 0x101D
	DEVICE_GLOBAL_MEM_CACHE_SIZE DeviceInfo = 0x101E
	DEVICE_GLOBAL_MEM_SIZE DeviceInfo = 0x101F
	DEVICE_MAX_CONSTANT_BUFFER_SIZE DeviceInfo = 0x1020
	DEVICE_MAX_CONSTANT_ARGS DeviceInfo = 0x1021
	DEVICE_LOCAL_MEM_TYPE DeviceInfo = 0x1022
	DEVICE_LOCAL_MEM_SIZE DeviceInfo = 0x1023
	DEVICE_ERROR_CORRECTION_SUPPORT DeviceInfo = 0x1024
	DEVICE_PROFILING_TIMER_RESOLUTION DeviceInfo = 0x1025
	DEVICE_ENDIAN_LITTLE DeviceInfo = 0x1026
	DEVICE_AVAILABLE DeviceInfo = 0x1027
	DEVICE_COMPILER_AVAILABLE DeviceInfo = 0x1028
	DEVICE_EXECUTION_CAPABILITIES DeviceInfo = 0x1029
	DEVICE_QUEUE_PROPERTIES DeviceInfo = 0x102A    /* deprecated */
	DEVICE_QUEUE_ON_HOST_PROPERTIES DeviceInfo = 0x102A
	DEVICE_NAME DeviceInfo = 0x102B
	DEVICE_VENDOR DeviceInfo = 0x102C
	DRIVER_VERSION DeviceInfo = 0x102D
	DEVICE_PROFILE DeviceInfo = 0x102E
	DEVICE_VERSION DeviceInfo = 0x102F
	DEVICE_EXTENSIONS DeviceInfo = 0x1030
	DEVICE_PLATFORM DeviceInfo = 0x1031
	DEVICE_DOUBLE_FP_CONFIG DeviceInfo = 0x1032
	DEVICE_PREFERRED_VECTOR_WIDTH_HALF DeviceInfo = 0x1034
	DEVICE_HOST_UNIFIED_MEMORY DeviceInfo = 0x1035
	DEVICE_NATIVE_VECTOR_WIDTH_CHAR DeviceInfo = 0x1036
	DEVICE_NATIVE_VECTOR_WIDTH_SHORT DeviceInfo = 0x1037
	DEVICE_NATIVE_VECTOR_WIDTH_INT DeviceInfo = 0x1038
	DEVICE_NATIVE_VECTOR_WIDTH_LONG DeviceInfo = 0x1039
	DEVICE_NATIVE_VECTOR_WIDTH_FLOAT DeviceInfo = 0x103A
	DEVICE_NATIVE_VECTOR_WIDTH_DOUBLE DeviceInfo = 0x103B
	DEVICE_NATIVE_VECTOR_WIDTH_HALF DeviceInfo = 0x103C
	DEVICE_OPENCL_C_VERSION DeviceInfo = 0x103D
	DEVICE_LINKER_AVAILABLE DeviceInfo = 0x103E
	DEVICE_BUILT_IN_KERNELS DeviceInfo = 0x103F
	DEVICE_IMAGE_MAX_BUFFER_SIZE DeviceInfo = 0x1040
	DEVICE_IMAGE_MAX_ARRAY_SIZE DeviceInfo = 0x1041
	DEVICE_PARENT_DEVICE DeviceInfo = 0x1042
	DEVICE_PARTITION_MAX_SUB_DEVICES DeviceInfo = 0x1043
	DEVICE_PARTITION_PROPERTIES DeviceInfo = 0x1044
	DEVICE_PARTITION_AFFINITY_DOMAIN DeviceInfo = 0x1045
	DEVICE_PARTITION_TYPE DeviceInfo = 0x1046
	DEVICE_REFERENCE_COUNT DeviceInfo = 0x1047
	DEVICE_PREFERRED_INTEROP_USER_SYNC DeviceInfo = 0x1048
	DEVICE_PRINTF_BUFFER_SIZE DeviceInfo = 0x1049
	DEVICE_IMAGE_PITCH_ALIGNMENT DeviceInfo = 0x104A
	DEVICE_IMAGE_BASE_ADDRESS_ALIGNMENT DeviceInfo = 0x104B
	DEVICE_MAX_READ_WRITE_IMAGE_ARGS DeviceInfo = 0x104C
	DEVICE_MAX_GLOBAL_VARIABLE_SIZE DeviceInfo = 0x104D
	DEVICE_QUEUE_ON_DEVICE_PROPERTIES DeviceInfo = 0x104E
	DEVICE_QUEUE_ON_DEVICE_PREFERRED_SIZE DeviceInfo = 0x104F
	DEVICE_QUEUE_ON_DEVICE_MAX_SIZE DeviceInfo = 0x1050
	DEVICE_MAX_ON_DEVICE_QUEUES DeviceInfo = 0x1051
	DEVICE_MAX_ON_DEVICE_EVENTS DeviceInfo = 0x1052
	DEVICE_SVM_CAPABILITIES DeviceInfo = 0x1053
	DEVICE_GLOBAL_VARIABLE_PREFERRED_TOTAL_SIZE DeviceInfo = 0x1054
	DEVICE_MAX_PIPE_ARGS DeviceInfo = 0x1055
	DEVICE_PIPE_MAX_ACTIVE_RESERVATIONS DeviceInfo = 0x1056
	DEVICE_PIPE_MAX_PACKET_SIZE DeviceInfo = 0x1057
	DEVICE_PREFERRED_PLATFORM_ATOMIC_ALIGNMENT DeviceInfo = 0x1058
	DEVICE_PREFERRED_GLOBAL_ATOMIC_ALIGNMENT DeviceInfo = 0x1059
	DEVICE_PREFERRED_LOCAL_ATOMIC_ALIGNMENT DeviceInfo = 0x105A
	DEVICE_IL_VERSION DeviceInfo = 0x105B
	DEVICE_MAX_NUM_SUB_GROUPS DeviceInfo = 0x105C
	DEVICE_SUB_GROUP_INDEPENDENT_FORWARD_PROGRESS DeviceInfo = 0x105D
	DEVICE_NUMERIC_VERSION DeviceInfo = 0x105E
	DEVICE_EXTENSIONS_WITH_VERSION DeviceInfo = 0x1060
	DEVICE_ILS_WITH_VERSION DeviceInfo = 0x1061
	DEVICE_BUILT_IN_KERNELS_WITH_VERSION DeviceInfo = 0x1062
	DEVICE_ATOMIC_MEMORY_CAPABILITIES DeviceInfo = 0x1063
	DEVICE_ATOMIC_FENCE_CAPABILITIES DeviceInfo = 0x1064
	DEVICE_NON_UNIFORM_WORK_GROUP_SUPPORT DeviceInfo = 0x1065
	DEVICE_OPENCL_C_ALL_VERSIONS DeviceInfo = 0x1066
	DEVICE_PREFERRED_WORK_GROUP_SIZE_MULTIPLE DeviceInfo = 0x1067
	DEVICE_WORK_GROUP_COLLECTIVE_FUNCTIONS_SUPPORT DeviceInfo = 0x1068
	DEVICE_GENERIC_ADDRESS_SPACE_SUPPORT DeviceInfo = 0x1069
	DEVICE_UUID DeviceInfo = 0x106A
	DRIVER_UUID DeviceInfo = 0x106B
	DEVICE_LUID_VALID DeviceInfo = 0x106C
	DEVICE_LUID DeviceInfo = 0x106D
	DEVICE_NODE_MASK DeviceInfo = 0x106E
	DEVICE_OPENCL_C_FEATURES DeviceInfo = 0x106F
	DEVICE_DEVICE_ENQUEUE_CAPABILITIES DeviceInfo = 0x1070
	DEVICE_PIPE_SUPPORT DeviceInfo = 0x1071
	DEVICE_LATEST_CONFORMANCE_VERSION_PASSED DeviceInfo = 0x1072
	DEVICE_INTEGER_DOT_PRODUCT_CAPABILITIES DeviceInfo = 0x1073
	DEVICE_INTEGER_DOT_PRODUCT_ACCELERATION_PROPERTIES_8BIT DeviceInfo = 0x1074
	DEVICE_INTEGER_DOT_PRODUCT_ACCELERATION_PROPERTIES_4x8BIT_PACKED DeviceInfo = 0x1075
	DEVICE_SPIRV_EXTENDED_INSTRUCTION_SETS DeviceInfo = 0x12B9
	DEVICE_SPIRV_EXTENSIONS DeviceInfo = 0x12BA
	DEVICE_SPIRV_CAPABILITIES DeviceInfo = 0x12BB
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceFpConfig
	FP_DENORM DeviceFpConfig = (1 << 0)
	FP_INF_NAN DeviceFpConfig = (1 << 1)
	FP_ROUND_TO_NEAREST DeviceFpConfig = (1 << 2)
	FP_ROUND_TO_ZERO DeviceFpConfig = (1 << 3)
	FP_ROUND_TO_INF DeviceFpConfig = (1 << 4)
	FP_FMA DeviceFpConfig = (1 << 5)
	FP_SOFT_FLOAT DeviceFpConfig = (1 << 6)
	FP_CORRECTLY_ROUNDED_DIVIDE_SQRT DeviceFpConfig = (1 << 7)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceMemCacheType
	NONE DeviceMemCacheType = 0x0
	READ_ONLY_CACHE DeviceMemCacheType = 0x1
	READ_WRITE_CACHE DeviceMemCacheType = 0x2
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceLocalMemType
	LOCAL DeviceLocalMemType = 0x1
	GLOBAL DeviceLocalMemType = 0x2
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceExecCapabilities
	EXEC_KERNEL DeviceExecCapabilities = (1 << 0)
	EXEC_NATIVE_KERNEL DeviceExecCapabilities = (1 << 1)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // CommandQueueProperties
	QUEUE_OUT_OF_ORDER_EXEC_MODE_ENABLE CommandQueueProperties = (1 << 0)
	QUEUE_PROFILING_ENABLE CommandQueueProperties = (1 << 1)
	QUEUE_ON_DEVICE CommandQueueProperties = (1 << 2)
	QUEUE_ON_DEVICE_DEFAULT CommandQueueProperties = (1 << 3)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ContextInfo
	CONTEXT_REFERENCE_COUNT ContextInfo = 0x1080
	CONTEXT_DEVICES ContextInfo = 0x1081
	CONTEXT_PROPERTIES ContextInfo = 0x1082
	CONTEXT_NUM_DEVICES ContextInfo = 0x1083
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ContextProperties
	CONTEXT_PLATFORM ContextProperties = 0x1084
	CONTEXT_INTEROP_USER_SYNC ContextProperties = 0x1085
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DevicePartitionProperty
	DEVICE_PARTITION_EQUALLY DevicePartitionProperty = 0x1086
	DEVICE_PARTITION_BY_COUNTS DevicePartitionProperty = 0x1087
	DEVICE_PARTITION_BY_COUNTS_LIST_END DevicePartitionProperty = 0x0
	DEVICE_PARTITION_BY_AFFINITY_DOMAIN DevicePartitionProperty = 0x1088
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceAffinityDomain
	DEVICE_AFFINITY_DOMAIN_NUMA DeviceAffinityDomain = (1 << 0)
	DEVICE_AFFINITY_DOMAIN_L4_CACHE DeviceAffinityDomain = (1 << 1)
	DEVICE_AFFINITY_DOMAIN_L3_CACHE DeviceAffinityDomain = (1 << 2)
	DEVICE_AFFINITY_DOMAIN_L2_CACHE DeviceAffinityDomain = (1 << 3)
	DEVICE_AFFINITY_DOMAIN_L1_CACHE DeviceAffinityDomain = (1 << 4)
	DEVICE_AFFINITY_DOMAIN_NEXT_PARTITIONABLE DeviceAffinityDomain = (1 << 5)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceSvmCapabilities
	DEVICE_SVM_COARSE_GRAIN_BUFFER DeviceSvmCapabilities = (1 << 0)
	DEVICE_SVM_FINE_GRAIN_BUFFER DeviceSvmCapabilities = (1 << 1)
	DEVICE_SVM_FINE_GRAIN_SYSTEM DeviceSvmCapabilities = (1 << 2)
	DEVICE_SVM_ATOMICS DeviceSvmCapabilities = (1 << 3)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // CommandQueueInfo
	QUEUE_CONTEXT CommandQueueInfo = 0x1090
	QUEUE_DEVICE CommandQueueInfo = 0x1091
	QUEUE_REFERENCE_COUNT CommandQueueInfo = 0x1092
	QUEUE_PROPERTIES CommandQueueInfo = 0x1093
	QUEUE_SIZE CommandQueueInfo = 0x1094
	QUEUE_DEVICE_DEFAULT CommandQueueInfo = 0x1095
	QUEUE_PROPERTIES_ARRAY CommandQueueInfo = 0x1098
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // MemFlags
	MEM_READ_WRITE MemFlags = (1 << 0)
	MEM_WRITE_ONLY MemFlags = (1 << 1)
	MEM_READ_ONLY MemFlags = (1 << 2)
	MEM_USE_HOST_PTR MemFlags = (1 << 3)
	MEM_ALLOC_HOST_PTR MemFlags = (1 << 4)
	MEM_COPY_HOST_PTR MemFlags = (1 << 5)
	MEM_HOST_WRITE_ONLY MemFlags = (1 << 7)
	MEM_HOST_READ_ONLY MemFlags = (1 << 8)
	MEM_HOST_NO_ACCESS MemFlags = (1 << 9)
	MEM_SVM_FINE_GRAIN_BUFFER MemFlags = (1 << 10)   /* used by cl_svm_mem_flags only */
	MEM_SVM_ATOMICS MemFlags = (1 << 11)   /* used by cl_svm_mem_flags only */
	MEM_KERNEL_READ_AND_WRITE MemFlags = (1 << 12)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // MemMigrationFlags
	MIGRATE_MEM_OBJECT_HOST MemMigrationFlags = (1 << 0)
	MIGRATE_MEM_OBJECT_CONTENT_UNDEFINED MemMigrationFlags = (1 << 1)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ChannelOrder
	R ChannelOrder = 0x10B0
	A ChannelOrder = 0x10B1
	RG ChannelOrder = 0x10B2
	RA ChannelOrder = 0x10B3
	RGB ChannelOrder = 0x10B4
	RGBA ChannelOrder = 0x10B5
	BGRA ChannelOrder = 0x10B6
	ARGB ChannelOrder = 0x10B7
	INTENSITY ChannelOrder = 0x10B8
	LUMINANCE ChannelOrder = 0x10B9
	Rx ChannelOrder = 0x10BA
	RGx ChannelOrder = 0x10BB
	RGBx ChannelOrder = 0x10BC
	DEPTH ChannelOrder = 0x10BD
	sRGB ChannelOrder = 0x10BF
	sRGBx ChannelOrder = 0x10C0
	sRGBA ChannelOrder = 0x10C1
	sBGRA ChannelOrder = 0x10C2
	ABGR ChannelOrder = 0x10C3
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ChannelType
	SNORM_INT8 ChannelType = 0x10D0
	SNORM_INT16 ChannelType = 0x10D1
	UNORM_INT8 ChannelType = 0x10D2
	UNORM_INT16 ChannelType = 0x10D3
	UNORM_SHORT_565 ChannelType = 0x10D4
	UNORM_SHORT_555 ChannelType = 0x10D5
	UNORM_INT_101010 ChannelType = 0x10D6
	SIGNED_INT8 ChannelType = 0x10D7
	SIGNED_INT16 ChannelType = 0x10D8
	SIGNED_INT32 ChannelType = 0x10D9
	UNSIGNED_INT8 ChannelType = 0x10DA
	UNSIGNED_INT16 ChannelType = 0x10DB
	UNSIGNED_INT32 ChannelType = 0x10DC
	HALF_FLOAT ChannelType = 0x10DD
	FLOAT ChannelType = 0x10DE
	UNORM_INT_101010_2 ChannelType = 0x10E0
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // MemObjectType
	MEM_OBJECT_BUFFER MemObjectType = 0x10F0
	MEM_OBJECT_IMAGE2D MemObjectType = 0x10F1
	MEM_OBJECT_IMAGE3D MemObjectType = 0x10F2
	MEM_OBJECT_IMAGE2D_ARRAY MemObjectType = 0x10F3
	MEM_OBJECT_IMAGE1D MemObjectType = 0x10F4
	MEM_OBJECT_IMAGE1D_ARRAY MemObjectType = 0x10F5
	MEM_OBJECT_IMAGE1D_BUFFER MemObjectType = 0x10F6
	MEM_OBJECT_PIPE MemObjectType = 0x10F7
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // MemInfo
	MEM_TYPE MemInfo = 0x1100
	MEM_FLAGS MemInfo = 0x1101
	MEM_SIZE MemInfo = 0x1102
	MEM_HOST_PTR MemInfo = 0x1103
	MEM_MAP_COUNT MemInfo = 0x1104
	MEM_REFERENCE_COUNT MemInfo = 0x1105
	MEM_CONTEXT MemInfo = 0x1106
	MEM_ASSOCIATED_MEMOBJECT MemInfo = 0x1107
	MEM_OFFSET MemInfo = 0x1108
	MEM_USES_SVM_POINTER MemInfo = 0x1109
	MEM_PROPERTIES MemInfo = 0x110A
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ImageInfo
	IMAGE_FORMAT ImageInfo = 0x1110
	IMAGE_ELEMENT_SIZE ImageInfo = 0x1111
	IMAGE_ROW_PITCH ImageInfo = 0x1112
	IMAGE_SLICE_PITCH ImageInfo = 0x1113
	IMAGE_WIDTH ImageInfo = 0x1114
	IMAGE_HEIGHT ImageInfo = 0x1115
	IMAGE_DEPTH ImageInfo = 0x1116
	IMAGE_ARRAY_SIZE ImageInfo = 0x1117
	IMAGE_BUFFER ImageInfo = 0x1118
	IMAGE_NUM_MIP_LEVELS ImageInfo = 0x1119
	IMAGE_NUM_SAMPLES ImageInfo = 0x111A
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // PipeInfo
	PIPE_PACKET_SIZE PipeInfo = 0x1120
	PIPE_MAX_PACKETS PipeInfo = 0x1121
	PIPE_PROPERTIES PipeInfo = 0x1122
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // AddressingMode
	ADDRESS_NONE AddressingMode = 0x1130
	ADDRESS_CLAMP_TO_EDGE AddressingMode = 0x1131
	ADDRESS_CLAMP AddressingMode = 0x1132
	ADDRESS_REPEAT AddressingMode = 0x1133
	ADDRESS_MIRRORED_REPEAT AddressingMode = 0x1134
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // FilterMode
	FILTER_NEAREST FilterMode = 0x1140
	FILTER_LINEAR FilterMode = 0x1141
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // SamplerInfo
	SAMPLER_REFERENCE_COUNT SamplerInfo = 0x1150
	SAMPLER_CONTEXT SamplerInfo = 0x1151
	SAMPLER_NORMALIZED_COORDS SamplerInfo = 0x1152
	SAMPLER_ADDRESSING_MODE SamplerInfo = 0x1153
	SAMPLER_FILTER_MODE SamplerInfo = 0x1154
	SAMPLER_MIP_FILTER_MODE SamplerInfo = 0x1155
	SAMPLER_LOD_MIN SamplerInfo = 0x1156
	SAMPLER_LOD_MAX SamplerInfo = 0x1157
	SAMPLER_PROPERTIES SamplerInfo = 0x1158
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // MapFlags
	MAP_READ MapFlags = (1 << 0)
	MAP_WRITE MapFlags = (1 << 1)
	MAP_WRITE_INVALIDATE_REGION MapFlags = (1 << 2)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ProgramInfo
	PROGRAM_REFERENCE_COUNT ProgramInfo = 0x1160
	PROGRAM_CONTEXT ProgramInfo = 0x1161
	PROGRAM_NUM_DEVICES ProgramInfo = 0x1162
	PROGRAM_DEVICES ProgramInfo = 0x1163
	PROGRAM_SOURCE ProgramInfo = 0x1164
	PROGRAM_BINARY_SIZES ProgramInfo = 0x1165
	PROGRAM_BINARIES ProgramInfo = 0x1166
	PROGRAM_NUM_KERNELS ProgramInfo = 0x1167
	PROGRAM_KERNEL_NAMES ProgramInfo = 0x1168
	PROGRAM_IL ProgramInfo = 0x1169
	PROGRAM_SCOPE_GLOBAL_CTORS_PRESENT ProgramInfo = 0x116A
	PROGRAM_SCOPE_GLOBAL_DTORS_PRESENT ProgramInfo = 0x116B
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ProgramBuildInfo
	PROGRAM_BUILD_STATUS ProgramBuildInfo = 0x1181
	PROGRAM_BUILD_OPTIONS ProgramBuildInfo = 0x1182
	PROGRAM_BUILD_LOG ProgramBuildInfo = 0x1183
	PROGRAM_BINARY_TYPE ProgramBuildInfo = 0x1184
	PROGRAM_BUILD_GLOBAL_VARIABLE_TOTAL_SIZE ProgramBuildInfo = 0x1185
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ProgramBinaryType
	PROGRAM_BINARY_TYPE_NONE ProgramBinaryType = 0x0
	PROGRAM_BINARY_TYPE_COMPILED_OBJECT ProgramBinaryType = 0x1
	PROGRAM_BINARY_TYPE_LIBRARY ProgramBinaryType = 0x2
	PROGRAM_BINARY_TYPE_EXECUTABLE ProgramBinaryType = 0x4
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // BuildStatus
	BUILD_SUCCESS BuildStatus = 0
	BUILD_NONE BuildStatus = -1
	BUILD_ERROR BuildStatus = -2
	BUILD_IN_PROGRESS BuildStatus = -3
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KernelInfo
	KERNEL_FUNCTION_NAME KernelInfo = 0x1190
	KERNEL_NUM_ARGS KernelInfo = 0x1191
	KERNEL_REFERENCE_COUNT KernelInfo = 0x1192
	KERNEL_CONTEXT KernelInfo = 0x1193
	KERNEL_PROGRAM KernelInfo = 0x1194
	KERNEL_ATTRIBUTES KernelInfo = 0x1195
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KernelArgInfo
	KERNEL_ARG_ADDRESS_QUALIFIER KernelArgInfo = 0x1196
	KERNEL_ARG_ACCESS_QUALIFIER KernelArgInfo = 0x1197
	KERNEL_ARG_TYPE_NAME KernelArgInfo = 0x1198
	KERNEL_ARG_TYPE_QUALIFIER KernelArgInfo = 0x1199
	KERNEL_ARG_NAME KernelArgInfo = 0x119A
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KernelArgAddressQualifier
	KERNEL_ARG_ADDRESS_GLOBAL KernelArgAddressQualifier = 0x119B
	KERNEL_ARG_ADDRESS_LOCAL KernelArgAddressQualifier = 0x119C
	KERNEL_ARG_ADDRESS_CONSTANT KernelArgAddressQualifier = 0x119D
	KERNEL_ARG_ADDRESS_PRIVATE KernelArgAddressQualifier = 0x119E
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KernelArgAccessQualifier
	KERNEL_ARG_ACCESS_READ_ONLY KernelArgAccessQualifier = 0x11A0
	KERNEL_ARG_ACCESS_WRITE_ONLY KernelArgAccessQualifier = 0x11A1
	KERNEL_ARG_ACCESS_READ_WRITE KernelArgAccessQualifier = 0x11A2
	KERNEL_ARG_ACCESS_NONE KernelArgAccessQualifier = 0x11A3
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KernelArgTypeQualifier
	KERNEL_ARG_TYPE_NONE KernelArgTypeQualifier = 0
	KERNEL_ARG_TYPE_CONST KernelArgTypeQualifier = (1 << 0)
	KERNEL_ARG_TYPE_RESTRICT KernelArgTypeQualifier = (1 << 1)
	KERNEL_ARG_TYPE_VOLATILE KernelArgTypeQualifier = (1 << 2)
	KERNEL_ARG_TYPE_PIPE KernelArgTypeQualifier = (1 << 3)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KernelWorkGroupInfo
	KERNEL_WORK_GROUP_SIZE KernelWorkGroupInfo = 0x11B0
	KERNEL_COMPILE_WORK_GROUP_SIZE KernelWorkGroupInfo = 0x11B1
	KERNEL_LOCAL_MEM_SIZE KernelWorkGroupInfo = 0x11B2
	KERNEL_PREFERRED_WORK_GROUP_SIZE_MULTIPLE KernelWorkGroupInfo = 0x11B3
	KERNEL_PRIVATE_MEM_SIZE KernelWorkGroupInfo = 0x11B4
	KERNEL_GLOBAL_WORK_SIZE KernelWorkGroupInfo = 0x11B5
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KernelSubGroupInfo
	KERNEL_MAX_SUB_GROUP_SIZE_FOR_NDRANGE KernelSubGroupInfo = 0x2033
	KERNEL_SUB_GROUP_COUNT_FOR_NDRANGE KernelSubGroupInfo = 0x2034
	KERNEL_LOCAL_SIZE_FOR_SUB_GROUP_COUNT KernelSubGroupInfo = 0x11B8
	KERNEL_MAX_NUM_SUB_GROUPS KernelSubGroupInfo = 0x11B9
	KERNEL_COMPILE_NUM_SUB_GROUPS KernelSubGroupInfo = 0x11BA
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KernelExecInfo
	KERNEL_EXEC_INFO_SVM_PTRS KernelExecInfo = 0x11B6
	KERNEL_EXEC_INFO_SVM_FINE_GRAIN_SYSTEM KernelExecInfo = 0x11B7
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // EventInfo
	EVENT_COMMAND_QUEUE EventInfo = 0x11D0
	EVENT_COMMAND_TYPE EventInfo = 0x11D1
	EVENT_REFERENCE_COUNT EventInfo = 0x11D2
	EVENT_COMMAND_EXECUTION_STATUS EventInfo = 0x11D3
	EVENT_CONTEXT EventInfo = 0x11D4
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // CommandType
	COMMAND_NDRANGE_KERNEL CommandType = 0x11F0
	COMMAND_TASK CommandType = 0x11F1
	COMMAND_NATIVE_KERNEL CommandType = 0x11F2
	COMMAND_READ_BUFFER CommandType = 0x11F3
	COMMAND_WRITE_BUFFER CommandType = 0x11F4
	COMMAND_COPY_BUFFER CommandType = 0x11F5
	COMMAND_READ_IMAGE CommandType = 0x11F6
	COMMAND_WRITE_IMAGE CommandType = 0x11F7
	COMMAND_COPY_IMAGE CommandType = 0x11F8
	COMMAND_COPY_IMAGE_TO_BUFFER CommandType = 0x11F9
	COMMAND_COPY_BUFFER_TO_IMAGE CommandType = 0x11FA
	COMMAND_MAP_BUFFER CommandType = 0x11FB
	COMMAND_MAP_IMAGE CommandType = 0x11FC
	COMMAND_UNMAP_MEM_OBJECT CommandType = 0x11FD
	COMMAND_MARKER CommandType = 0x11FE
	COMMAND_ACQUIRE_GL_OBJECTS CommandType = 0x11FF
	COMMAND_RELEASE_GL_OBJECTS CommandType = 0x1200
	COMMAND_READ_BUFFER_RECT CommandType = 0x1201
	COMMAND_WRITE_BUFFER_RECT CommandType = 0x1202
	COMMAND_COPY_BUFFER_RECT CommandType = 0x1203
	COMMAND_USER CommandType = 0x1204
	COMMAND_BARRIER CommandType = 0x1205
	COMMAND_MIGRATE_MEM_OBJECTS CommandType = 0x1206
	COMMAND_FILL_BUFFER CommandType = 0x1207
	COMMAND_FILL_IMAGE CommandType = 0x1208
	COMMAND_SVM_FREE CommandType = 0x1209
	COMMAND_SVM_MEMCPY CommandType = 0x120A
	COMMAND_SVM_MEMFILL CommandType = 0x120B
	COMMAND_SVM_MAP CommandType = 0x120C
	COMMAND_SVM_UNMAP CommandType = 0x120D
	COMMAND_SVM_MIGRATE_MEM CommandType = 0x120E
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // BufferCreateType
	BUFFER_CREATE_TYPE_REGION BufferCreateType = 0x1220
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // ProfilingInfo
	PROFILING_COMMAND_QUEUED ProfilingInfo = 0x1280
	PROFILING_COMMAND_SUBMIT ProfilingInfo = 0x1281
	PROFILING_COMMAND_START ProfilingInfo = 0x1282
	PROFILING_COMMAND_END ProfilingInfo = 0x1283
	PROFILING_COMMAND_COMPLETE ProfilingInfo = 0x1284
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceAtomicCapabilities
	DEVICE_ATOMIC_ORDER_RELAXED DeviceAtomicCapabilities = (1 << 0)
	DEVICE_ATOMIC_ORDER_ACQ_REL DeviceAtomicCapabilities = (1 << 1)
	DEVICE_ATOMIC_ORDER_SEQ_CST DeviceAtomicCapabilities = (1 << 2)
	DEVICE_ATOMIC_SCOPE_WORK_ITEM DeviceAtomicCapabilities = (1 << 3)
	DEVICE_ATOMIC_SCOPE_WORK_GROUP DeviceAtomicCapabilities = (1 << 4)
	DEVICE_ATOMIC_SCOPE_DEVICE DeviceAtomicCapabilities = (1 << 5)
	DEVICE_ATOMIC_SCOPE_ALL_DEVICES DeviceAtomicCapabilities = (1 << 6)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceDeviceEnqueueCapabilities
	DEVICE_QUEUE_SUPPORTED DeviceDeviceEnqueueCapabilities = (1 << 0)
	DEVICE_QUEUE_REPLACEABLE_DEFAULT DeviceDeviceEnqueueCapabilities = (1 << 1)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // KhronosVendorId
	KHRONOS_VENDOR_ID_CODEPLAY KhronosVendorId = 0x10004
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // Version
	VERSION_MAJOR_BITS Version = (10)
	VERSION_MINOR_BITS Version = (10)
	VERSION_PATCH_BITS Version = (12)
)
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/enums.html
const ( // DeviceIntegerDotProductCapabilities
	DEVICE_INTEGER_DOT_PRODUCT_INPUT_4x8BIT_PACKED DeviceIntegerDotProductCapabilities = (1 << 0)
	DEVICE_INTEGER_DOT_PRODUCT_INPUT_4x8BIT DeviceIntegerDotProductCapabilities = (1 << 1)
)
func (v ErrorCode) String() string {
	switch v {
	case SUCCESS: return "SUCCESS"
	case DEVICE_NOT_FOUND: return "DEVICE_NOT_FOUND"
	case DEVICE_NOT_AVAILABLE: return "DEVICE_NOT_AVAILABLE"
	case COMPILER_NOT_AVAILABLE: return "COMPILER_NOT_AVAILABLE"
	case MEM_OBJECT_ALLOCATION_FAILURE: return "MEM_OBJECT_ALLOCATION_FAILURE"
	case OUT_OF_RESOURCES: return "OUT_OF_RESOURCES"
	case OUT_OF_HOST_MEMORY: return "OUT_OF_HOST_MEMORY"
	case PROFILING_INFO_NOT_AVAILABLE: return "PROFILING_INFO_NOT_AVAILABLE"
	case MEM_COPY_OVERLAP: return "MEM_COPY_OVERLAP"
	case IMAGE_FORMAT_MISMATCH: return "IMAGE_FORMAT_MISMATCH"
	case IMAGE_FORMAT_NOT_SUPPORTED: return "IMAGE_FORMAT_NOT_SUPPORTED"
	case BUILD_PROGRAM_FAILURE: return "BUILD_PROGRAM_FAILURE"
	case MAP_FAILURE: return "MAP_FAILURE"
	case MISALIGNED_SUB_BUFFER_OFFSET: return "MISALIGNED_SUB_BUFFER_OFFSET"
	case EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST: return "EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST"
	case COMPILE_PROGRAM_FAILURE: return "COMPILE_PROGRAM_FAILURE"
	case LINKER_NOT_AVAILABLE: return "LINKER_NOT_AVAILABLE"
	case LINK_PROGRAM_FAILURE: return "LINK_PROGRAM_FAILURE"
	case DEVICE_PARTITION_FAILED: return "DEVICE_PARTITION_FAILED"
	case KERNEL_ARG_INFO_NOT_AVAILABLE: return "KERNEL_ARG_INFO_NOT_AVAILABLE"
	case INVALID_VALUE: return "INVALID_VALUE"
	case INVALID_DEVICE_TYPE: return "INVALID_DEVICE_TYPE"
	case INVALID_PLATFORM: return "INVALID_PLATFORM"
	case INVALID_DEVICE: return "INVALID_DEVICE"
	case INVALID_CONTEXT: return "INVALID_CONTEXT"
	case INVALID_QUEUE_PROPERTIES: return "INVALID_QUEUE_PROPERTIES"
	case INVALID_COMMAND_QUEUE: return "INVALID_COMMAND_QUEUE"
	case INVALID_HOST_PTR: return "INVALID_HOST_PTR"
	case INVALID_MEM_OBJECT: return "INVALID_MEM_OBJECT"
	case INVALID_IMAGE_FORMAT_DESCRIPTOR: return "INVALID_IMAGE_FORMAT_DESCRIPTOR"
	case INVALID_IMAGE_SIZE: return "INVALID_IMAGE_SIZE"
	case INVALID_SAMPLER: return "INVALID_SAMPLER"
	case INVALID_BINARY: return "INVALID_BINARY"
	case INVALID_BUILD_OPTIONS: return "INVALID_BUILD_OPTIONS"
	case INVALID_PROGRAM: return "INVALID_PROGRAM"
	case INVALID_PROGRAM_EXECUTABLE: return "INVALID_PROGRAM_EXECUTABLE"
	case INVALID_KERNEL_NAME: return "INVALID_KERNEL_NAME"
	case INVALID_KERNEL_DEFINITION: return "INVALID_KERNEL_DEFINITION"
	case INVALID_KERNEL: return "INVALID_KERNEL"
	case INVALID_ARG_INDEX: return "INVALID_ARG_INDEX"
	case INVALID_ARG_VALUE: return "INVALID_ARG_VALUE"
	case INVALID_ARG_SIZE: return "INVALID_ARG_SIZE"
	case INVALID_KERNEL_ARGS: return "INVALID_KERNEL_ARGS"
	case INVALID_WORK_DIMENSION: return "INVALID_WORK_DIMENSION"
	case INVALID_WORK_GROUP_SIZE: return "INVALID_WORK_GROUP_SIZE"
	case INVALID_WORK_ITEM_SIZE: return "INVALID_WORK_ITEM_SIZE"
	case INVALID_GLOBAL_OFFSET: return "INVALID_GLOBAL_OFFSET"
	case INVALID_EVENT_WAIT_LIST: return "INVALID_EVENT_WAIT_LIST"
	case INVALID_EVENT: return "INVALID_EVENT"
	case INVALID_OPERATION: return "INVALID_OPERATION"
	case INVALID_GL_OBJECT: return "INVALID_GL_OBJECT"
	case INVALID_BUFFER_SIZE: return "INVALID_BUFFER_SIZE"
	case INVALID_MIP_LEVEL: return "INVALID_MIP_LEVEL"
	case INVALID_GLOBAL_WORK_SIZE: return "INVALID_GLOBAL_WORK_SIZE"
	case INVALID_PROPERTY: return "INVALID_PROPERTY"
	case INVALID_IMAGE_DESCRIPTOR: return "INVALID_IMAGE_DESCRIPTOR"
	case INVALID_COMPILER_OPTIONS: return "INVALID_COMPILER_OPTIONS"
	case INVALID_LINKER_OPTIONS: return "INVALID_LINKER_OPTIONS"
	case INVALID_DEVICE_PARTITION_COUNT: return "INVALID_DEVICE_PARTITION_COUNT"
	case INVALID_PIPE_SIZE: return "INVALID_PIPE_SIZE"
	case INVALID_DEVICE_QUEUE: return "INVALID_DEVICE_QUEUE"
	case INVALID_SPEC_ID: return "INVALID_SPEC_ID"
	case MAX_SIZE_RESTRICTION_EXCEEDED: return "MAX_SIZE_RESTRICTION_EXCEEDED"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v PlatformInfo) String() string {
	switch v {
	case PLATFORM_PROFILE: return "PLATFORM_PROFILE"
	case PLATFORM_VERSION: return "PLATFORM_VERSION"
	case PLATFORM_NAME: return "PLATFORM_NAME"
	case PLATFORM_VENDOR: return "PLATFORM_VENDOR"
	case PLATFORM_EXTENSIONS: return "PLATFORM_EXTENSIONS"
	case PLATFORM_HOST_TIMER_RESOLUTION: return "PLATFORM_HOST_TIMER_RESOLUTION"
	case PLATFORM_NUMERIC_VERSION: return "PLATFORM_NUMERIC_VERSION"
	case PLATFORM_EXTENSIONS_WITH_VERSION: return "PLATFORM_EXTENSIONS_WITH_VERSION"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceType) String() string {
	switch v {
	case DEVICE_TYPE_DEFAULT: return "DEVICE_TYPE_DEFAULT"
	case DEVICE_TYPE_CPU: return "DEVICE_TYPE_CPU"
	case DEVICE_TYPE_GPU: return "DEVICE_TYPE_GPU"
	case DEVICE_TYPE_ACCELERATOR: return "DEVICE_TYPE_ACCELERATOR"
	case DEVICE_TYPE_CUSTOM: return "DEVICE_TYPE_CUSTOM"
	case DEVICE_TYPE_ALL: return "DEVICE_TYPE_ALL"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceInfo) String() string {
	switch v {
	case DEVICE_TYPE: return "DEVICE_TYPE"
	case DEVICE_VENDOR_ID: return "DEVICE_VENDOR_ID"
	case DEVICE_MAX_COMPUTE_UNITS: return "DEVICE_MAX_COMPUTE_UNITS"
	case DEVICE_MAX_WORK_ITEM_DIMENSIONS: return "DEVICE_MAX_WORK_ITEM_DIMENSIONS"
	case DEVICE_MAX_WORK_GROUP_SIZE: return "DEVICE_MAX_WORK_GROUP_SIZE"
	case DEVICE_MAX_WORK_GROUP_SIZES: return "DEVICE_MAX_WORK_GROUP_SIZES"
	case DEVICE_PREFERRED_VECTOR_WIDTH_CHAR: return "DEVICE_PREFERRED_VECTOR_WIDTH_CHAR"
	case DEVICE_PREFERRED_VECTOR_WIDTH_SHORT: return "DEVICE_PREFERRED_VECTOR_WIDTH_SHORT"
	case DEVICE_PREFERRED_VECTOR_WIDTH_INT: return "DEVICE_PREFERRED_VECTOR_WIDTH_INT"
	case DEVICE_PREFERRED_VECTOR_WIDTH_LONG: return "DEVICE_PREFERRED_VECTOR_WIDTH_LONG"
	case DEVICE_PREFERRED_VECTOR_WIDTH_FLOAT: return "DEVICE_PREFERRED_VECTOR_WIDTH_FLOAT"
	case DEVICE_PREFERRED_VECTOR_WIDTH_DOUBLE: return "DEVICE_PREFERRED_VECTOR_WIDTH_DOUBLE"
	case DEVICE_MAX_CLOCK_FREQUENCY: return "DEVICE_MAX_CLOCK_FREQUENCY"
	case DEVICE_ADDRESS_BITS: return "DEVICE_ADDRESS_BITS"
	case DEVICE_MAX_READ_IMAGE_ARGS: return "DEVICE_MAX_READ_IMAGE_ARGS"
	case DEVICE_MAX_WRITE_IMAGE_ARGS: return "DEVICE_MAX_WRITE_IMAGE_ARGS"
	case DEVICE_MAX_MEM_ALLOC_SIZE: return "DEVICE_MAX_MEM_ALLOC_SIZE"
	case DEVICE_IMAGE2D_MAX_WIDTH: return "DEVICE_IMAGE2D_MAX_WIDTH"
	case DEVICE_IMAGE2D_MAX_HEIGHT: return "DEVICE_IMAGE2D_MAX_HEIGHT"
	case DEVICE_IMAGE3D_MAX_WIDTH: return "DEVICE_IMAGE3D_MAX_WIDTH"
	case DEVICE_IMAGE3D_MAX_HEIGHT: return "DEVICE_IMAGE3D_MAX_HEIGHT"
	case DEVICE_IMAGE3D_MAX_DEPTH: return "DEVICE_IMAGE3D_MAX_DEPTH"
	case DEVICE_IMAGE_SUPPORT: return "DEVICE_IMAGE_SUPPORT"
	case DEVICE_MAX_PARAMETER_SIZE: return "DEVICE_MAX_PARAMETER_SIZE"
	case DEVICE_MAX_SAMPLERS: return "DEVICE_MAX_SAMPLERS"
	case DEVICE_MEM_BASE_ADDR_ALIGN: return "DEVICE_MEM_BASE_ADDR_ALIGN"
	case DEVICE_MIN_DATA_TYPE_ALIGN_SIZE: return "DEVICE_MIN_DATA_TYPE_ALIGN_SIZE"
	case DEVICE_SINGLE_FP_CONFIG: return "DEVICE_SINGLE_FP_CONFIG"
	case DEVICE_GLOBAL_MEM_CACHE_TYPE: return "DEVICE_GLOBAL_MEM_CACHE_TYPE"
	case DEVICE_GLOBAL_MEM_CACHELINE_SIZE: return "DEVICE_GLOBAL_MEM_CACHELINE_SIZE"
	case DEVICE_GLOBAL_MEM_CACHE_SIZE: return "DEVICE_GLOBAL_MEM_CACHE_SIZE"
	case DEVICE_GLOBAL_MEM_SIZE: return "DEVICE_GLOBAL_MEM_SIZE"
	case DEVICE_MAX_CONSTANT_BUFFER_SIZE: return "DEVICE_MAX_CONSTANT_BUFFER_SIZE"
	case DEVICE_MAX_CONSTANT_ARGS: return "DEVICE_MAX_CONSTANT_ARGS"
	case DEVICE_LOCAL_MEM_TYPE: return "DEVICE_LOCAL_MEM_TYPE"
	case DEVICE_LOCAL_MEM_SIZE: return "DEVICE_LOCAL_MEM_SIZE"
	case DEVICE_ERROR_CORRECTION_SUPPORT: return "DEVICE_ERROR_CORRECTION_SUPPORT"
	case DEVICE_PROFILING_TIMER_RESOLUTION: return "DEVICE_PROFILING_TIMER_RESOLUTION"
	case DEVICE_ENDIAN_LITTLE: return "DEVICE_ENDIAN_LITTLE"
	case DEVICE_AVAILABLE: return "DEVICE_AVAILABLE"
	case DEVICE_COMPILER_AVAILABLE: return "DEVICE_COMPILER_AVAILABLE"
	case DEVICE_EXECUTION_CAPABILITIES: return "DEVICE_EXECUTION_CAPABILITIES"
	case DEVICE_QUEUE_ON_HOST_PROPERTIES: return "DEVICE_QUEUE_ON_HOST_PROPERTIES"
	case DEVICE_NAME: return "DEVICE_NAME"
	case DEVICE_VENDOR: return "DEVICE_VENDOR"
	case DRIVER_VERSION: return "DRIVER_VERSION"
	case DEVICE_PROFILE: return "DEVICE_PROFILE"
	case DEVICE_VERSION: return "DEVICE_VERSION"
	case DEVICE_EXTENSIONS: return "DEVICE_EXTENSIONS"
	case DEVICE_PLATFORM: return "DEVICE_PLATFORM"
	case DEVICE_DOUBLE_FP_CONFIG: return "DEVICE_DOUBLE_FP_CONFIG"
	case DEVICE_PREFERRED_VECTOR_WIDTH_HALF: return "DEVICE_PREFERRED_VECTOR_WIDTH_HALF"
	case DEVICE_HOST_UNIFIED_MEMORY: return "DEVICE_HOST_UNIFIED_MEMORY"
	case DEVICE_NATIVE_VECTOR_WIDTH_CHAR: return "DEVICE_NATIVE_VECTOR_WIDTH_CHAR"
	case DEVICE_NATIVE_VECTOR_WIDTH_SHORT: return "DEVICE_NATIVE_VECTOR_WIDTH_SHORT"
	case DEVICE_NATIVE_VECTOR_WIDTH_INT: return "DEVICE_NATIVE_VECTOR_WIDTH_INT"
	case DEVICE_NATIVE_VECTOR_WIDTH_LONG: return "DEVICE_NATIVE_VECTOR_WIDTH_LONG"
	case DEVICE_NATIVE_VECTOR_WIDTH_FLOAT: return "DEVICE_NATIVE_VECTOR_WIDTH_FLOAT"
	case DEVICE_NATIVE_VECTOR_WIDTH_DOUBLE: return "DEVICE_NATIVE_VECTOR_WIDTH_DOUBLE"
	case DEVICE_NATIVE_VECTOR_WIDTH_HALF: return "DEVICE_NATIVE_VECTOR_WIDTH_HALF"
	case DEVICE_OPENCL_C_VERSION: return "DEVICE_OPENCL_C_VERSION"
	case DEVICE_LINKER_AVAILABLE: return "DEVICE_LINKER_AVAILABLE"
	case DEVICE_BUILT_IN_KERNELS: return "DEVICE_BUILT_IN_KERNELS"
	case DEVICE_IMAGE_MAX_BUFFER_SIZE: return "DEVICE_IMAGE_MAX_BUFFER_SIZE"
	case DEVICE_IMAGE_MAX_ARRAY_SIZE: return "DEVICE_IMAGE_MAX_ARRAY_SIZE"
	case DEVICE_PARENT_DEVICE: return "DEVICE_PARENT_DEVICE"
	case DEVICE_PARTITION_MAX_SUB_DEVICES: return "DEVICE_PARTITION_MAX_SUB_DEVICES"
	case DEVICE_PARTITION_PROPERTIES: return "DEVICE_PARTITION_PROPERTIES"
	case DEVICE_PARTITION_AFFINITY_DOMAIN: return "DEVICE_PARTITION_AFFINITY_DOMAIN"
	case DEVICE_PARTITION_TYPE: return "DEVICE_PARTITION_TYPE"
	case DEVICE_REFERENCE_COUNT: return "DEVICE_REFERENCE_COUNT"
	case DEVICE_PREFERRED_INTEROP_USER_SYNC: return "DEVICE_PREFERRED_INTEROP_USER_SYNC"
	case DEVICE_PRINTF_BUFFER_SIZE: return "DEVICE_PRINTF_BUFFER_SIZE"
	case DEVICE_IMAGE_PITCH_ALIGNMENT: return "DEVICE_IMAGE_PITCH_ALIGNMENT"
	case DEVICE_IMAGE_BASE_ADDRESS_ALIGNMENT: return "DEVICE_IMAGE_BASE_ADDRESS_ALIGNMENT"
	case DEVICE_MAX_READ_WRITE_IMAGE_ARGS: return "DEVICE_MAX_READ_WRITE_IMAGE_ARGS"
	case DEVICE_MAX_GLOBAL_VARIABLE_SIZE: return "DEVICE_MAX_GLOBAL_VARIABLE_SIZE"
	case DEVICE_QUEUE_ON_DEVICE_PROPERTIES: return "DEVICE_QUEUE_ON_DEVICE_PROPERTIES"
	case DEVICE_QUEUE_ON_DEVICE_PREFERRED_SIZE: return "DEVICE_QUEUE_ON_DEVICE_PREFERRED_SIZE"
	case DEVICE_QUEUE_ON_DEVICE_MAX_SIZE: return "DEVICE_QUEUE_ON_DEVICE_MAX_SIZE"
	case DEVICE_MAX_ON_DEVICE_QUEUES: return "DEVICE_MAX_ON_DEVICE_QUEUES"
	case DEVICE_MAX_ON_DEVICE_EVENTS: return "DEVICE_MAX_ON_DEVICE_EVENTS"
	case DEVICE_SVM_CAPABILITIES: return "DEVICE_SVM_CAPABILITIES"
	case DEVICE_GLOBAL_VARIABLE_PREFERRED_TOTAL_SIZE: return "DEVICE_GLOBAL_VARIABLE_PREFERRED_TOTAL_SIZE"
	case DEVICE_MAX_PIPE_ARGS: return "DEVICE_MAX_PIPE_ARGS"
	case DEVICE_PIPE_MAX_ACTIVE_RESERVATIONS: return "DEVICE_PIPE_MAX_ACTIVE_RESERVATIONS"
	case DEVICE_PIPE_MAX_PACKET_SIZE: return "DEVICE_PIPE_MAX_PACKET_SIZE"
	case DEVICE_PREFERRED_PLATFORM_ATOMIC_ALIGNMENT: return "DEVICE_PREFERRED_PLATFORM_ATOMIC_ALIGNMENT"
	case DEVICE_PREFERRED_GLOBAL_ATOMIC_ALIGNMENT: return "DEVICE_PREFERRED_GLOBAL_ATOMIC_ALIGNMENT"
	case DEVICE_PREFERRED_LOCAL_ATOMIC_ALIGNMENT: return "DEVICE_PREFERRED_LOCAL_ATOMIC_ALIGNMENT"
	case DEVICE_IL_VERSION: return "DEVICE_IL_VERSION"
	case DEVICE_MAX_NUM_SUB_GROUPS: return "DEVICE_MAX_NUM_SUB_GROUPS"
	case DEVICE_SUB_GROUP_INDEPENDENT_FORWARD_PROGRESS: return "DEVICE_SUB_GROUP_INDEPENDENT_FORWARD_PROGRESS"
	case DEVICE_NUMERIC_VERSION: return "DEVICE_NUMERIC_VERSION"
	case DEVICE_EXTENSIONS_WITH_VERSION: return "DEVICE_EXTENSIONS_WITH_VERSION"
	case DEVICE_ILS_WITH_VERSION: return "DEVICE_ILS_WITH_VERSION"
	case DEVICE_BUILT_IN_KERNELS_WITH_VERSION: return "DEVICE_BUILT_IN_KERNELS_WITH_VERSION"
	case DEVICE_ATOMIC_MEMORY_CAPABILITIES: return "DEVICE_ATOMIC_MEMORY_CAPABILITIES"
	case DEVICE_ATOMIC_FENCE_CAPABILITIES: return "DEVICE_ATOMIC_FENCE_CAPABILITIES"
	case DEVICE_NON_UNIFORM_WORK_GROUP_SUPPORT: return "DEVICE_NON_UNIFORM_WORK_GROUP_SUPPORT"
	case DEVICE_OPENCL_C_ALL_VERSIONS: return "DEVICE_OPENCL_C_ALL_VERSIONS"
	case DEVICE_PREFERRED_WORK_GROUP_SIZE_MULTIPLE: return "DEVICE_PREFERRED_WORK_GROUP_SIZE_MULTIPLE"
	case DEVICE_WORK_GROUP_COLLECTIVE_FUNCTIONS_SUPPORT: return "DEVICE_WORK_GROUP_COLLECTIVE_FUNCTIONS_SUPPORT"
	case DEVICE_GENERIC_ADDRESS_SPACE_SUPPORT: return "DEVICE_GENERIC_ADDRESS_SPACE_SUPPORT"
	case DEVICE_UUID: return "DEVICE_UUID"
	case DRIVER_UUID: return "DRIVER_UUID"
	case DEVICE_LUID_VALID: return "DEVICE_LUID_VALID"
	case DEVICE_LUID: return "DEVICE_LUID"
	case DEVICE_NODE_MASK: return "DEVICE_NODE_MASK"
	case DEVICE_OPENCL_C_FEATURES: return "DEVICE_OPENCL_C_FEATURES"
	case DEVICE_DEVICE_ENQUEUE_CAPABILITIES: return "DEVICE_DEVICE_ENQUEUE_CAPABILITIES"
	case DEVICE_PIPE_SUPPORT: return "DEVICE_PIPE_SUPPORT"
	case DEVICE_LATEST_CONFORMANCE_VERSION_PASSED: return "DEVICE_LATEST_CONFORMANCE_VERSION_PASSED"
	case DEVICE_INTEGER_DOT_PRODUCT_CAPABILITIES: return "DEVICE_INTEGER_DOT_PRODUCT_CAPABILITIES"
	case DEVICE_INTEGER_DOT_PRODUCT_ACCELERATION_PROPERTIES_8BIT: return "DEVICE_INTEGER_DOT_PRODUCT_ACCELERATION_PROPERTIES_8BIT"
	case DEVICE_INTEGER_DOT_PRODUCT_ACCELERATION_PROPERTIES_4x8BIT_PACKED: return "DEVICE_INTEGER_DOT_PRODUCT_ACCELERATION_PROPERTIES_4x8BIT_PACKED"
	case DEVICE_SPIRV_EXTENDED_INSTRUCTION_SETS: return "DEVICE_SPIRV_EXTENDED_INSTRUCTION_SETS"
	case DEVICE_SPIRV_EXTENSIONS: return "DEVICE_SPIRV_EXTENSIONS"
	case DEVICE_SPIRV_CAPABILITIES: return "DEVICE_SPIRV_CAPABILITIES"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceFpConfig) String() string {
	switch v {
	case FP_DENORM: return "FP_DENORM"
	case FP_INF_NAN: return "FP_INF_NAN"
	case FP_ROUND_TO_NEAREST: return "FP_ROUND_TO_NEAREST"
	case FP_ROUND_TO_ZERO: return "FP_ROUND_TO_ZERO"
	case FP_ROUND_TO_INF: return "FP_ROUND_TO_INF"
	case FP_FMA: return "FP_FMA"
	case FP_SOFT_FLOAT: return "FP_SOFT_FLOAT"
	case FP_CORRECTLY_ROUNDED_DIVIDE_SQRT: return "FP_CORRECTLY_ROUNDED_DIVIDE_SQRT"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceMemCacheType) String() string {
	switch v {
	case NONE: return "NONE"
	case READ_ONLY_CACHE: return "READ_ONLY_CACHE"
	case READ_WRITE_CACHE: return "READ_WRITE_CACHE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceLocalMemType) String() string {
	switch v {
	case LOCAL: return "LOCAL"
	case GLOBAL: return "GLOBAL"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceExecCapabilities) String() string {
	switch v {
	case EXEC_KERNEL: return "EXEC_KERNEL"
	case EXEC_NATIVE_KERNEL: return "EXEC_NATIVE_KERNEL"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v CommandQueueProperties) String() string {
	switch v {
	case QUEUE_OUT_OF_ORDER_EXEC_MODE_ENABLE: return "QUEUE_OUT_OF_ORDER_EXEC_MODE_ENABLE"
	case QUEUE_PROFILING_ENABLE: return "QUEUE_PROFILING_ENABLE"
	case QUEUE_ON_DEVICE: return "QUEUE_ON_DEVICE"
	case QUEUE_ON_DEVICE_DEFAULT: return "QUEUE_ON_DEVICE_DEFAULT"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v ContextInfo) String() string {
	switch v {
	case CONTEXT_REFERENCE_COUNT: return "CONTEXT_REFERENCE_COUNT"
	case CONTEXT_DEVICES: return "CONTEXT_DEVICES"
	case CONTEXT_PROPERTIES: return "CONTEXT_PROPERTIES"
	case CONTEXT_NUM_DEVICES: return "CONTEXT_NUM_DEVICES"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v ContextProperties) String() string {
	switch v {
	case CONTEXT_PLATFORM: return "CONTEXT_PLATFORM"
	case CONTEXT_INTEROP_USER_SYNC: return "CONTEXT_INTEROP_USER_SYNC"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DevicePartitionProperty) String() string {
	switch v {
	case DEVICE_PARTITION_EQUALLY: return "DEVICE_PARTITION_EQUALLY"
	case DEVICE_PARTITION_BY_COUNTS: return "DEVICE_PARTITION_BY_COUNTS"
	case DEVICE_PARTITION_BY_COUNTS_LIST_END: return "DEVICE_PARTITION_BY_COUNTS_LIST_END"
	case DEVICE_PARTITION_BY_AFFINITY_DOMAIN: return "DEVICE_PARTITION_BY_AFFINITY_DOMAIN"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceAffinityDomain) String() string {
	switch v {
	case DEVICE_AFFINITY_DOMAIN_NUMA: return "DEVICE_AFFINITY_DOMAIN_NUMA"
	case DEVICE_AFFINITY_DOMAIN_L4_CACHE: return "DEVICE_AFFINITY_DOMAIN_L4_CACHE"
	case DEVICE_AFFINITY_DOMAIN_L3_CACHE: return "DEVICE_AFFINITY_DOMAIN_L3_CACHE"
	case DEVICE_AFFINITY_DOMAIN_L2_CACHE: return "DEVICE_AFFINITY_DOMAIN_L2_CACHE"
	case DEVICE_AFFINITY_DOMAIN_L1_CACHE: return "DEVICE_AFFINITY_DOMAIN_L1_CACHE"
	case DEVICE_AFFINITY_DOMAIN_NEXT_PARTITIONABLE: return "DEVICE_AFFINITY_DOMAIN_NEXT_PARTITIONABLE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceSvmCapabilities) String() string {
	switch v {
	case DEVICE_SVM_COARSE_GRAIN_BUFFER: return "DEVICE_SVM_COARSE_GRAIN_BUFFER"
	case DEVICE_SVM_FINE_GRAIN_BUFFER: return "DEVICE_SVM_FINE_GRAIN_BUFFER"
	case DEVICE_SVM_FINE_GRAIN_SYSTEM: return "DEVICE_SVM_FINE_GRAIN_SYSTEM"
	case DEVICE_SVM_ATOMICS: return "DEVICE_SVM_ATOMICS"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v CommandQueueInfo) String() string {
	switch v {
	case QUEUE_CONTEXT: return "QUEUE_CONTEXT"
	case QUEUE_DEVICE: return "QUEUE_DEVICE"
	case QUEUE_REFERENCE_COUNT: return "QUEUE_REFERENCE_COUNT"
	case QUEUE_PROPERTIES: return "QUEUE_PROPERTIES"
	case QUEUE_SIZE: return "QUEUE_SIZE"
	case QUEUE_DEVICE_DEFAULT: return "QUEUE_DEVICE_DEFAULT"
	case QUEUE_PROPERTIES_ARRAY: return "QUEUE_PROPERTIES_ARRAY"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v MemFlags) String() string {
	var s strings.Builder
	if v&MEM_READ_WRITE != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_READ_WRITE") }
	if v&MEM_WRITE_ONLY != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_WRITE_ONLY") }
	if v&MEM_READ_ONLY != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_READ_ONLY") }
	if v&MEM_USE_HOST_PTR != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_USE_HOST_PTR") }
	if v&MEM_ALLOC_HOST_PTR != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_ALLOC_HOST_PTR") }
	if v&MEM_COPY_HOST_PTR != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_COPY_HOST_PTR") }
	if v&MEM_HOST_WRITE_ONLY != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_HOST_WRITE_ONLY") }
	if v&MEM_HOST_READ_ONLY != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_HOST_READ_ONLY") }
	if v&MEM_HOST_NO_ACCESS != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_HOST_NO_ACCESS") }
	if v&MEM_SVM_FINE_GRAIN_BUFFER != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_SVM_FINE_GRAIN_BUFFER") }
	if v&MEM_SVM_ATOMICS != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_SVM_ATOMICS") }
	if v&MEM_KERNEL_READ_AND_WRITE != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MEM_KERNEL_READ_AND_WRITE") }
	return s.String()
}
func (v MemMigrationFlags) String() string {
	var s strings.Builder
	if v&MIGRATE_MEM_OBJECT_HOST != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MIGRATE_MEM_OBJECT_HOST") }
	if v&MIGRATE_MEM_OBJECT_CONTENT_UNDEFINED != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MIGRATE_MEM_OBJECT_CONTENT_UNDEFINED") }
	return s.String()
}
func (v ChannelOrder) String() string {
	switch v {
	case R: return "R"
	case A: return "A"
	case RG: return "RG"
	case RA: return "RA"
	case RGB: return "RGB"
	case RGBA: return "RGBA"
	case BGRA: return "BGRA"
	case ARGB: return "ARGB"
	case INTENSITY: return "INTENSITY"
	case LUMINANCE: return "LUMINANCE"
	case Rx: return "Rx"
	case RGx: return "RGx"
	case RGBx: return "RGBx"
	case DEPTH: return "DEPTH"
	case sRGB: return "sRGB"
	case sRGBx: return "sRGBx"
	case sRGBA: return "sRGBA"
	case sBGRA: return "sBGRA"
	case ABGR: return "ABGR"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v ChannelType) String() string {
	switch v {
	case SNORM_INT8: return "SNORM_INT8"
	case SNORM_INT16: return "SNORM_INT16"
	case UNORM_INT8: return "UNORM_INT8"
	case UNORM_INT16: return "UNORM_INT16"
	case UNORM_SHORT_565: return "UNORM_SHORT_565"
	case UNORM_SHORT_555: return "UNORM_SHORT_555"
	case UNORM_INT_101010: return "UNORM_INT_101010"
	case SIGNED_INT8: return "SIGNED_INT8"
	case SIGNED_INT16: return "SIGNED_INT16"
	case SIGNED_INT32: return "SIGNED_INT32"
	case UNSIGNED_INT8: return "UNSIGNED_INT8"
	case UNSIGNED_INT16: return "UNSIGNED_INT16"
	case UNSIGNED_INT32: return "UNSIGNED_INT32"
	case HALF_FLOAT: return "HALF_FLOAT"
	case FLOAT: return "FLOAT"
	case UNORM_INT_101010_2: return "UNORM_INT_101010_2"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v MemObjectType) String() string {
	switch v {
	case MEM_OBJECT_BUFFER: return "MEM_OBJECT_BUFFER"
	case MEM_OBJECT_IMAGE2D: return "MEM_OBJECT_IMAGE2D"
	case MEM_OBJECT_IMAGE3D: return "MEM_OBJECT_IMAGE3D"
	case MEM_OBJECT_IMAGE2D_ARRAY: return "MEM_OBJECT_IMAGE2D_ARRAY"
	case MEM_OBJECT_IMAGE1D: return "MEM_OBJECT_IMAGE1D"
	case MEM_OBJECT_IMAGE1D_ARRAY: return "MEM_OBJECT_IMAGE1D_ARRAY"
	case MEM_OBJECT_IMAGE1D_BUFFER: return "MEM_OBJECT_IMAGE1D_BUFFER"
	case MEM_OBJECT_PIPE: return "MEM_OBJECT_PIPE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v MemInfo) String() string {
	switch v {
	case MEM_TYPE: return "MEM_TYPE"
	case MEM_FLAGS: return "MEM_FLAGS"
	case MEM_SIZE: return "MEM_SIZE"
	case MEM_HOST_PTR: return "MEM_HOST_PTR"
	case MEM_MAP_COUNT: return "MEM_MAP_COUNT"
	case MEM_REFERENCE_COUNT: return "MEM_REFERENCE_COUNT"
	case MEM_CONTEXT: return "MEM_CONTEXT"
	case MEM_ASSOCIATED_MEMOBJECT: return "MEM_ASSOCIATED_MEMOBJECT"
	case MEM_OFFSET: return "MEM_OFFSET"
	case MEM_USES_SVM_POINTER: return "MEM_USES_SVM_POINTER"
	case MEM_PROPERTIES: return "MEM_PROPERTIES"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v ImageInfo) String() string {
	switch v {
	case IMAGE_FORMAT: return "IMAGE_FORMAT"
	case IMAGE_ELEMENT_SIZE: return "IMAGE_ELEMENT_SIZE"
	case IMAGE_ROW_PITCH: return "IMAGE_ROW_PITCH"
	case IMAGE_SLICE_PITCH: return "IMAGE_SLICE_PITCH"
	case IMAGE_WIDTH: return "IMAGE_WIDTH"
	case IMAGE_HEIGHT: return "IMAGE_HEIGHT"
	case IMAGE_DEPTH: return "IMAGE_DEPTH"
	case IMAGE_ARRAY_SIZE: return "IMAGE_ARRAY_SIZE"
	case IMAGE_BUFFER: return "IMAGE_BUFFER"
	case IMAGE_NUM_MIP_LEVELS: return "IMAGE_NUM_MIP_LEVELS"
	case IMAGE_NUM_SAMPLES: return "IMAGE_NUM_SAMPLES"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v PipeInfo) String() string {
	switch v {
	case PIPE_PACKET_SIZE: return "PIPE_PACKET_SIZE"
	case PIPE_MAX_PACKETS: return "PIPE_MAX_PACKETS"
	case PIPE_PROPERTIES: return "PIPE_PROPERTIES"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v AddressingMode) String() string {
	switch v {
	case ADDRESS_NONE: return "ADDRESS_NONE"
	case ADDRESS_CLAMP_TO_EDGE: return "ADDRESS_CLAMP_TO_EDGE"
	case ADDRESS_CLAMP: return "ADDRESS_CLAMP"
	case ADDRESS_REPEAT: return "ADDRESS_REPEAT"
	case ADDRESS_MIRRORED_REPEAT: return "ADDRESS_MIRRORED_REPEAT"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v FilterMode) String() string {
	switch v {
	case FILTER_NEAREST: return "FILTER_NEAREST"
	case FILTER_LINEAR: return "FILTER_LINEAR"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v SamplerInfo) String() string {
	switch v {
	case SAMPLER_REFERENCE_COUNT: return "SAMPLER_REFERENCE_COUNT"
	case SAMPLER_CONTEXT: return "SAMPLER_CONTEXT"
	case SAMPLER_NORMALIZED_COORDS: return "SAMPLER_NORMALIZED_COORDS"
	case SAMPLER_ADDRESSING_MODE: return "SAMPLER_ADDRESSING_MODE"
	case SAMPLER_FILTER_MODE: return "SAMPLER_FILTER_MODE"
	case SAMPLER_MIP_FILTER_MODE: return "SAMPLER_MIP_FILTER_MODE"
	case SAMPLER_LOD_MIN: return "SAMPLER_LOD_MIN"
	case SAMPLER_LOD_MAX: return "SAMPLER_LOD_MAX"
	case SAMPLER_PROPERTIES: return "SAMPLER_PROPERTIES"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v MapFlags) String() string {
	var s strings.Builder
	if v&MAP_READ != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MAP_READ") }
	if v&MAP_WRITE != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MAP_WRITE") }
	if v&MAP_WRITE_INVALIDATE_REGION != 0 { if s.Len() != 0 { s.WriteByte('|') }; s.WriteString("MAP_WRITE_INVALIDATE_REGION") }
	return s.String()
}
func (v ProgramInfo) String() string {
	switch v {
	case PROGRAM_REFERENCE_COUNT: return "PROGRAM_REFERENCE_COUNT"
	case PROGRAM_CONTEXT: return "PROGRAM_CONTEXT"
	case PROGRAM_NUM_DEVICES: return "PROGRAM_NUM_DEVICES"
	case PROGRAM_DEVICES: return "PROGRAM_DEVICES"
	case PROGRAM_SOURCE: return "PROGRAM_SOURCE"
	case PROGRAM_BINARY_SIZES: return "PROGRAM_BINARY_SIZES"
	case PROGRAM_BINARIES: return "PROGRAM_BINARIES"
	case PROGRAM_NUM_KERNELS: return "PROGRAM_NUM_KERNELS"
	case PROGRAM_KERNEL_NAMES: return "PROGRAM_KERNEL_NAMES"
	case PROGRAM_IL: return "PROGRAM_IL"
	case PROGRAM_SCOPE_GLOBAL_CTORS_PRESENT: return "PROGRAM_SCOPE_GLOBAL_CTORS_PRESENT"
	case PROGRAM_SCOPE_GLOBAL_DTORS_PRESENT: return "PROGRAM_SCOPE_GLOBAL_DTORS_PRESENT"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v ProgramBuildInfo) String() string {
	switch v {
	case PROGRAM_BUILD_STATUS: return "PROGRAM_BUILD_STATUS"
	case PROGRAM_BUILD_OPTIONS: return "PROGRAM_BUILD_OPTIONS"
	case PROGRAM_BUILD_LOG: return "PROGRAM_BUILD_LOG"
	case PROGRAM_BINARY_TYPE: return "PROGRAM_BINARY_TYPE"
	case PROGRAM_BUILD_GLOBAL_VARIABLE_TOTAL_SIZE: return "PROGRAM_BUILD_GLOBAL_VARIABLE_TOTAL_SIZE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v ProgramBinaryType) String() string {
	switch v {
	case PROGRAM_BINARY_TYPE_NONE: return "PROGRAM_BINARY_TYPE_NONE"
	case PROGRAM_BINARY_TYPE_COMPILED_OBJECT: return "PROGRAM_BINARY_TYPE_COMPILED_OBJECT"
	case PROGRAM_BINARY_TYPE_LIBRARY: return "PROGRAM_BINARY_TYPE_LIBRARY"
	case PROGRAM_BINARY_TYPE_EXECUTABLE: return "PROGRAM_BINARY_TYPE_EXECUTABLE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v BuildStatus) String() string {
	switch v {
	case BUILD_SUCCESS: return "BUILD_SUCCESS"
	case BUILD_NONE: return "BUILD_NONE"
	case BUILD_ERROR: return "BUILD_ERROR"
	case BUILD_IN_PROGRESS: return "BUILD_IN_PROGRESS"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KernelInfo) String() string {
	switch v {
	case KERNEL_FUNCTION_NAME: return "KERNEL_FUNCTION_NAME"
	case KERNEL_NUM_ARGS: return "KERNEL_NUM_ARGS"
	case KERNEL_REFERENCE_COUNT: return "KERNEL_REFERENCE_COUNT"
	case KERNEL_CONTEXT: return "KERNEL_CONTEXT"
	case KERNEL_PROGRAM: return "KERNEL_PROGRAM"
	case KERNEL_ATTRIBUTES: return "KERNEL_ATTRIBUTES"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KernelArgInfo) String() string {
	switch v {
	case KERNEL_ARG_ADDRESS_QUALIFIER: return "KERNEL_ARG_ADDRESS_QUALIFIER"
	case KERNEL_ARG_ACCESS_QUALIFIER: return "KERNEL_ARG_ACCESS_QUALIFIER"
	case KERNEL_ARG_TYPE_NAME: return "KERNEL_ARG_TYPE_NAME"
	case KERNEL_ARG_TYPE_QUALIFIER: return "KERNEL_ARG_TYPE_QUALIFIER"
	case KERNEL_ARG_NAME: return "KERNEL_ARG_NAME"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KernelArgAddressQualifier) String() string {
	switch v {
	case KERNEL_ARG_ADDRESS_GLOBAL: return "KERNEL_ARG_ADDRESS_GLOBAL"
	case KERNEL_ARG_ADDRESS_LOCAL: return "KERNEL_ARG_ADDRESS_LOCAL"
	case KERNEL_ARG_ADDRESS_CONSTANT: return "KERNEL_ARG_ADDRESS_CONSTANT"
	case KERNEL_ARG_ADDRESS_PRIVATE: return "KERNEL_ARG_ADDRESS_PRIVATE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KernelArgAccessQualifier) String() string {
	switch v {
	case KERNEL_ARG_ACCESS_READ_ONLY: return "KERNEL_ARG_ACCESS_READ_ONLY"
	case KERNEL_ARG_ACCESS_WRITE_ONLY: return "KERNEL_ARG_ACCESS_WRITE_ONLY"
	case KERNEL_ARG_ACCESS_READ_WRITE: return "KERNEL_ARG_ACCESS_READ_WRITE"
	case KERNEL_ARG_ACCESS_NONE: return "KERNEL_ARG_ACCESS_NONE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KernelArgTypeQualifier) String() string {
	switch v {
	case KERNEL_ARG_TYPE_NONE: return "KERNEL_ARG_TYPE_NONE"
	case KERNEL_ARG_TYPE_CONST: return "KERNEL_ARG_TYPE_CONST"
	case KERNEL_ARG_TYPE_RESTRICT: return "KERNEL_ARG_TYPE_RESTRICT"
	case KERNEL_ARG_TYPE_VOLATILE: return "KERNEL_ARG_TYPE_VOLATILE"
	case KERNEL_ARG_TYPE_PIPE: return "KERNEL_ARG_TYPE_PIPE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KernelWorkGroupInfo) String() string {
	switch v {
	case KERNEL_WORK_GROUP_SIZE: return "KERNEL_WORK_GROUP_SIZE"
	case KERNEL_COMPILE_WORK_GROUP_SIZE: return "KERNEL_COMPILE_WORK_GROUP_SIZE"
	case KERNEL_LOCAL_MEM_SIZE: return "KERNEL_LOCAL_MEM_SIZE"
	case KERNEL_PREFERRED_WORK_GROUP_SIZE_MULTIPLE: return "KERNEL_PREFERRED_WORK_GROUP_SIZE_MULTIPLE"
	case KERNEL_PRIVATE_MEM_SIZE: return "KERNEL_PRIVATE_MEM_SIZE"
	case KERNEL_GLOBAL_WORK_SIZE: return "KERNEL_GLOBAL_WORK_SIZE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KernelSubGroupInfo) String() string {
	switch v {
	case KERNEL_MAX_SUB_GROUP_SIZE_FOR_NDRANGE: return "KERNEL_MAX_SUB_GROUP_SIZE_FOR_NDRANGE"
	case KERNEL_SUB_GROUP_COUNT_FOR_NDRANGE: return "KERNEL_SUB_GROUP_COUNT_FOR_NDRANGE"
	case KERNEL_LOCAL_SIZE_FOR_SUB_GROUP_COUNT: return "KERNEL_LOCAL_SIZE_FOR_SUB_GROUP_COUNT"
	case KERNEL_MAX_NUM_SUB_GROUPS: return "KERNEL_MAX_NUM_SUB_GROUPS"
	case KERNEL_COMPILE_NUM_SUB_GROUPS: return "KERNEL_COMPILE_NUM_SUB_GROUPS"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KernelExecInfo) String() string {
	switch v {
	case KERNEL_EXEC_INFO_SVM_PTRS: return "KERNEL_EXEC_INFO_SVM_PTRS"
	case KERNEL_EXEC_INFO_SVM_FINE_GRAIN_SYSTEM: return "KERNEL_EXEC_INFO_SVM_FINE_GRAIN_SYSTEM"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v EventInfo) String() string {
	switch v {
	case EVENT_COMMAND_QUEUE: return "EVENT_COMMAND_QUEUE"
	case EVENT_COMMAND_TYPE: return "EVENT_COMMAND_TYPE"
	case EVENT_REFERENCE_COUNT: return "EVENT_REFERENCE_COUNT"
	case EVENT_COMMAND_EXECUTION_STATUS: return "EVENT_COMMAND_EXECUTION_STATUS"
	case EVENT_CONTEXT: return "EVENT_CONTEXT"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v CommandType) String() string {
	switch v {
	case COMMAND_NDRANGE_KERNEL: return "COMMAND_NDRANGE_KERNEL"
	case COMMAND_TASK: return "COMMAND_TASK"
	case COMMAND_NATIVE_KERNEL: return "COMMAND_NATIVE_KERNEL"
	case COMMAND_READ_BUFFER: return "COMMAND_READ_BUFFER"
	case COMMAND_WRITE_BUFFER: return "COMMAND_WRITE_BUFFER"
	case COMMAND_COPY_BUFFER: return "COMMAND_COPY_BUFFER"
	case COMMAND_READ_IMAGE: return "COMMAND_READ_IMAGE"
	case COMMAND_WRITE_IMAGE: return "COMMAND_WRITE_IMAGE"
	case COMMAND_COPY_IMAGE: return "COMMAND_COPY_IMAGE"
	case COMMAND_COPY_IMAGE_TO_BUFFER: return "COMMAND_COPY_IMAGE_TO_BUFFER"
	case COMMAND_COPY_BUFFER_TO_IMAGE: return "COMMAND_COPY_BUFFER_TO_IMAGE"
	case COMMAND_MAP_BUFFER: return "COMMAND_MAP_BUFFER"
	case COMMAND_MAP_IMAGE: return "COMMAND_MAP_IMAGE"
	case COMMAND_UNMAP_MEM_OBJECT: return "COMMAND_UNMAP_MEM_OBJECT"
	case COMMAND_MARKER: return "COMMAND_MARKER"
	case COMMAND_ACQUIRE_GL_OBJECTS: return "COMMAND_ACQUIRE_GL_OBJECTS"
	case COMMAND_RELEASE_GL_OBJECTS: return "COMMAND_RELEASE_GL_OBJECTS"
	case COMMAND_READ_BUFFER_RECT: return "COMMAND_READ_BUFFER_RECT"
	case COMMAND_WRITE_BUFFER_RECT: return "COMMAND_WRITE_BUFFER_RECT"
	case COMMAND_COPY_BUFFER_RECT: return "COMMAND_COPY_BUFFER_RECT"
	case COMMAND_USER: return "COMMAND_USER"
	case COMMAND_BARRIER: return "COMMAND_BARRIER"
	case COMMAND_MIGRATE_MEM_OBJECTS: return "COMMAND_MIGRATE_MEM_OBJECTS"
	case COMMAND_FILL_BUFFER: return "COMMAND_FILL_BUFFER"
	case COMMAND_FILL_IMAGE: return "COMMAND_FILL_IMAGE"
	case COMMAND_SVM_FREE: return "COMMAND_SVM_FREE"
	case COMMAND_SVM_MEMCPY: return "COMMAND_SVM_MEMCPY"
	case COMMAND_SVM_MEMFILL: return "COMMAND_SVM_MEMFILL"
	case COMMAND_SVM_MAP: return "COMMAND_SVM_MAP"
	case COMMAND_SVM_UNMAP: return "COMMAND_SVM_UNMAP"
	case COMMAND_SVM_MIGRATE_MEM: return "COMMAND_SVM_MIGRATE_MEM"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v BufferCreateType) String() string {
	switch v {
	case BUFFER_CREATE_TYPE_REGION: return "BUFFER_CREATE_TYPE_REGION"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v ProfilingInfo) String() string {
	switch v {
	case PROFILING_COMMAND_QUEUED: return "PROFILING_COMMAND_QUEUED"
	case PROFILING_COMMAND_SUBMIT: return "PROFILING_COMMAND_SUBMIT"
	case PROFILING_COMMAND_START: return "PROFILING_COMMAND_START"
	case PROFILING_COMMAND_END: return "PROFILING_COMMAND_END"
	case PROFILING_COMMAND_COMPLETE: return "PROFILING_COMMAND_COMPLETE"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceAtomicCapabilities) String() string {
	switch v {
	case DEVICE_ATOMIC_ORDER_RELAXED: return "DEVICE_ATOMIC_ORDER_RELAXED"
	case DEVICE_ATOMIC_ORDER_ACQ_REL: return "DEVICE_ATOMIC_ORDER_ACQ_REL"
	case DEVICE_ATOMIC_ORDER_SEQ_CST: return "DEVICE_ATOMIC_ORDER_SEQ_CST"
	case DEVICE_ATOMIC_SCOPE_WORK_ITEM: return "DEVICE_ATOMIC_SCOPE_WORK_ITEM"
	case DEVICE_ATOMIC_SCOPE_WORK_GROUP: return "DEVICE_ATOMIC_SCOPE_WORK_GROUP"
	case DEVICE_ATOMIC_SCOPE_DEVICE: return "DEVICE_ATOMIC_SCOPE_DEVICE"
	case DEVICE_ATOMIC_SCOPE_ALL_DEVICES: return "DEVICE_ATOMIC_SCOPE_ALL_DEVICES"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceDeviceEnqueueCapabilities) String() string {
	switch v {
	case DEVICE_QUEUE_SUPPORTED: return "DEVICE_QUEUE_SUPPORTED"
	case DEVICE_QUEUE_REPLACEABLE_DEFAULT: return "DEVICE_QUEUE_REPLACEABLE_DEFAULT"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v KhronosVendorId) String() string {
	switch v {
	case KHRONOS_VENDOR_ID_CODEPLAY: return "KHRONOS_VENDOR_ID_CODEPLAY"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}
func (v DeviceIntegerDotProductCapabilities) String() string {
	switch v {
	case DEVICE_INTEGER_DOT_PRODUCT_INPUT_4x8BIT_PACKED: return "DEVICE_INTEGER_DOT_PRODUCT_INPUT_4x8BIT_PACKED"
	case DEVICE_INTEGER_DOT_PRODUCT_INPUT_4x8BIT: return "DEVICE_INTEGER_DOT_PRODUCT_INPUT_4x8BIT"
	default: return fmt.Sprintf("UNKNOWN (%d)", v)
	}
}

// Functions

// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetPlatformIDs.html
func GetPlatformIDs() (_platforms []PlatformId, _err error) {
	var platforms_actual_len C.cl_uint
	C.clGetPlatformIDs(0, nil, &platforms_actual_len)
	platforms_1 := make([]PlatformId, platforms_actual_len)
	platforms, num_entries, platforms_fin := sliceToC(platforms_1)
	defer platforms_fin()
	res := C.clGetPlatformIDs(C.cl_uint(num_entries), (*C.cl_platform_id)(platforms), nil)
	return platforms_1, makeError(ErrorCode(res))
}
// Simplified binding for clGetPlatformInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetPlatformInfo.html
func GetPlatformInfo(platform PlatformId, param_name PlatformInfo, param_value any) (_err error) {
	platform_1 := C.cl_platform_id(platform)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetPlatformInfo(platform_1, C.cl_platform_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetPlatformInfo(platform_1, C.cl_platform_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetDeviceIDs.html
func GetDeviceIDs(platform PlatformId, device_type DeviceType) (_devices []DeviceId, _err error) {
	platform_1 := C.cl_platform_id(platform)
	device_type_1 := C.cl_device_type(device_type)
	var devices_actual_len C.cl_uint
	C.clGetDeviceIDs(platform_1, device_type_1, 0, nil, &devices_actual_len)
	devices_1 := make([]DeviceId, devices_actual_len)
	devices, num_entries, devices_fin := sliceToC(devices_1)
	defer devices_fin()
	res := C.clGetDeviceIDs(platform_1, device_type_1, C.cl_uint(num_entries), (*C.cl_device_id)(devices), nil)
	return devices_1, makeError(ErrorCode(res))
}
// Simplified binding for clGetDeviceInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetDeviceInfo.html
func GetDeviceInfo(device DeviceId, param_name DeviceInfo, param_value any) (_err error) {
	device_1 := C.cl_device_id(device)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetDeviceInfo(device_1, C.cl_device_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetDeviceInfo(device_1, C.cl_device_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateSubDevices.html
func CreateSubDevices(in_device DeviceId, properties *DevicePartitionProperty, out_devices []DeviceId, num_devices_ret *uint32) (_err error) {
	out_devices_1, num_devices_1, out_devices_fin := sliceToC(out_devices)
	defer out_devices_fin()
	in_device_1 := C.cl_device_id(in_device)
	properties_1 := (*C.cl_device_partition_property)(properties)
	num_devices_2 := C.cl_uint(num_devices_1)
	out_devices_2 := (*C.cl_device_id)(out_devices_1)
	num_devices_ret_1 := (*C.cl_uint)(num_devices_ret)
	res := C.clCreateSubDevices(in_device_1, properties_1, num_devices_2, out_devices_2, num_devices_ret_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clRetainDevice.html
func RetainDevice(device DeviceId) (_err error) {
	device_1 := C.cl_device_id(device)
	res := C.clRetainDevice(device_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clReleaseDevice.html
func ReleaseDevice(device DeviceId) (_err error) {
	device_1 := C.cl_device_id(device)
	res := C.clReleaseDevice(device_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetDefaultDeviceCommandQueue.html
func SetDefaultDeviceCommandQueue(context Context, device DeviceId, command_queue CommandQueue) (_err error) {
	context_1 := C.cl_context(context)
	device_1 := C.cl_device_id(device)
	command_queue_1 := C.cl_command_queue(command_queue)
	res := C.clSetDefaultDeviceCommandQueue(context_1, device_1, command_queue_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetDeviceAndHostTimer.html
func GetDeviceAndHostTimer(device DeviceId) (_device_timestamp uint64, _host_timestamp uint64, _err error) {
	var device_timestamp_1 C.cl_ulong
	var host_timestamp_1 C.cl_ulong
	device_1 := C.cl_device_id(device)
	res := C.clGetDeviceAndHostTimer(device_1, &device_timestamp_1, &host_timestamp_1)
	device_timestamp_2 := uint64(device_timestamp_1)
	host_timestamp_2 := uint64(host_timestamp_1)
	return device_timestamp_2, host_timestamp_2, makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetHostTimer.html
func GetHostTimer(device DeviceId) (_host_timestamp uint64, _err error) {
	var host_timestamp_1 C.cl_ulong
	device_1 := C.cl_device_id(device)
	res := C.clGetHostTimer(device_1, &host_timestamp_1)
	host_timestamp_2 := uint64(host_timestamp_1)
	return host_timestamp_2, makeError(ErrorCode(res))
}
//export go_cl_callback_clCreateContext
func go_cl_callback_clCreateContext(errinfo *C.char, private_info *C.void, cb C.size_t, user_data *C.void) {
	errinfo_1 := (*int8)(errinfo)
	private_info_1 := (unsafe.Pointer)(private_info)
	cb_1 := uint64(cb)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(errinfo *int8, private_info unsafe.Pointer, cb uint64)))(errinfo_1, private_info_1, cb_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateContext.html
func CreateContext(properties []ContextProperties, devices []DeviceId, pfn_notify func(errinfo *int8, private_info unsafe.Pointer, cb uint64)) (_res Context, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	properties_1, properties_fin := sliceToCZeroTerm(properties)
	defer properties_fin()
	devices_1, num_devices_1, devices_fin := sliceToC(devices)
	defer devices_fin()
	properties_2 := (*C.cl_context_properties)(properties_1)
	num_devices_2 := C.cl_uint(num_devices_1)
	devices_2 := (*C.cl_device_id)(devices_1)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clCreateContext)
	}
	res := C.clCreateContext(properties_2, num_devices_2, devices_2, callback, callback_uid, &errcode_ret_1)
	res_1 := Context(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
//export go_cl_callback_clCreateContextFromType
func go_cl_callback_clCreateContextFromType(errinfo *C.char, private_info *C.void, cb C.size_t, user_data *C.void) {
	errinfo_1 := (*int8)(errinfo)
	private_info_1 := (unsafe.Pointer)(private_info)
	cb_1 := uint64(cb)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(errinfo *int8, private_info unsafe.Pointer, cb uint64)))(errinfo_1, private_info_1, cb_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateContextFromType.html
func CreateContextFromType(properties []ContextProperties, device_type DeviceType, pfn_notify func(errinfo *int8, private_info unsafe.Pointer, cb uint64)) (_res Context, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	properties_1, properties_fin := sliceToCZeroTerm(properties)
	defer properties_fin()
	properties_2 := (*C.cl_context_properties)(properties_1)
	device_type_1 := C.cl_device_type(device_type)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clCreateContextFromType)
	}
	res := C.clCreateContextFromType(properties_2, device_type_1, callback, callback_uid, &errcode_ret_1)
	res_1 := Context(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clRetainContext.html
func RetainContext(context Context) (_err error) {
	context_1 := C.cl_context(context)
	res := C.clRetainContext(context_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clReleaseContext.html
func ReleaseContext(context Context) (_err error) {
	context_1 := C.cl_context(context)
	res := C.clReleaseContext(context_1)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetContextInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetContextInfo.html
func GetContextInfo(context Context, param_name ContextInfo, param_value any) (_err error) {
	context_1 := C.cl_context(context)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetContextInfo(context_1, C.cl_context_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetContextInfo(context_1, C.cl_context_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
//export go_cl_callback_clSetContextDestructorCallback
func go_cl_callback_clSetContextDestructorCallback(context C.cl_context, user_data *C.void) {
	context_1 := Context(context)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(context Context)))(context_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetContextDestructorCallback.html
func SetContextDestructorCallback(context Context, pfn_notify func(context Context)) (_err error) {
	context_1 := C.cl_context(context)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clSetContextDestructorCallback)
	}
	res := C.clSetContextDestructorCallback(context_1, callback, callback_uid)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateCommandQueueWithProperties.html
func CreateCommandQueueWithProperties(context Context, device DeviceId, properties []QueueProperties) (_res CommandQueue, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	properties_1, properties_fin := sliceToCZeroTerm(properties)
	defer properties_fin()
	context_1 := C.cl_context(context)
	device_1 := C.cl_device_id(device)
	properties_2 := (*C.cl_queue_properties)(properties_1)
	res := C.clCreateCommandQueueWithProperties(context_1, device_1, properties_2, &errcode_ret_1)
	res_1 := CommandQueue(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clRetainCommandQueue.html
func RetainCommandQueue(command_queue CommandQueue) (_err error) {
	command_queue_1 := C.cl_command_queue(command_queue)
	res := C.clRetainCommandQueue(command_queue_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clReleaseCommandQueue.html
func ReleaseCommandQueue(command_queue CommandQueue) (_err error) {
	command_queue_1 := C.cl_command_queue(command_queue)
	res := C.clReleaseCommandQueue(command_queue_1)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetCommandQueueInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetCommandQueueInfo.html
func GetCommandQueueInfo(command_queue CommandQueue, param_name CommandQueueInfo, param_value any) (_err error) {
	command_queue_1 := C.cl_command_queue(command_queue)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetCommandQueueInfo(command_queue_1, C.cl_command_queue_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetCommandQueueInfo(command_queue_1, C.cl_command_queue_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateBuffer.html
func CreateBuffer(context Context, flags MemFlags, size uint64, host_ptr unsafe.Pointer) (_res Mem, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	flags_1 := C.cl_mem_flags(flags)
	size_1 := C.size_t(size)
	host_ptr_1 := host_ptr
	res := C.clCreateBuffer(context_1, flags_1, size_1, host_ptr_1, &errcode_ret_1)
	res_1 := Mem(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateSubBuffer.html
func CreateSubBuffer(buffer Mem, flags MemFlags, buffer_create_type BufferCreateType, buffer_create_info unsafe.Pointer) (_res Mem, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	buffer_1 := C.cl_mem(buffer)
	flags_1 := C.cl_mem_flags(flags)
	buffer_create_type_1 := C.cl_buffer_create_type(buffer_create_type)
	buffer_create_info_1 := buffer_create_info
	res := C.clCreateSubBuffer(buffer_1, flags_1, buffer_create_type_1, buffer_create_info_1, &errcode_ret_1)
	res_1 := Mem(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateImage.html
func CreateImage(context Context, flags MemFlags, image_format *ImageFormat, image_desc *ImageDesc, host_ptr unsafe.Pointer) (_res Mem, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	flags_1 := C.cl_mem_flags(flags)
	image_format_1 := (*C.cl_image_format)(image_format)
	image_desc_1 := (*C.cl_image_desc)(image_desc)
	host_ptr_1 := host_ptr
	res := C.clCreateImage(context_1, flags_1, image_format_1, image_desc_1, host_ptr_1, &errcode_ret_1)
	res_1 := Mem(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreatePipe.html
func CreatePipe(context Context, flags MemFlags, pipe_packet_size uint32, pipe_max_packets uint32, properties *PipeProperties) (_res Mem, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	flags_1 := C.cl_mem_flags(flags)
	pipe_packet_size_1 := C.cl_uint(pipe_packet_size)
	pipe_max_packets_1 := C.cl_uint(pipe_max_packets)
	properties_1 := (*C.cl_pipe_properties)(properties)
	res := C.clCreatePipe(context_1, flags_1, pipe_packet_size_1, pipe_max_packets_1, properties_1, &errcode_ret_1)
	res_1 := Mem(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateBufferWithProperties.html
func CreateBufferWithProperties(context Context, properties []MemProperties, flags MemFlags, size uint64, host_ptr unsafe.Pointer) (_res Mem, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	properties_1, properties_fin := sliceToCZeroTerm(properties)
	defer properties_fin()
	context_1 := C.cl_context(context)
	properties_2 := (*C.cl_mem_properties)(properties_1)
	flags_1 := C.cl_mem_flags(flags)
	size_1 := C.size_t(size)
	host_ptr_1 := host_ptr
	res := C.clCreateBufferWithProperties(context_1, properties_2, flags_1, size_1, host_ptr_1, &errcode_ret_1)
	res_1 := Mem(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateImageWithProperties.html
func CreateImageWithProperties(context Context, properties []MemProperties, flags MemFlags, image_format *ImageFormat, image_desc *ImageDesc, host_ptr unsafe.Pointer) (_res Mem, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	properties_1, properties_fin := sliceToCZeroTerm(properties)
	defer properties_fin()
	context_1 := C.cl_context(context)
	properties_2 := (*C.cl_mem_properties)(properties_1)
	flags_1 := C.cl_mem_flags(flags)
	image_format_1 := (*C.cl_image_format)(image_format)
	image_desc_1 := (*C.cl_image_desc)(image_desc)
	host_ptr_1 := host_ptr
	res := C.clCreateImageWithProperties(context_1, properties_2, flags_1, image_format_1, image_desc_1, host_ptr_1, &errcode_ret_1)
	res_1 := Mem(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clRetainMemObject.html
func RetainMemObject(memobj Mem) (_err error) {
	memobj_1 := C.cl_mem(memobj)
	res := C.clRetainMemObject(memobj_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clReleaseMemObject.html
func ReleaseMemObject(memobj Mem) (_err error) {
	memobj_1 := C.cl_mem(memobj)
	res := C.clReleaseMemObject(memobj_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetSupportedImageFormats.html
func GetSupportedImageFormats(context Context, flags MemFlags, image_type MemObjectType) (_image_formats []ImageFormat, _err error) {
	context_1 := C.cl_context(context)
	flags_1 := C.cl_mem_flags(flags)
	image_type_1 := C.cl_mem_object_type(image_type)
	var image_formats_actual_len C.cl_uint
	C.clGetSupportedImageFormats(context_1, flags_1, image_type_1, 0, nil, &image_formats_actual_len)
	image_formats_1 := make([]ImageFormat, image_formats_actual_len)
	image_formats, num_entries, image_formats_fin := sliceToC(image_formats_1)
	defer image_formats_fin()
	res := C.clGetSupportedImageFormats(context_1, flags_1, image_type_1, C.cl_uint(num_entries), (*C.cl_image_format)(image_formats), nil)
	return image_formats_1, makeError(ErrorCode(res))
}
// Simplified binding for clGetMemObjectInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetMemObjectInfo.html
func GetMemObjectInfo(memobj Mem, param_name MemInfo, param_value any) (_err error) {
	memobj_1 := C.cl_mem(memobj)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetMemObjectInfo(memobj_1, C.cl_mem_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetMemObjectInfo(memobj_1, C.cl_mem_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetImageInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetImageInfo.html
func GetImageInfo(image Mem, param_name ImageInfo, param_value any) (_err error) {
	image_1 := C.cl_mem(image)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetImageInfo(image_1, C.cl_image_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetImageInfo(image_1, C.cl_image_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetPipeInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetPipeInfo.html
func GetPipeInfo(pipe Mem, param_name PipeInfo, param_value any) (_err error) {
	pipe_1 := C.cl_mem(pipe)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetPipeInfo(pipe_1, C.cl_pipe_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetPipeInfo(pipe_1, C.cl_pipe_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
//export go_cl_callback_clSetMemObjectDestructorCallback
func go_cl_callback_clSetMemObjectDestructorCallback(memobj C.cl_mem, user_data *C.void) {
	memobj_1 := Mem(memobj)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(memobj Mem)))(memobj_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetMemObjectDestructorCallback.html
func SetMemObjectDestructorCallback(memobj Mem, pfn_notify func(memobj Mem)) (_err error) {
	memobj_1 := C.cl_mem(memobj)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clSetMemObjectDestructorCallback)
	}
	res := C.clSetMemObjectDestructorCallback(memobj_1, callback, callback_uid)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSVMAlloc.html
func SVMAlloc(context Context, flags SvmMemFlags, size uint64, alignment uint32) (_res unsafe.Pointer) {
	context_1 := C.cl_context(context)
	flags_1 := C.cl_svm_mem_flags(flags)
	size_1 := C.size_t(size)
	alignment_1 := C.cl_uint(alignment)
	res := C.clSVMAlloc(context_1, flags_1, size_1, alignment_1)
	res_1 := (unsafe.Pointer)(res)
	return res_1
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSVMFree.html
func SVMFree(context Context, svm_pointer unsafe.Pointer) {
	context_1 := C.cl_context(context)
	svm_pointer_1 := svm_pointer
	C.clSVMFree(context_1, svm_pointer_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateSamplerWithProperties.html
func CreateSamplerWithProperties(context Context, sampler_properties []SamplerProperties) (_res Sampler, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	sampler_properties_1, sampler_properties_fin := sliceToCZeroTerm(sampler_properties)
	defer sampler_properties_fin()
	context_1 := C.cl_context(context)
	sampler_properties_2 := (*C.cl_sampler_properties)(sampler_properties_1)
	res := C.clCreateSamplerWithProperties(context_1, sampler_properties_2, &errcode_ret_1)
	res_1 := Sampler(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clRetainSampler.html
func RetainSampler(sampler Sampler) (_err error) {
	sampler_1 := C.cl_sampler(sampler)
	res := C.clRetainSampler(sampler_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clReleaseSampler.html
func ReleaseSampler(sampler Sampler) (_err error) {
	sampler_1 := C.cl_sampler(sampler)
	res := C.clReleaseSampler(sampler_1)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetSamplerInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetSamplerInfo.html
func GetSamplerInfo(sampler Sampler, param_name SamplerInfo, param_value any) (_err error) {
	sampler_1 := C.cl_sampler(sampler)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetSamplerInfo(sampler_1, C.cl_sampler_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetSamplerInfo(sampler_1, C.cl_sampler_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateProgramWithSource.html
func CreateProgramWithSource(context Context, strings []string) (_res Program, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	strings_1, lengths_1, strings_1_fin := stringsToC(strings, true)
	count_1 := len(strings)
	defer strings_1_fin()
	context_1 := C.cl_context(context)
	count_2 := C.cl_uint(count_1)
	res := C.clCreateProgramWithSource(context_1, count_2, strings_1, lengths_1, &errcode_ret_1)
	res_1 := Program(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateProgramWithBinary.html
func CreateProgramWithBinary(context Context, device_list []DeviceId, binaries [][]byte, binary_status *int32) (_res Program, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	device_list_1, num_devices_1, device_list_fin := sliceToC(device_list)
	defer device_list_fin()
	binaries_1, lengths_1, binaries_1_fin := byteSlicesToC(binaries)
	defer binaries_1_fin()
	context_1 := C.cl_context(context)
	num_devices_2 := C.cl_uint(num_devices_1)
	device_list_2 := (*C.cl_device_id)(device_list_1)
	binary_status_1 := (*C.cl_int)(binary_status)
	res := C.clCreateProgramWithBinary(context_1, num_devices_2, device_list_2, lengths_1, binaries_1, binary_status_1, &errcode_ret_1)
	res_1 := Program(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateProgramWithBuiltInKernels.html
func CreateProgramWithBuiltInKernels(context Context, device_list []DeviceId, kernel_names string) (_res Program, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	device_list_1, num_devices_1, device_list_fin := sliceToC(device_list)
	defer device_list_fin()
	kernel_names_1, kernel_names_1_fin := stringToC(kernel_names)
	defer kernel_names_1_fin()
	context_1 := C.cl_context(context)
	num_devices_2 := C.cl_uint(num_devices_1)
	device_list_2 := (*C.cl_device_id)(device_list_1)
	res := C.clCreateProgramWithBuiltInKernels(context_1, num_devices_2, device_list_2, kernel_names_1, &errcode_ret_1)
	res_1 := Program(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateProgramWithIL.html
func CreateProgramWithIL(context Context, il unsafe.Pointer, length uint64) (_res Program, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	il_1 := il
	length_1 := C.size_t(length)
	res := C.clCreateProgramWithIL(context_1, il_1, length_1, &errcode_ret_1)
	res_1 := Program(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clRetainProgram.html
func RetainProgram(program Program) (_err error) {
	program_1 := C.cl_program(program)
	res := C.clRetainProgram(program_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clReleaseProgram.html
func ReleaseProgram(program Program) (_err error) {
	program_1 := C.cl_program(program)
	res := C.clReleaseProgram(program_1)
	return makeError(ErrorCode(res))
}
//export go_cl_callback_clBuildProgram
func go_cl_callback_clBuildProgram(program C.cl_program, user_data *C.void) {
	program_1 := Program(program)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(program Program)))(program_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clBuildProgram.html
func BuildProgram(program Program, device_list []DeviceId, options string, pfn_notify func(program Program)) (_err error) {
	options_1, options_1_fin := stringToC(options)
	defer options_1_fin()
	device_list_1, num_devices_1, device_list_fin := sliceToC(device_list)
	defer device_list_fin()
	program_1 := C.cl_program(program)
	num_devices_2 := C.cl_uint(num_devices_1)
	device_list_2 := (*C.cl_device_id)(device_list_1)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clBuildProgram)
	}
	res := C.clBuildProgram(program_1, num_devices_2, device_list_2, options_1, callback, callback_uid)
	return makeError(ErrorCode(res))
}
//export go_cl_callback_clCompileProgram
func go_cl_callback_clCompileProgram(program C.cl_program, user_data *C.void) {
	program_1 := Program(program)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(program Program)))(program_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCompileProgram.html
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
//export go_cl_callback_clLinkProgram
func go_cl_callback_clLinkProgram(program C.cl_program, user_data *C.void) {
	program_1 := Program(program)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(program Program)))(program_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clLinkProgram.html
func LinkProgram(context Context, device_list []DeviceId, options string, input_programs []Program, pfn_notify func(program Program)) (_res Program, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	options_1, options_1_fin := stringToC(options)
	defer options_1_fin()
	device_list_1, num_devices_1, device_list_fin := sliceToC(device_list)
	defer device_list_fin()
	input_programs_1, num_input_programs_1, input_programs_fin := sliceToC(input_programs)
	defer input_programs_fin()
	context_1 := C.cl_context(context)
	num_devices_2 := C.cl_uint(num_devices_1)
	device_list_2 := (*C.cl_device_id)(device_list_1)
	num_input_programs_2 := C.cl_uint(num_input_programs_1)
	input_programs_2 := (*C.cl_program)(input_programs_1)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clLinkProgram)
	}
	res := C.clLinkProgram(context_1, num_devices_2, device_list_2, options_1, num_input_programs_2, input_programs_2, callback, callback_uid, &errcode_ret_1)
	res_1 := Program(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
//export go_cl_callback_clSetProgramReleaseCallback
func go_cl_callback_clSetProgramReleaseCallback(program C.cl_program, user_data *C.void) {
	program_1 := Program(program)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(program Program)))(program_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetProgramReleaseCallback.html
func SetProgramReleaseCallback(program Program, pfn_notify func(program Program)) (_err error) {
	program_1 := C.cl_program(program)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clSetProgramReleaseCallback)
	}
	res := C.clSetProgramReleaseCallback(program_1, callback, callback_uid)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetProgramSpecializationConstant.html
func SetProgramSpecializationConstant(program Program, spec_id uint32, spec_size uint64, spec_value unsafe.Pointer) (_err error) {
	program_1 := C.cl_program(program)
	spec_id_1 := C.cl_uint(spec_id)
	spec_size_1 := C.size_t(spec_size)
	spec_value_1 := spec_value
	res := C.clSetProgramSpecializationConstant(program_1, spec_id_1, spec_size_1, spec_value_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clUnloadPlatformCompiler.html
func UnloadPlatformCompiler(platform PlatformId) (_err error) {
	platform_1 := C.cl_platform_id(platform)
	res := C.clUnloadPlatformCompiler(platform_1)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetProgramInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetProgramInfo.html
func GetProgramInfo(program Program, param_name ProgramInfo, param_value any) (_err error) {
	program_1 := C.cl_program(program)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetProgramInfo(program_1, C.cl_program_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetProgramInfo(program_1, C.cl_program_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetProgramBuildInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetProgramBuildInfo.html
func GetProgramBuildInfo(program Program, device DeviceId, param_name ProgramBuildInfo, param_value any) (_err error) {
	program_1 := C.cl_program(program)
	device_1 := C.cl_device_id(device)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetProgramBuildInfo(program_1, device_1, C.cl_program_build_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetProgramBuildInfo(program_1, device_1, C.cl_program_build_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateKernel.html
func CreateKernel(program Program, kernel_name string) (_res Kernel, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	kernel_name_1, kernel_name_1_fin := stringToC(kernel_name)
	defer kernel_name_1_fin()
	program_1 := C.cl_program(program)
	res := C.clCreateKernel(program_1, kernel_name_1, &errcode_ret_1)
	res_1 := Kernel(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateKernelsInProgram.html
func CreateKernelsInProgram(program Program) (_kernels []Kernel, _err error) {
	program_1 := C.cl_program(program)
	var kernels_actual_len C.cl_uint
	C.clCreateKernelsInProgram(program_1, 0, nil, &kernels_actual_len)
	kernels_1 := make([]Kernel, kernels_actual_len)
	kernels, num_kernels, kernels_fin := sliceToC(kernels_1)
	defer kernels_fin()
	res := C.clCreateKernelsInProgram(program_1, C.cl_uint(num_kernels), (*C.cl_kernel)(kernels), nil)
	return kernels_1, makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCloneKernel.html
func CloneKernel(source_kernel Kernel) (_res Kernel, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	source_kernel_1 := C.cl_kernel(source_kernel)
	res := C.clCloneKernel(source_kernel_1, &errcode_ret_1)
	res_1 := Kernel(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clRetainKernel.html
func RetainKernel(kernel Kernel) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	res := C.clRetainKernel(kernel_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clReleaseKernel.html
func ReleaseKernel(kernel Kernel) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	res := C.clReleaseKernel(kernel_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetKernelArg.html
func SetKernelArg(kernel Kernel, arg_index uint32, arg_size uint64, arg_value unsafe.Pointer) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	arg_index_1 := C.cl_uint(arg_index)
	arg_size_1 := C.size_t(arg_size)
	arg_value_1 := arg_value
	res := C.clSetKernelArg(kernel_1, arg_index_1, arg_size_1, arg_value_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetKernelArgSVMPointer.html
func SetKernelArgSVMPointer(kernel Kernel, arg_index uint32, arg_value unsafe.Pointer) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	arg_index_1 := C.cl_uint(arg_index)
	arg_value_1 := arg_value
	res := C.clSetKernelArgSVMPointer(kernel_1, arg_index_1, arg_value_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetKernelExecInfo.html
func SetKernelExecInfo(kernel Kernel, param_name KernelExecInfo, param_value_size uint64, param_value unsafe.Pointer) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	param_name_1 := C.cl_kernel_exec_info(param_name)
	param_value_size_1 := C.size_t(param_value_size)
	param_value_1 := param_value
	res := C.clSetKernelExecInfo(kernel_1, param_name_1, param_value_size_1, param_value_1)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetKernelInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetKernelInfo.html
func GetKernelInfo(kernel Kernel, param_name KernelInfo, param_value any) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetKernelInfo(kernel_1, C.cl_kernel_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetKernelInfo(kernel_1, C.cl_kernel_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetKernelArgInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetKernelArgInfo.html
func GetKernelArgInfo(kernel Kernel, arg_indx uint32, param_name KernelArgInfo, param_value any) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	arg_indx_1 := C.cl_uint(arg_indx)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetKernelArgInfo(kernel_1, arg_indx_1, C.cl_kernel_arg_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetKernelArgInfo(kernel_1, arg_indx_1, C.cl_kernel_arg_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetKernelWorkGroupInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetKernelWorkGroupInfo.html
func GetKernelWorkGroupInfo(kernel Kernel, device DeviceId, param_name KernelWorkGroupInfo, param_value any) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	device_1 := C.cl_device_id(device)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetKernelWorkGroupInfo(kernel_1, device_1, C.cl_kernel_work_group_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetKernelWorkGroupInfo(kernel_1, device_1, C.cl_kernel_work_group_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetKernelSubGroupInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetKernelSubGroupInfo.html
func GetKernelSubGroupInfo(kernel Kernel, device DeviceId, param_name KernelSubGroupInfo, input_value_size uint64, input_value unsafe.Pointer, param_value any) (_err error) {
	kernel_1 := C.cl_kernel(kernel)
	device_1 := C.cl_device_id(device)
	input_value_size_1 := C.size_t(input_value_size)
	input_value_1 := input_value
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetKernelSubGroupInfo(kernel_1, device_1, C.cl_kernel_sub_group_info(param_name), input_value_size_1, input_value_1, 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetKernelSubGroupInfo(kernel_1, device_1, C.cl_kernel_sub_group_info(param_name), input_value_size_1, input_value_1, param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clWaitForEvents.html
func WaitForEvents(event_list []Event) (_err error) {
	event_list_1, num_events_1, event_list_fin := sliceToC(event_list)
	defer event_list_fin()
	num_events_2 := C.cl_uint(num_events_1)
	event_list_2 := (*C.cl_event)(event_list_1)
	res := C.clWaitForEvents(num_events_2, event_list_2)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetEventInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetEventInfo.html
func GetEventInfo(event Event, param_name EventInfo, param_value any) (_err error) {
	event_1 := C.cl_event(event)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetEventInfo(event_1, C.cl_event_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetEventInfo(event_1, C.cl_event_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateUserEvent.html
func CreateUserEvent(context Context) (_res Event, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	res := C.clCreateUserEvent(context_1, &errcode_ret_1)
	res_1 := Event(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clRetainEvent.html
func RetainEvent(event Event) (_err error) {
	event_1 := C.cl_event(event)
	res := C.clRetainEvent(event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clReleaseEvent.html
func ReleaseEvent(event Event) (_err error) {
	event_1 := C.cl_event(event)
	res := C.clReleaseEvent(event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetUserEventStatus.html
func SetUserEventStatus(event Event, execution_status int32) (_err error) {
	event_1 := C.cl_event(event)
	execution_status_1 := C.cl_int(execution_status)
	res := C.clSetUserEventStatus(event_1, execution_status_1)
	return makeError(ErrorCode(res))
}
//export go_cl_callback_clSetEventCallback
func go_cl_callback_clSetEventCallback(event C.cl_event, event_command_status C.cl_int, user_data *C.void) {
	event_1 := Event(event)
	event_command_status_1 := int32(event_command_status)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(event Event, event_command_status int32)))(event_1, event_command_status_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clSetEventCallback.html
func SetEventCallback(event Event, command_exec_callback_type int32, pfn_notify func(event Event, event_command_status int32)) (_err error) {
	event_1 := C.cl_event(event)
	command_exec_callback_type_1 := C.cl_int(command_exec_callback_type)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_notify != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_notify)))
		callback = (*[0]byte)(C.go_cl_callback_clSetEventCallback)
	}
	res := C.clSetEventCallback(event_1, command_exec_callback_type_1, callback, callback_uid)
	return makeError(ErrorCode(res))
}
// Simplified binding for clGetEventProfilingInfo.
//
// value must be a pointer to the result variable.
// The result variable must be typed according to param_name.
// The Go types allowed for the C types are:
//	- numerical/struct (e.g. cl_uint, size_t, cl_device_type): equivalent Go type (e.g. uint32, uint64, DeviceType)
//	- string (e.g. char[]): Go string or []byte (either is accepted)
//	- array (e.g. size_t[]): slice of equivalent Go type (e.g. []uint64)
//
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetEventProfilingInfo.html
func GetEventProfilingInfo(event Event, param_name ProfilingInfo, param_value any) (_err error) {
	event_1 := C.cl_event(event)
	var pin runtime.Pinner
	defer pin.Unpin()
	var param_actual_size C.size_t
	param_value_1 := reflect.ValueOf(param_value)
	if param_value_1.Kind() != reflect.Pointer {
		panic("expected param_value to be pointer to value, but got "+param_value_1.Kind().String())
	}
	isString := param_value_1.Elem().Kind() == reflect.String
	if isString || param_value_1.Elem().Kind() == reflect.Slice {
		param_value_1 = param_value_1.Elem() // need to take the address of the slice, not the pointer
		C.clGetEventProfilingInfo(event_1, C.cl_profiling_info(param_name), 0, nil, &param_actual_size)
		var elemTyp reflect.Type
		if isString { elemTyp = reflect.TypeFor[byte]() } else { elemTyp = param_value_1.Type().Elem() }
		sliceLen := int(param_actual_size)/int(elemTyp.Size())
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemTyp), sliceLen, sliceLen)
		if isString {
			outVal := param_value_1
			param_value_1 = newSlice
			defer func() {
				outVal.Set(param_value_1.Convert(reflect.TypeFor[string]()))
				if outVal.Len() > 0 && outVal.Index(outVal.Len()-1).IsZero() { outVal.Set(outVal.Slice(0, outVal.Len()-1)) } // strip null terminator
			}()
		} else {
			param_value_1.Set(newSlice)
		}
	} else {
		param_actual_size = C.size_t(param_value_1.Type().Size())
	}
	pin.Pin(param_value_1.UnsafePointer())
	res := C.clGetEventProfilingInfo(event_1, C.cl_profiling_info(param_name), param_actual_size, param_value_1.UnsafePointer(), &param_actual_size)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clFlush.html
func Flush(command_queue CommandQueue) (_err error) {
	command_queue_1 := C.cl_command_queue(command_queue)
	res := C.clFlush(command_queue_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clFinish.html
func Finish(command_queue CommandQueue) (_err error) {
	command_queue_1 := C.cl_command_queue(command_queue)
	res := C.clFinish(command_queue_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueReadBuffer.html
func EnqueueReadBuffer(command_queue CommandQueue, buffer Mem, blocking_read bool, offset uint64, size uint64, ptr unsafe.Pointer, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	buffer_1 := C.cl_mem(buffer)
	blocking_read_1 := boolToClBool(blocking_read)
	offset_1 := C.size_t(offset)
	size_1 := C.size_t(size)
	ptr_1 := ptr
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueReadBuffer(command_queue_1, buffer_1, blocking_read_1, offset_1, size_1, ptr_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueReadBufferRect.html
func EnqueueReadBufferRect(command_queue CommandQueue, buffer Mem, blocking_read bool, buffer_origin *uint64, host_origin *uint64, region *uint64, buffer_row_pitch uint64, buffer_slice_pitch uint64, host_row_pitch uint64, host_slice_pitch uint64, ptr unsafe.Pointer, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	buffer_1 := C.cl_mem(buffer)
	blocking_read_1 := boolToClBool(blocking_read)
	buffer_origin_1 := (*C.size_t)(buffer_origin)
	host_origin_1 := (*C.size_t)(host_origin)
	region_1 := (*C.size_t)(region)
	buffer_row_pitch_1 := C.size_t(buffer_row_pitch)
	buffer_slice_pitch_1 := C.size_t(buffer_slice_pitch)
	host_row_pitch_1 := C.size_t(host_row_pitch)
	host_slice_pitch_1 := C.size_t(host_slice_pitch)
	ptr_1 := ptr
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueReadBufferRect(command_queue_1, buffer_1, blocking_read_1, buffer_origin_1, host_origin_1, region_1, buffer_row_pitch_1, buffer_slice_pitch_1, host_row_pitch_1, host_slice_pitch_1, ptr_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueWriteBuffer.html
func EnqueueWriteBuffer(command_queue CommandQueue, buffer Mem, blocking_write bool, offset uint64, size uint64, ptr unsafe.Pointer, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	buffer_1 := C.cl_mem(buffer)
	blocking_write_1 := boolToClBool(blocking_write)
	offset_1 := C.size_t(offset)
	size_1 := C.size_t(size)
	ptr_1 := ptr
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueWriteBuffer(command_queue_1, buffer_1, blocking_write_1, offset_1, size_1, ptr_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueWriteBufferRect.html
func EnqueueWriteBufferRect(command_queue CommandQueue, buffer Mem, blocking_write bool, buffer_origin *uint64, host_origin *uint64, region *uint64, buffer_row_pitch uint64, buffer_slice_pitch uint64, host_row_pitch uint64, host_slice_pitch uint64, ptr unsafe.Pointer, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	buffer_1 := C.cl_mem(buffer)
	blocking_write_1 := boolToClBool(blocking_write)
	buffer_origin_1 := (*C.size_t)(buffer_origin)
	host_origin_1 := (*C.size_t)(host_origin)
	region_1 := (*C.size_t)(region)
	buffer_row_pitch_1 := C.size_t(buffer_row_pitch)
	buffer_slice_pitch_1 := C.size_t(buffer_slice_pitch)
	host_row_pitch_1 := C.size_t(host_row_pitch)
	host_slice_pitch_1 := C.size_t(host_slice_pitch)
	ptr_1 := ptr
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueWriteBufferRect(command_queue_1, buffer_1, blocking_write_1, buffer_origin_1, host_origin_1, region_1, buffer_row_pitch_1, buffer_slice_pitch_1, host_row_pitch_1, host_slice_pitch_1, ptr_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueFillBuffer.html
func EnqueueFillBuffer(command_queue CommandQueue, buffer Mem, pattern unsafe.Pointer, pattern_size uint64, offset uint64, size uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	buffer_1 := C.cl_mem(buffer)
	pattern_1 := pattern
	pattern_size_1 := C.size_t(pattern_size)
	offset_1 := C.size_t(offset)
	size_1 := C.size_t(size)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueFillBuffer(command_queue_1, buffer_1, pattern_1, pattern_size_1, offset_1, size_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueCopyBuffer.html
func EnqueueCopyBuffer(command_queue CommandQueue, src_buffer Mem, dst_buffer Mem, src_offset uint64, dst_offset uint64, size uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	src_buffer_1 := C.cl_mem(src_buffer)
	dst_buffer_1 := C.cl_mem(dst_buffer)
	src_offset_1 := C.size_t(src_offset)
	dst_offset_1 := C.size_t(dst_offset)
	size_1 := C.size_t(size)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueCopyBuffer(command_queue_1, src_buffer_1, dst_buffer_1, src_offset_1, dst_offset_1, size_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueCopyBufferRect.html
func EnqueueCopyBufferRect(command_queue CommandQueue, src_buffer Mem, dst_buffer Mem, src_origin *uint64, dst_origin *uint64, region *uint64, src_row_pitch uint64, src_slice_pitch uint64, dst_row_pitch uint64, dst_slice_pitch uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	src_buffer_1 := C.cl_mem(src_buffer)
	dst_buffer_1 := C.cl_mem(dst_buffer)
	src_origin_1 := (*C.size_t)(src_origin)
	dst_origin_1 := (*C.size_t)(dst_origin)
	region_1 := (*C.size_t)(region)
	src_row_pitch_1 := C.size_t(src_row_pitch)
	src_slice_pitch_1 := C.size_t(src_slice_pitch)
	dst_row_pitch_1 := C.size_t(dst_row_pitch)
	dst_slice_pitch_1 := C.size_t(dst_slice_pitch)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueCopyBufferRect(command_queue_1, src_buffer_1, dst_buffer_1, src_origin_1, dst_origin_1, region_1, src_row_pitch_1, src_slice_pitch_1, dst_row_pitch_1, dst_slice_pitch_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueReadImage.html
func EnqueueReadImage(command_queue CommandQueue, image Mem, blocking_read bool, origin *uint64, region *uint64, row_pitch uint64, slice_pitch uint64, ptr unsafe.Pointer, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	image_1 := C.cl_mem(image)
	blocking_read_1 := boolToClBool(blocking_read)
	origin_1 := (*C.size_t)(origin)
	region_1 := (*C.size_t)(region)
	row_pitch_1 := C.size_t(row_pitch)
	slice_pitch_1 := C.size_t(slice_pitch)
	ptr_1 := ptr
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueReadImage(command_queue_1, image_1, blocking_read_1, origin_1, region_1, row_pitch_1, slice_pitch_1, ptr_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueWriteImage.html
func EnqueueWriteImage(command_queue CommandQueue, image Mem, blocking_write bool, origin *uint64, region *uint64, input_row_pitch uint64, input_slice_pitch uint64, ptr unsafe.Pointer, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	image_1 := C.cl_mem(image)
	blocking_write_1 := boolToClBool(blocking_write)
	origin_1 := (*C.size_t)(origin)
	region_1 := (*C.size_t)(region)
	input_row_pitch_1 := C.size_t(input_row_pitch)
	input_slice_pitch_1 := C.size_t(input_slice_pitch)
	ptr_1 := ptr
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueWriteImage(command_queue_1, image_1, blocking_write_1, origin_1, region_1, input_row_pitch_1, input_slice_pitch_1, ptr_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueFillImage.html
func EnqueueFillImage(command_queue CommandQueue, image Mem, fill_color unsafe.Pointer, origin *uint64, region *uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	image_1 := C.cl_mem(image)
	fill_color_1 := fill_color
	origin_1 := (*C.size_t)(origin)
	region_1 := (*C.size_t)(region)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueFillImage(command_queue_1, image_1, fill_color_1, origin_1, region_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueCopyImage.html
func EnqueueCopyImage(command_queue CommandQueue, src_image Mem, dst_image Mem, src_origin *uint64, dst_origin *uint64, region *uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	src_image_1 := C.cl_mem(src_image)
	dst_image_1 := C.cl_mem(dst_image)
	src_origin_1 := (*C.size_t)(src_origin)
	dst_origin_1 := (*C.size_t)(dst_origin)
	region_1 := (*C.size_t)(region)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueCopyImage(command_queue_1, src_image_1, dst_image_1, src_origin_1, dst_origin_1, region_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueCopyImageToBuffer.html
func EnqueueCopyImageToBuffer(command_queue CommandQueue, src_image Mem, dst_buffer Mem, src_origin *uint64, region *uint64, dst_offset uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	src_image_1 := C.cl_mem(src_image)
	dst_buffer_1 := C.cl_mem(dst_buffer)
	src_origin_1 := (*C.size_t)(src_origin)
	region_1 := (*C.size_t)(region)
	dst_offset_1 := C.size_t(dst_offset)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueCopyImageToBuffer(command_queue_1, src_image_1, dst_buffer_1, src_origin_1, region_1, dst_offset_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueCopyBufferToImage.html
func EnqueueCopyBufferToImage(command_queue CommandQueue, src_buffer Mem, dst_image Mem, src_offset uint64, dst_origin *uint64, region *uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	src_buffer_1 := C.cl_mem(src_buffer)
	dst_image_1 := C.cl_mem(dst_image)
	src_offset_1 := C.size_t(src_offset)
	dst_origin_1 := (*C.size_t)(dst_origin)
	region_1 := (*C.size_t)(region)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueCopyBufferToImage(command_queue_1, src_buffer_1, dst_image_1, src_offset_1, dst_origin_1, region_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueMapBuffer.html
func EnqueueMapBuffer(command_queue CommandQueue, buffer Mem, blocking_map bool, map_flags MapFlags, offset uint64, size uint64, event_wait_list []Event, event *Event) (_res unsafe.Pointer, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	buffer_1 := C.cl_mem(buffer)
	blocking_map_1 := boolToClBool(blocking_map)
	map_flags_1 := C.cl_map_flags(map_flags)
	offset_1 := C.size_t(offset)
	size_1 := C.size_t(size)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueMapBuffer(command_queue_1, buffer_1, blocking_map_1, map_flags_1, offset_1, size_1, num_events_in_wait_list_2, event_wait_list_2, event_1, &errcode_ret_1)
	res_1 := (unsafe.Pointer)(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueMapImage.html
func EnqueueMapImage(command_queue CommandQueue, image Mem, blocking_map bool, map_flags MapFlags, origin *uint64, region *uint64, image_row_pitch *uint64, image_slice_pitch *uint64, event_wait_list []Event, event *Event) (_res unsafe.Pointer, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	image_1 := C.cl_mem(image)
	blocking_map_1 := boolToClBool(blocking_map)
	map_flags_1 := C.cl_map_flags(map_flags)
	origin_1 := (*C.size_t)(origin)
	region_1 := (*C.size_t)(region)
	image_row_pitch_1 := (*C.size_t)(image_row_pitch)
	image_slice_pitch_1 := (*C.size_t)(image_slice_pitch)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueMapImage(command_queue_1, image_1, blocking_map_1, map_flags_1, origin_1, region_1, image_row_pitch_1, image_slice_pitch_1, num_events_in_wait_list_2, event_wait_list_2, event_1, &errcode_ret_1)
	res_1 := (unsafe.Pointer)(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueUnmapMemObject.html
func EnqueueUnmapMemObject(command_queue CommandQueue, memobj Mem, mapped_ptr unsafe.Pointer, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	memobj_1 := C.cl_mem(memobj)
	mapped_ptr_1 := mapped_ptr
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueUnmapMemObject(command_queue_1, memobj_1, mapped_ptr_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueMigrateMemObjects.html
func EnqueueMigrateMemObjects(command_queue CommandQueue, mem_objects []Mem, flags MemMigrationFlags, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	mem_objects_1, num_mem_objects_1, mem_objects_fin := sliceToC(mem_objects)
	defer mem_objects_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	num_mem_objects_2 := C.cl_uint(num_mem_objects_1)
	mem_objects_2 := (*C.cl_mem)(mem_objects_1)
	flags_1 := C.cl_mem_migration_flags(flags)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueMigrateMemObjects(command_queue_1, num_mem_objects_2, mem_objects_2, flags_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueNDRangeKernel.html
func EnqueueNDRangeKernel(command_queue CommandQueue, kernel Kernel, work_dim uint32, global_work_offset []uint64, global_work_size []uint64, local_work_size []uint64, event_wait_list []Event, event *Event) (_err error) {
	global_work_offset_1, _, global_work_offset_fin := sliceToC(global_work_offset)
	defer global_work_offset_fin()
	global_work_size_1, _, global_work_size_fin := sliceToC(global_work_size)
	defer global_work_size_fin()
	local_work_size_1, _, local_work_size_fin := sliceToC(local_work_size)
	defer local_work_size_fin()
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	kernel_1 := C.cl_kernel(kernel)
	work_dim_1 := C.cl_uint(work_dim)
	global_work_offset_2 := (*C.size_t)(global_work_offset_1)
	global_work_size_2 := (*C.size_t)(global_work_size_1)
	local_work_size_2 := (*C.size_t)(local_work_size_1)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueNDRangeKernel(command_queue_1, kernel_1, work_dim_1, global_work_offset_2, global_work_size_2, local_work_size_2, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueMarkerWithWaitList.html
func EnqueueMarkerWithWaitList(command_queue CommandQueue, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueMarkerWithWaitList(command_queue_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueBarrierWithWaitList.html
func EnqueueBarrierWithWaitList(command_queue CommandQueue, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueBarrierWithWaitList(command_queue_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
//export go_cl_callback_clEnqueueSVMFree
func go_cl_callback_clEnqueueSVMFree(queue C.cl_command_queue, num_svm_pointers C.cl_uint, svm_pointers **C.void, user_data *C.void) {
	queue_1 := CommandQueue(queue)
	num_svm_pointers_1 := uint32(num_svm_pointers)
	svm_pointers_1 := (unsafe.Pointer)(svm_pointers)
	uid := int(uintptr(unsafe.Pointer(user_data)))
	defer callbackUnregister(uid)
	(callbackFn(uid).(func(queue CommandQueue, num_svm_pointers uint32, svm_pointers unsafe.Pointer)))(queue_1, num_svm_pointers_1, svm_pointers_1)
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueSVMFree.html
func EnqueueSVMFree(command_queue CommandQueue, num_svm_pointers uint32, svm_pointers unsafe.Pointer, pfn_free_func func(queue CommandQueue, num_svm_pointers uint32, svm_pointers unsafe.Pointer), event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	num_svm_pointers_1 := C.cl_uint(num_svm_pointers)
	svm_pointers_1 := (*unsafe.Pointer)(svm_pointers)
	var callback_uid unsafe.Pointer
	var callback *[0]byte
	if pfn_free_func != nil {
		callback_uid = unsafe.Pointer(uintptr(callbackRegister(pfn_free_func)))
		callback = (*[0]byte)(C.go_cl_callback_clEnqueueSVMFree)
	}
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueSVMFree(command_queue_1, num_svm_pointers_1, svm_pointers_1, callback, callback_uid, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueSVMMemcpy.html
func EnqueueSVMMemcpy(command_queue CommandQueue, blocking_copy bool, dst_ptr unsafe.Pointer, src_ptr unsafe.Pointer, size uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	blocking_copy_1 := boolToClBool(blocking_copy)
	dst_ptr_1 := dst_ptr
	src_ptr_1 := src_ptr
	size_1 := C.size_t(size)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueSVMMemcpy(command_queue_1, blocking_copy_1, dst_ptr_1, src_ptr_1, size_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueSVMMemFill.html
func EnqueueSVMMemFill(command_queue CommandQueue, svm_ptr unsafe.Pointer, pattern unsafe.Pointer, pattern_size uint64, size uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	svm_ptr_1 := svm_ptr
	pattern_1 := pattern
	pattern_size_1 := C.size_t(pattern_size)
	size_1 := C.size_t(size)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueSVMMemFill(command_queue_1, svm_ptr_1, pattern_1, pattern_size_1, size_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueSVMMap.html
func EnqueueSVMMap(command_queue CommandQueue, blocking_map bool, flags MapFlags, svm_ptr unsafe.Pointer, size uint64, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	blocking_map_1 := boolToClBool(blocking_map)
	flags_1 := C.cl_map_flags(flags)
	svm_ptr_1 := svm_ptr
	size_1 := C.size_t(size)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueSVMMap(command_queue_1, blocking_map_1, flags_1, svm_ptr_1, size_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueSVMUnmap.html
func EnqueueSVMUnmap(command_queue CommandQueue, svm_ptr unsafe.Pointer, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	svm_ptr_1 := svm_ptr
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueSVMUnmap(command_queue_1, svm_ptr_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueSVMMigrateMem.html
func EnqueueSVMMigrateMem(command_queue CommandQueue, num_svm_pointers uint32, svm_pointers unsafe.Pointer, sizes *uint64, flags MemMigrationFlags, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	num_svm_pointers_1 := C.cl_uint(num_svm_pointers)
	svm_pointers_1 := (*unsafe.Pointer)(svm_pointers)
	sizes_1 := (*C.size_t)(sizes)
	flags_1 := C.cl_mem_migration_flags(flags)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueSVMMigrateMem(command_queue_1, num_svm_pointers_1, svm_pointers_1, sizes_1, flags_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetKernelSuggestedLocalWorkSize.html
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
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetExtensionFunctionAddressForPlatform.html
func GetExtensionFunctionAddressForPlatform(platform PlatformId, func_name string) (_res unsafe.Pointer) {
	func_name_1, func_name_1_fin := stringToC(func_name)
	defer func_name_1_fin()
	platform_1 := C.cl_platform_id(platform)
	res := C.clGetExtensionFunctionAddressForPlatform(platform_1, func_name_1)
	res_1 := (unsafe.Pointer)(res)
	return res_1
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateImage2D.html
func CreateImage2D(context Context, flags MemFlags, image_format *ImageFormat, image_width uint64, image_height uint64, image_row_pitch uint64, host_ptr unsafe.Pointer) (_res Mem, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	flags_1 := C.cl_mem_flags(flags)
	image_format_1 := (*C.cl_image_format)(image_format)
	image_width_1 := C.size_t(image_width)
	image_height_1 := C.size_t(image_height)
	image_row_pitch_1 := C.size_t(image_row_pitch)
	host_ptr_1 := host_ptr
	res := C.clCreateImage2D(context_1, flags_1, image_format_1, image_width_1, image_height_1, image_row_pitch_1, host_ptr_1, &errcode_ret_1)
	res_1 := Mem(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateImage3D.html
func CreateImage3D(context Context, flags MemFlags, image_format *ImageFormat, image_width uint64, image_height uint64, image_depth uint64, image_row_pitch uint64, image_slice_pitch uint64, host_ptr unsafe.Pointer) (_res Mem, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	flags_1 := C.cl_mem_flags(flags)
	image_format_1 := (*C.cl_image_format)(image_format)
	image_width_1 := C.size_t(image_width)
	image_height_1 := C.size_t(image_height)
	image_depth_1 := C.size_t(image_depth)
	image_row_pitch_1 := C.size_t(image_row_pitch)
	image_slice_pitch_1 := C.size_t(image_slice_pitch)
	host_ptr_1 := host_ptr
	res := C.clCreateImage3D(context_1, flags_1, image_format_1, image_width_1, image_height_1, image_depth_1, image_row_pitch_1, image_slice_pitch_1, host_ptr_1, &errcode_ret_1)
	res_1 := Mem(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueMarker.html
func EnqueueMarker(command_queue CommandQueue, event *Event) (_err error) {
	command_queue_1 := C.cl_command_queue(command_queue)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueMarker(command_queue_1, event_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueWaitForEvents.html
func EnqueueWaitForEvents(command_queue CommandQueue, event_list []Event) (_err error) {
	event_list_1, num_events_1, event_list_fin := sliceToC(event_list)
	defer event_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	num_events_2 := C.cl_uint(num_events_1)
	event_list_2 := (*C.cl_event)(event_list_1)
	res := C.clEnqueueWaitForEvents(command_queue_1, num_events_2, event_list_2)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueBarrier.html
func EnqueueBarrier(command_queue CommandQueue) (_err error) {
	command_queue_1 := C.cl_command_queue(command_queue)
	res := C.clEnqueueBarrier(command_queue_1)
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clUnloadCompiler.html
func UnloadCompiler() (_err error) {
	res := C.clUnloadCompiler()
	return makeError(ErrorCode(res))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clGetExtensionFunctionAddress.html
func GetExtensionFunctionAddress(func_name string) (_res unsafe.Pointer) {
	func_name_1, func_name_1_fin := stringToC(func_name)
	defer func_name_1_fin()
	res := C.clGetExtensionFunctionAddress(func_name_1)
	res_1 := (unsafe.Pointer)(res)
	return res_1
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateCommandQueue.html
func CreateCommandQueue(context Context, device DeviceId, properties CommandQueueProperties) (_res CommandQueue, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	device_1 := C.cl_device_id(device)
	properties_1 := C.cl_command_queue_properties(properties)
	res := C.clCreateCommandQueue(context_1, device_1, properties_1, &errcode_ret_1)
	res_1 := CommandQueue(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clCreateSampler.html
func CreateSampler(context Context, normalized_coords bool, addressing_mode AddressingMode, filter_mode FilterMode) (_res Sampler, _errcode_ret error) {
	var errcode_ret_1 C.cl_int
	context_1 := C.cl_context(context)
	normalized_coords_1 := boolToClBool(normalized_coords)
	addressing_mode_1 := C.cl_addressing_mode(addressing_mode)
	filter_mode_1 := C.cl_filter_mode(filter_mode)
	res := C.clCreateSampler(context_1, normalized_coords_1, addressing_mode_1, filter_mode_1, &errcode_ret_1)
	res_1 := Sampler(res)
	return res_1, makeError(ErrorCode(errcode_ret_1))
}
// See https://registry.khronos.org/OpenCL/specs/unified/refpages/man/html/clEnqueueTask.html
func EnqueueTask(command_queue CommandQueue, kernel Kernel, event_wait_list []Event, event *Event) (_err error) {
	event_wait_list_1, num_events_in_wait_list_1, event_wait_list_fin := sliceToC(event_wait_list)
	defer event_wait_list_fin()
	command_queue_1 := C.cl_command_queue(command_queue)
	kernel_1 := C.cl_kernel(kernel)
	num_events_in_wait_list_2 := C.cl_uint(num_events_in_wait_list_1)
	event_wait_list_2 := (*C.cl_event)(event_wait_list_1)
	event_1 := (*C.cl_event)(event)
	res := C.clEnqueueTask(command_queue_1, kernel_1, num_events_in_wait_list_2, event_wait_list_2, event_1)
	return makeError(ErrorCode(res))
}
// (CUSTOM)
// Sets a kernel arg using a value.
//
func SetKernelArgValue[T any](kernel Kernel, arg_index uint32, value T) (_err error) {
	var pin runtime.Pinner
	defer pin.Unpin()
	pin.Pin(&value)
	return SetKernelArg(kernel, arg_index, uint64(unsafe.Sizeof(value)), unsafe.Pointer(&value))
}
// (CUSTOM)
// Set multiple kernel args. arg_offset specifies the arg index of the first value.
//
// Slower than [SetKernelArgValue] due to its use of reflection and the need
// to make (temporary) allocations.
//
// Returns the first error it encounters (if any).
//
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
// (CUSTOM)
// The [CreateBufferSlice] of [CreateBufferWithProperties].
//
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
// (CUSTOM)
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
func CreateBufferSlice[E any, S ~[]E](context Context, flags MemFlags, items S) (_res Mem, _errcode_ret error) {
	return CreateBufferSliceWithProperties(context, nil, flags, items)
}
// (CUSTOM)
// The [CreateBufferEmpty] of [CreateBufferWithProperties].
//
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
// (CUSTOM)
// Like [CreateBuffer], but determines memory requirements from data type and item count.
//
// Flags MUST NOT CONTAIN MEM_COPY_HOST_PTR or MEM_USE_HOST_PTR (will panic otherwise).
// 
// To allocate a buffer from a slice, see [CreateBufferSlice].
//
func CreateBufferEmpty[E any](context Context, flags MemFlags, num_items int) (_res Mem, _errcode_ret error) {
	return CreateBufferEmptyWithProperties[E](context, nil, flags, num_items)
}
// (CUSTOM)
// Like [EnqueueReadBuffer], but accepts a slice and automatically determines the memory region size.
//
// offset is now the number of ITEMS, not bytes (unlike [EnqueueReadBuffer]).
//
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
// (CUSTOM)
// Like [EnqueueWriteBuffer], but accepts a slice and automatically determines the memory region size.
//
// offset is now the number of ITEMS, not bytes (unlike [EnqueueWriteBuffer]).
//
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
// (CUSTOM)
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
// (CUSTOM)
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
// (CUSTOM)
// The [CreateBackedBuffer] of [CreateBufferWithProperties].
//
func CreateBackedBufferWithProperties[E any](context Context, properties []MemProperties, flags MemFlags, items []E) (res BackedBuffer[E], errcode_ret error) {
	if flags & MEM_ALLOC_HOST_PTR != 0 {
		panic("MEM_ALLOC_HOST_PTR not allowed")
	}
	if len(items) <= 0 {
		return BackedBuffer[E]{}, makeError(INVALID_VALUE)
	}
	itemSize := uint64(unsafe.Sizeof(&items[0]))
	res.Pinner.Pin(&items[0])
	ptr := unsafe.Pointer(&items[0])
	if flags&MEM_COPY_HOST_PTR == 0 && flags&MEM_USE_HOST_PTR == 0 {
		// Prevent CL_INVALID_HOST_PTR
		ptr = nil
	}
	res.Items = items
	if len(properties) == 0 {
		// BUG: CreateBufferWithProperties seems to cause a segfault on some system, so I'll just
		// use regular CreateBuffer if possible until I have figured this out.
		res.Mem, errcode_ret = CreateBuffer(context, flags, uint64(len(items))*itemSize, ptr)
	} else {
		res.Mem, errcode_ret = CreateBufferWithProperties(context, properties, flags, uint64(len(items))*itemSize, ptr)
	}
	if errcode_ret != nil {
		return BackedBuffer[E]{}, errcode_ret
	}
	return
}
// (CUSTOM)
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
func CreateBackedBuffer[E any](context Context, flags MemFlags, items []E) (res BackedBuffer[E], errcode_ret error) {
	return CreateBackedBufferWithProperties(context, nil, flags, items)
}
// (CUSTOM)
// Calls [ReleaseMemObject] and unpins the memory.
//
func (m *BackedBuffer[E]) Release() (err error) {
	err = ReleaseMemObject(m.Mem)
	m.Pinner.Unpin()
	return
}
// (CUSTOM)
// Like [EnqueueReadBuffer].
//
// items from indexing from start_offset to end_offset (exclusive) will be
// read back into the host buffer.
//
// end_offset may be set to -1 to indicate the range over all items.
//
// Returns [INVALID_VALUE] if start_offset >= end_offset, or any start_offset is negative.
//
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
// (CUSTOM)
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
// (CUSTOM)
// Calls [EnqueueFillBufferSlice] (see for documentation).
//
// Additionally, size may be -1 to indicate len(Items).
//
func (m BackedBuffer[E]) EnqueueFill(command_queue CommandQueue, pattern []E, offset int, size int, event_wait_list []Event, event *Event) (_err error) {
	if size == -1 {
		size = len(m.Items)
	}
	return EnqueueFillBufferSlice(command_queue, m.Mem, pattern, offset, size, event_wait_list, event)
}

// END cl-3.1/inc/CL/cl.h //

