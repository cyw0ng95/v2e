package capec

// These variables are injected at build time via ldflags
var (
	buildCAPECXMLPath  = "assets/capec_contents_latest.xml" // Default CAPEC XML path, can be overridden with -ldflags "-X capec.buildCAPECXMLPath=assets/capec_contents_latest.xml"
	buildCAPECXSDPath  = "assets/capec_schema_latest.xsd"   // Default CAPEC XSD path, can be overridden with -ldflags "-X capec.buildCAPECXSDPath=assets/capec_schema_latest.xsd"
	buildXSDValidation = false                              // Default XSD validation setting, can be overridden with -ldflags "-X capec.buildXSDValidation=true"
)

// DefaultBuildCAPECXMLPath returns the default CAPEC XML path based on build configuration
func DefaultBuildCAPECXMLPath() string {
	return buildCAPECXMLPath
}

// DefaultBuildCAPECXSDPath returns the default CAPEC XSD path based on build configuration
func DefaultBuildCAPECXSDPath() string {
	return buildCAPECXSDPath
}

// DefaultBuildXSDValidation returns the default XSD validation setting based on build configuration
func DefaultBuildXSDValidation() bool {
	return buildXSDValidation
}
