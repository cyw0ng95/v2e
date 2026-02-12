'use client';

/**
 * CVSS Metric Tooltip Component
 * Provides contextual help and explanations for CVSS metrics
 */

import { Info, AlertCircle, Shield, Target, Lock, User, Zap } from 'lucide-react';
import { useState } from 'react';

interface MetricTooltipProps {
  metric: string;
  version: '3.0' | '3.1' | '4.0';
  children: React.ReactNode;
}

interface MetricDefinition {
  title: string;
  description: string;
  values: Array<{ value: string; label: string; explanation: string; example?: string }>;
  icon: React.ReactNode;
}

const METRIC_DEFINITIONS: Record<string, MetricDefinition> = {
  // Attack Vector
  AV: {
    title: 'Attack Vector (AV)',
    description: 'This metric reflects the context by which vulnerability exploitation is possible.',
    values: [
      {
        value: 'N',
        label: 'Network',
        explanation: 'A vulnerability exploitable across the network.',
        example: 'Remote code execution via network protocol (e.g., SMB, HTTP)'
      },
      {
        value: 'A',
        label: 'Adjacent',
        explanation: 'A vulnerability exploitable across adjacent networks.',
        example: 'ARP spoofing, local subnet broadcast attacks'
      },
      {
        value: 'L',
        label: 'Local',
        explanation: 'A vulnerability exploitable only from local access.',
        example: 'Local privilege escalation, USB device exploitation'
      },
      {
        value: 'P',
        label: 'Physical',
        explanation: 'A vulnerability requiring physical access to the device.',
        example: 'Cold boot attack, physical port exploitation'
      }
    ],
    icon: <Target className="h-5 w-5" />
  },
  // Attack Complexity
  AC: {
    title: 'Attack Complexity (AC)',
    description: 'This metric describes the conditions beyond the attacker\'s control that must exist in order to exploit the vulnerability.',
    values: [
      {
        value: 'L',
        label: 'Low',
        explanation: 'Specialized access conditions or extenuating circumstances do not exist.',
        example: 'Attacker can expect repeatable success against the vulnerable component'
      },
      {
        value: 'H',
        label: 'High',
        explanation: 'A successful attack depends on conditions beyond the attacker\'s control.',
        example: 'Requires race conditions, specific system configurations, or timing-dependent attacks'
      }
    ],
    icon: <Zap className="h-5 w-5" />
  },
  // Attack Requirements (v4.0)
  AT: {
    title: 'Attack Requirements (AT)',
    description: 'This metric describes the conditions beyond the attacker\'s control that must exist in order to exploit the vulnerability.',
    values: [
      {
        value: 'N',
        label: 'None',
        explanation: 'No attack requirements are needed.',
        example: 'Vulnerability can be exploited without any special conditions'
      },
      {
        value: 'P',
        label: 'Present',
        explanation: 'The attacker must have some capability or preparation before exploitation.',
        example: 'Requires specific system configuration, software version, or prior access'
      },
      {
        value: 'R',
        label: 'Required',
        explanation: 'The attack requires specific conditions that are difficult to achieve.',
        example: 'Requires multiple systems, specific timing, or environmental conditions'
      }
    ],
    icon: <AlertCircle className="h-5 w-5" />
  },
  // Privileges Required
  PR: {
    title: 'Privileges Required (PR)',
    description: 'This metric describes the level of privileges an attacker must possess before successfully exploiting the vulnerability.',
    values: [
      {
        value: 'N',
        label: 'None',
        explanation: 'The attacker is unauthorized prior to attack.',
        example: 'Unauthenticated remote code execution'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'The attacker requires privileges that provide basic user capabilities.',
        example: 'Requires standard user account, guest access, or unprivileged session'
      },
      {
        value: 'H',
        label: 'High',
        explanation: 'The attacker requires privileges that provide significant control over the vulnerable component.',
        example: 'Requires administrator or root access to exploit'
      }
    ],
    icon: <Lock className="h-5 w-5" />
  },
  // User Interaction
  UI: {
    title: 'User Interaction (UI)',
    description: 'This metric captures the requirement for a human user, other than the attacker, to participate in the successful compromise of the vulnerable component.',
    values: [
      {
        value: 'N',
        label: 'None',
        explanation: 'The vulnerable system can be exploited without interaction from any user.',
        example: 'Wormable exploit, automated drive-by download'
      },
      {
        value: 'R',
        label: 'Required',
        explanation: 'Successful exploitation requires a user to take some action before the vulnerability can be exploited.',
        example: 'Opening malicious file, clicking link, visiting website'
      },
      {
        value: 'P',
        label: 'Passive',
        explanation: 'The user action does not actively enable the attack, but is necessary for the attack to succeed.',
        example: 'Reading an email that triggers exploit in preview pane'
      },
      {
        value: 'A',
        label: 'Active',
        explanation: 'The user must perform an action that directly enables the attack.',
        example: 'Clicking a button, entering credentials, accepting dialog'
      }
    ],
    icon: <User className="h-5 w-5" />
  },
  // Scope
  S: {
    title: 'Scope (S)',
    description: 'This metric measures whether a successful attack impacts components beyond the vulnerable component.',
    values: [
      {
        value: 'U',
        label: 'Unchanged',
        explanation: 'An exploited vulnerability only impacts resources managed by the same authority.',
        example: 'Vulnerability only affects the application where it exists'
      },
      {
        value: 'C',
        label: 'Changed',
        explanation: 'An exploited vulnerability can impact resources beyond the vulnerable component.',
        example: 'Browser vulnerability affecting OS, or app vulnerability affecting other apps'
      }
    ],
    icon: <Shield className="h-5 w-5" />
  },
  // Impact Metrics (C, I, A)
  C: {
    title: 'Confidentiality (C)',
    description: 'This metric measures the impact to the confidentiality of the information resources managed by a software component.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of confidentiality, resulting in all resources within the impacted component being disclosed to the attacker.',
        example: 'All user data, all database records, complete file system access'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'There is some loss of confidentiality.',
        example: 'Some records disclosed, partial information leak, non-sensitive data'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'There is no loss of confidentiality.',
        example: 'No information disclosure, or information already public'
      }
    ],
    icon: <Lock className="h-5 w-5" />
  },
  I: {
    title: 'Integrity (I)',
    description: 'This metric measures the impact to integrity of a successfully exploited vulnerability.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of integrity, or a complete loss of protection.',
        example: 'Attacker can modify any/all files, data, or configuration'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'Modification of data is possible, but the attacker does not have control over what can be modified.',
        example: 'Limited modification capability, partial data alteration'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'There is no loss of integrity.',
        example: 'No data modification possible or modification meaningless'
      }
    ],
    icon: <Shield className="h-5 w-5" />
  },
  A: {
    title: 'Availability (A)',
    description: 'This metric measures the impact to the availability of the impacted component.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of availability, resulting in the attacker being able to fully deny access to resources.',
        example: 'Complete system shutdown, permanent DoS, all services unavailable'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'Performance is reduced or there are interruptions in resource availability.',
        example: 'Periodic DoS, slow response, temporary unavailability'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'There is no impact to availability.',
        example: 'Service remains fully operational despite exploit'
      }
    ],
    icon: <AlertCircle className="h-5 w-5" />
  },
  // CVSS v4.0 Specific Metrics
  VC: {
    title: 'Vulnerable System Confidentiality (VC)',
    description: 'This metric measures the impact to the confidentiality of the vulnerable system.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of confidentiality for the vulnerable system.',
        example: 'All data in the vulnerable system exposed'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'Partial loss of confidentiality.',
        example: 'Some sensitive data exposed from vulnerable system'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'No confidentiality impact on the vulnerable system.',
        example: 'No data disclosed from vulnerable system'
      }
    ],
    icon: <Lock className="h-5 w-5" />
  },
  VI: {
    title: 'Vulnerable System Integrity (VI)',
    description: 'This metric measures the impact to the integrity of the vulnerable system.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of integrity for the vulnerable system.',
        example: 'Attacker can modify any data in the vulnerable system'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'Partial loss of integrity.',
        example: 'Attacker can modify some data in the vulnerable system'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'No integrity impact on the vulnerable system.',
        example: 'No data modification possible in vulnerable system'
      }
    ],
    icon: <Shield className="h-5 w-5" />
  },
  VA: {
    title: 'Vulnerable System Availability (VA)',
    description: 'This metric measures the impact to the availability of the vulnerable system.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of availability for the vulnerable system.',
        example: 'Vulnerable system becomes completely unavailable'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'Partial loss of availability.',
        example: 'Vulnerable system experiences reduced performance'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'No availability impact on the vulnerable system.',
        example: 'Vulnerable system remains fully operational'
      }
    ],
    icon: <AlertCircle className="h-5 w-5" />
  },
  SC: {
    title: 'Subsequent System Confidentiality (SC)',
    description: 'This metric measures the impact to the confidentiality of systems subsequent to the exploited vulnerability.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of confidentiality for subsequent systems.',
        example: 'All data in connected systems exposed'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'Partial loss of confidentiality.',
        example: 'Some sensitive data exposed from connected systems'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'No confidentiality impact on subsequent systems.',
        example: 'No data disclosed from connected systems'
      }
    ],
    icon: <Lock className="h-5 w-5" />
  },
  SI: {
    title: 'Subsequent System Integrity (SI)',
    description: 'This metric measures the impact to the integrity of systems subsequent to the exploited vulnerability.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of integrity for subsequent systems.',
        example: 'Attacker can modify any data in connected systems'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'Partial loss of integrity.',
        example: 'Attacker can modify some data in connected systems'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'No integrity impact on subsequent systems.',
        example: 'No data modification possible in connected systems'
      }
    ],
    icon: <Shield className="h-5 w-5" />
  },
  SA: {
    title: 'Subsequent System Availability (SA)',
    description: 'This metric measures the impact to the availability of systems subsequent to the exploited vulnerability.',
    values: [
      {
        value: 'H',
        label: 'High',
        explanation: 'Total loss of availability for subsequent systems.',
        example: 'Connected systems become completely unavailable'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'Partial loss of availability.',
        example: 'Connected systems experience reduced performance'
      },
      {
        value: 'N',
        label: 'None',
        explanation: 'No availability impact on subsequent systems.',
        example: 'Connected systems remain fully operational'
      }
    ],
    icon: <AlertCircle className="h-5 w-5" />
  },
  // Temporal Metrics
  E: {
    title: 'Exploit Maturity (E)',
    description: 'This metric measures the likelihood of the vulnerability being attacked, based on the current state of exploit techniques.',
    values: [
      {
        value: 'X',
        label: 'Not Defined',
        explanation: 'Value is not defined or is being omitted.',
        example: 'Used when exploit status is unknown or not applicable'
      },
      {
        value: 'U',
        label: 'Unproven',
        explanation: 'No exploit code is available, or an exploit is theoretical.',
        example: 'Vulnerability reported but no public PoC available'
      },
      {
        value: 'P',
        label: 'Proof-of-Concept',
        explanation: 'Proof-of-concept exploit code is available, or no reliable exploit is available.',
        example: 'PoC available on exploit-db or GitHub'
      },
      {
        value: 'F',
        label: 'Functional',
        explanation: 'Functional exploit code is available.',
        example: 'Working exploit in exploit frameworks (Metasploit, etc.)'
      },
      {
        value: 'H',
        label: 'High',
        explanation: 'Reliable exploit code is available that works consistently.',
        example: 'Automated exploits, widespread exploitation in the wild'
      },
      {
        value: 'R',
        label: 'Official',
        explanation: 'The vendor has confirmed the vulnerability and an official fix is available.',
        example: 'Vendor-confirmed exploit with patch released'
      },
      {
        value: 'A',
        label: 'Attacked',
        explanation: 'The vulnerability is being actively exploited in the wild.',
        example: 'Active campaigns, ransomware using this vulnerability'
      }
    ],
    icon: <Target className="h-5 w-5" />
  },
  RL: {
    title: 'Remediation Level (RL)',
    description: 'This metric measures the type of remediation currently available.',
    values: [
      {
        value: 'X',
        label: 'Not Defined',
        explanation: 'Value is not defined or is being omitted.',
        example: 'Used when remediation status is unknown'
      },
      {
        value: 'U',
        label: 'Unavailable',
        explanation: 'No official or unofficial fix is available.',
        example: 'Zero-day vulnerability, no vendor fix available'
      },
      {
        value: 'O',
        label: 'Workaround',
        explanation: 'An unofficial or temporary fix is available.',
        example: 'Configuration changes, disabling vulnerable feature'
      },
      {
        value: 'T',
        label: 'Temporary',
        explanation: 'An official fix is available but is not a complete solution.',
        example: 'Patch that mitigates but does not fully fix the issue'
      },
      {
        value: 'W',
        label: 'Official',
        explanation: 'A complete vendor fix is available.',
        example: 'Official security patch or software update released'
      }
    ],
    icon: <Shield className="h-5 w-5" />
  },
  RC: {
    title: 'Report Confidence (RC)',
    description: 'This metric measures the degree of confidence in the existence of the vulnerability.',
    values: [
      {
        value: 'X',
        label: 'Not Defined',
        explanation: 'Value is not defined or is being omitted.',
        example: 'Used when report confidence is unknown'
      },
      {
        value: 'U',
        label: 'Unknown',
        explanation: 'The report is unverified or the source is unreliable.',
        example: 'Unconfirmed rumors, third-party reports without verification'
      },
      {
        value: 'C',
        label: 'Reasonable',
        explanation: 'There is some confidence in the validity of the report.',
        example: 'Vendor acknowledgment, technical analysis available'
      },
      {
        value: 'R',
        label: 'Confirmed',
        explanation: 'The vulnerability has been confirmed.',
        example: 'Independent verification, vendor confirmation, PoC tested'
      }
    ],
    icon: <Info className="h-5 w-5" />
  },
  // Safety (v4.0)
  S: {
    title: 'Safety (S)',
    description: 'This metric measures the potential for harm to people or property if the vulnerability is exploited.',
    values: [
      {
        value: 'X',
        label: 'Not Defined',
        explanation: 'Value is not defined or is being omitted.',
        example: 'Used when safety impact is not applicable (e.g., pure software)'
      },
      {
        value: 'N',
        label: 'Negligible',
        explanation: 'There is no impact to safety.',
        example: 'Vulnerability has no physical safety implications'
      },
      {
        value: 'P',
        label: 'Present',
        explanation: 'Exploitation could result in harm to people or property.',
        example: 'Medical device vulnerability, industrial control system exploit'
      }
    ],
    icon: <AlertCircle className="h-5 w-5" />
  },
  // Automation (v4.0)
  AU: {
    title: 'Automation (AU)',
    description: 'This metric measures the level of automation that can be applied to exploit this vulnerability.',
    values: [
      {
        value: 'N',
        label: 'No',
        explanation: 'The vulnerability cannot be exploited by automated means.',
        example: 'Requires manual interaction at each exploitation step'
      },
      {
        value: 'L',
        label: 'Low',
        explanation: 'The vulnerability can be exploited with some automation.',
        example: 'Semi-automated exploits requiring human oversight'
      },
      {
        value: 'H',
        label: 'High',
        explanation: 'The vulnerability can be fully exploited by automated means.',
        example: 'Wormable exploits, fully automated attack campaigns'
      }
    ],
    icon: <Zap className="h-5 w-5" />
  }
};

