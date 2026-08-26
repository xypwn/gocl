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

//BEGIN:DELETE//

func _forceusepkgs() {
	// Prevent gopls from removing imports
	fmt.Println()
	reflect.TypeFor[int]()
	_ := strings.Builder{}
}

//END:DELETE//

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
