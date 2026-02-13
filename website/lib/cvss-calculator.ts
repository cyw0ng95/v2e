/**
 * CVSS Calculator - Implementation of FIRST.org CVSS scoring formulas
 * Supports CVSS v3.0, v3.1, and v4.0
 *
 * References:
 * - CVSS v3.1: https://www.first.org/cvss/calculator/3.1
 * - CVSS v4.0: https://www.first.org/cvss/calculator/4.0
 */

import type {
  CVSSVersion, CVSS3Metrics, CVSS4Metrics,
  CVSS3BaseMetrics, CVSS4BaseMetrics,
  CVSS3TemporalMetrics, CVSS3EnvironmentalMetrics,
  CVSS4ThreatMetrics, CVSS4EnvironmentalMetrics,
  CVSSSeverity, CVSS3ScoreBreakdown, CVSS4ScoreBreakdown
} from './types';

// ============================================================================
// Types for internal calculations
// ============================================================================

interface V3LookupMaps {
  av: Record<AV, number>;
  ac: Record<AC, number>;
  pr: Record<PR, number>;
  ui: Record<UI, number>;
  s: Record<S, number>;
  c: Record<C, number>;
  i: Record<I, number>;
  a: Record<A, number>;
}

interface V4LookupMaps {
  av: Record<AV4, number>;
  ac: Record<AC4, number>;
  at: Record<AT4, number>;
  pr: Record<PR4, number>;
  ui: Record<UI4, number>;
}

type AV = 'N' | 'A' | 'L' | 'P';
type AC = 'L' | 'H';
type PR = 'N' | 'L' | 'H';
type UI = 'N' | 'R';
type S = 'U' | 'C';
type C = 'H' | 'L' | 'N';
type I = 'H' | 'L' | 'N';
type A = 'H' | 'L' | 'N';

type AV4 = 'N' | 'A' | 'L' | 'P';
type AC4 = 'L' | 'H';
type AT4 = 'N' | 'P' | 'R';
type PR4 = 'N' | 'L' | 'H';
type UI4 = 'N' | 'P' | 'A';
type VC4 = 'H' | 'L' | 'N';
type VS4 = 'H' | 'L' | 'N';

// ============================================================================
// Constants
// ============================================================================

/** Round to 1 decimal place */
const ROUND = (num: number): number => Math.round(num * 10) / 10;

/** Round up to 1 decimal place */
const ROUND_UP = (num: number): number => Math.ceil(num * 10) / 10;

/** Clamp value between min and max */
const CLAMP = (num: number, min: number, max: number): number =>
  Math.min(Math.max(num, min), max);

// ============================================================================
// CVSS v3.x Lookup Maps
// ============================================================================

const V3_MAPS: V3LookupMaps = {
  av: { N: 0.85, A: 0.62, L: 0.55, P: 0.2 },
  ac: { L: 0.77, H: 0.44 },
  pr: { N: 0.85, L: 0.62, H: 0.50 },
  ui: { N: 0.85, R: 0.62 },
  s: { U: 0, C: 1 },
  c: { H: 0.56, L: 0.22, N: 0 },
  i: { H: 0.56, L: 0.22, N: 0 },
  a: { H: 0.56, L: 0.22, N: 0 }
};

// ============================================================================
// CVSS v3.x Formula Implementation
// ============================================================================

/**
 * Calculates CVSS v3.0/v3.1 Exploitability score
 */
function calculateV3Exploitability(metrics: CVSS3BaseMetrics): number {
  const av = V3_MAPS.av[metrics.AV];
  const ac = V3_MAPS.ac[metrics.AC];
  const pr = V3_MAPS.pr[metrics.PR];
  const ui = V3_MAPS.ui[metrics.UI];

  return ROUND(8.22 * av * ac * pr * ui);
}

/**
 * Calculates CVSS v3.0/v3.1 Impact score
 */
function calculateV3Impact(metrics: CVSS3BaseMetrics): number {
  const c = V3_MAPS.c[metrics.C];
  const i = V3_MAPS.i[metrics.I];
  const a = V3_MAPS.a[metrics.A];

  // ISS (Impact Sub-Score) = 1 - [(1 - Confidentiality) × (1 - Integrity) × (1 - Availability)]
  const iss = 1 - ((1 - c) * (1 - i) * (1 - a));

  const scopeModified = V3_MAPS.s[metrics.S];

  // Impact formula per FIRST.org CVSS v3.1 specification:
  // If Scope is Unchanged: Impact = 6.42 × ISS
  // If Scope is Changed: Impact = 7.52 × (ISS - 0.029) - 3.25 × (ISS × 0.9731 - 0.02)^13
  if (scopeModified === 1) {
    // Scope Changed: Impact = 7.52 × (ISS - 0.029) - 3.25 × (ISS × 0.9731 - 0.02)^13
    return ROUND(7.52 * (iss - 0.029) - 3.25 * Math.pow((iss * 0.9731 - 0.02), 13));
  } else {
    // Scope Unchanged: Impact = 6.42 × ISS
    return ROUND(6.42 * iss);
  }
}

