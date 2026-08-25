package main

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	cl "github.com/xypwn/gocl/cl-3.1"
)

func main() {
	log.SetFlags(log.Lshortfile)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	platforms, err := cl.GetPlatformIDs()
	if err != nil {
		log.Fatal(err)
	}
	if len(platforms) == 0 {
		log.Fatal("no platforms")
	}
	platform := platforms[0]

	devices, err := cl.GetDeviceIDs(platform, cl.DEVICE_TYPE_GPU)
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		log.Fatal("no devices")
	}
	device := devices[0]

	var deviceNameLen uint64
	var deviceNameArr [64]byte
	if err := cl.GetDeviceInfo(device, cl.DEVICE_NAME, uint64(len(deviceNameArr)), unsafe.Pointer(&deviceNameArr[0]), &deviceNameLen); err != nil {
		log.Fatal(err)
	}
	deviceName := string(deviceNameArr[:deviceNameLen])

	fmt.Println("Device:", deviceName)

	ctx, err := cl.CreateContext(nil, []cl.DeviceId{device}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.ReleaseContext(ctx)

	prog, err := cl.CreateProgramWithSource(ctx, []string{
		`__kernel void helloworld(__global float* in, __global float* out, int count) {
	int id = get_global_id(0);
	if (id >= count) return;
	out[id] = in[id] * in[id];
}`,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cl.ReleaseProgram(prog)

	if err := cl.BuildProgram(prog, []cl.DeviceId{device}, nil, func(program cl.Program) {
		fmt.Println("Done building program (callback)")
	}); err != nil {
		var logStr string
		{
			var logLen uint64
			cl.GetProgramBuildInfo(prog, device, cl.PROGRAM_BUILD_LOG, 0, nil, &logLen)
			logBuf := make([]byte, logLen)
			if err := cl.GetProgramBuildInfo(prog, device, cl.PROGRAM_BUILD_LOG, logLen, unsafe.Pointer(&logBuf[0]), &logLen); err != nil {
				log.Fatal(err)
			}
			runtime.KeepAlive(logBuf)
			logStr = string(logBuf)
		}
		log.Fatal(err, logStr)
	}

	kernel, err := cl.CreateKernel(prog, "helloworld")
	if err != nil {
		log.Fatal(err)
	}
	defer cl.ReleaseKernel(kernel)

	data := [8]float32{1, 2, 3, 4, 5, 6, 7, 8}
	data_size := uint64(int(unsafe.Sizeof(float32(0))) * len(data))

	memIn, err := cl.CreateBuffer(ctx, cl.MEM_READ_ONLY, data_size, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.ReleaseMemObject(memIn)
	memOut, err := cl.CreateBuffer(ctx, cl.MEM_WRITE_ONLY, data_size, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.ReleaseMemObject(memOut)

	queue, err := cl.CreateCommandQueueWithProperties(ctx, device, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.ReleaseCommandQueue(queue)

	if err := cl.EnqueueWriteBuffer(queue, memIn, true, 0, data_size, unsafe.Pointer(&data[0]), nil, nil); err != nil {
		log.Fatal(err)
	}

	if err := cl.SetKernelArg(kernel, 0, uint64(unsafe.Sizeof(memIn)), unsafe.Pointer(&memIn)); err != nil {
		log.Fatal(err)
	}
	if err := cl.SetKernelArg(kernel, 1, uint64(unsafe.Sizeof(memOut)), unsafe.Pointer(&memOut)); err != nil {
		log.Fatal(err)
	}
	count := uint32(len(data))
	if err := cl.SetKernelArg(kernel, 2, uint64(unsafe.Sizeof(count)), unsafe.Pointer(&count)); err != nil {
		log.Fatal(err)
	}

	count64 := uint64(count)
	if err := cl.EnqueueNDRangeKernel(queue, kernel, 1, nil, &count64, nil, nil, nil); err != nil {
		log.Fatal(err)
	}

	if err := cl.Finish(queue); err != nil {
		log.Fatal(err)
	}

	if err := cl.EnqueueReadBuffer(queue, memOut, true, 0, data_size, unsafe.Pointer(&data[0]), nil, nil); err != nil {
		log.Fatal(err)
	}

	fmt.Println(data)
}
