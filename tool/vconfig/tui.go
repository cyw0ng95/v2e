package main

import (
	"encoding/json"
	"fmt"
	"os"
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

	// Create UI elements
	termWidth, termHeight := termui.TerminalDimensions()

	// Create a title
	title := widgets.NewParagraph()
	title.Text = "vconfig - Configuration Editor"
	title.TextStyle.Fg = termui.ColorGreen
	title.Border = false

	// Create three columns and a tweak panel
	majorClassList := widgets.NewList()
	majorClassList.Title = "Major Classes"
	majorClassList.SelectedRowStyle = termui.NewStyle(termui.ColorWhite, termui.ColorRed)
	majorClassList.WrapText = false

	minorClassList := widgets.NewList()
	minorClassList.Title = "Minor Classes"
	minorClassList.SelectedRowStyle = termui.NewStyle(termui.ColorWhite, termui.ColorYellow)
	minorClassList.WrapText = false

	optionsList := widgets.NewList()
	optionsList.Title = "Options"
	optionsList.SelectedRowStyle = termui.NewStyle(termui.ColorWhite, termui.ColorBlue)
	optionsList.WrapText = false

	tweakPanel := widgets.NewParagraph()
	tweakPanel.Title = "Tweak Panel"
	tweakPanel.WrapText = true
	tweakPanel.Border = true

	instructions := widgets.NewParagraph()
	instructions.Text = "ARROW KEYS: Navigate | SPACE: Toggle/Edit | q: Quit | s: Save & Exit | p: Preview Config"
	instructions.Title = "Instructions"
	instructions.Border = false

	// State variables
	selectedMajor := 0
	selectedMinor := 0
	selectedOption := 0
	activePane := 0 // 0: major, 1: minor, 2: options

	// Populate major class list
	for _, majorClass := range groupOrder {
		majorClassList.Rows = append(majorClassList.Rows, majorClass)
	}

	// Helper function to get status string
	getStatusString := func(key string, option ConfigOption) string {
		switch v := option.Default.(type) {
		case bool:
			if v {
				return "enabled"
			} else {
				return "disabled"
			}
		case string:
			return v
		case int:
			return fmt.Sprintf("%d", v)
		case float64:
			return fmt.Sprintf("%g", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	// Update the minor class list based on selected major class
	updateMinorClasses := func() {
		if selectedMajor >= 0 && selectedMajor < len(groupOrder) {
			majorClass := groupOrder[selectedMajor]
			minorClassList.Rows = minorOrder[majorClass]
		} else {
			minorClassList.Rows = []string{}
		}
		minorClassList.SelectedRow = 0
		selectedMinor = 0
	}

	// Update the options list based on selected major and minor class
	updateOptions := func() {
		if selectedMajor >= 0 && selectedMajor < len(groupOrder) &&
			selectedMinor >= 0 && selectedMinor < len(minorOrder[groupOrder[selectedMajor]]) {
			majorClass := groupOrder[selectedMajor]
			minorClass := minorOrder[majorClass][selectedMinor]
			optionsList.Rows = []string{}
			for _, key := range groups[majorClass][minorClass] {
				status := getStatusString(key, config.Features[key])
				optionsList.Rows = append(optionsList.Rows, fmt.Sprintf("%s [%s]", key, status))
			}
		} else {
			optionsList.Rows = []string{}
		}
		selectedOption = 0
		optionsList.SelectedRow = 0
	}

	// Update the tweak panel with information about the selected option and editing interface
	updateTweakPanel := func() {
		if selectedMajor >= 0 && selectedMajor < len(groupOrder) &&
			selectedMinor >= 0 && selectedMinor < len(minorOrder[groupOrder[selectedMajor]]) &&
			selectedOption >= 0 && selectedOption < len(groups[groupOrder[selectedMajor]][minorOrder[groupOrder[selectedMajor]][selectedMinor]]) {
			majorClass := groupOrder[selectedMajor]
			minorClass := minorOrder[majorClass][selectedMinor]
			key := groups[majorClass][minorClass][selectedOption]
			option := config.Features[key]

			// Create different editing interfaces based on the option type
			switch option.Type {
			case "bool":
				value := "off"
				if val, ok := option.Default.(bool); ok && val {
					value = "on"
				}
				tweakPanel.Text = fmt.Sprintf(
					"NAME: %s\n\nDESCRIPTION:\n%s\n\nTYPE: %s\nCURRENT VALUE: %s\n\nACTION: Press SPACE to toggle (currently: %s)\n\nMETHOD: %s\nTARGET: %s",
					key, option.Description, option.Type, option.Default, value, option.Method, option.Target,
				)
			case "string":
				if len(option.Values) > 0 {
					// If there are predefined values, show cycling options
					var availableValues string
					for i, v := range option.Values {
						if i == 0 {
							availableValues = v
						} else {
							availableValues = availableValues + ", " + v
						}
					}
					tweakPanel.Text = fmt.Sprintf(
						"NAME: %s\n\nDESCRIPTION:\n%s\n\nTYPE: %s\nCURRENT VALUE: %s\n\nAVAILABLE VALUES: %s\n\nACTION: Press SPACE to cycle through values\n\nMETHOD: %s\nTARGET: %s",
						key, option.Description, option.Type, option.Default, availableValues, option.Method, option.Target,
					)
				} else {
					// For freeform strings, suggest an edit interface
					tweakPanel.Text = fmt.Sprintf(
						"NAME: %s\n\nDESCRIPTION:\n%s\n\nTYPE: %s\nCURRENT VALUE: %s\n\nACTION: String value (editing not implemented in TUI)\n\nMETHOD: %s\nTARGET: %s",
						key, option.Description, option.Type, option.Default, option.Method, option.Target,
					)
				}
			case "int":
				tweakPanel.Text = fmt.Sprintf(
					"NAME: %s\n\nDESCRIPTION:\n%s\n\nTYPE: %s\nCURRENT VALUE: %v\n\nACTION: Integer value (editing not implemented in TUI)\n\nMETHOD: %s\nTARGET: %s",
					key, option.Description, option.Type, option.Default, option.Method, option.Target,
				)
			default:
				tweakPanel.Text = fmt.Sprintf(
					"NAME: %s\n\nDESCRIPTION:\n%s\n\nTYPE: %s\nCURRENT VALUE: %v\n\nACTION: Unsupported type\n\nMETHOD: %s\nTARGET: %s",
					key, option.Description, option.Type, option.Default, option.Method, option.Target,
				)
			}
		} else {
			tweakPanel.Text = "No option selected\n\nUse arrow keys to navigate:\n  - Left/Right: Move between columns\n  - Up/Down: Select items\n  - Space: Toggle/edit value"
		}
	}

	// Update grid layout to show/hide details panel
	updateLayout := func() {
		// Update borders to highlight active pane
		majorClassList.BorderStyle = termui.NewStyle(termui.ColorWhite)
		minorClassList.BorderStyle = termui.NewStyle(termui.ColorWhite)
		optionsList.BorderStyle = termui.NewStyle(termui.ColorWhite)

		switch activePane {
		case 0: // Major Classes active
			majorClassList.BorderStyle = termui.NewStyle(termui.ColorRed)
		case 1: // Minor Classes active
			minorClassList.BorderStyle = termui.NewStyle(termui.ColorYellow)
		case 2: // Options active
			optionsList.BorderStyle = termui.NewStyle(termui.ColorCyan)
		}

		// Always show the tweak panel below the selectors
		majorClassList.SetRect(0, 1, int(float64(termWidth)*0.4), int(float64(termHeight)*0.6))
		minorClassList.SetRect(int(float64(termWidth)*0.4), 1, int(float64(termWidth)*0.7), int(float64(termHeight)*0.6))
		optionsList.SetRect(int(float64(termWidth)*0.7), 1, termWidth, int(float64(termHeight)*0.6))
		tweakPanel.SetRect(0, int(float64(termHeight)*0.6), termWidth, termHeight-2)
		instructions.SetRect(0, termHeight-2, termWidth, termHeight)
		termui.Render(title, majorClassList, minorClassList, optionsList, tweakPanel, instructions)
	}

	// Initial population
	if len(groupOrder) > 0 {
		updateMinorClasses()
		updateOptions()
		updateTweakPanel()
	}

	updateLayout()

	// Handle events
	uiEvents := termui.PollEvents()

	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return nil
		case "<Right>":
			if activePane < 2 {
				activePane++
			}
		case "<Left>":
			if activePane > 0 {
				activePane--
			}
		case "j", "<Down>":
			switch activePane {
			case 0: // Major class list
				if selectedMajor < len(groupOrder)-1 {
					selectedMajor++
					majorClassList.SelectedRow = selectedMajor
					updateMinorClasses()
					updateOptions()
					updateTweakPanel()
				}
			case 1: // Minor class list
				if selectedMinor < len(minorClassList.Rows)-1 {
					selectedMinor++
					minorClassList.SelectedRow = selectedMinor
					updateOptions()
					updateTweakPanel()
				}
			case 2: // Options list
				if selectedOption < len(optionsList.Rows)-1 {
					selectedOption++
					optionsList.SelectedRow = selectedOption
					updateTweakPanel()
				}
			}
			updateLayout()
		case "k", "<Up>":
			switch activePane {
			case 0: // Major class list
				if selectedMajor > 0 {
					selectedMajor--
					majorClassList.SelectedRow = selectedMajor
					updateMinorClasses()
					updateOptions()
					updateTweakPanel()
				}
			case 1: // Minor class list
				if selectedMinor > 0 {
					selectedMinor--
					minorClassList.SelectedRow = selectedMinor
					updateOptions()
					updateTweakPanel()
				}
			case 2: // Options list
				if selectedOption > 0 {
					selectedOption--
					optionsList.SelectedRow = selectedOption
					updateTweakPanel()
				}
			}
			updateLayout()
		case " ", "<Enter>":
			if activePane == 2 && selectedOption >= 0 &&
				selectedOption < len(groups[groupOrder[selectedMajor]][minorOrder[groupOrder[selectedMajor]][selectedMinor]]) {
				// Toggle/edit the selected option
				majorClass := groupOrder[selectedMajor]
				minorClass := minorOrder[majorClass][selectedMinor]
				key := groups[majorClass][minorClass][selectedOption]
				option := config.Features[key]

				// Toggle based on type
				if option.Type == "bool" {
					if val, ok := option.Default.(bool); ok {
						newVal := !val
						option.Default = newVal
						config.Features[key] = option
						// Update the display
						updateOptions()
						updateTweakPanel()
					} else {
						// Convert to bool if needed
						strVal := fmt.Sprintf("%v", option.Default)
						if strVal == "true" || strVal == "1" || strVal == "y" {
							option.Default = false
						} else {
							option.Default = true
						}
						config.Features[key] = option
						updateOptions()
						updateTweakPanel()
					}
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
						config.Features[key] = option
						updateOptions()
						updateTweakPanel()
					} else {
						// For string options without predefined values, we could implement an edit interface
						// For now, we'll update the tweak panel to show the option information
						updateTweakPanel()
					}
				} else if option.Type == "int" {
					// For integer options, we could implement an edit interface
					updateTweakPanel()
				}
				updateLayout()
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
				majorClassList.SetRect(0, 1, int(float64(termWidth)*0.4), termHeight-2)
				minorClassList.SetRect(int(float64(termWidth)*0.4), 1, int(float64(termWidth)*0.7), termHeight-2)
				optionsList.SetRect(int(float64(termWidth)*0.7), 1, termWidth, termHeight-2)
				errorMsg.SetRect(0, termHeight-2, termWidth, termHeight)
				instructions.SetRect(0, termHeight-1, termWidth, termHeight)
				termui.Render(title, majorClassList, minorClassList, optionsList, errorMsg, instructions)

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
				majorClassList.SetRect(0, 1, int(float64(termWidth)*0.4), termHeight-2)
				minorClassList.SetRect(int(float64(termWidth)*0.4), 1, int(float64(termWidth)*0.7), termHeight-2)
				optionsList.SetRect(int(float64(termWidth)*0.7), 1, termWidth, termHeight-2)
				successMsg.SetRect(0, termHeight-2, termWidth, termHeight)
				instructions.SetRect(0, termHeight-1, termWidth, termHeight)
				termui.Render(title, majorClassList, minorClassList, optionsList, successMsg, instructions)

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
				majorClassList.SetRect(0, 1, int(float64(termWidth)*0.4), termHeight-2)
				minorClassList.SetRect(int(float64(termWidth)*0.4), 1, int(float64(termWidth)*0.7), termHeight-2)
				optionsList.SetRect(int(float64(termWidth)*0.7), 1, termWidth, termHeight-2)
				errorMsg.SetRect(0, termHeight-2, termWidth, termHeight)
				instructions.SetRect(0, termHeight-1, termWidth, termHeight)
				termui.Render(title, majorClassList, minorClassList, optionsList, errorMsg, instructions)

				// Wait a moment then restore
				go func() {
					time.Sleep(time.Second * 3)
					updateLayout()
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
					majorClassList.SetRect(0, 1, int(float64(termWidth)*0.4), termHeight-2)
					minorClassList.SetRect(int(float64(termWidth)*0.4), 1, int(float64(termWidth)*0.7), termHeight-2)
					optionsList.SetRect(int(float64(termWidth)*0.7), 1, termWidth, termHeight-2)
					errorMsg.SetRect(0, termHeight-2, termWidth, termHeight)
					instructions.SetRect(0, termHeight-1, termWidth, termHeight)
					termui.Render(title, majorClassList, minorClassList, optionsList, errorMsg, instructions)

					// Wait a moment then restore
					go func() {
						time.Sleep(time.Second * 3)
						updateLayout()
					}()
				} else {
					// Show the config content
					configDisplay := widgets.NewParagraph()
					configDisplay.Text = string(content)
					configDisplay.Title = "Current Configuration (.config format)"
					configDisplay.TextStyle.Fg = termui.ColorCyan
					configDisplay.Border = true

					// Temporarily show config
					configDisplay.SetRect(0, 1, termWidth, termHeight-1)
					instructions.SetRect(0, termHeight-1, termWidth, termHeight)
					termui.Render(title, configDisplay, instructions)

					// Wait a moment then restore
					go func() {
						time.Sleep(time.Second * 5)
						updateLayout()
					}()
				}
			}
			// Clean up temp file
			os.Remove(tempConfigPath)
		}
	}
}
