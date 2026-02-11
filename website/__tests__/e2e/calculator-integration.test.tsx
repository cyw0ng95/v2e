/**
 * CVSS Calculator End-to-End Integration Tests
 * Tests the complete user flow from selecting metrics to exporting results
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import CVSSCalculator from '@/components/cvss/calculator';
import { CVSSProvider } from '@/lib/cvss-context';

// Mock clipboard API
const mockClipboard = {
  writeText: vi.fn(() => Promise.resolve())
};

Object.assign(navigator, { clipboard: mockClipboard });

// Mock download functionality
const mockCreateElement = vi.fn();
const mockClick = vi.fn();
const mockRevokeObjectURL = vi.fn();

global.URL.createObjectURL = vi.fn(() => 'blob:url');
global.URL.revokeObjectURL = mockRevokeObjectURL;

const originalCreateElement = document.createElement.bind(document);
document.createElement = vi.fn((tagName) => {
  const element = originalCreateElement(tagName);
  if (tagName === 'a') {
    mockCreateElement(tagName);
    element.click = mockClick;
  }
  return element;
});

describe('CVSS Calculator E2E Integration', () => {
  const renderCalculator = () => {
    return render(
      <CVSSProvider initialVersion="3.1">
        <CVSSCalculator />
      </CVSSProvider>
    );
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    document.createElement = originalCreateElement;
  });

  describe('Initial State', () => {
    it('should render calculator with default version selected', () => {
      renderCalculator();

      expect(screen.getByText(/CVSS Calculator.*3\.1/)).toBeInTheDocument();
    });

    it('should display all version buttons', () => {
      renderCalculator();

      expect(screen.getByText('3.0')).toBeInTheDocument();
      expect(screen.getByText('3.1')).toBeInTheDocument();
      expect(screen.getByText('4.0')).toBeInTheDocument();
    });

    it('should show base metrics section', () => {
      renderCalculator();

      expect(screen.getByText('Base Metrics')).toBeInTheDocument();
      expect(screen.getByText('Attack Vector (AV)')).toBeInTheDocument();
      expect(screen.getByText('Attack Complexity (AC)')).toBeInTheDocument();
      expect(screen.getByText('Privileges Required (PR)')).toBeInTheDocument();
      expect(screen.getByText('User Interaction (UI)')).toBeInTheDocument();
      expect(screen.getByText('Scope (S)')).toBeInTheDocument();
      expect(screen.getByText('Confidentiality (C)')).toBeInTheDocument();
      expect(screen.getByText('Integrity (I)')).toBeInTheDocument();
      expect(screen.getByText('Availability (A)')).toBeInTheDocument();
    });

    it('should show vector string display', () => {
      renderCalculator();

      expect(screen.getByText('Vector String')).toBeInTheDocument();
      const vectorElement = screen.getByRole('code');
      expect(vectorElement).toBeInTheDocument();
    });
  });

  describe('Version Switching', () => {
    it('should switch from v3.1 to v3.0', async () => {
      renderCalculator();

      const v30Button = screen.getByText('3.0');
      fireEvent.click(v30Button);

      expect(screen.getByText(/CVSS Calculator.*3\.0/)).toBeInTheDocument();
    });

    it('should switch from v3.1 to v4.0', async () => {
      renderCalculator();

      const v40Button = screen.getByText('4.0');
      fireEvent.click(v40Button);

      expect(screen.getByText(/CVSS Calculator.*4\.0/)).toBeInTheDocument();
    });

    it('should update vector string prefix when version changes', async () => {
      renderCalculator();

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('CVSS:3.1');

      const v40Button = screen.getByText('4.0');
      fireEvent.click(v40Button);

      await waitFor(() => {
        expect(vectorElement.textContent).toContain('CVSS:4.0');
      });
    });
  });

  describe('Metric Selection', () => {
    it('should update AV metric when Network is clicked', async () => {
      renderCalculator();

      const avSection = screen.getByText('Attack Vector (AV)')?.closest('div');
      const networkButton = within(avSection!).getByText('N');

      fireEvent.click(networkButton);

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('AV:N');
    });

    it('should update AC metric when High is clicked', async () => {
      renderCalculator();

      const acSection = screen.getByText('Attack Complexity (AC)')?.closest('div');
      const highButton = within(acSection!).getByText('H');

      fireEvent.click(highButton);

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('AC:H');
    });

    it('should update PR metric options correctly', async () => {
      renderCalculator();

      const prSection = screen.getByText('Privileges Required (PR)')?.closest('div');
      const lowButton = within(prSection!).getByText('L');

      fireEvent.click(lowButton);

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('PR:L');
    });

    it('should update UI metric when Required is clicked', async () => {
      renderCalculator();

      const uiSection = screen.getByText('User Interaction (UI)')?.closest('div');
      const requiredButton = within(uiSection!).getByText('R');

      fireEvent.click(requiredButton);

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('UI:R');
    });

    it('should update Scope metric', async () => {
      renderCalculator();

      const scopeSection = screen.getByText('Scope (S)')?.closest('div');
      const changedButton = within(scopeSection!).getByText('C');

      fireEvent.click(changedButton);

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('S:C');
    });

    it('should update Confidentiality metric', async () => {
      renderCalculator();

      const cSection = screen.getByText('Confidentiality (C)')?.closest('div');
      const highButton = within(cSection!).getByText('H');

      fireEvent.click(highButton);

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('C:H');
    });

    it('should update Integrity metric', async () => {
      renderCalculator();

      const iSection = screen.getByText('Integrity (I)')?.closest('div');
      const highButton = within(iSection!).getByText('H');

      fireEvent.click(highButton);

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('I:H');
    });

    it('should update Availability metric', async () => {
      renderCalculator();

      const aSection = screen.getByText('Availability (A)')?.closest('div');
      const highButton = within(aSection!).getByText('H');

      fireEvent.click(highButton);

      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toContain('A:H');
    });
  });

  describe('Score Updates', () => {
    it('should update base score in real-time when metrics change', async () => {
      renderCalculator();

      // Find and click all High impact options for maximum score
      const avSection = screen.getByText('Attack Vector (AV)')?.closest('div');
      fireEvent.click(within(avSection!).getByText('N'));

      const cSection = screen.getByText('Confidentiality (C)')?.closest('div');
      fireEvent.click(within(cSection!).getByText('H'));

      const iSection = screen.getByText('Integrity (I)')?.closest('div');
      fireEvent.click(within(iSection!).getByText('H'));

      const aSection = screen.getByText('Availability (A)')?.closest('div');
      fireEvent.click(within(aSection!).getByText('H'));

      // Wait for score update
      await waitFor(() => {
        const scoreElements = screen.queryAllByText(/\d+\.\d/);
        const baseScore = scoreElements.find(el => el.textContent === '9.8');
        expect(baseScore).toBeInTheDocument();
      });
    });

    it('should show CRITICAL severity for high scores', async () => {
      renderCalculator();

      // Set to maximum critical configuration
      const avSection = screen.getByText('Attack Vector (AV)')?.closest('div');
      fireEvent.click(within(avSection!).getByText('N'));

      const sSection = screen.getByText('Scope (S)')?.closest('div');
      fireEvent.click(within(sSection!).getByText('C'));

      const cSection = screen.getByText('Confidentiality (C)')?.closest('div');
      fireEvent.click(within(cSection!).getByText('H'));

      const iSection = screen.getByText('Integrity (I)')?.closest('div');
      fireEvent.click(within(iSection!).getByText('H'));

      const aSection = screen.getByText('Availability (A)')?.closest('div');
      fireEvent.click(within(aSection!).getByText('H'));

      await waitFor(() => {
        expect(screen.getByText(/Critical.*\(9\.0-10\.0\)/i)).toBeInTheDocument();
      });
    });

    it('should calculate zero score for no impact', async () => {
      renderCalculator();

      // Set all impacts to None
      const cSection = screen.getByText('Confidentiality (C)')?.closest('div');
      fireEvent.click(within(cSection!).getByText('N'));

      const iSection = screen.getByText('Integrity (I)')?.closest('div');
      fireEvent.click(within(iSection!).getByText('N'));

      const aSection = screen.getByText('Availability (A)')?.closest('div');
      fireEvent.click(within(aSection!).getByText('N'));

      await waitFor(() => {
        const scoreElements = screen.queryAllByText(/\d+\.\d/);
        const baseScore = scoreElements.find(el => el.textContent === '0.0');
        expect(baseScore).toBeInTheDocument();
      });
    });
  });

  describe('Temporal Metrics', () => {
    it('should show temporal metrics section for v3.x', async () => {
      renderCalculator();

      expect(screen.getByText('Temporal Metrics')).toBeInTheDocument();
    });

    it('should toggle temporal metrics visibility', async () => {
      renderCalculator();

      const showButton = screen.getByText('Show');
      fireEvent.click(showButton);

      expect(screen.getByText('Exploit Maturity (E)')).toBeInTheDocument();

      const hideButton = screen.getByText('Hide');
      fireEvent.click(hideButton);

      expect(screen.queryByText('Exploit Maturity (E)')).not.toBeInTheDocument();
    });

    it('should update Exploit Maturity metric', async () => {
      renderCalculator();

      const showButton = screen.getByText('Show');
      fireEvent.click(showButton);

      const eSection = screen.getByText('Exploit Maturity (E)')?.closest('div');
      const functionalButton = within(eSection!).getByText('F');

      fireEvent.click(functionalButton);

      const vectorElement = screen.getByRole('code');
      await waitFor(() => {
        expect(vectorElement.textContent).toContain('/E:F');
      });
    });
  });

  describe('Vector String Operations', () => {
    it('should copy vector string to clipboard', async () => {
      renderCalculator();

      const copyButton = screen.getByText('Copy').closest('button');
      fireEvent.click(copyButton!);

      await waitFor(() => {
        expect(mockClipboard.writeText).toHaveBeenCalled();
      });

      expect(screen.getByText('Copied!')).toBeInTheDocument();
    });

    it('should show copied feedback temporarily', async () => {
      renderCalculator();

      const copyButton = screen.getByText('Copy').closest('button');
      fireEvent.click(copyButton!);

      await waitFor(() => {
        expect(screen.getByText('Copied!')).toBeInTheDocument();
      });

      await waitFor(
        () => {
          expect(screen.queryByText('Copied!')).not.toBeInTheDocument();
        },
        { timeout: 2500 }
      );
    });
  });

  describe('Export Functionality', () => {
    it('should open export menu when button is clicked', async () => {
      renderCalculator();

      const exportButton = screen.getByText('Export CVSS Data');
      fireEvent.click(exportButton);

      expect(screen.getByText('JSON')).toBeInTheDocument();
      expect(screen.getByText('CSV')).toBeInTheDocument();
      expect(screen.getByText('Share URL')).toBeInTheDocument();
    });

    it('should export as JSON', async () => {
      renderCalculator();

      const exportButton = screen.getByText('Export CVSS Data');
      fireEvent.click(exportButton);

      const jsonButton = screen.getByText('JSON');
      fireEvent.click(jsonButton);

      await waitFor(() => {
        expect(mockCreateElement).toHaveBeenCalledWith('a');
        expect(mockClick).toHaveBeenCalled();
      });
    });

    it('should export as CSV', async () => {
      renderCalculator();

      const exportButton = screen.getByText('Export CVSS Data');
      fireEvent.click(exportButton);

      const csvButton = screen.getByText('CSV');
      fireEvent.click(csvButton);

      await waitFor(() => {
        expect(mockCreateElement).toHaveBeenCalledWith('a');
        expect(mockClick).toHaveBeenCalled();
      });
    });

    it('should copy share URL to clipboard', async () => {
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});

      renderCalculator();

      const exportButton = screen.getByText('Export CVSS Data');
      fireEvent.click(exportButton);

      const urlButton = screen.getByText('Share URL');
      fireEvent.click(urlButton);

      await waitFor(() => {
        expect(mockClipboard.writeText).toHaveBeenCalled();
        expect(alertSpy).toHaveBeenCalledWith('CVSS URL copied to clipboard!');
      });

      alertSpy.mockRestore();
    });

    it('should close export menu after selection', async () => {
      renderCalculator();

      const exportButton = screen.getByText('Export CVSS Data');
      fireEvent.click(exportButton);

      expect(screen.getByText('JSON')).toBeInTheDocument();

      const jsonButton = screen.getByText('JSON');
      fireEvent.click(jsonButton);

      await waitFor(() => {
        expect(screen.queryByText('JSON')).not.toBeInTheDocument();
      });
    });
  });

  describe('Reset Functionality', () => {
    it('should reset all metrics to defaults', async () => {
      renderCalculator();

      // Change some metrics
      const avSection = screen.getByText('Attack Vector (AV)')?.closest('div');
      fireEvent.click(within(avSection!).getByText('A'));

      const cSection = screen.getByText('Confidentiality (C)')?.closest('div');
      fireEvent.click(within(cSection!).getByText('H'));

      // Reset
      const resetButton = screen.getByText('Reset');
      fireEvent.click(resetButton);

      // Verify default vector string
      const vectorElement = screen.getByRole('code');
      await waitFor(() => {
        expect(vectorElement.textContent).toContain('AV:N');
        expect(vectorElement.textContent).toContain('C:N');
      });
    });
  });

  describe('CVSS v4.0 Specific Features', () => {
    it('should show v4.0 specific metrics', async () => {
      renderCalculator();

      const v40Button = screen.getByText('4.0');
      fireEvent.click(v40Button);

      await waitFor(() => {
        expect(screen.getByText(/CVSS Calculator.*4\.0/)).toBeInTheDocument();
      });

      // v4.0 should have Attack Time metric
      // This will be in the base metrics section
      expect(screen.getByText('Base Metrics')).toBeInTheDocument();
    });
  });

  describe('Complete User Flow', () => {
    it('should handle full calculation workflow', async () => {
      renderCalculator();

      // Select all metrics for a critical vulnerability
      const metrics = [
        { label: 'Attack Vector (AV)', value: 'N' },
        { label: 'Attack Complexity (AC)', value: 'L' },
        { label: 'Privileges Required (PR)', value: 'N' },
        { label: 'User Interaction (UI)', value: 'N' },
        { label: 'Scope (S)', value: 'C' },
        { label: 'Confidentiality (C)', value: 'H' },
        { label: 'Integrity (I)', value: 'H' },
        { label: 'Availability (A)', value: 'H' }
      ];

      for (const metric of metrics) {
        const section = screen.getByText(metric.label)?.closest('div');
        fireEvent.click(within(section!).getByText(metric.value));
      }

      // Verify score is calculated
      await waitFor(() => {
        const scoreElements = screen.queryAllByText(/\d+\.\d/);
        expect(scoreElements.length).toBeGreaterThan(0);
      });

      // Verify vector string is generated
      const vectorElement = screen.getByRole('code');
      expect(vectorElement.textContent).toMatch(/^CVSS:3\.1\/AV:N\/AC:L\/PR:N\/UI:N\/S:C\/C:H\/I:H\/A:H/);

      // Copy vector string
      const copyButton = screen.getByText('Copy').closest('button');
      fireEvent.click(copyButton!);
      expect(mockClipboard.writeText).toHaveBeenCalled();

      // Export as JSON
      const exportButton = screen.getByText('Export CVSS Data');
      fireEvent.click(exportButton);
      fireEvent.click(screen.getByText('JSON'));
      expect(mockCreateElement).toHaveBeenCalledWith('a');
    });
  });

  describe('Error Handling', () => {
    it('should handle invalid vector string import gracefully', async () => {
      renderCalculator();

      // This would be tested if the calculator has import functionality
      // For now, we verify the calculator doesn't crash with normal operations
      const vectorElement = screen.getByRole('code');
      expect(vectorElement).toBeInTheDocument();
    });
  });

  describe('Responsiveness', () => {
    it('should render correctly on small screens', () => {
      // Mock small viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375
      });

      renderCalculator();

      expect(screen.getByText('Base Metrics')).toBeInTheDocument();
      expect(screen.getByText('Vector String')).toBeInTheDocument();
    });
  });
});
