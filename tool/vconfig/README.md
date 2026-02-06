# vconfig - Configuration Management Tool

vconfig is a configuration management tool that allows you to define structured configuration files and generate different compilation flavors during build time. It features a TUI interface similar to the Linux kernel's kconfig system.

## Features

- JSON-based configuration files with schema validation
- Terminal User Interface (TUI) for interactive configuration
- Generation of build flags for Go and C projects
- Profile-based configurations (development, production, etc.)
- Type-safe configuration options with descriptions

## Usage

### Command Line Options

```bash
Usage of ./vconfig:
  -config string
        Path to the configuration file (default "config.json")
  -generate-defaults
        Generate configuration with default values
  -get-build-flags
        Output build flags based on configuration
  -tui
        Run in TUI mode
```

### Generate Default Configuration

```bash
./vconfig -generate-defaults
```

### Run Interactive TUI

```bash
./vconfig -tui
```

### Get Build Flags

```bash
./vconfig -get-build-flags
```

## Integration with Build System

The build system integrates vconfig via the `-c` flag:

```bash
./build.sh -c  # Runs vconfig TUI if no .config exists in .build directory
```

## Configuration Schema

The configuration file uses the following JSON schema:

```json
{
  "build": {
    "config_file": ".config"
  },
  "features": {
    "FEATURE_NAME": {
      "description": "Feature description",
      "type": "bool|string|int",
      "default": true,
      "values": ["available", "values", "for", "selection"]
    }
  },
  "profiles": {
    "development": {
      "FEATURE_NAME": true
    },
    "production": {
      "FEATURE_NAME": false
    }
  }
}
```

## Build Integration

The tool supports generating various build artifacts:

- **Go Build Tags**: Generated as `-tags FEATURE1,FEATURE2,...`
- **C Header Files**: Generated as `#define FEATURE_NAME value` pairs
- **Build Flags**: Comma-separated list of enabled features
- **Simple Config Format**: Generates `.config` files with `CONFIG_XXX=YYY` format

## Simple Config Format

The tool generates a simple configuration file format for build systems:

```
CONFIG_MIN_LOG_LEVEL="INFO"
```

## License

MIT