export function MetricTooltip({ metric, version, children }: MetricTooltipProps) {
  const [isOpen, setIsOpen] = useState(false);
  const definition = METRIC_DEFINITIONS[metric];

  if (!definition) {
    return <>{children}</>;
  }

  return (
    <div className="relative inline-block">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="inline-flex items-center gap-1 text-blue-600 hover:text-blue-700 focus:outline-none"
        aria-label={`Show help for ${definition.title}`}
      >
        {children}
        <Info className="h-4 w-4" />
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setIsOpen(false)}
            aria-hidden="true"
          />
          <div className="absolute z-50 left-full top-0 ml-2 w-96 bg-white rounded-lg shadow-xl border border-slate-200 p-4">
            <div className="flex items-start gap-2 mb-3">
              <div className="text-blue-600 mt-0.5">{definition.icon}</div>
              <div className="flex-1">
                <h3 className="font-semibold text-slate-900 text-sm">{definition.title}</h3>
                <p className="text-xs text-slate-600 mt-1">{definition.description}</p>
              </div>
            </div>

            <div className="space-y-2">
              <h4 className="text-xs font-semibold text-slate-700">Possible Values:</h4>
              {definition.values.map((val) => (
                <div key={val.value} className="bg-slate-50 rounded-md p-2">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-mono text-sm font-semibold text-blue-600 bg-blue-100 px-1.5 py-0.5 rounded">
                      {val.value}
                    </span>
                    <span className="text-xs font-medium text-slate-900">{val.label}</span>
                  </div>
                  <p className="text-xs text-slate-600 ml-1">{val.explanation}</p>
                  {val.example && (
                    <p className="text-xs text-slate-500 mt-1 ml-1 italic">
                      Example: {val.example}
                    </p>
                  )}
                </div>
              ))}
            </div>

            <button
              type="button"
              onClick={() => setIsOpen(false)}
              className="mt-3 w-full py-2 text-xs font-medium text-slate-600 hover:text-slate-900 border border-slate-200 rounded-md hover:bg-slate-50 transition-colors"
            >
              Close
            </button>
          </div>
        </>
      )}
    </div>
  );
}