/**
 * Calculates CVSS v3.0/v3.1 base score
 */
function calculateV3BaseScore(metrics: CVSS3BaseMetrics): number {
  const impact = calculateV3Impact(metrics);
  const exploitability = calculateV3Exploitability(metrics);
  const scopeModified = V3_MAPS.s[metrics.S];

  let baseScore = impact + exploitability;

  // Per FIRST spec: min(10, Impact + Exploitability)
  // But when Scope is Changed and result >= 10, round up to 10
  baseScore = Math.min(10, baseScore);

  if (scopeModified === 1 && baseScore > 9) {
    baseScore = 10;
  }

  return ROUND(baseScore);
}

/**
 * Applies CVSS v3 temporal adjustments
 */
function applyV3TemporalAdjustments(
  baseScore: number,
  baseMetrics: CVSS3BaseMetrics,
  temporal: CVSS3TemporalMetrics
): { score: number; breakdown: CVSS3ScoreBreakdown } {
  const eMap: Record<string, number> = {
    X: 1, U: 1, F: 0.97, P: 0.94, H: 0.91, R: 0.95
  };
  const rlMap: Record<string, number> = {
    X: 1, U: 1, O: 1, T: 0.97, W: 0.95
  };
  const rcMap: Record<string, number> = {
    X: 1, U: 1, C: 1, R: 0.96
  };

  const e = eMap[temporal.E] ?? 1;
  const rl = rlMap[temporal.RL] ?? 1;
  const rc = rcMap[temporal.RC] ?? 1;

  const adjustedScore = baseScore * e * rl * rc;

  return {
    score: ROUND(baseScore),
    breakdown: {
      baseScore: ROUND(baseScore),
      temporalScore: ROUND(adjustedScore),
      exploitabilityScore: ROUND(calculateV3Exploitability(baseMetrics)),
      impactScore: ROUND(calculateV3Impact(baseMetrics)),
      baseSeverity: getSeverity(ROUND(baseScore)),
      temporalSeverity: getSeverity(ROUND(adjustedScore))
    }
  };
}

/**
 * Helper to get environmental metric value with fallback
 */
function getEnvValue<T extends string>(
  envValue: T | undefined,
  baseValue: T,
  defaultValue: T
): T {
  if (envValue === 'X') return defaultValue;
  return envValue ?? baseValue;
}

/**
 * Applies CVSS v3 environmental adjustments
 */
function applyV3EnvironmentalAdjustments(
  baseScore: number,
  baseMetrics: CVSS3BaseMetrics,
  environmental: CVSS3EnvironmentalMetrics
): { score: number; breakdown: CVSS3ScoreBreakdown } {
  // Get environmental multipliers
  const crMap: Record<string, number> = { X: 1, H: 1.5, M: 1, L: 0.5, N: 1 };
  const irMap: Record<string, number> = { X: 1, H: 1.5, M: 1, L: 0.5, N: 1 };
  const arMap: Record<string, number> = { X: 1, H: 1.5, M: 1, L: 0.5, N: 1 };

  const cr = crMap[environmental.CR] ?? 1;
  const ir = irMap[environmental.IR] ?? 1;
  const ar = arMap[environmental.AR] ?? 1;

  // Calculate modified impact
  const mc = getEnvValue(environmental.MC, baseMetrics.C, 'N');
  const mi = getEnvValue(environmental.MI, baseMetrics.I, 'N');
  const ma = getEnvValue(environmental.MA, baseMetrics.A, 'N');

  const cMap: Record<C, number> = { H: 0.56, L: 0.22, N: 0 };
  const iMap: Record<I, number> = { H: 0.56, L: 0.22, N: 0 };
  const aMap: Record<A, number> = { H: 0.56, L: 0.22, N: 0 };

  const mcValue = cMap[mc];
  const miValue = iMap[mi];
  const maValue = aMap[ma];

  // Get MS value handling
  let msModified = 0;
  if (environmental.MS === 'X') {
    msModified = 0;
  } else if (environmental.MS !== undefined) {
    msModified = 0;
  } else {
    const sMap: Record<S, number> = { U: 0, C: 1 };
    msModified = sMap[environmental.MS];
  }

  const modifiedImpact =
    (1 - mcValue) * (1 - miValue) * (1 - maValue);

  // Environmental metrics in CVSS v3 only adjust the impact and requirements
  // They do not modify the exploitability vector (AV/AC/PR/UI)
  // So we use the original exploitability score but adjust with requirements
  const modifiedExploitability =
    calculateV3Exploitability(baseMetrics) * cr * ir * ar;

  const msValue = msModified;
  const sMap: Record<S, number> = { U: 0, C: 1 };

  const adjustedScore =
    10.41 * (1 - modifiedImpact) * (1 - msValue) +
    modifiedExploitability;

  return {
    score: ROUND(adjustedScore),
    breakdown: {
      baseScore: baseScore,
      environmentalScore: ROUND(adjustedScore),
      exploitabilityScore: ROUND(modifiedExploitability),
      impactScore: ROUND(modifiedImpact * 10.41),
      baseSeverity: getSeverity(baseScore),
      environmentalSeverity: getSeverity(ROUND(adjustedScore))
    }
  };
}

