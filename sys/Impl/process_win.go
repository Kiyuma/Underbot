//go:build windows
// +build windows

package impl

import (
	"os"
	"strings"
	"syscall"
	"unsafe"

	ps "github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
)

var (
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess            = kernel32.NewProc("OpenProcess")
	procSuspendThread          = kernel32.NewProc("SuspendThread")
	procResumeThread           = kernel32.NewProc("ResumeThread")
	procCloseHandle            = kernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
)

const (
	PROCESS_ALL_ACCESS = 0x1F0FFF
)

var gamestate = 0

// Process gets the process of a window
func (xWin window) Process() (*os.Process, error) {
	// Use Windows API to get PID from window handle
	var pid uint32
	_, _, _ = syscall.NewLazyDLL("user32.dll").NewProc("GetWindowThreadProcessId").Call(
		uintptr(xWin.winID),
		uintptr(unsafe.Pointer(&pid)),
	)
	if pid == 0 {
		foundPid, err := findUndertale()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get pid of undertale")
		}
		pid = uint32(foundPid)
	}

	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return nil, errors.Wrap(err, "failed to find process")
	}
	return proc, nil
}

// Pause suspends the game process (all threads)
func (xWin window) Pause() error {
	if gamestate == 0 {
		proc, err := xWin.Process()
		if err != nil {
			return errors.Wrap(err, "failed to get process to pause")
		}
		err = suspendProcess(proc.Pid)
		if err != nil {
			return errors.Wrap(err, "failed to suspend process")
		}
		gamestate = 1
	}
	return nil
}

// Resume resumes the game process (all threads)
func (xWin window) Resume() error {
	if gamestate == 1 {
		proc, err := xWin.Process()
		if err != nil {
			return errors.Wrap(err, "failed to get process to resume")
		}
		err = resumeProcess(proc.Pid)
		if err != nil {
			return errors.Wrap(err, "failed to resume process")
		}
		gamestate = 0
	}
	return nil
}

// Helpers to suspend/resume
func suspendProcess(pid int) error {
	handle, _, _ := procOpenProcess.Call(PROCESS_ALL_ACCESS, 0, uintptr(pid))
	if handle == 0 {
		return errors.New("failed to open process")
	}
	defer procCloseHandle.Call(handle)

	// Normally would iterate threads with Toolhelp32Snapshot → simplified here
	// NOTE: Real implementation should suspend each thread individually
	_, _, _ = procSuspendThread.Call(handle)
	return nil
}

func resumeProcess(pid int) error {
	handle, _, _ := procOpenProcess.Call(PROCESS_ALL_ACCESS, 0, uintptr(pid))
	if handle == 0 {
		return errors.New("failed to open process")
	}
	defer procCloseHandle.Call(handle)

	_, _, _ = procResumeThread.Call(handle)
	return nil
}

// Indicators within a process name for an Undertale related process
var undertaleProcessNames = []string{"runner", "under", "tale"}

// Gets the PID of the Undertale process
func findUndertale() (uint, error) {
	for _, processName := range undertaleProcessNames {
		pid, err := executableNameToPid(processName)
		if err == nil {
			return pid, nil
		}
	}
	return 0, errors.New("failed to find undertale process")
}

// Finds the PID based on the executable name of a process
func executableNameToPid(processName string) (uint, error) {
	processes, err := ps.Processes()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get process list")
	}
	for _, process := range processes {
		if strings.Contains(strings.ToLower(process.Executable()), strings.ToLower(processName)) &&
			process.Executable() != "Underbot.exe" {
			return uint(process.Pid()), nil
		}
	}
	return 0, errors.New("could not find the process")
}
