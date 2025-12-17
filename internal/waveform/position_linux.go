//go:build linux

package waveform

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// positionWindow positions the window in the bottom-right corner of the screen
// and sets it to always-on-top. This function should be called after the window
// is created and visible.
func positionWindow(windowTitle string, width, height int) {
	// Give the window time to appear
	time.Sleep(100 * time.Millisecond)

	// Get screen dimensions using xdotool
	screenWidth, screenHeight := getScreenSize()
	if screenWidth == 0 || screenHeight == 0 {
		return
	}

	// Calculate position (bottom-right corner with padding)
	x := screenWidth - width - 20
	y := screenHeight - height - 60 // Account for taskbar

	// Find window by title and move it
	cmd := exec.Command("xdotool", "search", "--name", windowTitle)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	windowIDs := strings.Fields(string(output))
	if len(windowIDs) == 0 {
		return
	}

	windowID := windowIDs[0]

	// Move window to position
	moveCmd := exec.Command("xdotool", "windowmove", windowID, strconv.Itoa(x), strconv.Itoa(y))
	moveCmd.Run()

	// Try to set always-on-top using wmctrl
	wmctrlCmd := exec.Command("wmctrl", "-i", "-r", windowID, "-b", "add,above")
	if err := wmctrlCmd.Run(); err != nil {
		// wmctrl might not be installed, try xprop alternative
		xpropCmd := exec.Command("xprop", "-id", windowID, "-f", "_NET_WM_STATE", "32a",
			"-set", "_NET_WM_STATE", "_NET_WM_STATE_ABOVE")
		xpropCmd.Run()
	}
}

// getScreenSize returns the screen dimensions using xdotool.
func getScreenSize() (width, height int) {
	cmd := exec.Command("xdotool", "getdisplaygeometry")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	parts := strings.Fields(string(output))
	if len(parts) != 2 {
		return 0, 0
	}

	width, _ = strconv.Atoi(parts[0])
	height, _ = strconv.Atoi(parts[1])
	return width, height
}
