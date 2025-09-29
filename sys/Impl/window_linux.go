//go:build windows
// +build windows

package impl

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"github.com/pkg/errors"
)

// window is a Windows implementation of Window
type window struct {
	parent Server   // The server in charge of the window
	hwnd   syscall.Handle // Windows handle to the window
}

// newWindow creates a new window instance
func newWindow(x Server, hwnd syscall.Handle) (window, error) {
	win := window{parent: x, hwnd: hwnd}
	return win, win.check()
}

// check ensures the window is valid
func (w window) check() error {
	if w.hwnd == 0 {
		return errors.New("invalid window handle")
	}
	return nil
}

// rect gets the bounding rectangle of the window
func (w window) rect() (windows.RECT, error) {
	var r windows.RECT
	ret := windows.GetWindowRect(w.hwnd, &r)
	if ret != nil {
		return r, errors.Wrap(ret, "failed to get window rect")
	}
	return r, nil
}

// Center finds the point in the middle of the window
func (w window) Center() (image.Point, error) {
	r, err := w.rect()
	if err != nil {
		return image.Point{}, err
	}
	return image.Point{
		X: int((r.Left + r.Right) / 2),
		Y: int((r.Top + r.Bottom) / 2),
	}, nil
}

// WxH gets the width and height of the window
func (w window) WxH() (int, int, error) {
	r, err := w.rect()
	if err != nil {
		return 0, 0, err
	}
	return int(r.Right - r.Left), int(r.Bottom - r.Top), nil
}

// GetImage captures a screenshot of the window
func (w window) GetImage() (image.RGBA, error) {
	r, err := w.rect()
	if err != nil {
		return image.RGBA{}, err
	}
	width := int(r.Right - r.Left)
	height := int(r.Bottom - r.Top)

	hdcWindow, err := windows.GetDC(w.hwnd)
	if err != nil {
		return image.RGBA{}, errors.Wrap(err, "GetDC failed")
	}
	defer windows.ReleaseDC(w.hwnd, hdcWindow)

	hdcMem, err := windows.CreateCompatibleDC(hdcWindow)
	if err != nil {
		return image.RGBA{}, errors.Wrap(err, "CreateCompatibleDC failed")
	}
	defer windows.DeleteDC(hdcMem)

	hbm, err := windows.CreateCompatibleBitmap(hdcWindow, int32(width), int32(height))
	if err != nil {
		return image.RGBA{}, errors.Wrap(err, "CreateCompatibleBitmap failed")
	}
	defer windows.DeleteObject(windows.HGDIOBJ(hbm))

	old := windows.SelectObject(hdcMem, windows.HGDIOBJ(hbm))
	defer windows.SelectObject(hdcMem, old)

	if !windows.BitBlt(hdcMem, 0, 0, int32(width), int32(height), hdcWindow, 0, 0, windows.SRCCOPY) {
		return image.RGBA{}, errors.New("BitBlt failed")
	}

	var bmi windows.BITMAPINFO
	bmi.BmiHeader.Size = uint32(unsafe.Sizeof(bmi.BmiHeader))
	bmi.BmiHeader.Width = int32(width)
	bmi.BmiHeader.Height = -int32(height) // top-down DIB
	bmi.BmiHeader.Planes = 1
	bmi.BmiHeader.BitCount = 32
	bmi.BmiHeader.Compression = windows.BI_RGB

	buf := make([]byte, width*height*4)
	if windows.GetDIBits(hdcMem, hbm, 0, uint32(height), unsafe.Pointer(&buf[0]), &bmi, windows.DIB_RGB_COLORS) == 0 {
		return image.RGBA{}, errors.New("GetDIBits failed")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, buf)

	return *img, nil
}

// Name gets the window title
func (w window) Name() (string, error) {
	buf := make([]uint16, 256)
	n, err := windows.GetWindowText(w.hwnd, &buf[0], int32(len(buf)))
	if err != nil {
		return "", errors.Wrap(err, "GetWindowText failed")
	}
	return windows.UTF16ToString(buf[:n]), nil
}

// Resize adjusts the window size
func (w window) Resize(height, width int) error {
	return windows.SetWindowPos(w.hwnd, 0, 0, 0, int32(width), int32(height), windows.SWP_NOMOVE|windows.SWP_NOZORDER)
}

// SetActive brings the window to the foreground
func (w window) SetActive() error {
	if !windows.SetForegroundWindow(w.hwnd) {
		return errors.New("failed to set foreground window")
	}
	return nil
}

// ID returns the HWND as int
func (w window) ID() (int, error) {
	return int(w.hwnd), nil
}
