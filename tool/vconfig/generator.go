package main

import (
	"fmt"
	"os"
	"strings"
)

// GenerateBuildFlags creates build flags from the configuration
func GenerateBuildFlags(config *Config) (string, error) {
	// Preallocate slice for flags for performance
	flags := make([]string, 0, len(config.Features))
	for key, option := range config.Features {
		// For boolean options, add the flag if true
		if val, ok := option.Default.(bool); ok {
			if val {
				flags = append(flags, key)
			}
		} else {
			// For string options, check the method property to decide whether to include in build flags
			if option.Method != "build-tag" {
				continue
			}
			// If explicitly marked as build-tag, add it using the target pattern or key as fallback
			if option.Target != "" {
				flags = append(flags, option.Target)
			} else {
				flags = append(flags, key)
			}
		}
	}
	if len(flags) == 0 {
		return "", nil
	}
	// Use strings.Builder for efficient concatenation
	var sb strings.Builder
	for i, flag := range flags {
		sb.WriteString(flag)
		if i < len(flags)-1 {
			sb.WriteByte(',')
		}
	}
	return sb.String(), nil
}

// GenerateGoBuildTags creates Go build tags from the configuration
func GenerateGoBuildTags(config *Config) (string, error) {
	flags, err := GenerateBuildFlags(config)
	if err != nil {
		return "", err
	}

	if flags == "" {
		return "", nil
	}

	return "-tags " + flags, nil
}

// GenerateLdflags creates linker flags from the configuration
func GenerateLdflags(config *Config) (string, error) {
	// Preallocate slice for ldflags for performance
	ldflags := make([]string, 0, len(config.Features))
	for _, option := range config.Features {
		// Handle string options that should be injected via ldflags
		if strVal, ok := option.Default.(string); ok {
			if option.Method == "ldflags" && option.Target != "" {
				ldflags = append(ldflags, "-X '")
				ldflags = append(ldflags, option.Target)
				ldflags = append(ldflags, "=")
				ldflags = append(ldflags, strVal)
				ldflags = append(ldflags, "'")
			}
		}
	}
	if len(ldflags) == 0 {
		return "", nil
	}
	// Use strings.Builder for efficient concatenation
	var sb strings.Builder
	for i := 0; i < len(ldflags); i += 5 {
		sb.WriteString(ldflags[i])   // -X '
		sb.WriteString(ldflags[i+1]) // target
		sb.WriteString(ldflags[i+2]) // =
		sb.WriteString(ldflags[i+3]) // value
		sb.WriteString(ldflags[i+4]) // '
		if i+5 < len(ldflags) {
			sb.WriteByte(' ')
		}
	}
	return sb.String(), nil
}

// GenerateCHeader creates a C header file from the configuration
func GenerateCHeader(config *Config, outputPath string) error {
	var content strings.Builder

	content.WriteString("#ifndef _VCONFIG_H_\n")
	content.WriteString("#define _VCONFIG_H_\n\n")

	for key, option := range config.Features {
		content.WriteString("// " + option.Description + "\n")
		if val, ok := option.Default.(bool); ok {
			if val {
				content.WriteString("#define " + key + " 1\n")
			} else {
				content.WriteString("#undef " + key + "\n")
			}
		} else {
			// Handle non-boolean values
			switch v := option.Default.(type) {
			case string:
				content.WriteString("#define " + key + " \"" + v + "\"\n")
			case int, int64, float64:
				content.WriteString("#define " + key + " " + fmt.Sprintf("%v", v) + "\n")
			default:
				content.WriteString("#define " + key + " " + fmt.Sprintf("%v", v) + "\n")
			}
		}
		content.WriteString("\n")
	}

	content.WriteString("#endif // _VCONFIG_H_\n")

	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

// GenerateSimpleConfig creates a simple config file with CONFIG_XXX=YYY format
func GenerateSimpleConfig(config *Config, outputPath string) error {
	var content strings.Builder

	for key, option := range config.Features {
		if val, ok := option.Default.(bool); ok {
			if val {
				content.WriteString(fmt.Sprintf("%s=y\n", key))
			} else {
				content.WriteString(fmt.Sprintf("%s=n\n", key))
			}
		} else {
			// Handle non-boolean values
			switch v := option.Default.(type) {
			case string:
				content.WriteString(fmt.Sprintf("%s=\"%s\"\n", key, v))
			case int, int64, float64:
				content.WriteString(fmt.Sprintf("%s=%v\n", key, v))
			default:
				content.WriteString(fmt.Sprintf("%s=%v\n", key, v))
			}
		}
	}

	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

// GenerateDetailedConfigMapping creates a detailed mapping of config options to build flags
func GenerateDetailedConfigMapping(config *Config) (string, error) {
	var result strings.Builder
	result.WriteString("Config to Build Flags Mapping:\n")
	result.WriteString("==============================\n")

	for key, option := range config.Features {
		if val, ok := option.Default.(bool); ok {
			if val {
				result.WriteString(fmt.Sprintf("%s (%s) -> INCLUDED in build flags\n", key, option.Description))
			} else {
				result.WriteString(fmt.Sprintf("%s (%s) -> NOT included in build flags\n", key, option.Description))
			}
		} else {
			// Check if it's a string option that gets converted based on method
			if strVal, ok := option.Default.(string); ok {
				method := option.Method
				if method == "" {
					method = "not-used" // default if no method specified
				}

				switch option.Method {
				case "ldflags":
					result.WriteString(fmt.Sprintf("%s (%s) -> Type: %T, Value: %s -> Method: %s, Target: %s\n", key, option.Description, option.Default, strVal, option.Method, option.Target))
				case "build-tag":
					result.WriteString(fmt.Sprintf("%s (%s) -> Type: %T, Value: %s -> Method: %s, Target: %s\n", key, option.Description, option.Default, strVal, option.Method, option.Target))
				default:
					result.WriteString(fmt.Sprintf("%s (%s) -> Type: %T, Value: %v -> Method: %s (not used in build flags)\n", key, option.Description, option.Default, option.Default, method))
				}
			} else {
				result.WriteString(fmt.Sprintf("%s (%s) -> Type: %T, Value: %v -> Method: %s (not used in build flags)\n", key, option.Description, option.Default, option.Default, option.Method))
			}
		}
	}

	return result.String(), nil
}
