package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	configFile      = flag.String("config", "config.json", "Path to the configuration file")
	generateDefaults = flag.Bool("generate-defaults", false, "Generate configuration with default values")
	tuiMode         = flag.Bool("tui", false, "Run in TUI mode")
	getBuildFlags   = flag.Bool("get-build-flags", false, "Output build flags based on configuration")
)

func main() {
	flag.Parse()

	if *generateDefaults {
		config, err := GetDefaultConfigFromFile()
		if err != nil {
			fmt.Printf("Error loading config spec: %v\n", err)
			config = GetDefaultConfig() // fallback to hardcoded defaults
		}
		err = GenerateSimpleConfig(&config, *configFile)
		if err != nil {
			fmt.Printf("Error generating simple config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated simple configuration file: %s\n", *configFile)
		return
	}

	if *tuiMode {
		err := runTUIInteractive()
		if err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *getBuildFlags {
		// Try to load simple config format first
		simpleConfig, err := LoadSimpleConfig(*configFile)
		if err != nil {
			// If simple config fails, try full config format
			config, err := LoadConfig(*configFile)
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
			
			flags, err := GenerateBuildFlags(config)
			if err != nil {
				fmt.Printf("Error generating build flags: %v\n", err)
				os.Exit(1)
			}
			
			fmt.Print(flags)
			return
		}
		
		// Convert simple config to full config
		config, err := ConvertSimpleToFullConfig(simpleConfig)
		if err != nil {
			fmt.Printf("Error converting simple config: %v\n", err)
			os.Exit(1)
		}
		
		flags, err := GenerateBuildFlags(config)
		if err != nil {
			fmt.Printf("Error generating build flags: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Print(flags)
		return
	}

	// Default behavior: show help
	flag.Usage()
}
