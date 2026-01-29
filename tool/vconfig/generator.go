package main

import (
	"fmt"
	"os"
	"strings"
)

// GenerateBuildFlags creates build flags from the configuration
func GenerateBuildFlags(config *Config) (string, error) {
	var flags []string

	for key, option := range config.Features {
		// For boolean options, add the flag if true
		if val, ok := option.Default.(bool); ok {
			if val {
				flags = append(flags, key)
			}
		} else {
			// For non-boolean options, we might create flags differently
			// For now, we'll skip non-boolean options in build flags
			continue
		}
	}

	if len(flags) == 0 {
		return "", nil
	}

	return strings.Join(flags, ","), nil
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
