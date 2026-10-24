// Package operatingsystem provides helper function to get the operating system
// name for different platforms.
package operatingsystem

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	// file to check to determine Operating System
	etcOsRelease = "/etc/os-release"

	// used by stateless systems like Clear Linux
	altOsRelease = "/usr/lib/os-release"

	// used by RancherOS
	rosOsRelease = "/usr/share/ros/os-release"
)

// GetOperatingSystem gets the name of the current operating system.
func GetOperatingSystem() (string, error) {
	osReleaseFile, err := os.Open(etcOsRelease)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("Error opening %s: %v", etcOsRelease, err)
		}
		osReleaseFile, err = os.Open(altOsRelease)
		if os.IsNotExist(err) {
			osReleaseFile, err = os.Open(rosOsRelease)
			if err != nil {
				return "", fmt.Errorf("Error opening %s: %v", rosOsRelease, err)
			}
		} else {
			if err != nil {
				return "", fmt.Errorf("Error opening %s: %v", altOsRelease, err)
			}
		}
	}
	defer osReleaseFile.Close()

	var prettyName string
	scanner := bufio.NewScanner(osReleaseFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName = unquote(strings.TrimPrefix(line, "PRETTY_NAME="))
		}
	}
	if prettyName != "" {
		return prettyName, nil
	}
	// If not set, defaults to PRETTY_NAME="Linux"
	// c.f. http://www.freedesktop.org/software/systemd/man/os-release.html
	return "Linux", nil
}

// unquote strips the shell-style quotes os-release values may carry.
func unquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
		return v[1 : len(v)-1]
	}
	if u, err := strconv.Unquote(v); err == nil {
		return u
	}
	return v
}
