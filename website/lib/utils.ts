import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import type { EntityType } from "@/components/entity-detail-modal";
import type { CVEItem } from "@/lib/types";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Generate standardized URN for v2e entities
 */
export function generateURN(entityType: EntityType, entityId: string): string {
  switch (entityType) {
    case 'CVE':
      return `v2e::nvd::cve::${entityId}`;
    case 'CWE':
      return `v2e::mitre::cwe::${entityId}`;
    case 'CAPEC':
      return `v2e::mitre::capec::${entityId}`;
    case 'ATTACK':
      return `v2e::mitre::attack::${entityId}`;
    case 'ASVS':
      return `v2e::asvs::requirement::${entityId}`;
    case 'SSG':
      return `v2e::ssg::rule::${entityId}`;
    default:
      return `v2e::unknown::${entityType}::${entityId}`;
  }
}

/**
 * Get badge variant based on CVSS severity score
 */
export const getSeverityVariant = (score?: number): "default" | "secondary" | "destructive" | "outline" => {
  if (!score && score !== 0) return "outline";
  if (score >= 9.0) return "destructive";
  if (score >= 7.0) return "default";
  if (score >= 4.0) return "secondary";
  return "outline";
};

/**
 * Get severity label based on CVSS score
 */
export const getSeverityLabel = (score?: number): string => {
  if (!score && score !== 0) return "Unknown";
  if (score >= 9.0) return "Critical";
  if (score >= 7.0) return "High";
  if (score >= 4.0) return "Medium";
  return "Low";
};

/**
 * Get hex color for severity level
 * Returns hex colors to avoid Tailwind palette variations
 */
export const getSeverityColorHex = (score?: number): string => {
  if (!score && score !== 0) return '#9CA3AF'; // gray-400
  if (score >= 9.0) return '#ef4444'; // red-500
  if (score >= 7.0) return '#f59e0b'; // amber-500 (high)
  if (score >= 4.0) return '#fbbf24'; // yellow-400 (medium)
  return '#10b981'; // green-500 (low)
};

/**
 * Extract CVSS score from CVE metrics
 * Prefers newer metrics (4.0 -> 3.1 -> 3.0 -> 2.0) and returns the highest baseScore
 */
export const getCVSSScore = (cve: CVEItem): number | undefined => {
  const metricLists = [
    cve.metrics?.cvssMetricV40,
    cve.metrics?.cvssMetricV31,
    // some payloads may use cvssMetricV3 or cvssMetricV30
    (cve.metrics as any)?.cvssMetricV30 || (cve.metrics as any)?.cvssMetricV3,
    cve.metrics?.cvssMetricV2,
  ];

  let maxScore: number | undefined = undefined;
  for (const list of metricLists) {
    if (!Array.isArray(list)) continue;
    for (const item of list) {
      const s = item?.cvssData?.baseScore;
      if (typeof s === 'number') {
        if (maxScore === undefined || s > maxScore) maxScore = s;
      }
    }
    // if we already found a score from a higher-preference metric family, prefer it
    if (maxScore !== undefined) break;
  }
  return maxScore;
};

/**
 * Get English description from CVE item
 * Returns truncated description with limit characters
 */
export const getCVEDescription = (cve: CVEItem, limit: number = 100): string => {
  const desc = cve.descriptions?.find(d => d.lang === 'en') || cve.descriptions?.[0];
  if (!desc) return 'No description available';
  return desc.value.length > limit ? desc.value.substring(0, limit) + '...' : desc.value;
};
