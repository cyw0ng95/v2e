package ssg

import (
	"archive/tar"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// ExtractSSGPackage extracts the SSG tar.gz package to the target directory
func ExtractSSGPackage(tarGzData []byte, targetDir string, logger *common.Logger) error {
	logger.Info(LogMsgSSGPackageExtracting, targetDir)

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create a gzip reader
	gzr, err := gzip.NewReader(strings.NewReader(string(tarGzData)))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create a tar reader
	tr := tar.NewReader(gzr)

	// Extract each file
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Construct target path
		target := filepath.Join(targetDir, header.Name)

		// Ensure the target is within targetDir (prevent path traversal)
		if !strings.HasPrefix(target, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}

		case tar.TypeReg:
			// Create file
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", target, err)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", target, err)
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("failed to write file %s: %w", target, err)
			}
			f.Close()
		}
	}

	logger.Info(LogMsgSSGPackageExtracted)
	return nil
}

// ParseSSGBenchmark parses an SSG DataStream XML file and extracts the XCCDF Benchmark
func ParseSSGBenchmark(xmlPath string, logger *common.Logger) (*SSGBenchmark, error) {
	logger.Info(LogMsgSSGParsingFile, xmlPath)

	// Open the XML file
	f, err := os.Open(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open XML file: %w", err)
	}
	defer f.Close()

	// Since SSG files are SCAP DataStream files, we need to extract the embedded XCCDF
	// The Benchmark is embedded within ds:component elements
	// For simplicity, we'll parse the entire file and extract Benchmark elements
	
	// Read file content
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read XML file: %w", err)
	}

	// Find the Benchmark element in the DataStream
	// The Benchmark is within a ds:component element
	// We'll use a simple approach: extract the Benchmark XML and parse it
	
	benchmarkStart := strings.Index(string(data), "<xccdf-1.2:Benchmark")
	benchmarkEnd := strings.Index(string(data), "</xccdf-1.2:Benchmark>")
	
	if benchmarkStart == -1 || benchmarkEnd == -1 {
		return nil, fmt.Errorf("no XCCDF Benchmark found in file")
	}

	// Extract Benchmark XML (include closing tag)
	benchmarkXML := string(data[benchmarkStart : benchmarkEnd+len("</xccdf-1.2:Benchmark>")])
	
	// Replace xccdf-1.2: prefix with empty string for easier parsing
	benchmarkXML = strings.ReplaceAll(benchmarkXML, "xccdf-1.2:", "")
	benchmarkXML = strings.ReplaceAll(benchmarkXML, "html:", "")
	
	// Parse the Benchmark
	var benchmark SSGBenchmark
	if err := xml.Unmarshal([]byte(benchmarkXML), &benchmark); err != nil {
		return nil, fmt.Errorf("failed to parse Benchmark XML: %w", err)
	}

	logger.Info(LogMsgSSGBenchmarkParsed, benchmark.ID)
	logger.Info(LogMsgSSGProfilesLoaded, len(benchmark.Profiles), xmlPath)
	
	// Count rules (including nested groups)
	ruleCount := countRules(&benchmark)
	logger.Info(LogMsgSSGRulesLoaded, ruleCount, xmlPath)

	return &benchmark, nil
}

// countRules recursively counts all rules in a benchmark
func countRules(benchmark *SSGBenchmark) int {
	count := len(benchmark.Rules)
	for _, group := range benchmark.Groups {
		count += countGroupRules(&group)
	}
	return count
}

// countGroupRules recursively counts rules in a group
func countGroupRules(group *SSGGroup) int {
	count := len(group.Rules)
	for _, subGroup := range group.Groups {
		count += countGroupRules(&subGroup)
	}
	return count
}

// FindSSGXMLFiles finds all SSG DataStream XML files in a directory
func FindSSGXMLFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasPrefix(info.Name(), "ssg-") && strings.HasSuffix(info.Name(), "-ds.xml") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}
