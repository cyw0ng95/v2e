package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config represents the configuration structure
type Config struct {
	Build    BuildSection            `json:"build"`
	Features map[string]ConfigOption `json:"features"`
	Profiles map[string]interface{}  `json:"profiles,omitempty"`
}

// BuildSection represents build-related configuration
type BuildSection struct {
	ConfigFile string `json:"config_file"`
}

// ConfigOption represents a single configuration option
type ConfigOption struct {
	Description string      `json:"description"`
	Type        string      `json:"type"` // "bool", "string", "int", etc.
	Default     interface{} `json:"default"`
	Values      []string    `json:"values,omitempty"`      // Available values for selection
	Method      string      `json:"method,omitempty"`      // "build-tag", "ldflags", "c-header", etc.
	Target      string      `json:"target,omitempty"`      // For ldflags: variable path, for build-tag: tag name pattern
	MajorClass  string      `json:"major_class,omitempty"` // Major classification of the config option
	MinorClass  string      `json:"minor_class,omitempty"` // Minor classification of the config option
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(path string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func GetDefaultConfigFromFile() (Config, error) {
	// Try multiple possible locations for the config spec file
	possiblePaths := []string{
		"config_spec.json",          // Current directory
		"../config_spec.json",       // One level up
		"../../config_spec.json",    // Two levels up (project root from tool/vconfig)
		"../../../config_spec.json", // Three levels up
	}

	var data []byte
	var err error

	// Try each path until we find the file
	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		// If the file doesn't exist anywhere, return an empty config
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func GetDefaultConfig() Config {
	// Try to load from spec file first
	specConfig, err := GetDefaultConfigFromFile()
	if err != nil {
		panic("Failed to load default config: " + err.Error())
	}

	return specConfig
}

// ValidTypes defines allowed configuration option types
var ValidTypes = map[string]bool{
	"bool":   true,
	"string": true,
	"int":    true,
}

// ValidMethods defines allowed configuration methods
var ValidMethods = map[string]bool{
	"build-tag": true,
	"ldflags":   true,
	"c-header":  true,
}

// ValidOptimizerPolicies defines valid values for optimizer policy
var ValidOptimizerPolicies = map[string]bool{
	"drop":   true,
	"wait":   true,
	"reject": true,
}

// ValidLogLevels defines valid log level values
var ValidLogLevels = map[string]bool{
	"DEBUG": true,
	"INFO":  true,
	"WARN":  true,
	"ERROR": true,
}

// validatePort validates that a string represents a valid port number or host:port
func validatePort(addr string) error {
	// Try to parse as host:port
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// Might be just a port number
		portNum, err := strconv.Atoi(addr)
		if err != nil {
			return fmt.Errorf("invalid port format: %s", addr)
		}
		if portNum < 1 || portNum > 65535 {
			return fmt.Errorf("port number out of range (1-65535): %d", portNum)
		}
		return nil
	}

	// Validate port part
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port number out of range (1-65535): %d", portNum)
	}

	// Validate host part (empty means bind to all interfaces)
	if host != "" && host != "0.0.0.0" && host != "localhost" && host != "127.0.0.1" {
		// Could be a hostname, we'll accept it but could add more validation
	}
	return nil
}

// validateOptionType validates type field and returns descriptive error
func validateOptionType(key string, optionType string) error {
	if optionType == "" {
		return fmt.Errorf("option '%s' is missing required field 'type'\n  Fix: Add \"type\": \"bool|string|int\" to the option definition", key)
	}
	if !ValidTypes[optionType] {
		return fmt.Errorf("option '%s' has invalid type '%s'\n  Valid types are: bool, string, int\n  Fix: Change \"type\" to one of the valid types", key, optionType)
	}
	return nil
}

// validateOptionDescription validates description field
func validateOptionDescription(key string, description string) error {
	if description == "" {
		return fmt.Errorf("option '%s' is missing required field 'description'\n  Fix: Add a descriptive \"description\" field to explain what this option does", key)
	}
	if len(description) < 10 {
		return fmt.Errorf("option '%s' description is too short (must be at least 10 characters)\n  Current: \"%s\"\n  Fix: Provide a more detailed description", key, description)
	}
	return nil
}

// validateDefaultValue validates that default value matches type
func validateDefaultValue(key string, optionType string, defaultValue interface{}) error {
	if defaultValue == nil {
		return nil // nil default is allowed
	}

	switch optionType {
	case "bool":
		if _, ok := defaultValue.(bool); !ok {
			return fmt.Errorf("option '%s' has type 'bool' but default value is not a boolean\n  Current default: %v (%T)\n  Fix: Set default to true or false", key, defaultValue, defaultValue)
		}
	case "int":
		// JSON numbers are float64 by default
		if f, ok := defaultValue.(float64); ok {
			if f != float64(int64(f)) {
				return fmt.Errorf("option '%s' has type 'int' but default value %v is not an integer\n  Fix: Use a whole number for default", key, defaultValue)
			}
		} else if _, ok := defaultValue.(int); !ok {
			return fmt.Errorf("option '%s' has type 'int' but default value is not an integer\n  Current default: %v (%T)\n  Fix: Set default to an integer value", key, defaultValue, defaultValue)
		}
	case "string":
		if _, ok := defaultValue.(string); !ok {
			return fmt.Errorf("option '%s' has type 'string' but default value is not a string\n  Current default: %v (%T)\n  Fix: Set default to a string value", key, defaultValue, defaultValue)
		}
	}

	return nil
}

// validateValuesEnum validates that values array contains valid enum values
func validateValuesEnum(key string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	// Check for duplicate values
	seen := make(map[string]bool)
	for _, v := range values {
		if seen[v] {
			return fmt.Errorf("option '%s' has duplicate value '%s' in values array\n  Fix: Remove duplicate entries", key, v)
		}
		seen[v] = true
	}

	return nil
}

// validateMethod validates method field if present
func validateMethod(key string, method string) error {
	if method == "" {
		return nil // method is optional
	}
	if !ValidMethods[method] {
		return fmt.Errorf("option '%s' has invalid method '%s'\n  Valid methods are: build-tag, ldflags, c-header\n  Fix: Change \"method\" to one of the valid methods or omit it", key, method)
	}
	return nil
}

// validateTarget validates target field for specific methods
func validateTarget(key string, method string, target string) error {
	if method == "ldflags" && target == "" {
		return fmt.Errorf("option '%s' uses method 'ldflags' but missing 'target' field\n  Fix: Add \"target\": \"import/path.to.Variable\" pointing to the Go variable to set", key)
	}
	if method == "c-header" && target == "" {
		return fmt.Errorf("option '%s' uses method 'c-header' but missing 'target' field\n  Fix: Add \"target\" specifying the header guard or identifier", key)
	}
	return nil
}

// validateSpecificConstraints validates option-specific constraints
func validateSpecificConstraints(key string, option ConfigOption) error {
	switch key {
	case "CONFIG_ACCESS_SERVERADDR":
		if addr, ok := option.Default.(string); ok {
			if err := validatePort(addr); err != nil {
				return fmt.Errorf("option '%s' has invalid server address: %w\n  Fix: Use format \"host:port\" where port is 1-65535 (e.g., \"0.0.0.0:8080\")", key, err)
			}
		}
	case "CONFIG_MIN_LOG_LEVEL":
		if level, ok := option.Default.(string); ok {
			if !ValidLogLevels[level] {
				return fmt.Errorf("option '%s' has invalid default value '%s'\n  Valid values: DEBUG, INFO, WARN, ERROR\n  Fix: Set default to one of the valid log levels", key, level)
			}
		}
		if len(option.Values) > 0 {
			for _, v := range option.Values {
				if !ValidLogLevels[v] {
					return fmt.Errorf("option '%s' has invalid value '%s' in values array\n  Valid values: DEBUG, INFO, WARN, ERROR\n  Fix: Remove invalid log level from values array", key, v)
				}
			}
		}
	case "CONFIG_OPTIMIZER_POLICY":
		if policy, ok := option.Default.(string); ok {
			if !ValidOptimizerPolicies[policy] {
				return fmt.Errorf("option '%s' has invalid default value '%s'\n  Valid values: drop, wait, reject\n  Fix: Set default to one of the valid policy values", key, policy)
			}
		}
	case "CONFIG_OPTIMIZER_BATCH", "CONFIG_OPTIMIZER_BUFFER", "CONFIG_OPTIMIZER_FLUSH",
		"CONFIG_WORKERPOOL_INITIAL_SIZE":
		// These should be non-negative integers
		if val, ok := option.Default.(float64); ok && val < 0 {
			return fmt.Errorf("option '%s' must have a non-negative default value\n  Current default: %v\n  Fix: Set default to 0 or a positive integer", key, val)
		}
		if val, ok := option.Default.(int); ok && val < 0 {
			return fmt.Errorf("option '%s' must have a non-negative default value\n  Current default: %v\n  Fix: Set default to 0 or a positive integer", key, val)
		}
	case "CONFIG_OPTIMIZER_WORKERS":
		if val, ok := option.Default.(float64); ok {
			if val < 1 {
				return fmt.Errorf("option '%s' must have at least 1 worker\n  Current default: %v\n  Fix: Set default to 1 or higher", key, val)
			}
		}
		if val, ok := option.Default.(int); ok {
			if val < 1 {
				return fmt.Errorf("option '%s' must have at least 1 worker\n  Current default: %v\n  Fix: Set default to 1 or higher", key, val)
			}
		}
	case "CONFIG_WORKERPOOL_MIN_SIZE", "CONFIG_WORKERPOOL_MAX_SIZE":
		// Min should be <= Max (need to check cross-field when both present)
		if val, ok := option.Default.(float64); ok && val > 10000 {
			return fmt.Errorf("option '%s' has unreasonably large default value %v\n  Fix: Use a reasonable value (suggested range: 0-1000)", key, val)
		}
		if val, ok := option.Default.(int); ok && val > 10000 {
			return fmt.Errorf("option '%s' has unreasonably large default value %v\n  Fix: Use a reasonable value (suggested range: 0-1000)", key, val)
		}
	}

	return nil
}

// validatePath validates that a path field is reasonable
func validatePath(key string, path string) error {
	if path == "" {
		return fmt.Errorf("option '%s' has empty path value\n  Fix: Provide a valid file or directory path", key)
	}

	// Check for obviously invalid paths
	if strings.Contains(path, "\\\\") {
		return fmt.Errorf("option '%s' contains invalid path separators '\\\\'\n  Fix: Use forward slashes '/' for paths", key)
	}

	// For absolute paths, basic validation
	if filepath.IsAbs(path) {
		// Check path doesn't contain suspicious patterns
		if strings.Contains(path, "..") {
			return fmt.Errorf("option '%s' contains parent directory reference '..' in absolute path\n  Current: %s\n  Fix: Use a clean absolute path without '..'", key, path)
		}
	}

	return nil
}

func (c *Config) Validate() error {
	// Validate build section
	if c.Build.ConfigFile == "" {
		return fmt.Errorf("build.config_file is empty\n  Fix: Set \"config_file\" to a valid filename (e.g., \".config\")")
	}

	// Validate each feature option
	for key, option := range c.Features {
		// Validate required fields
		if err := validateOptionType(key, option.Type); err != nil {
			return err
		}
		if err := validateOptionDescription(key, option.Description); err != nil {
			return err
		}

		// Validate default value matches type
		if err := validateDefaultValue(key, option.Type, option.Default); err != nil {
			return err
		}

		// Validate values array if present
		if len(option.Values) > 0 {
			if err := validateValuesEnum(key, option.Values); err != nil {
				return err
			}
		}

		// Validate method if present
		if err := validateMethod(key, option.Method); err != nil {
			return err
		}

		// Validate target for specific methods
		if err := validateTarget(key, option.Method, option.Target); err != nil {
			return err
		}

		// Validate specific constraints for known options
		if err := validateSpecificConstraints(key, option); err != nil {
			return err
		}

		// Validate path-type options
		if strings.Contains(key, "_DBPATH") || strings.Contains(key, "_DIR") || strings.Contains(key, "_STATICDIR") {
			if pathStr, ok := option.Default.(string); ok {
				if err := validatePath(key, pathStr); err != nil {
					return err
				}
			}
		}
	}

	// Cross-field validation: worker pool size constraints
	if minSizeOpt, hasMin := c.Features["CONFIG_WORKERPOOL_MIN_SIZE"]; hasMin {
		if maxSizeOpt, hasMax := c.Features["CONFIG_WORKERPOOL_MAX_SIZE"]; hasMax {
			minVal, minOk := minSizeOpt.Default.(float64)
			maxVal, maxOk := maxSizeOpt.Default.(float64)
			if minOk && maxOk && maxVal > 0 && minVal > maxVal {
				return fmt.Errorf("CONFIG_WORKERPOOL_MIN_SIZE (%v) cannot be greater than CONFIG_WORKERPOOL_MAX_SIZE (%v)\n  Fix: Set MIN_SIZE <= MAX_SIZE, or set MAX_SIZE to 0 for unlimited", minVal, maxVal)
			}
		}
	}

	return nil
}

// LoadSimpleConfig loads a simple config file with CONFIG_XXX=YYY format
func LoadSimpleConfig(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	result := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result, nil
}

// ConvertSimpleToFullConfig converts a simple config map to a full Config struct
// Uses the default config as a template and applies the simple config values
func ConvertSimpleToFullConfig(simpleConfig map[string]string) (*Config, error) {
	// Start with the default config as a template
	fullConfig := GetDefaultConfig()

	// Apply values from simple config to the full config
	for key, value := range simpleConfig {
		if option, exists := fullConfig.Features[key]; exists {
			// Convert the string value to the appropriate type based on the option type
			switch option.Type {
			case "bool":
				if value == "y" || value == "true" || value == "1" {
					option.Default = true
				} else {
					option.Default = false
				}
			case "string":
				// Remove quotes if present
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
					value = value[1 : len(value)-1]
				}
				option.Default = value
			case "int":
				if intValue, err := strconv.Atoi(value); err == nil {
					option.Default = intValue
				}
			default:
				// For unknown types, treat as string
				option.Default = value
			}
			fullConfig.Features[key] = option
		}
	}

	return &fullConfig, nil
}
