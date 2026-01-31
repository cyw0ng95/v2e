package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// runTUIInteractive runs the terminal user interface for configuration
func runTUIInteractive() error {
	if err := termui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %w", err)
	}
	defer termui.Close()

	// Load existing config or create default
	configPath := ".build/.config"

	var config *Config
	if _, err := os.Stat(configPath); err == nil {
		// Try to load simple config format first
		simpleConfig, err := LoadSimpleConfig(configPath)
		if err == nil {
			// Convert simple config to full config
			loadedConfig, err := ConvertSimpleToFullConfig(simpleConfig)
			if err != nil {
				return fmt.Errorf("failed to convert simple config: %w", err)
			}
			config = loadedConfig
		} else {
			// If simple config fails, try full config format
			loadedConfig, err := LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}
			config = loadedConfig
		}
	} else {
		// Create default config
		defaultConfig, err := GetDefaultConfigFromFile()
		if err != nil {
			// If config_spec.json is not found, fall back to hardcoded defaults
			defaultConfig = GetDefaultConfig()
		}
		// Convert the map to the Config struct
		jsonData, err := json.Marshal(defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
		}

		config = &Config{}
		if err := json.Unmarshal(jsonData, config); err != nil {
			return fmt.Errorf("failed to unmarshal default config: %w", err)
		}

		// Create .build directory if it doesn't exist
		if err := os.MkdirAll(".build", 0755); err != nil {
			return fmt.Errorf("failed to create .build directory: %w", err)
		}

		// Save the default config
		if err := GenerateSimpleConfig(config, configPath); err != nil {
			return fmt.Errorf("failed to save default config: %w", err)
		}

		fmt.Printf("Created default config at %s\n", configPath)
	}

	// Create UI elements
	grid := termui.NewGrid()
	termWidth, termHeight := termui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	// Create a title
	title := widgets.NewParagraph()
	title.Text = "vconfig - Configuration Editor"
	title.TextStyle.Fg = termui.ColorGreen
	title.Border = false

	// Group features by major and minor class
	groups := make(map[string]map[string][]string)
	groupOrder := make([]string, 0)         // Track order of major classes
	minorOrder := make(map[string][]string) // Track order of minor classes within each major class

	// Create original ordered list of feature keys to maintain consistent ordering
	originalFeatureKeys := make([]string, 0, len(config.Features))
	for key := range config.Features {
		originalFeatureKeys = append(originalFeatureKeys, key)
	}

	for _, key := range originalFeatureKeys {
		option := config.Features[key]
		majorClass := option.MajorClass
		if majorClass == "" {
			majorClass = "Uncategorized" // Default group if no major class specified
		}
		minorClass := option.MinorClass
		if minorClass == "" {
			minorClass = "General" // Default minor class if none specified
		}

		// Initialize major class group if it doesn't exist
		if groups[majorClass] == nil {
			groups[majorClass] = make(map[string][]string)
			groupOrder = append(groupOrder, majorClass)
		}

		// Add to minor class within major class
		groups[majorClass][minorClass] = append(groups[majorClass][minorClass], key)

		// Track minor class order within major class
		found := false
		for _, existingMinor := range minorOrder[majorClass] {
			if existingMinor == minorClass {
				found = true
				break
			}
		}
		if !found {
			minorOrder[majorClass] = append(minorOrder[majorClass], minorClass)
		}
	}

	// Create a list of configuration options organized by major/minor class
	list := widgets.NewList()
	list.Title = "Configuration Options (Grouped by Class)"
	list.Rows = make([]string, 0)

	// Add features grouped by major and minor class
	for _, majorClass := range groupOrder {
		list.Rows = append(list.Rows, fmt.Sprintf("[=== %s ===](fg:green)", majorClass))
		for _, minorClass := range minorOrder[majorClass] {
			list.Rows = append(list.Rows, fmt.Sprintf("  [--- %s ---](fg:cyan)", minorClass))
			for _, key := range groups[majorClass][minorClass] {
				option := config.Features[key]
				status := "disabled"
				if val, ok := option.Default.(bool); ok && val {
					status = "enabled"
				} else if strVal, ok := option.Default.(string); ok {
					status = strVal
				} else if intVal, ok := option.Default.(int); ok {
					status = fmt.Sprintf("%d", intVal)
				}
				list.Rows = append(list.Rows, fmt.Sprintf("    [%s](fg:blue) - %s [%s](fg:yellow)", key, option.Description, status))
			}
		}
		list.Rows = append(list.Rows, "") // Add blank line between major groups
	}

	// Update featureKeys to match the new ordering for navigation purposes
	featureKeys := make([]string, 0)
	for _, majorClass := range groupOrder {
		for _, minorClass := range minorOrder[majorClass] {
			for _, key := range groups[majorClass][minorClass] {
				featureKeys = append(featureKeys, key)
			}
		}
	}

	list.SelectedRowStyle = termui.NewStyle(termui.ColorWhite, termui.ColorBlue)
	list.WrapText = false

	// Create instructions
	instructions := widgets.NewParagraph()
	instructions.Text = "Press q to quit, j/k to navigate, space to toggle, s to save & exit & show config, p to print config"
	instructions.Title = "Instructions"
	instructions.Border = false

	// Set up the grid layout
	grid.Set(
		termui.NewRow(1.0/10, title),
		termui.NewRow(8.0/10, list),
		termui.NewRow(1.0/10, instructions),
	)

	termui.Render(grid)

	// Handle events
	uiEvents := termui.PollEvents()
	selectedIndex := 0

	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return nil
		case "j", "<Down>":
			if selectedIndex < len(list.Rows)-1 {
				selectedIndex++
				list.SelectedRow = selectedIndex
				termui.Render(grid)
			}
		case "k", "<Up>":
			if selectedIndex > 0 {
				selectedIndex--
				list.SelectedRow = selectedIndex
				termui.Render(grid)
			}
		case " ", "<Enter>":
			// Toggle the selected option using the ordered featureKeys
			if len(list.Rows) > 0 && selectedIndex < len(featureKeys) {
				// Filter out the group header entries to get the actual config key
				actualConfigIndex := 0
				for i, rowText := range list.Rows {
					if i == selectedIndex {
						break
					}
					// Skip group headers and empty lines
					if !strings.HasPrefix(rowText, "[=== ") && !strings.HasPrefix(rowText, "  [--- ") && rowText != "" {
						actualConfigIndex++
					}
				}

				// Get the key from our reordered list
				if actualConfigIndex >= 0 && actualConfigIndex < len(featureKeys) {
					key := featureKeys[actualConfigIndex]
					option := config.Features[key]

					// Toggle based on type
					if val, ok := option.Default.(bool); ok {
						newVal := !val
						option.Default = newVal

						// Update the list display
						status := "disabled"
						if newVal {
							status = "enabled"
						}
						// Find the correct row to update based on the actual config index
						rowIndex := findActualRowIndex(list.Rows, actualConfigIndex)
						if rowIndex >= 0 && rowIndex < len(list.Rows) {
							list.Rows[rowIndex] = fmt.Sprintf("[%s](fg:blue) - %s [%s](fg:yellow)", key, option.Description, status)
						}

						// Update the config
						config.Features[key] = option
					} else if option.Type == "string" {
						// For string options, cycle through available values if any
						if len(option.Values) > 0 {
							currentStr := option.Default.(string)
							nextIndex := 0
							// Find current value index
							for i, v := range option.Values {
								if v == currentStr {
									nextIndex = (i + 1) % len(option.Values)
									break
								}
							}
							newValue := option.Values[nextIndex]
							option.Default = newValue

							// Update the list display
							rowIndex := findActualRowIndex(list.Rows, actualConfigIndex)
							if rowIndex >= 0 && rowIndex < len(list.Rows) {
								list.Rows[rowIndex] = fmt.Sprintf("[%s](fg:blue) - %s [%s](fg:yellow)", key, option.Description, newValue)
							}

							// Update the config
							config.Features[key] = option
						}
					}
					termui.Render(grid)
				}
			}
		case "s":
			// Save the configuration
			if err := GenerateSimpleConfig(config, configPath); err != nil {
				errorMsg := widgets.NewParagraph()
				errorMsg.Text = fmt.Sprintf("Error saving config: %v", err)
				errorMsg.Title = "Error"
				errorMsg.TextStyle.Fg = termui.ColorRed
				errorMsg.Border = true

				// Temporarily show error
				grid.Set(
					termui.NewRow(1.0/10, title),
					termui.NewRow(7.0/10, list),
					termui.NewRow(1.0/10, errorMsg),
					termui.NewRow(1.0/10, instructions),
				)
				termui.Render(grid)

				// Close termui and show error immediately
				termui.Close()
				fmt.Printf("Error saving config: %v\n", err)
				os.Exit(1)

				// This return will never be reached due to os.Exit, but included for compiler
				return fmt.Errorf("failed to save config: %w", err)
			} else {
				// Show success message and then print the config
				successMsg := widgets.NewParagraph()
				successMsg.Text = "Configuration saved successfully! Printing config..."
				successMsg.Title = "Success"
				successMsg.TextStyle.Fg = termui.ColorGreen
				successMsg.Border = true

				// Temporarily show success
				grid.Set(
					termui.NewRow(1.0/10, title),
					termui.NewRow(7.0/10, list),
					termui.NewRow(1.0/10, successMsg),
					termui.NewRow(1.0/10, instructions),
				)
				termui.Render(grid)

				// Now print the config that was written
				content, err := os.ReadFile(configPath)
				if err != nil {
					termui.Close()
					return fmt.Errorf("failed to read config file after saving: %w", err)
				}

				// Print the config content to stdout
				// First print the content, then close termui to avoid terminal issues
				fmt.Println("Current configuration:")
				fmt.Println("=====================")
				fmt.Println(string(content))

				// Close termui and exit immediately to ensure clean termination
				termui.Close()
				os.Exit(0)

				// This return will never be reached due to os.Exit, but included for compiler
				return nil
			}
		case "p":
			// Print the current config to stdout after exiting the TUI
			// Generate the config to a temporary location
			tempConfigPath := configPath + ".preview"
			if err := GenerateSimpleConfig(config, tempConfigPath); err != nil {
				// Show error if generation fails
				errorMsg := widgets.NewParagraph()
				errorMsg.Text = fmt.Sprintf("Error generating config preview: %v", err)
				errorMsg.Title = "Error"
				errorMsg.TextStyle.Fg = termui.ColorRed
				errorMsg.Border = true

				// Temporarily show error
				grid.Set(
					termui.NewRow(1.0/10, title),
					termui.NewRow(7.0/10, list),
					termui.NewRow(1.0/10, errorMsg),
					termui.NewRow(1.0/10, instructions),
				)
				termui.Render(grid)

				// Wait a moment then restore
				go func() {
					time.Sleep(time.Second * 3)
					grid.Set(
						termui.NewRow(1.0/10, title),
						termui.NewRow(8.0/10, list),
						termui.NewRow(1.0/10, instructions),
					)
					termui.Render(grid)
				}()
			} else {
				// Read and display the config content in the TUI
				content, err := os.ReadFile(tempConfigPath)
				if err != nil {
					// Handle read error
					errorMsg := widgets.NewParagraph()
					errorMsg.Text = fmt.Sprintf("Error reading config preview: %v", err)
					errorMsg.Title = "Error"
					errorMsg.TextStyle.Fg = termui.ColorRed
					errorMsg.Border = true

					// Temporarily show error
					grid.Set(
						termui.NewRow(1.0/10, title),
						termui.NewRow(7.0/10, list),
						termui.NewRow(1.0/10, errorMsg),
						termui.NewRow(1.0/10, instructions),
					)
					termui.Render(grid)

					// Wait a moment then restore
					go func() {
						time.Sleep(time.Second * 3)
						grid.Set(
							termui.NewRow(1.0/10, title),
							termui.NewRow(8.0/10, list),
							termui.NewRow(1.0/10, instructions),
						)
						termui.Render(grid)
					}()
				} else {
					// Show the config content
					configDisplay := widgets.NewParagraph()
					configDisplay.Text = string(content)
					configDisplay.Title = "Current Configuration (.config format)"
					configDisplay.TextStyle.Fg = termui.ColorCyan
					configDisplay.Border = true

					// Temporarily show config
					grid.Set(
						termui.NewRow(1.0/10, title),
						termui.NewRow(7.0/10, configDisplay),
						termui.NewRow(2.0/10, instructions),
					)
					termui.Render(grid)

					// Wait a moment then restore
					go func() {
						time.Sleep(time.Second * 5)
						grid.Set(
							termui.NewRow(1.0/10, title),
							termui.NewRow(8.0/10, list),
							termui.NewRow(1.0/10, instructions),
						)
						termui.Render(grid)
					}()
				}
				// Clean up temp file
				os.Remove(tempConfigPath)
			}
		}
	}
}

// findActualRowIndex finds the actual row index in the list for a given config option index
func findActualRowIndex(rows []string, configIndex int) int {
	actualIndex := 0
	for i, rowText := range rows {
		// Skip group headers and empty lines
		if !strings.HasPrefix(rowText, "[=== ") && !strings.HasPrefix(rowText, "  [--- ") && rowText != "" {
			if actualIndex == configIndex {
				return i
			}
			actualIndex++
		}
	}
	return -1 // Not found
}
