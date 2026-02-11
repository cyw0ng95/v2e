/**
 * Severity Gauge Component Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import SeverityGauge from '../severity-gauge';
import { CVSSProvider } from '@/lib/cvss-context';

// Mock the cvss-calculator module
vi.mock('@/lib/cvss-calculator', () => ({
  calculateCVSS: vi.fn(() => ({
    vectorString: 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
    breakdown: {
      baseScore: 9.8,
      finalScore: 9.8,
      baseSeverity: 'CRITICAL',
      finalSeverity: 'CRITICAL'
    }
  })),
  getDefaultMetrics: vi.fn(() => ({
    AV: 'N', AC: 'L', PR: 'N', UI: 'N', S: 'U', C: 'N', I: 'N', A: 'N'
  })),
  getCVSSMetadata: vi.fn(() => ({
    version: '3.1',
    name: 'CVSS v3.1',
    specUrl: 'https://www.first.org/cvss/calculator/3.1',
    releaseDate: '2023-11-02'
  }))
}));

describe('SeverityGauge Component', () => {
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <CVSSProvider>{children}</CVSSProvider>
  );

  describe('Rendering', () => {
    it('should render gauge container', () => {
      render(<SeverityGauge score={7.5} severity="HIGH" />, { wrapper });

      const gauge = screen.getByRole('meter');
      expect(gauge).toBeInTheDocument();
    });

    it('should display current score', () => {
      render(<SeverityGauge score={7.5} severity="HIGH" />, { wrapper });

      expect(screen.getByText('7.5')).toBeInTheDocument();
    });

    it('should display severity label', () => {
      render(<SeverityGauge score={9.8} severity="CRITICAL" />, { wrapper });

      expect(screen.getByText('Critical')).toBeInTheDocument();
    });
  });

  describe('Severity Labels', () => {
    it('should display CRITICAL label for scores >= 9.0', () => {
      render(<SeverityGauge score={9.0} severity="CRITICAL" />, { wrapper });
      expect(screen.getByText('Critical')).toBeInTheDocument();

      render(<SeverityGauge score={10.0} severity="CRITICAL" />, { wrapper });
      expect(screen.getByText('Critical')).toBeInTheDocument();
    });

    it('should display HIGH label for scores >= 7.0 and < 9.0', () => {
      render(<SeverityGauge score={7.0} severity="HIGH" />, { wrapper });
      expect(screen.getByText('High')).toBeInTheDocument();

      render(<SeverityGauge score={8.9} severity="HIGH" />, { wrapper });
      expect(screen.getByText('High')).toBeInTheDocument();
    });

    it('should display MEDIUM label for scores >= 4.0 and < 7.0', () => {
      render(<SeverityGauge score={4.0} severity="MEDIUM" />, { wrapper });
      expect(screen.getByText('Medium')).toBeInTheDocument();

      render(<SeverityGauge score={6.9} severity="MEDIUM" />, { wrapper });
      expect(screen.getByText('Medium')).toBeInTheDocument();
    });

    it('should display LOW label for scores > 0.0 and < 4.0', () => {
      render(<SeverityGauge score={0.1} severity="LOW" />, { wrapper });
      expect(screen.getByText('Low')).toBeInTheDocument();

      render(<SeverityGauge score={3.9} severity="LOW" />, { wrapper });
      expect(screen.getByText('Low')).toBeInTheDocument();
    });

    it('should display NONE label for score 0.0', () => {
      render(<SeverityGauge score={0.0} severity="NONE" />, { wrapper });
      expect(screen.getByText('None')).toBeInTheDocument();
    });
  });

  describe('Color Coding', () => {
    it('should use purple gradient for CRITICAL severity', () => {
      const { container } = render(
        <SeverityGauge score={9.8} severity="CRITICAL" />,
        { wrapper }
      );

      const gaugeFill = container.querySelector('[data-severity="CRITICAL"]');
      expect(gaugeFill).toHaveClass('from-purple-500', 'to-purple-700');
    });

    it('should use red gradient for HIGH severity', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" />,
        { wrapper }
      );

      const gaugeFill = container.querySelector('[data-severity="HIGH"]');
      expect(gaugeFill).toHaveClass('from-red-400', 'to-red-600');
    });

    it('should use orange gradient for MEDIUM severity', () => {
      const { container } = render(
        <SeverityGauge score={5.5} severity="MEDIUM" />,
        { wrapper }
      );

      const gaugeFill = container.querySelector('[data-severity="MEDIUM"]');
      expect(gaugeFill).toHaveClass('from-orange-400', 'to-orange-600');
    });

    it('should use yellow gradient for LOW severity', () => {
      const { container } = render(
        <SeverityGauge score={2.5} severity="LOW" />,
        { wrapper }
      );

      const gaugeFill = container.querySelector('[data-severity="LOW"]');
      expect(gaugeFill).toHaveClass('from-yellow-400', 'to-yellow-600');
    });

    it('should use gray gradient for NONE severity', () => {
      const { container } = render(
        <SeverityGauge score={0.0} severity="NONE" />,
        { wrapper }
      );

      const gaugeFill = container.querySelector('[data-severity="NONE"]');
      expect(gaugeFill).toHaveClass('from-gray-400', 'to-gray-500');
    });
  });

  describe('Progress/Fill Indication', () => {
    it('should show correct fill percentage for score', () => {
      const { container } = render(
        <SeverityGauge score={5.0} severity="MEDIUM" />,
        { wrapper }
      );

      const gaugeElement = container.querySelector('[role="meter"]');
      expect(gaugeElement).toHaveAttribute('aria-valuenow', '5.0');
      expect(gaugeElement).toHaveAttribute('aria-valuemin', '0');
      expect(gaugeElement).toHaveAttribute('aria-valuemax', '10');
    });

    it('should calculate fill width based on score', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" />,
        { wrapper }
      );

      const fillElement = container.querySelector('.gauge-fill');
      expect(fillElement).toHaveStyle({ width: '75%' });
    });

    it('should handle edge cases', () => {
      const { container: container0 } = render(
        <SeverityGauge score={0} severity="NONE" />,
        { wrapper }
      );
      const fillElement0 = container0.querySelector('.gauge-fill');
      expect(fillElement0).toHaveStyle({ width: '0%' });

      const { container: container10 } = render(
        <SeverityGauge score={10} severity="CRITICAL" />,
        { wrapper }
      );
      const fillElement10 = container10.querySelector('.gauge-fill');
      expect(fillElement10).toHaveStyle({ width: '100%' });
    });
  });

  describe('Size Variants', () => {
    it('should render small size', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" size="small" />,
        { wrapper }
      );

      const gauge = container.querySelector('.severity-gauge-small');
      expect(gauge).toBeInTheDocument();
    });

    it('should render medium size (default)', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" size="medium" />,
        { wrapper }
      );

      const gauge = container.querySelector('.severity-gauge-medium');
      expect(gauge).toBeInTheDocument();
    });

    it('should render large size', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" size="large" />,
        { wrapper }
      );

      const gauge = container.querySelector('.severity-gauge-large');
      expect(gauge).toBeInTheDocument();
    });
  });

  describe('Orientation', () => {
    it('should render horizontal orientation by default', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" />,
        { wrapper }
      );

      const gauge = container.querySelector('.orientation-horizontal');
      expect(gauge).toBeInTheDocument();
    });

    it('should render vertical orientation when specified', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" orientation="vertical" />,
        { wrapper }
      );

      const gauge = container.querySelector('.orientation-vertical');
      expect(gauge).toBeInTheDocument();
    });
  });

  describe('Animation', () => {
    it('should animate score changes', () => {
      const { rerender } = render(
        <SeverityGauge score={5.0} severity="MEDIUM" />,
        { wrapper }
      );

      rerender(<SeverityGauge score={9.8} severity="CRITICAL" />);

      const fillElement = document.querySelector('.gauge-fill');
      expect(fillElement).toHaveClass('transition-all');
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA attributes', () => {
      render(<SeverityGauge score={7.5} severity="HIGH" />, { wrapper });

      const meter = screen.getByRole('meter');
      expect(meter).toHaveAttribute('aria-label');
      expect(meter).toHaveAttribute('aria-valuenow');
      expect(meter).toHaveAttribute('aria-valuemin');
      expect(meter).toHaveAttribute('aria-valuemax');
      expect(meter).toHaveAttribute('aria-describedby');
    });

    it('should announce severity to screen readers', () => {
      render(<SeverityGauge score={9.8} severity="CRITICAL" />, { wrapper });

      const description = screen.getByTestId('severity-description');
      expect(description).toHaveTextContent(/Critical/i);
    });

    it('should provide score range description', () => {
      render(<SeverityGauge score={7.5} severity="HIGH" />, { wrapper });

      const description = screen.getByTestId('severity-description');
      expect(description).toHaveTextContent(/7\.0\s*-\s*8\.9/i);
    });
  });

  describe('With Labels', () => {
    it('should show min/max labels when enabled', () => {
      render(<SeverityGauge score={7.5} severity="HIGH" showLabels />, { wrapper });

      expect(screen.getByText('0')).toBeInTheDocument();
      expect(screen.getByText('10')).toBeInTheDocument();
    });

    it('should show custom min/max labels', () => {
      render(
        <SeverityGauge score={7.5} severity="HIGH" showLabels minLabel="Low" maxLabel="High" />,
        { wrapper }
      );

      expect(screen.getByText('Low')).toBeInTheDocument();
      expect(screen.getByText('High')).toBeInTheDocument();
    });
  });

  describe('Threshold Markers', () => {
    it('should show severity threshold markers when enabled', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" showThresholds />,
        { wrapper }
      );

      const markers = container.querySelectorAll('.threshold-marker');
      expect(markers.length).toBeGreaterThan(0);
    });

    it('should mark critical threshold at 9.0', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" showThresholds />,
        { wrapper }
      );

      const criticalMarker = container.querySelector('[data-threshold="critical"]');
      expect(criticalMarker).toBeInTheDocument();
    });

    it('should mark high threshold at 7.0', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" showThresholds />,
        { wrapper }
      );

      const highMarker = container.querySelector('[data-threshold="high"]');
      expect(highMarker).toBeInTheDocument();
    });

    it('should mark medium threshold at 4.0', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" showThresholds />,
        { wrapper }
      );

      const mediumMarker = container.querySelector('[data-threshold="medium"]');
      expect(mediumMarker).toBeInTheDocument();
    });
  });

  describe('Interactive Mode', () => {
    it('should be interactive when onClick provided', () => {
      const handleClick = vi.fn();
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" onClick={handleClick} />,
        { wrapper }
      );

      const gauge = container.querySelector('[role="meter"]');
      expect(gauge).toHaveClass('cursor-pointer');

      gauge?.click();
      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    it('should not be interactive when no onClick', () => {
      const { container } = render(
        <SeverityGauge score={7.5} severity="HIGH" />,
        { wrapper }
      );

      const gauge = container.querySelector('[role="meter"]');
      expect(gauge).not.toHaveClass('cursor-pointer');
    });
  });

  describe('Edge Cases', () => {
    it('should handle negative scores gracefully', () => {
      render(<SeverityGauge score={-1} severity="NONE" />, { wrapper });
      const meter = screen.getByRole('meter');
      expect(meter).toHaveAttribute('aria-valuenow', '0');
    });

    it('should handle scores above 10 gracefully', () => {
      render(<SeverityGauge score={12} severity="CRITICAL" />, { wrapper });
      const meter = screen.getByRole('meter');
      expect(meter).toHaveAttribute('aria-valuenow', '10');
    });

    it('should handle undefined severity', () => {
      const { container } = render(
        <SeverityGauge score={5.5} severity={undefined as any} />,
        { wrapper }
      );

      const gauge = container.querySelector('.severity-gauge');
      expect(gauge).toBeInTheDocument();
    });
  });
});