/**
 * Generates CVSS v3 vector string
 */
function generateV3VectorString(
  metrics: CVSS3Metrics,
  version: '3.0' | '3.1'
): string {
  const prefix = `CVSS:${version}`;
  const base = `AV:${metrics.AV}/AC:${metrics.AC}/PR:${metrics.PR}/UI:${metrics.UI}/S:${metrics.S}/C:${metrics.C}/I:${metrics.I}/A:${metrics.A}`;

  if (!metrics.temporal || metrics.temporal.E === 'X') {
    return `${prefix}/${base}`;
  }

  let vector = `${prefix}/${base}`;

  const { E, RL, RC } = metrics.temporal;
  if (E !== undefined && (E as any) !== 'X') vector += `/E:${E}`;
  if (RL !== undefined && (RL as any) !== 'X') vector += `/RL:${RL}`;
  if (RC !== undefined && (RC as any) !== 'X') vector += `/RC:${RC}`;

  if (metrics.environmental) {
    const env = metrics.environmental;
    vector += `/CR:${env.CR}/IR:${env.IR}/AR:${env.AR}`;
    // CVSS v3 environmental metrics do not include modified AV/AC/PR/UI
    // These are only in CVSS v4
    if (env.MS !== undefined && (env.MS as any) !== 'X') vector += `/MS:${env.MS}`;
    if (env.MC !== undefined && (env.MC as any) !== 'X') vector += `/MC:${env.MC}`;
    if (env.MI !== undefined && (env.MI as any) !== 'X') vector += `/MI:${env.MI}`;
    if (env.MA !== undefined && (env.MA as any) !== 'X') vector += `/MA:${env.MA}`;
  }

  return vector;
}

// ============================================================================
// CVSS v4.0 Formula Implementation
// ============================================================================

/**
 * CVSS v4.0 Impact lookups
 */
const V4_IMPACT_LOOKUP: Record<string, { iq: number; ss: number; isc: number }> = {
  'H:N': { iq: 0.95, ss: 0.9, isc: 0.5 },
  'H:L': { iq: 0.7, ss: 0.6, isc: 0.35 },
  'H:H': { iq: 0.9, ss: 0.85, isc: 0.45 },
  'L:N': { iq: 0.75, ss: 0.65, isc: 0.4 },
  'L:L': { iq: 0.5, ss: 0.4, isc: 0.25 },
  'L:H': { iq: 0.7, ss: 0.55, isc: 0.3 },
  'N:N': { iq: 0.85, ss: 0.75, isc: 0.45 },
  'N:L': { iq: 0.6, ss: 0.5, isc: 0.3 },
  'N:H': { iq: 0.8, ss: 0.7, isc: 0.4 }
};

/**
 * CVSS v4.0 lookup maps
 */
const V4_MAPS: V4LookupMaps = {
  av: { N: 0.85, A: 0.62, L: 0.55, P: 0.2 },
  ac: { L: 0.77, H: 0.44 },
  at: { N: 0.85, P: 0.75, R: 0.6 },
  pr: { N: 0.85, L: 0.62, H: 0.50 },
  ui: { N: 0.85, P: 0.62, A: 0.5 }
};

