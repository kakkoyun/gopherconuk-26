//go:build linux

package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// platformChecks returns Linux-specific diagnostic checks.
func platformChecks() []Check {
	return []Check{
		checkSMT(),
		checkGovernor(),
		checkTurbo(),
		checkLoadAvgLinux(),
		checkContainer(),
	}
}

func checkSMT() Check {
	const path = "/sys/devices/system/cpu/smt/control"
	data, err := os.ReadFile(path)
	if err != nil {
		return Check{
			Name:   "SMT control",
			Status: statusUnavailable,
			Detail: fmt.Sprintf("cannot read %s: %v (may be a VM or single-core CPU)", path, err),
		}
	}
	status, detail, remedy := smtResult(strings.TrimSpace(string(data)))
	return Check{
		Name:   "SMT control",
		Status: status,
		Detail: detail,
		Remedy: remedy,
	}
}

func checkGovernor() Check {
	const path = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor"
	data, err := os.ReadFile(path)
	if err != nil {
		return Check{
			Name:   "CPU frequency governor",
			Status: statusUnavailable,
			Detail: fmt.Sprintf("cannot read %s: %v (cpufreq driver may not be loaded, or running in a VM)", path, err),
		}
	}
	status, detail, remedy := governorResult(strings.TrimSpace(string(data)))
	return Check{
		Name:   "CPU frequency governor",
		Status: status,
		Detail: detail,
		Remedy: remedy,
	}
}

func checkTurbo() Check {
	// Intel pstate driver.
	const intelPath = "/sys/devices/system/cpu/intel_pstate/no_turbo"
	if data, err := os.ReadFile(intelPath); err == nil {
		status, detail, remedy := turboIntelResult(strings.TrimSpace(string(data)))
		return Check{
			Name:   "Turbo Boost (Intel)",
			Status: status,
			Detail: detail,
			Remedy: remedy,
		}
	}

	// AMD cpufreq boost knob.
	const amdPath = "/sys/devices/system/cpu/cpufreq/boost"
	if data, err := os.ReadFile(amdPath); err == nil {
		status, detail, remedy := turboAMDResult(strings.TrimSpace(string(data)))
		return Check{
			Name:   "Turbo Boost (AMD)",
			Status: status,
			Detail: detail,
			Remedy: remedy,
		}
	}

	return Check{
		Name:   "Turbo Boost",
		Status: statusUnavailable,
		Detail: "neither intel_pstate nor AMD cpufreq boost knob found; may be a VM or unsupported CPU",
	}
}

func checkLoadAvgLinux() Check {
	const path = "/proc/loadavg"
	data, err := os.ReadFile(path)
	if err != nil {
		return Check{
			Name:   "load average",
			Status: statusUnavailable,
			Detail: fmt.Sprintf("cannot read %s: %v", path, err),
		}
	}
	load, err := parseLoadAvg(strings.TrimSpace(string(data)))
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

func checkContainer() Check {
	// Docker leaves a marker file at the root.
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return Check{
			Name:   "container environment",
			Status: statusOK,
			Detail: "running inside Docker — use --cpuset-cpus and --cpus to pin and cap resources",
		}
	}

	// cgroup v1/v2: check for container runtime hints in /proc/1/cgroup.
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		for _, hint := range []string{"docker", "containerd", "kubepods", "lxc"} {
			if strings.Contains(content, hint) {
				return Check{
					Name:   "container environment",
					Status: statusOK,
					Detail: fmt.Sprintf("running inside a container (%s cgroup hint) — use --cpuset-cpus to pin resources", hint),
				}
			}
		}
	}

	return Check{
		Name:   "container environment",
		Status: statusOK,
		Detail: "not running in a container — direct host environment",
	}
}
