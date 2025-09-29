//go:build windows
// +build windows

package impl

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/pkg/errors"
)

var keycodes map[string]string

func keyInit() error {
	keycodes = make(map[string]string)

	// Define needed keys and their robotgo equivalents
	neededKeys := []string{"z", "x", "up", "left", "right", "down", "enter"}

	for _, k := range neededKeys {
		keycodes[k] = k
	}
	return nil
}

// Press key in the Undertale window (should be lowercase)
func (win window) Press(key string) error {
	var res error
	go func() {
		fmt.Printf("Pressing %s\n", key)
		activeWin, err := win.parent.activeWindow()
		if err != nil {
			res = errors.Wrap(err, "failed to get the active window")
			return
		}
		acID, err := activeWin.ID()
		if err != nil {
			res = errors.Wrap(err, "failed to get the ID of the active window")
			return
		}
		winID, err := win.ID()
		if err != nil {
			res = errors.Wrap(err, "failed to get the ID of the window")
			return
		}

		if acID == winID {
			err := win.justPress(key)
			if err != nil {
				res = errors.Wrap(err, "failed to use justPress")
			}
		} else {
			fmt.Println("Refocusing")
			err = win.SetActive()
			if err != nil {
				res = errors.Wrap(err, "failed to set the active window")
			}
			time.Sleep(250 * time.Millisecond)
			err = win.justPress(key)
			if err != nil {
				res = errors.Wrap(err, "failed to use justPress")
			}
		}
	}()
	return res
}

// No active window handling, just pressing the key on whatever window is active
func (win window) justPress(key string) error {
	lower := strings.ToLower(key)
	k, ok := keycodes[lower]
	if !ok {
		return errors.New("the key given was not included in the keycodes map")
	}

	// Press + Release using robotgo
	robotgo.KeyDown(k)
	time.Sleep(40 * time.Millisecond)
	robotgo.KeyUp(k)

	return nil
}
