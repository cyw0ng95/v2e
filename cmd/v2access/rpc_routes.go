package main

// RPCRouteMapping maps path-based RPC endpoints to method/target
// Format: /rpc/{resource}/{action} -> {method, target}
type RPCRouteMapping struct {
	Method string // RPC method name (e.g., "RPCGetCVE")
	Target string // Target service ("local", "meta", "sysmon", "analysis")
}

// rpcRoutes maps URL paths to RPC method and target
// Key format: "{resource}/{action}" (e.g., "cve/get")
var rpcRoutes = map[string]RPCRouteMapping{
	// CVE endpoints (target: local)
	"cve/get":    {Method: "RPCGetCVE", Target: "local"},
	"cve/create": {Method: "RPCCreateCVE", Target: "local"},
	"cve/update": {Method: "RPCUpdateCVE", Target: "local"},
	"cve/delete": {Method: "RPCDeleteCVE", Target: "local"},
	"cve/list":   {Method: "RPCListCVEs", Target: "local"},
	"cve/count":  {Method: "RPCCountCVEs", Target: "local"},

	// CWE endpoints (target: local)
	"cwe/get":    {Method: "RPCGetCWEByID", Target: "local"},
	"cwe/list":   {Method: "RPCListCWEs", Target: "local"},
	"cwe/import": {Method: "RPCImportCWEs", Target: "local"},

	// CWE View endpoints (target: local)
	"cwe-view/save":      {Method: "RPCSaveCWEView", Target: "local"},
	"cwe-view/get":       {Method: "RPCGetCWEViewByID", Target: "local"},
	"cwe-view/list":      {Method: "RPCListCWEViews", Target: "local"},
	"cwe-view/delete":    {Method: "RPCDeleteCWEView", Target: "local"},
	"cwe-view/start-job": {Method: "RPCStartCWEViewJob", Target: "meta"},
	"cwe-view/stop-job":  {Method: "RPCStopCWEViewJob", Target: "meta"},

	// CAPEC endpoints (target: local)
	"capec/get":          {Method: "RPCGetCAPECByID", Target: "local"},
	"capec/list":         {Method: "RPCListCAPECs", Target: "local"},
	"capec/import":       {Method: "RPCImportCAPECs", Target: "local"},
	"capec/force-import": {Method: "RPCForceImportCAPECs", Target: "local"},
	"capec/metadata":     {Method: "RPCGetCAPECCatalogMeta", Target: "local"},

	// ATT&CK endpoints (target: local)
	"attack/technique":        {Method: "RPCGetAttackTechnique", Target: "local"},
	"attack/tactic":           {Method: "RPCGetAttackTactic", Target: "local"},
	"attack/mitigation":       {Method: "RPCGetAttackMitigation", Target: "local"},
	"attack/software":         {Method: "RPCGetAttackSoftware", Target: "local"},
	"attack/group":            {Method: "RPCGetAttackGroup", Target: "local"},
	"attack/technique-by-id":  {Method: "RPCGetAttackTechniqueByID", Target: "local"},
	"attack/tactic-by-id":     {Method: "RPCGetAttackTacticByID", Target: "local"},
	"attack/mitigation-by-id": {Method: "RPCGetAttackMitigationByID", Target: "local"},
	"attack/software-by-id":   {Method: "RPCGetAttackSoftwareByID", Target: "local"},
	"attack/group-by-id":      {Method: "RPCGetAttackGroupByID", Target: "local"},
	"attack/techniques":       {Method: "RPCListAttackTechniques", Target: "local"},
	"attack/tactics":          {Method: "RPCListAttackTactics", Target: "local"},
	"attack/mitigations":      {Method: "RPCListAttackMitigations", Target: "local"},
	"attack/softwares":        {Method: "RPCListAttackSoftware", Target: "local"},
	"attack/groups":           {Method: "RPCListAttackGroups", Target: "local"},
	"attack/import":           {Method: "RPCImportATTACKs", Target: "local"},
	"attack/import-metadata":  {Method: "RPCGetAttackImportMetadata", Target: "local"},

	// ASVS endpoints (target: local)
	"asvs/list":   {Method: "RPCListASVS", Target: "local"},
	"asvs/get":    {Method: "RPCGetASVSByID", Target: "local"},
	"asvs/import": {Method: "RPCImportASVS", Target: "local"},

	// CCE endpoints (target: local)
	"cce/get":        {Method: "RPCGetCCEByID", Target: "local"},
	"cce/list":       {Method: "RPCListCCEs", Target: "local"},
	"cce/import":     {Method: "RPCImportCCEs", Target: "local"},
	"cce/import-one": {Method: "RPCImportCCE", Target: "local"},
	"cce/count":      {Method: "RPCCountCCEs", Target: "local"},
	"cce/delete":     {Method: "RPCDeleteCCE", Target: "local"},
	"cce/update":     {Method: "RPCUpdateCCE", Target: "local"},

	// Session/Job endpoints (target: meta)
	"session/start":       {Method: "RPCStartSession", Target: "meta"},
	"session/start-typed": {Method: "RPCStartTypedSession", Target: "meta"},
	"session/stop":        {Method: "RPCStopSession", Target: "meta"},
	"session/status":      {Method: "RPCGetSessionStatus", Target: "meta"},
	"job/pause":           {Method: "RPCPauseJob", Target: "meta"},
	"job/resume":          {Method: "RPCResumeJob", Target: "meta"},

	// SSG endpoints (target: local)
	"ssg/import-guide":      {Method: "RPCSSGImportGuide", Target: "local"},
	"ssg/import-table":      {Method: "RPCSSGImportTable", Target: "local"},
	"ssg/guide":             {Method: "RPCSSGGetGuide", Target: "local"},
	"ssg/guides":            {Method: "RPCSSGListGuides", Target: "local"},
	"ssg/tables":            {Method: "RPCSSGListTables", Target: "local"},
	"ssg/table":             {Method: "RPCSSGGetTable", Target: "local"},
	"ssg/table-entries":     {Method: "RPCSSGGetTableEntries", Target: "local"},
	"ssg/tree":              {Method: "RPCSSGGetTree", Target: "local"},
	"ssg/tree-node":         {Method: "RPCSSGGetTreeNode", Target: "local"},
	"ssg/group":             {Method: "RPCSSGGetGroup", Target: "local"},
	"ssg/child-groups":      {Method: "RPCSSGGetChildGroups", Target: "local"},
	"ssg/rule":              {Method: "RPCSSGGetRule", Target: "local"},
	"ssg/rules":             {Method: "RPCSSGListRules", Target: "local"},
	"ssg/child-rules":       {Method: "RPCSSGGetChildRules", Target: "local"},
	"ssg/import-manifest":   {Method: "RPCSSGImportManifest", Target: "local"},
	"ssg/manifests":         {Method: "RPCSSGListManifests", Target: "local"},
	"ssg/manifest":          {Method: "RPCSSGGetManifest", Target: "local"},
	"ssg/profiles":          {Method: "RPCSSGListProfiles", Target: "local"},
	"ssg/profile":           {Method: "RPCSSGGetProfile", Target: "local"},
	"ssg/profile-rules":     {Method: "RPCSSGGetProfileRules", Target: "local"},
	"ssg/import-datastream": {Method: "RPCSSGImportDataStream", Target: "local"},
	"ssg/datastreams":       {Method: "RPCSSGListDataStreams", Target: "local"},
	"ssg/datastream":        {Method: "RPCSSGGetDataStream", Target: "local"},
	"ssg/ds-profiles":       {Method: "RPCSSGListDSProfiles", Target: "local"},
	"ssg/ds-profile":        {Method: "RPCSSGGetDSProfile", Target: "local"},
	"ssg/ds-profile-rules":  {Method: "RPCSSGGetDSProfileRules", Target: "local"},
	"ssg/ds-groups":         {Method: "RPCSSGListDSGroups", Target: "local"},
	"ssg/ds-rules":          {Method: "RPCSSGListDSRules", Target: "local"},
	"ssg/ds-rule":           {Method: "RPCSSGGetDSRule", Target: "local"},
	"ssg/cross-references":  {Method: "RPCSSGGetCrossReferences", Target: "local"},
	"ssg/find-related":      {Method: "RPCSSGFindRelatedObjects", Target: "local"},

	// SSG Job endpoints (target: meta)
	"ssg/job/start":  {Method: "RPCSSGStartImportJob", Target: "meta"},
	"ssg/job/stop":   {Method: "RPCSSGStopImportJob", Target: "meta"},
	"ssg/job/pause":  {Method: "RPCSSGPauseImportJob", Target: "meta"},
	"ssg/job/resume": {Method: "RPCSSGResumeImportJob", Target: "meta"},
	"ssg/job/status": {Method: "RPCSSGGetImportStatus", Target: "meta"},

	// Bookmark endpoints (target: local)
	"bookmark/create": {Method: "RPCCreateBookmark", Target: "local"},
	"bookmark/get":    {Method: "RPCGetBookmark", Target: "local"},
	"bookmark/update": {Method: "RPCUpdateBookmark", Target: "local"},
	"bookmark/delete": {Method: "RPCDeleteBookmark", Target: "local"},
	"bookmark/list":   {Method: "RPCListBookmarks", Target: "local"},

	// Note endpoints (target: local)
	"note/add":         {Method: "RPCAddNote", Target: "local"},
	"note/get":         {Method: "RPCGetNote", Target: "local"},
	"note/update":      {Method: "RPCUpdateNote", Target: "local"},
	"note/delete":      {Method: "RPCDeleteNote", Target: "local"},
	"note/by-bookmark": {Method: "RPCGetNotesByBookmark", Target: "local"},

	// Memory Card endpoints (target: local)
	"memory-card/create": {Method: "RPCCreateMemoryCard", Target: "local"},
	"memory-card/get":    {Method: "RPCGetMemoryCard", Target: "local"},
	"memory-card/update": {Method: "RPCUpdateMemoryCard", Target: "local"},
	"memory-card/delete": {Method: "RPCDeleteMemoryCard", Target: "local"},
	"memory-card/list":   {Method: "RPCListMemoryCards", Target: "local"},
	"memory-card/rate":   {Method: "RPCRateMemoryCard", Target: "local"},

	// GLC Graph endpoints (target: local)
	"glc/graph/create":      {Method: "RPCGLCGraphCreate", Target: "local"},
	"glc/graph/get":         {Method: "RPCGLCGraphGet", Target: "local"},
	"glc/graph/update":      {Method: "RPCGLCGraphUpdate", Target: "local"},
	"glc/graph/delete":      {Method: "RPCGLCGraphDelete", Target: "local"},
	"glc/graph/list":        {Method: "RPCGLCGraphList", Target: "local"},
	"glc/graph/list-recent": {Method: "RPCGLCGraphListRecent", Target: "local"},

	// GLC Version endpoints (target: local)
	"glc/version/get":     {Method: "RPCGLCVersionGet", Target: "local"},
	"glc/version/list":    {Method: "RPCGLCVersionList", Target: "local"},
	"glc/version/restore": {Method: "RPCGLCVersionRestore", Target: "local"},

	// GLC Preset endpoints (target: local)
	"glc/preset/create": {Method: "RPCGLCPresetCreate", Target: "local"},
	"glc/preset/get":    {Method: "RPCGLCPresetGet", Target: "local"},
	"glc/preset/update": {Method: "RPCGLCPresetUpdate", Target: "local"},
	"glc/preset/delete": {Method: "RPCGLCPresetDelete", Target: "local"},
	"glc/preset/list":   {Method: "RPCGLCPresetList", Target: "local"},

	// GLC Share endpoints (target: local)
	"glc/share/create": {Method: "RPCGLCShareCreateLink", Target: "local"},
	"glc/share/get":    {Method: "RPCGLCShareGetShared", Target: "local"},
	"glc/share/embed":  {Method: "RPCGLCShareGetEmbedData", Target: "local"},

	// Analysis endpoints (target: analysis)
	"analysis/stats":         {Method: "RPCGetGraphStats", Target: "analysis"},
	"analysis/node/add":      {Method: "RPCAddNode", Target: "analysis"},
	"analysis/edge/add":      {Method: "RPCAddEdge", Target: "analysis"},
	"analysis/node/get":      {Method: "RPCGetNode", Target: "analysis"},
	"analysis/neighbors":     {Method: "RPCGetNeighbors", Target: "analysis"},
	"analysis/path/find":     {Method: "RPCFindPath", Target: "analysis"},
	"analysis/nodes/by-type": {Method: "RPCGetNodesByType", Target: "analysis"},
	"analysis/status":        {Method: "RPCGetUEEStatus", Target: "analysis"},
	"analysis/graph/build":   {Method: "RPCBuildCVEGraph", Target: "analysis"},
	"analysis/graph/clear":   {Method: "RPCClearGraph", Target: "analysis"},
	"analysis/fsm/state":     {Method: "RPCGetFSMState", Target: "analysis"},
	"analysis/fsm/pause":     {Method: "RPCPauseAnalysis", Target: "analysis"},
	"analysis/fsm/resume":    {Method: "RPCResumeAnalysis", Target: "analysis"},
	"analysis/graph/save":    {Method: "RPCSaveGraph", Target: "analysis"},
	"analysis/graph/load":    {Method: "RPCLoadGraph", Target: "analysis"},

	// System endpoints (target: sysmon)
	"system/metrics": {Method: "RPCGetSysMetrics", Target: "sysmon"},

	// ETL endpoints (target: meta)
	"etl/tree":               {Method: "RPCGetEtlTree", Target: "meta"},
	"etl/provider/start":     {Method: "RPCStartProvider", Target: "meta"},
	"etl/provider/pause":     {Method: "RPCPauseProvider", Target: "meta"},
	"etl/provider/stop":      {Method: "RPCStopProvider", Target: "meta"},
	"etl/performance-policy": {Method: "RPCUpdatePerformancePolicy", Target: "meta"},
	"etl/kernel-metrics":     {Method: "RPCGetKernelMetrics", Target: "meta"},
}

// GetRPCRoute looks up the RPC method and target for a given path
// Returns the mapping and whether it was found
func GetRPCRoute(path string) (RPCRouteMapping, bool) {
	mapping, found := rpcRoutes[path]
	return mapping, found
}
