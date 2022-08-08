package main

import (
	"io"
	"os"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

var (
	argMemTotal         = pflag.String("mem-total", "0", "total memory to be consumed. Memory will be consumed via multiple allocations.")
	argMemStepSize      = pflag.String("mem-alloc-size", "4Ki", "amount of memory to be consumed in each allocation")
	argMemSleepDuration = pflag.Duration("mem-alloc-sleep", time.Millisecond, "duration to sleep between allocations")
	argCpus             = pflag.Int("cpus", 0, "total number of CPUs to utilize")
	buffer              [][]byte
)

func main() {
	pflag.Parse()
	total := resource.MustParse(*argMemTotal)
	stepSize := resource.MustParse(*argMemStepSize)
	// Acquire lock-------------------------
	klog.Info("In main, Attempting to lock file /mnt/mydir/foo.lock")
	file, err := os.OpenFile("/mnt/mydir/foo.lock", syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC, 0666)
	if err != nil {
		klog.Errorf("error opening file: %s", err)
		return
	}
	//defer file.Close() # this seems to release the lock, and a contending container on the same node is able to acquire exclusive lock.

	flockT := syscall.Flock_t{
		Type:   syscall.F_WRLCK,
		Whence: io.SeekStart,
		Start:  0,
		Len:    0,
	}
	err = syscall.FcntlFlock(file.Fd(), syscall.F_SETLK, &flockT)
	if err != nil {
		klog.Fatalf("error locking file: %s", err)
		return
	}

	klog.Infof("write lock accessed on file /mnt/mydir/foo.lock")
	// -------------------------------------
	allocateMemory(total, stepSize)
	klog.Infof("Allocated %q memory", total.String())
	select {}
}

func allocateMemory(total, stepSize resource.Quantity) {
	klog.Infof("Allocating %q memory, in %q chunks, with a %v sleep between allocations", total.String(), stepSize.String(), *argMemSleepDuration)
	for i := int64(1); i*stepSize.Value() <= total.Value(); i++ {
		klog.Infof("allocate stepsize = %d bytes", stepSize.Value())
		newBuffer := make([]byte, stepSize.Value())
		for i := range newBuffer {
			newBuffer[i] = 0
		}
		buffer = append(buffer, newBuffer)
		time.Sleep(*argMemSleepDuration)
	}
}
