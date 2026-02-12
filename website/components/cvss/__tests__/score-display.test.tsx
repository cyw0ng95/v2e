/**
 * Score Display Component Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreDisplay from '../score-display';
import { CVSSProvider } from '@/lib/cvss-context';

// Mock the cvss-calculator module
vi.mock('@/lib/cvss-calculator', () => ({
  calculateCVSS: vi.fn(() => ({
    vectorString: 'CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H',
    breakdown: {
      baseScore: 9.8,
      finalScore: 9.8,
      baseSeverity: 'CRITICAL',
      finalSeverity: 'CRITICAL',
      exploitabilityScore: 3.9,
      impactScore: 5.9
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

describe('ScoreDisplay Component', () => {
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <CVSSProvider>{children}</CVSSProvider>
  );

  describe('Rendering', () => {
    it('should render base score display', () => {
      render(<ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" />, { wrapper });

      expect(screen.getByText('Base Score')).toBeInTheDocument();
      expect(screen.getByText('9.8')).toBeInTheDocument();
    });

    it('should render severity label correctly', () => {
      const { rerender } = render(
        <ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" />,
        { wrapper }
      );

      expect(screen.getByText('Critical')).toBeInTheDocument();

      rerender(<ScoreDisplay score={5.5} severity="MEDIUM" label="Base Score" />);
      expect(screen.getByText('Medium')).toBeInTheDocument();

      rerender(<ScoreDisplay score={2.5} severity="LOW" label="Base Score" />);
      expect(screen.getByText('Low')).toBeInTheDocument();

      rerender(<ScoreDisplay score={0.0} severity="NONE" label="Base Score" />);
      expect(screen.getByText('None')).toBeInTheDocument();
    });

    it('should render custom label', () => {
      render(<ScoreDisplay score={7.5} severity="HIGH" label="Temporal Score" />, { wrapper });

      expect(screen.getByText('Temporal Score')).toBeInTheDocument();
    });

    it('should render score with one decimal place', () => {
      render(<ScoreDisplay score={9.87} severity="HIGH" label="Base Score" />, { wrapper });

      expect(screen.getByText('9.9')).toBeInTheDocument();
    });
  });

  describe('Styling', () => {
    it('should apply critical severity styles', () => {
      const { container } = render(
        <ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.bg-critical-gradient');
      expect(scoreElement).toBeInTheDocument();
    });

    it('should apply high severity styles', () => {
      const { container } = render(
        <ScoreDisplay score={8.5} severity="HIGH" label="Base Score" />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.bg-high-gradient');
      expect(scoreElement).toBeInTheDocument();
    });

    it('should apply medium severity styles', () => {
      const { container } = render(
        <ScoreDisplay score={5.5} severity="MEDIUM" label="Base Score" />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.bg-medium-gradient');
      expect(scoreElement).toBeInTheDocument();
    });

    it('should apply low severity styles', () => {
      const { container } = render(
        <ScoreDisplay score={2.5} severity="LOW" label="Base Score" />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.bg-low-gradient');
      expect(scoreElement).toBeInTheDocument();
    });

    it('should apply none severity styles', () => {
      const { container } = render(
        <ScoreDisplay score={0.0} severity="NONE" label="Base Score" />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.bg-none-gradient');
      expect(scoreElement).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have accessible label for score', () => {
      render(<ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" />, { wrapper });

      const scoreElement = screen.getByText('9.8');
      expect(scoreElement).toHaveAccessibleDescription();
    });

    it('should announce severity to screen readers', () => {
      render(<ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" />, { wrapper });

      expect(screen.getByText('Critical Severity')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle zero score', () => {
      render(<ScoreDisplay score={0} severity="NONE" label="Base Score" />, { wrapper });

      expect(screen.getByText('0.0')).toBeInTheDocument();
    });

    it('should handle maximum score', () => {
      render(<ScoreDisplay score={10} severity="CRITICAL" label="Base Score" />, { wrapper });

      expect(screen.getByText('10.0')).toBeInTheDocument();
    });

    it('should handle small decimal scores', () => {
      render(<ScoreDisplay score={0.1} severity="LOW" label="Base Score" />, { wrapper });

      expect(screen.getByText('0.1')).toBeInTheDocument();
    });
  });

  describe('Size Variants', () => {
    it('should render large size variant', () => {
      const { container } = render(
        <ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" size="large" />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.text-6xl');
      expect(scoreElement).toBeInTheDocument();
    });

    it('should render medium size variant', () => {
      const { container } = render(
        <ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" size="medium" />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.text-4xl');
      expect(scoreElement).toBeInTheDocument();
    });

    it('should render small size variant', () => {
      const { container } = render(
        <ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" size="small" />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.text-2xl');
      expect(scoreElement).toBeInTheDocument();
    });
  });

  describe('Compact Mode', () => {
    it('should render in compact mode', () => {
      const { container } = render(
        <ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" compact />,
        { wrapper }
      );

      const scoreElement = container.querySelector('.compact-mode');
      expect(scoreElement).toBeInTheDocument();
    });

    it('should hide label in compact mode', () => {
      render(<ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" compact />, { wrapper });

      expect(screen.queryByText('Base Score')).not.toBeInTheDocument();
    });
  });

  describe('Animation', () => {
    it('should animate score changes', () => {
      const { rerender } = render(
        <ScoreDisplay score={5.5} severity="MEDIUM" label="Base Score" />,
        { wrapper }
      );

      rerender(<ScoreDisplay score={9.8} severity="CRITICAL" label="Base Score" />);

      const scoreElement = screen.getByText('9.8');
      expect(scoreElement).toHaveClass('transition-all');
    });
  });

  describe('With Breakdown', () => {
    it('should show exploitability and impact scores when breakdown provided', () => {
      render(
        <ScoreDisplay
          score={9.8}
          severity="CRITICAL"
          label="Base Score"
          breakdown={{ exploitability: 3.9, impact: 5.9 }}
        />,
        { wrapper }
      );

      expect(screen.getByText(/exploitability/i)).toBeInTheDocument();
      expect(screen.getByText(/impact/i)).toBeInTheDocument();
    });

    it('should format breakdown scores correctly', () => {
      render(
        <ScoreDisplay
          score={9.8}
          severity="CRITICAL"
          label="Base Score"
          breakdown={{ exploitability: 3.9, impact: 5.9 }}
        />,
        { wrapper }
      );

      expect(screen.getByText('3.9')).toBeInTheDocument();
      expect(screen.getByText('5.9')).toBeInTheDocument();
    });
  });
});