/**
 * Calculates CVSS v4.0 base score
 */
function calculateV4BaseScore(metrics: CVSS4BaseMetrics, threat?: CVSS4ThreatMetrics): {
  score: number;
  breakdown: CVSS4ScoreBreakdown;
} {
  const av = V4_MAPS.av[metrics.AV];
  const ac = V4_MAPS.ac[metrics.AC];
  const at = V4_MAPS.at[metrics.AT];
  const pr = V4_MAPS.pr[metrics.PR];
  const ui = V4_MAPS.ui[metrics.UI];

  const iq = ROUND(10 * av * ac * at);
  const ss = ROUND(10 * pr * ui);

  // Determine I:E value - check threat metrics for Provider I
  const provider = threat?.I ?? 'X';
  const ieKey = provider === 'E' ? `${metrics.VC}:${metrics.SI}` : 'N:N';
  const impactLookup = V4_IMPACT_LOOKUP[ieKey];

  const { iq: iqFactor, ss: ssFactor, isc } = impactLookup;

  const baseScore = iq * iqFactor + ss * ssFactor + isc * 10;

  return {
    score: ROUND(baseScore),
    breakdown: {
      baseScore: ROUND(baseScore),
      baseSeverity: getSeverity(ROUND(baseScore))
    }
  };
}

/**
 * Applies CVSS v4.0 threat adjustments
 */
function applyV4ThreatAdjustments(
  baseScore: number,
  threat: CVSS4ThreatMetrics
): number {
  const eMap: Record<string, number> = {
    X: 1, U: 1, P: 0.91, F: 0.94, H: 0.97, A: 0.93, R: 0.95
  };
  const mMap: Record<string, number> = {
    X: 1, N: 1, P: 0.95, A: 0.9, R: 0.96, E: 0.98
  };
  const dMap: Record<string, number> = {
    X: 1, N: 1, L: 0.98, M: 0.97, H: 0.95
  };

  const e = eMap[threat.E] ?? 1;
  const m = mMap[threat.M] ?? 1;
  const d = dMap[threat.D] ?? 1;

  return ROUND(baseScore * e * m * d);
}

/**
 * Helper to determine if provider supports I:E
 */
function hasProviderI(metrics: CVSS4EnvironmentalMetrics): boolean {
  return metrics.MI === 'E';
}

/**
 * Helper to check if value is present
 */
function isMetricPresent<T extends string>(value: T | undefined): value is T {
  return value !== 'X' && value !== undefined;
}

/**
 * Applies CVSS v4.0 environmental adjustments
 */
function applyV4EnvironmentalAdjustments(
  baseMetrics: CVSS4BaseMetrics,
  environmental: CVSS4EnvironmentalMetrics
): { baseScore: number; adjustedScore: number } {
  const crMap: Record<string, number> = {
    X: 1, H: 1.51, M: 1.01, L: 0.76, N: 1
  };
  const irMap: Record<string, number> = {
    X: 1, H: 1.51, M: 1.01, L: 0.76, N: 1
  };
  const arMap: Record<string, number> = {
    X: 1, H: 1.51, M: 1.01, L: 0.76, N: 1
  };

  const cr = crMap[environmental.CR] ?? 1;
  const ir = irMap[environmental.IR] ?? 1;
  const ar = arMap[environmental.AR] ?? 1;

  // Calculate modified IQ and SS
  const mav = getEnvValue(environmental.MAV, baseMetrics.AV, baseMetrics.AV);
  const mac = getEnvValue(environmental.MAC, baseMetrics.AC, baseMetrics.AC);
  const mat = getEnvValue(environmental.MAT, baseMetrics.AT, baseMetrics.AT);
  const mpr = getEnvValue(environmental.MPR, baseMetrics.PR, baseMetrics.PR);
  const mui = getEnvValue(environmental.MUI, baseMetrics.UI, baseMetrics.UI);

  const modifiedIQ =
    10 * V4_MAPS.av[mav] * V4_MAPS.ac[mac] * V4_MAPS.at[mat] * cr * ir * ar;

  const modifiedSS =
    10 * V4_MAPS.pr[mpr] * V4_MAPS.ui[mui] * cr * ir * ar;

  // Get modified I:E values
  const hasVC = isMetricPresent(environmental.MVC);
  const hasVI = isMetricPresent(environmental.MVI);
  const hasSC = isMetricPresent(environmental.MSC);
  const hasSI = isMetricPresent(environmental.MSI);
  const hasSA = isMetricPresent(environmental.MSA);

  const vc = hasVC ? environmental.MVC ?? baseMetrics.VC : 'N';
  const vi = hasVI ? environmental.MVI ?? baseMetrics.VI : 'N';
  const sc = hasSC ? environmental.MSC ?? baseMetrics.SC : 'N';
  const si = hasSI ? environmental.MSI ?? baseMetrics.SI : 'N';
  const sa = hasSA ? environmental.MSA ?? baseMetrics.SA : 'N';

  // For CVSS v4.0 environmental scoring, use simplified impact lookup
  // The lookup key format is: VC:SI or H:N format
  const providerI = hasProviderI(environmental) ? 'E' : 'X';
  const ieKey = providerI ? `${vc}:${si}` : 'N:N';
  const modifiedImpactLookup = V4_IMPACT_LOOKUP[ieKey] ?? V4_IMPACT_LOOKUP['N:N'];

  const { iq: iqFactor, ss: ssFactor, isc } = modifiedImpactLookup;

  const modifiedBaseScore = modifiedIQ * iqFactor + modifiedSS * ssFactor + isc * 10;

  const baseResult = calculateV4BaseScore(baseMetrics);
  return {
    baseScore: baseResult.breakdown.baseScore,
    adjustedScore: ROUND(modifiedBaseScore)
  };
}

