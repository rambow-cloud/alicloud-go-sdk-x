//go:build !windows

package externalcreds

import "os/exec"

func hideProcessWindow(*exec.Cmd) {}
