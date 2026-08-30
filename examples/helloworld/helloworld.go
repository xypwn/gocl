package main

import (
	"fmt"
	"log"
	"runtime"

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

	var platformName string
	if err := cl.GetPlatformInfo(platform, cl.PLATFORM_NAME, &platformName); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Platform: %q\n", platformName)

	devices, err := cl.GetDeviceIDs(platform, cl.DEVICE_TYPE_GPU)
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		log.Fatal("no devices")
	}
	device := devices[0]

	var deviceName string
	if err := cl.GetDeviceInfo(device, cl.DEVICE_NAME, &deviceName); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Device: %q\n", deviceName)

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

	if err := cl.BuildProgram(prog, []cl.DeviceId{device}, "", func(program cl.Program) {
		fmt.Println("Done building program (callback)")
	}); err != nil {
		var logStr string
		if err := cl.GetProgramBuildInfo(prog, device, cl.PROGRAM_BUILD_LOG, &logStr); err != nil {
			log.Fatal(err)
		}
		log.Fatal(logStr)
	}

	kernel, err := cl.CreateKernel(prog, "helloworld")
	if err != nil {
		log.Fatal(err)
	}
	defer cl.ReleaseKernel(kernel)

	bufIn, err := cl.CreateBackedBuffer(ctx, cl.MEM_READ_ONLY|cl.MEM_COPY_HOST_PTR,
		[]float32{1, 2, 3, 4, 5, 6, 7, 8})
	if err != nil {
		log.Fatal(err)
	}
	defer bufIn.Release()

	bufOut, err := cl.CreateBackedBuffer(ctx, cl.MEM_WRITE_ONLY, make([]float32, 8))
	if err != nil {
		log.Fatal(err)
	}
	defer bufOut.Release()

	queue, err := cl.CreateCommandQueueWithProperties(ctx, device, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cl.ReleaseCommandQueue(queue)

	// Not necessary since we called CreateBufferSlice with cl.MEM_COPY_HOST_PTR, but this would
	// be used to update the buffer contents from the host side.
	//if err := memIn.EnqueueWrite(queue, true, 0, -1, nil, nil); err != nil {
	//	log.Fatal(err)
	//}

	if err := cl.SetKernelArgValues(kernel, 0, bufIn.Mem, bufOut.Mem, uint32(len(bufIn.Items))); err != nil {
		log.Fatal(err)
	}

	if err := cl.EnqueueNDRangeKernel(queue, kernel, 1, nil, []uint64{uint64(len(bufIn.Items))}, nil, nil, nil); err != nil {
		log.Fatal(err)
	}

	if err := cl.Finish(queue); err != nil {
		log.Fatal(err)
	}

	if err := bufOut.EnqueueRead(queue, true, 0, -1, nil, nil); err != nil {
		log.Fatal(err)
	}

	fmt.Println(bufOut.Items)
}