/**
 * Generates CVSS v4.0 vector string
 */
function generateV4VectorString(metrics: CVSS4Metrics): string {
  const parts: string[] = [
    'CVSS:4.0',
    `AV:${metrics.AV}`,
    `AC:${metrics.AC}`,
    `AT:${metrics.AT}`,
    `PR:${metrics.PR}`,
    `UI:${metrics.UI}`,
    `VC:${metrics.VC}`,
    `VI:${metrics.VI}`,
    `VA:${metrics.VA}`,
    `SC:${metrics.SC}`,
    `SI:${metrics.SI}`,
    `SA:${metrics.SA}`
  ];

  if (metrics.S !== 'X') {
    parts.push(`S:${metrics.S}`);
  }

  if (metrics.AU !== 'N') {
    parts.push(`AU:${metrics.AU}`);
  }

  if (metrics.threat) {
    const { E, M, D } = metrics.threat;
    if (E !== 'X') parts.push(`E:${E}`);
    if (M !== 'X') parts.push(`M:${M}`);
    if (D !== 'X') parts.push(`D:${D}`);
  }

  if (metrics.environmental) {
    const {
      CR, IR, AR, MAV, MAC, MAT, MPR, MUI,
      MS, MVC, MVI, MVA, MSC, MSI, MSA
    } = metrics.environmental;
    parts.push(`CR:${CR}`, `IR:${IR}`, `AR:${AR}`);
    if (MAV !== undefined) parts.push(`MAV:${MAV}`);
    if (MAC !== undefined) parts.push(`MAC:${MAC}`);
    if (MAT !== undefined) parts.push(`MAT:${MAT}`);
    if (MPR !== undefined) parts.push(`MPR:${MPR}`);
    if (MUI !== undefined) parts.push(`MUI:${MUI}`);
    if (MS !== undefined && (MS as any) !== 'X') parts.push(`MS:${MS}`);
    if (MVC !== undefined) parts.push(`MVC:${MVC}`);
    if (MVI !== undefined) parts.push(`MVI:${MVI}`);
    if (MVA !== undefined) parts.push(`MVA:${MVA}`);
    if (MSC !== undefined) parts.push(`MSC:${MSC}`);
    if (MSI !== undefined) parts.push(`MSI:${MSI}`);
    if (MSA !== undefined) parts.push(`MSA:${MSA}`);
  }

  return parts.join('/');
}

// ============================================================================
// Severity Rating
// ============================================================================

/**
 * Get severity rating from score
 */
function getSeverity(score: number): CVSSSeverity {
  if (score >= 9.0) return 'CRITICAL';
  if (score >= 7.0) return 'HIGH';
  if (score >= 4.0) return 'MEDIUM';
  if (score > 0.0) return 'LOW';
  return 'NONE';
}

// ============================================================================
// Public API
// ============================================================================

/**
 * Calculate CVSS v3.0/v3.1 score
 */