/**
 * Quick help tooltip for a specific metric value
 */
interface ValueTooltipProps {
  metric: string;
  value: string;
  children: React.ReactNode;
}

export function ValueTooltip({ metric, value, children }: ValueTooltipProps) {
  const [isOpen, setIsOpen] = useState(false);
  const definition = METRIC_DEFINITIONS[metric];

  if (!definition) {
    return <>{children}</>;
  }

  const valueInfo = definition.values.find(v => v.value === value);

  if (!valueInfo) {
    return <>{children}</>;
  }

  return (
    <div className="relative inline-block">
      <button
        type="button"
        onMouseEnter={() => setIsOpen(true)}
        onMouseLeave={() => setIsOpen(false)}
        className="focus:outline-none"
        aria-label={`Show details for ${valueInfo.label}`}
      >
        {children}
      </button>

      {isOpen && (
        <div className="absolute z-50 bottom-full left-1/2 -translate-x-1/2 mb-2 w-72 bg-slate-900 text-white rounded-lg shadow-xl p-3">
          <div className="text-xs font-semibold mb-1">{valueInfo.label}</div>
          <div className="text-xs text-slate-300 mb-2">{valueInfo.explanation}</div>
          {valueInfo.example && (
            <div className="text-xs text-slate-400 italic">
              Example: {valueInfo.example}
            </div>
          )}
          <div className="absolute bottom-0 left-1/2 -translate-x-1/2 translate-y-full">
            <div className="border-4 border-transparent border-t-slate-900" />
          </div>
        </div>
      )}
    </div>
  );
}

export default MetricTooltip;
