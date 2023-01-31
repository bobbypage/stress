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
	checkLock           = pflag.Bool("check-lock", false, "check lock status periodically")
	buffer              [][]byte
)

const (
	filePath = "/mnt/mydir/foo.lock"
)

func main() {
	pflag.Parse()
	//total := resource.MustParse(*argMemTotal)
	//stepSize := resource.MustParse(*argMemStepSize)
	// Acquire lock-------------------------
	klog.Info("In main, Attempting to lock file /mnt/mydir/foo.lock")
	file, err := os.OpenFile(filePath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC, 0666)
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
	if *checkLock {
		go checkLockStatus()
	}
	//allocateMemory(total, stepSize)
	//klog.Infof("Allocated %q memory", total.String())
	for {
		klog.Infof("Main sleep for 10s")
		time.Sleep(10 * time.Second)
	}
	//select {}
}

func checkLockStatus() {
	klog.Info("Start lock checker go routine")
	// fd, err := os.OpenFile(filePath, syscall.O_RDONLY, 0666)
	// fd, err := os.Open(filePath)
	fd, err := os.OpenFile(filePath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC, 0666)
	if err != nil {
		klog.Errorf("error opening file: %s", err)
		return
	}
	klog.Infof("Opened file %s", filePath)
	for {
		flock := syscall.Flock_t{
			Type:   syscall.F_WRLCK,
			Whence: io.SeekStart,
			Start:  0,
			Len:    0,
		}
		klog.Infof("sleep for 30 secs")
		time.Sleep(30 * time.Second)
		err = syscall.FcntlFlock(fd.Fd(), syscall.F_GETLK, &flock)
		//err = syscall.FcntlFlock(fd.Fd(), syscall.F_SETLK, &flock)
		if err != nil {
			klog.Errorf("error calling FcntlFlock file: %s", err)
			continue
		}
		switch flock.Type {
		case syscall.F_WRLCK:
			klog.Infof("write lock")
		case syscall.F_RDLCK:
			klog.Infof("read lock")
		default:
			klog.Infof("Unknown lock type %v", flock.Type)
		}
	}
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