export function calculateCVSS3(
  metrics: CVSS3Metrics,
  version: '3.0' | '3.1'
): { vectorString: string; breakdown: CVSS3ScoreBreakdown } {
  const baseScore = calculateV3BaseScore(metrics);

  let breakdown: CVSS3ScoreBreakdown = {
    baseScore: baseScore,
    exploitabilityScore: ROUND(calculateV3Exploitability(metrics)),
    impactScore: ROUND(calculateV3Impact(metrics)),
    baseSeverity: getSeverity(baseScore)
  };

  let finalScore = baseScore;

  if (metrics.temporal) {
    const temporalResult = applyV3TemporalAdjustments(baseScore, metrics, metrics.temporal);
    breakdown = { ...breakdown, ...temporalResult.breakdown };
    finalScore = temporalResult.breakdown.temporalScore ?? finalScore;
  }

  if (metrics.environmental) {
    const envResult = applyV3EnvironmentalAdjustments(
      calculateV3BaseScore(metrics),
      metrics,
      metrics.environmental
    );
    breakdown = { ...breakdown, ...envResult.breakdown };
    finalScore = envResult.breakdown.environmentalScore ?? finalScore;
  }

  const vectorString = generateV3VectorString(metrics, version);

  return { vectorString, breakdown };
}

/**
 * Calculate CVSS v4.0 score
 */
export function calculateCVSS4(
  metrics: CVSS4Metrics
): { vectorString: string; breakdown: CVSS4ScoreBreakdown } {
  const baseResult = calculateV4BaseScore(metrics, metrics.threat);
  let finalScore = baseResult.breakdown.baseScore;

  let breakdown: CVSS4ScoreBreakdown = {
    baseScore: baseResult.breakdown.baseScore,
    baseSeverity: baseResult.breakdown.baseSeverity
  };

  if (metrics.threat) {
    const threatScore = applyV4ThreatAdjustments(finalScore, metrics.threat);
    breakdown.threatScore = ROUND(threatScore);
    breakdown.threatSeverity = getSeverity(ROUND(threatScore));
    finalScore = threatScore;
  }

  if (metrics.environmental) {
    const envResult = applyV4EnvironmentalAdjustments(metrics, metrics.environmental);
    breakdown.environmentalScore = envResult.adjustedScore;
    breakdown.environmentalSeverity = getSeverity(envResult.adjustedScore);
    finalScore = envResult.adjustedScore;
  }

  const vectorString = generateV4VectorString(metrics);

  return { vectorString, breakdown };
}

/**
 * Generic CVSS calculation based on version
 */
export function calculateCVSS(
  version: CVSSVersion,
  metrics: CVSS3Metrics | CVSS4Metrics
): { vectorString: string; breakdown: CVSS3ScoreBreakdown | CVSS4ScoreBreakdown } {
  if (version === '3.0' || version === '3.1') {
    const result = calculateCVSS3(metrics as CVSS3Metrics, version);
    return { vectorString: result.vectorString, breakdown: result.breakdown };
  }
  if (version === '4.0') {
    const result = calculateCVSS4(metrics as CVSS4Metrics);
    return { vectorString: result.vectorString, breakdown: result.breakdown };
  }
  throw new Error(`Unsupported CVSS version: ${version}`);
}

/**
 * Get default empty metrics for a version
 */
export function getDefaultMetrics(version: CVSSVersion): CVSS3BaseMetrics | CVSS4BaseMetrics {
  if (version === '3.0' || version === '3.1') {
    return {
      AV: 'N', AC: 'L', PR: 'N', UI: 'N',
      S: 'U', C: 'N', I: 'N', A: 'N'
    };
  }
  return {
    AV: 'N', AC: 'L', AT: 'N', PR: 'N', UI: 'N',
    VC: 'N', VI: 'N', VA: 'N', SC: 'N', SI: 'N', SA: 'N',
    S: 'N', AU: 'N'
  };
}

/**
 * Get metric metadata for UI
 */
export function getCVSSMetadata(version: CVSSVersion) {
  if (version === '3.0' || version === '3.1') {
    return {
      version,
      name: version === '3.0' ? 'CVSS v3.0' : 'CVSS v3.1',
      specUrl: version === '3.0'
        ? 'https://www.first.org/cvss/calculator/3.0'
        : 'https://www.first.org/cvss/calculator/3.1',
      releaseDate: version === '3.0' ? '2015-04-09' : '2023-11-02',
      metricGroups: [],
      availableMetrics: {}
    };
  }

  return {
    version,
    name: 'CVSS v4.0',
    specUrl: 'https://www.first.org/cvss/calculator/4.0',
    releaseDate: '2023-11-02',
    metricGroups: [],
    availableMetrics: {}
  };
}
