package main

import (
	"testing"
)

// --- summarize ---

func TestSummarize(t *testing.T) {
	checks := []Check{
		{Status: statusOK},
		{Status: statusOK},
		{Status: statusWarn},
		{Status: statusUnavailable},
		{Status: statusUnavailable},
		{Status: statusUnavailable},
	}
	got := summarize(checks)
	if got.OK != 2 {
		t.Errorf("summarize OK = %d, want 2", got.OK)
	}
	if got.Warn != 1 {
		t.Errorf("summarize Warn = %d, want 1", got.Warn)
	}
	if got.Unavailable != 3 {
		t.Errorf("summarize Unavailable = %d, want 3", got.Unavailable)
	}
}

func TestSummarizeEmpty(t *testing.T) {
	got := summarize(nil)
	if got.OK != 0 || got.Warn != 0 || got.Unavailable != 0 {
		t.Errorf("summarize(nil) = %+v, want all zeros", got)
	}
}

// --- smtResult ---

func TestSMTResult(t *testing.T) {
	tests := []struct {
		value      string
		wantStatus Status
	}{
		{"off", statusOK},
		{"forceoff", statusOK},
		{"notsupported", statusUnavailable},
		{"on", statusWarn},
		{"notexist", statusWarn},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, _, _ := smtResult(tt.value)
			if got != tt.wantStatus {
				t.Errorf("smtResult(%q) status = %q, want %q", tt.value, got, tt.wantStatus)
			}
		})
	}
}

func TestSMTResultOnHasRemedy(t *testing.T) {
	_, _, remedy := smtResult("on")
	if remedy == "" {
		t.Error("smtResult(\"on\") remedy is empty; expected a sysfs command")
	}
}

// --- governorResult ---

func TestGovernorResult(t *testing.T) {
	tests := []struct {
		value      string
		wantStatus Status
	}{
		{"performance", statusOK},
		{"powersave", statusWarn},
		{"ondemand", statusWarn},
		{"schedutil", statusWarn},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, _, _ := governorResult(tt.value)
			if got != tt.wantStatus {
				t.Errorf("governorResult(%q) status = %q, want %q", tt.value, got, tt.wantStatus)
			}
		})
	}
}

func TestGovernorResultNonPerformanceHasRemedy(t *testing.T) {
	_, _, remedy := governorResult("powersave")
	if remedy == "" {
		t.Error("governorResult(\"powersave\") remedy is empty; expected a sysfs command")
	}
}

// --- turboIntelResult ---

func TestTurboIntelResult(t *testing.T) {
	tests := []struct {
		value      string
		wantStatus Status
	}{
		{"1", statusOK},   // 1 = turbo disabled = good
		{"0", statusWarn}, // 0 = turbo enabled = warn
		{"2", statusWarn}, // unexpected value
	}
	for _, tt := range tests {
		t.Run("no_turbo="+tt.value, func(t *testing.T) {
			got, _, _ := turboIntelResult(tt.value)
			if got != tt.wantStatus {
				t.Errorf("turboIntelResult(%q) status = %q, want %q", tt.value, got, tt.wantStatus)
			}
		})
	}
}

func TestTurboIntelEnabledHasRemedy(t *testing.T) {
	_, _, remedy := turboIntelResult("0")
	if remedy == "" {
		t.Error("turboIntelResult(\"0\") remedy is empty; expected a sysfs command")
	}
}

// --- turboAMDResult ---

func TestTurboAMDResult(t *testing.T) {
	tests := []struct {
		value      string
		wantStatus Status
	}{
		{"0", statusOK},   // 0 = boost disabled = good
		{"1", statusWarn}, // 1 = boost enabled = warn
		{"x", statusWarn}, // unexpected
	}
	for _, tt := range tests {
		t.Run("boost="+tt.value, func(t *testing.T) {
			got, _, _ := turboAMDResult(tt.value)
			if got != tt.wantStatus {
				t.Errorf("turboAMDResult(%q) status = %q, want %q", tt.value, got, tt.wantStatus)
			}
		})
	}
}

// --- parseLoadAvg (/proc/loadavg format) ---

func TestParseLoadAvg(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"0.52 0.41 0.30 1/423 12345", 0.52, false},
		{"2.00 1.50 1.20 2/512 99", 2.00, false},
		{"0.00 0.00 0.00 0/1 1", 0.00, false},
		{"", 0, true},
		{"notanumber 0 0", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseLoadAvg(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLoadAvg(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseLoadAvg(%q) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

// --- parseDarwinLoadAvg (sysctl vm.loadavg format) ---

func TestParseDarwinLoadAvg(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"{ 1.23 2.34 3.45 }", 1.23, false},
		{"{ 0.00 0.00 0.00 }", 0.00, false},
		{"{ 12.50 8.00 5.25 }", 12.50, false},
		{"{}", 0, true},
		{"{ }", 0, true},
		{"not valid", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDarwinLoadAvg(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDarwinLoadAvg(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseDarwinLoadAvg(%q) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

// --- loadAvgResult ---

func TestLoadAvgResult(t *testing.T) {
	tests := []struct {
		load       float64
		numCPU     int
		wantStatus Status
	}{
		// threshold = numCPU * 0.5
		{load: 0.1, numCPU: 8, wantStatus: statusOK},   // 0.1 < 4.0
		{load: 3.9, numCPU: 8, wantStatus: statusOK},   // 3.9 < 4.0
		{load: 4.1, numCPU: 8, wantStatus: statusWarn}, // 4.1 > 4.0
		{load: 0.6, numCPU: 1, wantStatus: statusWarn}, // 0.6 > 0.5
		{load: 0.5, numCPU: 1, wantStatus: statusOK},   // 0.5 == threshold → ok (not strictly greater)
		{load: 0.0, numCPU: 4, wantStatus: statusOK},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, _, _ := loadAvgResult(tt.load, tt.numCPU)
			if got != tt.wantStatus {
				t.Errorf("loadAvgResult(%.2f, %d) = %q, want %q", tt.load, tt.numCPU, got, tt.wantStatus)
			}
		})
	}
}

func TestLoadAvgResultHighHasRemedy(t *testing.T) {
	_, _, remedy := loadAvgResult(10.0, 1)
	if remedy == "" {
		t.Error("loadAvgResult(10.0, 1) remedy is empty; expected guidance on reducing background load")
	}
}
