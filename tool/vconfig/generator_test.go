package main

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateBuildFlags(t *testing.T) {
	config := GetDefaultConfig()

	flags, err := GenerateBuildFlags(&config)
	if err != nil {
		t.Fatalf("Failed to generate build flags: %v", err)
	}

	// Should include CONFIG_ENABLE_METRICS since it's true by default
	expectedFlag := "CONFIG_ENABLE_METRICS"
	if !strings.Contains(flags, expectedFlag) {
		t.Errorf("Expected build flags to contain '%s', got '%s'", expectedFlag, flags)
	}

	// Should not include CONFIG_DEBUG_MODE since it's false by default
	unexpectedFlag := "CONFIG_DEBUG_MODE"
	if strings.Contains(flags, unexpectedFlag) {
		t.Errorf("Build flags should not contain '%s', got '%s'", unexpectedFlag, flags)
	}
}

func TestGenerateGoBuildTags(t *testing.T) {
	config := GetDefaultConfig()

	tags, err := GenerateGoBuildTags(&config)
	if err != nil {
		t.Fatalf("Failed to generate Go build tags: %v", err)
	}

	// Should start with -tags if there are any flags
	if !strings.HasPrefix(tags, "-tags ") {
		t.Errorf("Expected Go build tags to start with '-tags ', got '%s'", tags)
	}

	expectedFlag := "CONFIG_ENABLE_METRICS"
	if !strings.Contains(tags, expectedFlag) {
		t.Errorf("Expected build tags to contain '%s', got '%s'", expectedFlag, tags)
	}
}

func TestGenerateCHeader(t *testing.T) {
	config := GetDefaultConfig()

	tempFile, err := os.CreateTemp("", "test-header-*.h")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	err = GenerateCHeader(&config, tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to generate C header: %v", err)
	}

	// Read the generated file
	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read generated header: %v", err)
	}

	headerStr := string(content)

	// Check for required elements
	if !strings.Contains(headerStr, "#ifndef _VCONFIG_H_") {
		t.Error("Generated header should contain '#ifndef _VCONFIG_H_'")
	}

	if !strings.Contains(headerStr, "#define _VCONFIG_H_") {
		t.Error("Generated header should contain '#define _VCONFIG_H_'")
	}

	if !strings.Contains(headerStr, "#define CONFIG_ENABLE_METRICS 1") {
		t.Error("Generated header should contain '#define CONFIG_ENABLE_METRICS 1'")
	}

	if !strings.Contains(headerStr, "#undef CONFIG_DEBUG_MODE") {
		t.Error("Generated header should contain '#undef CONFIG_DEBUG_MODE'")
	}

	if !strings.Contains(headerStr, "// Minimum log level for the application") {
		t.Error("Generated header should contain description comments")
	}
}

func TestGenerateBuildFlagsEmpty(t *testing.T) {
	// Create a config with no enabled boolean features
	config := GetDefaultConfig()

	// Disable all boolean features
	for key, option := range config.Features {
		if _, ok := option.Default.(bool); ok {
			option.Default = false
			config.Features[key] = option
		}
	}

	flags, err := GenerateBuildFlags(&config)
	if err != nil {
		t.Fatalf("Failed to generate build flags: %v", err)
	}

	if flags != "" {
		t.Errorf("Expected empty build flags for config with no enabled features, got '%s'", flags)
	}
}
