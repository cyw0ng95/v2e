package monitor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
)

type FlameGraphConfig struct {
	OutputPath       string
	Duration         time.Duration
	SampleRate       int
	Frequency        int
	IncludeKernel    bool
	IncludeGoroutine bool
	FilterProcesses  []string
}

type FlameGraphGenerator struct {
	config  FlameGraphConfig
	profile *os.File
	running bool
}

func NewFlameGraphGenerator(config FlameGraphConfig) *FlameGraphGenerator {
	if config.OutputPath == "" {
		config.OutputPath = "/tmp/v2e-flamegraph"
	}

	if config.Duration == 0 {
		config.Duration = 30 * time.Second
	}

	if config.SampleRate == 0 {
		config.SampleRate = 99
	}

	if config.Frequency == 0 {
		config.Frequency = 99
	}

	return &FlameGraphGenerator{
		config: config,
	}
}

func (fg *FlameGraphGenerator) Start() error {
	if fg.running {
		return fmt.Errorf("flame graph generation already running")
	}

	outputFile := filepath.Join(fg.config.OutputPath, fmt.Sprintf("profile-%d.prof", time.Now().Unix()))

	if err := os.MkdirAll(fg.config.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	profileFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}

	fg.profile = profileFile

	if err := pprof.StartCPUProfile(profileFile); err != nil {
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	fg.running = true
	return nil
}

func (fg *FlameGraphGenerator) Stop() error {
	if !fg.running {
		return fmt.Errorf("flame graph generation not running")
	}

	pprof.StopCPUProfile()
	fg.running = false

	if fg.profile != nil {
		if err := fg.profile.Close(); err != nil {
			return fmt.Errorf("failed to close profile file: %w", err)
		}
		fg.profile = nil
	}

	return nil
}

func (fg *FlameGraphGenerator) GenerateFlameGraph() (string, error) {
	profileFile := filepath.Join(fg.config.OutputPath, fmt.Sprintf("profile-%d.prof", time.Now().Unix()))
	svgFile := filepath.Join(fg.config.OutputPath, fmt.Sprintf("flamegraph-%d.svg", time.Now().Unix()))

	if _, err := os.Stat(profileFile); os.IsNotExist(err) {
		return "", fmt.Errorf("profile file not found: %s", profileFile)
	}

	if !hasFlameGraphTool() {
		if err := fg.generateBasicSVG(profileFile, svgFile); err != nil {
			return "", fmt.Errorf("failed to generate basic SVG: %w", err)
		}
	} else {
		if err := fg.generateFlameGraphSVG(profileFile, svgFile); err != nil {
			return "", fmt.Errorf("failed to generate flame graph: %w", err)
		}
	}

	return svgFile, nil
}

func (fg *FlameGraphGenerator) generateFlameGraphSVG(profileFile, svgFile string) error {
	cmd := exec.Command("flamegraph", "-o", svgFile, profileFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("flamegraph failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (fg *FlameGraphGenerator) generateBasicSVG(profileFile, svgFile string) error {
	svgContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="800">
  <rect width="100%%" height="100%%" fill="#ffffff"/>
  <text x="600" y="400" font-size="16" text-anchor="middle" fill="#000000">
    CPU Profile
  </text>
  <text x="600" y="425" font-size="12" text-anchor="middle" fill="#666666">
    Install flamegraph tool for detailed visualization
  </text>
</svg>
`)

	svgFileHandle, err := os.Create(svgFile)
	if err != nil {
		return fmt.Errorf("failed to create SVG file: %w", err)
	}
	defer svgFileHandle.Close()

	if _, err := svgFileHandle.WriteString(svgContent); err != nil {
		return fmt.Errorf("failed to write SVG: %w", err)
	}

	return nil
}

func (fg *FlameGraphGenerator) CaptureHeapProfile(duration time.Duration) (string, error) {
	outputFile := filepath.Join(fg.config.OutputPath, fmt.Sprintf("heap-%d.prof", time.Now().Unix()))

	if err := os.MkdirAll(fg.config.OutputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	profileFile, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to create heap profile file: %w", err)
	}
	defer profileFile.Close()

	if err := pprof.WriteHeapProfile(profileFile); err != nil {
		return "", fmt.Errorf("failed to write heap profile: %w", err)
	}

	return outputFile, nil
}

func (fg *FlameGraphGenerator) CaptureGoroutineProfile() (string, error) {
	outputFile := filepath.Join(fg.config.OutputPath, fmt.Sprintf("goroutine-%d.prof", time.Now().Unix()))

	if err := os.MkdirAll(fg.config.OutputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	profileFile, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to create goroutine profile file: %w", err)
	}
	defer profileFile.Close()

	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return "", fmt.Errorf("goroutine profile not available")
	}

	if err := profile.WriteTo(profileFile, 0); err != nil {
		return "", fmt.Errorf("failed to write goroutine profile: %w", err)
	}

	return outputFile, nil
}

func (fg *FlameGraphGenerator) CaptureBlockProfile() (string, error) {
	runtime.SetBlockProfileRate(1)

	outputFile := filepath.Join(fg.config.OutputPath, fmt.Sprintf("block-%d.prof", time.Now().Unix()))

	if err := os.MkdirAll(fg.config.OutputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	profileFile, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to create block profile file: %w", err)
	}
	defer profileFile.Close()

	profile := pprof.Lookup("block")
	if profile == nil {
		return "", fmt.Errorf("block profile not available")
	}

	if err := profile.WriteTo(profileFile, 0); err != nil {
		return "", fmt.Errorf("failed to write block profile: %w", err)
	}

	return outputFile, nil
}

func (fg *FlameGraphGenerator) CaptureMutexProfile() (string, error) {
	runtime.SetMutexProfileFraction(1)

	outputFile := filepath.Join(fg.config.OutputPath, fmt.Sprintf("mutex-%d.prof", time.Now().Unix()))

	if err := os.MkdirAll(fg.config.OutputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	profileFile, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to create mutex profile file: %w", err)
	}
	defer profileFile.Close()

	profile := pprof.Lookup("mutex")
	if profile == nil {
		return "", fmt.Errorf("mutex profile not available")
	}

	if err := profile.WriteTo(profileFile, 0); err != nil {
		return "", fmt.Errorf("failed to write mutex profile: %w", err)
	}

	return outputFile, nil
}

func (fg *FlameGraphGenerator) CaptureAllProfiles() (map[string]string, error) {
	profiles := make(map[string]string)

	profileFile, err := fg.GenerateFlameGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CPU flame graph: %w", err)
	}
	profiles["cpu"] = profileFile

	heapFile, err := fg.CaptureHeapProfile(fg.config.Duration)
	if err != nil {
		return nil, fmt.Errorf("failed to capture heap profile: %w", err)
	}
	profiles["heap"] = heapFile

	goroutineFile, err := fg.CaptureGoroutineProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to capture goroutine profile: %w", err)
	}
	profiles["goroutine"] = goroutineFile

	blockFile, err := fg.CaptureBlockProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to capture block profile: %w", err)
	}
	profiles["block"] = blockFile

	mutexFile, err := fg.CaptureMutexProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to capture mutex profile: %w", err)
	}
	profiles["mutex"] = mutexFile

	return profiles, nil
}

func (fg *FlameGraphGenerator) IsRunning() bool {
	return fg.running
}

func (fg *FlameGraphGenerator) GetConfig() FlameGraphConfig {
	return fg.config
}

func hasFlameGraphTool() bool {
	_, err := exec.LookPath("flamegraph")
	return err == nil
}

func NewDefaultFlameGraphConfig() FlameGraphConfig {
	return FlameGraphConfig{
		OutputPath:       "/tmp/v2e-flamegraph",
		Duration:         30 * time.Second,
		SampleRate:       99,
		Frequency:        99,
		IncludeKernel:    false,
		IncludeGoroutine: true,
		FilterProcesses:  []string{},
	}
}
