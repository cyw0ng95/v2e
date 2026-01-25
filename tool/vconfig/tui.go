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
		// Load existing config
		loadedConfig, err := LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		config = loadedConfig
	} else {
		// Create default config
		defaultConfig := GetDefaultConfig()
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

	// Create a list of configuration options
	list := widgets.NewList()
	list.Title = "Configuration Options"
	list.Rows = make([]string, 0, len(config.Features))
	
	// Add all features to the list
	for key, option := range config.Features {
		status := "disabled"
		if val, ok := option.Default.(bool); ok && val {
			status = "enabled"
		}
		list.Rows = append(list.Rows, fmt.Sprintf("[%s](fg:blue) - %s [%s](fg:yellow)", key, option.Description, status))
	}
	
	list.SelectedRowStyle = termui.NewStyle(termui.ColorWhite, termui.ColorBlue)
	list.WrapText = false

	// Create instructions
	instructions := widgets.NewParagraph()
	instructions.Text = "Press q to quit, j/k to navigate, space to toggle, s to save"
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
			// Toggle the selected option
			if len(config.Features) > 0 {
				i := 0
				for key, option := range config.Features {
					if i == selectedIndex {
						// Toggle boolean options
						if val, ok := option.Default.(bool); ok {
							option.Default = !val
							
							// Update the list display
							status := "disabled"
							if !val { // After toggling, !val is the new value
								status = "enabled"
							}
							list.Rows[i] = fmt.Sprintf("[%s](fg:blue) - %s [%s](fg:yellow)", key, option.Description, status)
							
							// Update the config
							config.Features[key] = option
							break
						}
					}
					i++
				}
				termui.Render(grid)
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
				
				// Wait a moment then restore
				go func() {
					select {
					case <-time.After(time.Second * 2):
						grid.Set(
							termui.NewRow(1.0/10, title),
							termui.NewRow(8.0/10, list),
							termui.NewRow(1.0/10, instructions),
						)
						termui.Render(grid)
					}
				}()
			} else {
				// Show success message
				successMsg := widgets.NewParagraph()
				successMsg.Text = "Configuration saved successfully!"
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
				
				// Wait a moment then restore
				go func() {
					select {
					case <-time.After(time.Second * 2):
						grid.Set(
							termui.NewRow(1.0/10, title),
							termui.NewRow(8.0/10, list),
							termui.NewRow(1.0/10, instructions),
						)
						termui.Render(grid)
					}
				}()
			}
		}
	}
}