//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// platformChecks returns macOS-specific diagnostic checks.
func platformChecks() []Check {
	return []Check{
		checkSMTDarwin(),
		checkGovernorDarwin(),
		checkTurboDarwin(),
		checkLoadAvgDarwin(),
		checkThermalDarwin(),
	}
}

const darwinLinuxRemedy = "use a Linux machine or bare-metal CI runner for publication-quality numbers"

func checkSMTDarwin() Check {
	return Check{
		Name:   "SMT control",
		Status: statusUnavailable,
		Detail: "macOS does not expose SMT control via sysfs — " + darwinLinuxRemedy,
	}
}

func checkGovernorDarwin() Check {
	return Check{
		Name:   "CPU frequency governor",
		Status: statusUnavailable,
		Detail: "macOS does not expose a CPU frequency governor — " + darwinLinuxRemedy,
	}
}

func checkTurboDarwin() Check {
	return Check{
		Name:   "Turbo Boost",
		Status: statusUnavailable,
		Detail: "macOS does not expose Turbo Boost control from user space — " + darwinLinuxRemedy,
	}
}

func checkLoadAvgDarwin() Check {
	// sysctl -n vm.loadavg returns "{ 1.23 2.34 3.45 }"
	out, err := exec.Command("sysctl", "-n", "vm.loadavg").Output()
	if err != nil {
		return Check{
			Name:   "load average",
			Status: statusUnavailable,
			Detail: fmt.Sprintf("sysctl vm.loadavg failed: %v", err),
		}
	}
	load, err := parseDarwinLoadAvg(strings.TrimSpace(string(out)))
	if err != nil {
		return Check{
			Name:   "load average",
			Status: statusUnavailable,
			Detail: fmt.Sprintf("parse error: %v", err),
		}
	}
	status, detail, remedy := loadAvgResult(load, runtime.NumCPU())
	return Check{
		Name:   "load average",
		Status: status,
		Detail: detail,
		Remedy: remedy,
	}
}

func checkThermalDarwin() Check {
	return Check{
		Name:   "thermal pressure",
		Status: statusUnavailable,
		Detail: "macOS thermal state is not accessible from user space — watch for CPU throttling on sustained benchmark runs",
	}
}
