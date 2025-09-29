//go:build windows
// +build windows

package impl

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"github.com/pkg/errors"
)

// Server represents a Windows window manager session
type Server struct {
	// no explicit connection needed in Windows
}

// NewServer creates a new server session
func NewServer() (Server, error) {
	return Server{}, nil
}

// check ensures the server is valid
func (s Server) check() error {
	// nothing to check, always valid
	return nil
}

// FindWindow finds a window by its title
func (s Server) FindWindow(name string) (window, error) {
	hwnd, err := windows.FindWindow(nil, syscall.StringToUTF16Ptr(name))
	if err != nil {
		return window{}, errors.Wrap(err, "failed to find window")
	}
	if hwnd == 0 {
		return window{}, errors.New("window not found")
	}
	return newWindow(s, hwnd)
}

// ActiveWindow returns the currently active window
func (s Server) ActiveWindow() (window, error) {
	hwnd := windows.GetForegroundWindow()
	if hwnd == 0 {
		return window{}, errors.New("no active window")
	}
	return newWindow(s, hwnd)
}

// EnumWindows enumerates all top-level windows
func (s Server) EnumWindows() ([]window, error) {
	var wins []window
	cb := syscall.NewCallback(func(h syscall.Handle, _ uintptr) uintptr {
		win, err := newWindow(s, h)
		if err == nil {
			wins = append(wins, win)
		}
		return 1 // continue enumeration
	})
	if err := windows.EnumWindows(cb, 0); err != nil {
		return nil, errors.Wrap(err, "EnumWindows failed")
	}
	return wins, nil
}

// KillWindow closes a window by handle
func (s Server) KillWindow(w window) error {
	if !windows.PostMessage(w.hwnd, windows.WM_CLOSE, 0, 0) {
		return errors.New("failed to send WM_CLOSE")
	}
	return nil
}

// BringToFront brings a window to the foreground
func (s Server) BringToFront(w window) error {
	if !windows.SetForegroundWindow(w.hwnd) {
		return errors.New("failed to set foreground window")
	}
	return nil
}

// GetWindowThreadProcessID returns the process ID of the window
func (s Server) GetWindowThreadProcessID(w window) (uint32, error) {
	var pid uint32
	windows.GetWindowThreadProcessId(w.hwnd, &pid)
	if pid == 0 {
		return 0, errors.New("failed to get process ID")
	}
	return pid, nil
}

// GetClassName returns the window class name
func (s Server) GetClassName(w window) (string, error) {
	buf := make([]uint16, 256)
	n, err := windows.GetClassName(w.hwnd, &buf[0], int32(len(buf)))
	if err != nil {
		return "", errors.Wrap(err, "GetClassName failed")
	}
	return windows.UTF16ToString(buf[:n]), nil
}
