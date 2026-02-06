import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import type { EntityType } from "@/components/entity-detail-modal";

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
