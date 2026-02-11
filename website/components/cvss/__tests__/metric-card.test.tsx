/**
 * Metric Card Component Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import MetricCard from '../metric-card';
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

describe('MetricCard Component', () => {
  const defaultProps = {
    metric: 'AV',
    label: 'Attack Vector',
    description: 'How vulnerable is the component?',
    value: 'N',
    onChange: vi.fn()
  };

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <CVSSProvider>{children}</CVSSProvider>
  );

  describe('Rendering', () => {
    it('should render metric label', () => {
      render(<MetricCard {...defaultProps} />, { wrapper });

      expect(screen.getByText('Attack Vector')).toBeInTheDocument();
    });

    it('should render metric description', () => {
      render(<MetricCard {...defaultProps} />, { wrapper });

      expect(screen.getByText('How vulnerable is the component?')).toBeInTheDocument();
    });

    it('should render metric code', () => {
      render(<MetricCard {...defaultProps} />, { wrapper });

      expect(screen.getByText('AV')).toBeInTheDocument();
    });

    it('should render current value', () => {
      render(<MetricCard {...defaultProps} value="N" />, { wrapper });

      expect(screen.getByText('N')).toBeInTheDocument();
    });
  });

  describe('Options Display', () => {
    const optionsProps = {
      ...defaultProps,
      options: [
        { value: 'N', label: 'Network', description: 'Network exploitable' },
        { value: 'A', label: 'Adjacent', description: 'Adjacent network' },
        { value: 'L', label: 'Local', description: 'Local access' },
        { value: 'P', label: 'Physical', description: 'Physical access' }
      ]
    };

    it('should render all option buttons', () => {
      render(<MetricCard {...optionsProps} />, { wrapper });

      expect(screen.getByText('Network')).toBeInTheDocument();
      expect(screen.getByText('Adjacent')).toBeInTheDocument();
      expect(screen.getByText('Local')).toBeInTheDocument();
      expect(screen.getByText('Physical')).toBeInTheDocument();
    });

    it('should highlight selected option', () => {
      const { container } = render(<MetricCard {...optionsProps} value="N" />, { wrapper });

      const selectedButton = container.querySelector('[data-value="N"].selected');
      expect(selectedButton).toBeInTheDocument();
    });

    it('should show option descriptions', () => {
      render(<MetricCard {...optionsProps} showDescriptions />, { wrapper });

      expect(screen.getByText('Network exploitable')).toBeInTheDocument();
      expect(screen.getByText('Adjacent network')).toBeInTheDocument();
    });

    it('should hide option descriptions by default', () => {
      const { container } = render(<MetricCard {...optionsProps} />, { wrapper });

      const descriptions = container.querySelectorAll('.option-description');
      expect(descriptions.length).toBe(0);
    });
  });

  describe('User Interactions', () => {
    it('should call onChange when option is clicked', () => {
      const handleChange = vi.fn();
      const props = {
        ...defaultProps,
        onChange: handleChange,
        options: [
          { value: 'N', label: 'Network', description: 'Network exploitable' },
          { value: 'A', label: 'Adjacent', description: 'Adjacent network' }
        ]
      };

      render(<MetricCard {...props} />, { wrapper });

      fireEvent.click(screen.getByText('Network'));
      expect(handleChange).toHaveBeenCalledWith('N');
    });

    it('should update visual state when selection changes', () => {
      const handleChange = vi.fn();
      const props = {
        ...defaultProps,
        onChange: handleChange,
        options: [
          { value: 'N', label: 'Network', description: 'Network exploitable' },
          { value: 'A', label: 'Adjacent', description: 'Adjacent network' }
        ]
      };

      const { container, rerender } = render(<MetricCard {...props} value="N" />, { wrapper });

      let selectedButton = container.querySelector('[data-value="N"].selected');
      expect(selectedButton).toBeInTheDocument();

      rerender(<MetricCard {...props} value="A" />);

      selectedButton = container.querySelector('[data-value="A"].selected');
      expect(selectedButton).toBeInTheDocument();
    });
  });

  describe('Compact Mode', () => {
    it('should render in compact mode', () => {
      const { container } = render(
        <MetricCard {...defaultProps} compact />,
        { wrapper }
      );

      const card = container.querySelector('.metric-card-compact');
      expect(card).toBeInTheDocument();
    });

    it('should hide description in compact mode', () => {
      render(<MetricCard {...defaultProps} compact />, { wrapper });

      expect(screen.queryByText(defaultProps.description)).not.toBeInTheDocument();
    });

    it('should use smaller buttons in compact mode', () => {
      const { container } = render(
        <MetricCard
          {...defaultProps}
          compact
          options={[
            { value: 'N', label: 'Network', description: 'Network exploitable' }
          ]}
        />,
        { wrapper }
      );

      const button = container.querySelector('.option-button-compact');
      expect(button).toBeInTheDocument();
    });
  });

  describe('Grouped Metrics', () => {
    it('should display group label when provided', () => {
      render(
        <MetricCard {...defaultProps} group="Base Metrics" groupColor="blue" />,
        { wrapper }
      );

      expect(screen.getByText('Base Metrics')).toBeInTheDocument();
    });

    it('should apply group color styling', () => {
      const { container } = render(
        <MetricCard {...defaultProps} group="Base Metrics" groupColor="blue" />,
        { wrapper }
      );

      const groupLabel = container.querySelector('.group-label-blue');
      expect(groupLabel).toBeInTheDocument();
    });
  });

  describe('Tooltips', () => {
    it('should show tooltip on hover', () => {
      render(<MetricCard {...defaultProps} tooltip="Custom tooltip text" />, { wrapper });

      const card = screen.getByText('Attack Vector').closest('.metric-card');
      fireEvent.mouseEnter(card!);

      expect(screen.getByText('Custom tooltip text')).toBeInTheDocument();
    });

    it('should hide tooltip on mouse leave', () => {
      render(<MetricCard {...defaultProps} tooltip="Custom tooltip text" />, { wrapper });

      const card = screen.getByText('Attack Vector').closest('.metric-card');
      fireEvent.mouseEnter(card!);
      expect(screen.getByText('Custom tooltip text')).toBeInTheDocument();

      fireEvent.mouseLeave(card!);
      expect(screen.queryByText('Custom tooltip text')).not.toBeInTheDocument();
    });
  });

  describe('Disabled State', () => {
    it('should not call onChange when disabled', () => {
      const handleChange = vi.fn();
      const props = {
        ...defaultProps,
        onChange: handleChange,
        disabled: true,
        options: [
          { value: 'N', label: 'Network', description: 'Network exploitable' }
        ]
      };

      render(<MetricCard {...props} />, { wrapper });

      fireEvent.click(screen.getByText('Network'));
      expect(handleChange).not.toHaveBeenCalled();
    });

    it('should apply disabled styling', () => {
      const { container } = render(
        <MetricCard {...defaultProps} disabled />,
        { wrapper }
      );

      const card = container.querySelector('.metric-card.disabled');
      expect(card).toBeInTheDocument();
    });
  });

  describe('Visual States', () => {
    it('should apply error state styling', () => {
      const { container } = render(
        <MetricCard {...defaultProps} error />,
        { wrapper }
      );

      const card = container.querySelector('.metric-card.error');
      expect(card).toBeInTheDocument();
    });

    it('should apply warning state styling', () => {
      const { container } = render(
        <MetricCard {...defaultProps} warning />,
        { wrapper }
      );

      const card = container.querySelector('.metric-card.warning');
      expect(card).toBeInTheDocument();
    });

    it('should apply success state styling', () => {
      const { container } = render(
        <MetricCard {...defaultProps} success />,
        { wrapper }
      );

      const card = container.querySelector('.metric-card.success');
      expect(card).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have accessible label', () => {
      render(<MetricCard {...defaultProps} />, { wrapper });

      const label = screen.getByText('Attack Vector');
      expect(label).toHaveAccessibleName();
    });

    it('should provide accessible options', () => {
      const props = {
        ...defaultProps,
        options: [
          { value: 'N', label: 'Network', description: 'Network exploitable' }
        ]
      };

      render(<MetricCard {...props} />, { wrapper });

      const button = screen.getByRole('button', { name: /Network/i });
      expect(button).toBeInTheDocument();
    });

    it('should announce changes to screen readers', () => {
      const handleChange = vi.fn();
      const props = {
        ...defaultProps,
        onChange: handleChange,
        options: [
          { value: 'N', label: 'Network', description: 'Network exploitable' }
        ],
        ariaLive: 'polite'
      };

      render(<MetricCard {...props} />, { wrapper });

      const liveRegion = document.querySelector('[aria-live="polite"]');
      expect(liveRegion).toBeInTheDocument();
    });

    it('should have proper role for metric selection', () => {
      const props = {
        ...defaultProps,
        options: [
          { value: 'N', label: 'Network', description: 'Network exploitable' }
        ]
      };

      render(<MetricCard {...props} />, { wrapper });

      const radiogroup = screen.getByRole('radiogroup');
      expect(radiogroup).toBeInTheDocument();
    });
  });

  describe('Different Metric Types', () => {
    it('should render Attack Vector metric', () => {
      const { container } = render(
        <MetricCard
          metric="AV"
          label="Attack Vector"
          value="N"
          onChange={vi.fn()}
          options={[
            { value: 'N', label: 'Network', description: 'Network exploitable' },
            { value: 'A', label: 'Adjacent', description: 'Adjacent network' },
            { value: 'L', label: 'Local', description: 'Local access' },
            { value: 'P', label: 'Physical', description: 'Physical access' }
          ]}
        />,
        { wrapper }
      );

      expect(container.querySelector('[data-metric="AV"]')).toBeInTheDocument();
    });

    it('should render Attack Complexity metric', () => {
      const { container } = render(
        <MetricCard
          metric="AC"
          label="Attack Complexity"
          value="L"
          onChange={vi.fn()}
          options={[
            { value: 'L', label: 'Low', description: 'Specialized access' },
            { value: 'H', label: 'High', description: 'Specialized conditions' }
          ]}
        />,
        { wrapper }
      );

      expect(container.querySelector('[data-metric="AC"]')).toBeInTheDocument();
    });

    it('should render Impact metrics (C/I/A)', () => {
      const { container: containerC } = render(
        <MetricCard
          metric="C"
          label="Confidentiality"
          value="H"
          onChange={vi.fn()}
          options={[
            { value: 'H', label: 'High', description: 'Total compromise' },
            { value: 'L', label: 'Low', description: 'Partial compromise' },
            { value: 'N', label: 'None', description: 'No impact' }
          ]}
        />,
        { wrapper }
      );

      expect(containerC.querySelector('[data-metric="C"]')).toBeInTheDocument();
    });
  });

  describe('Custom Styling', () => {
    it('should apply custom className', () => {
      const { container } = render(
        <MetricCard {...defaultProps} className="custom-class" />,
        { wrapper }
      );

      const card = container.querySelector('.custom-class');
      expect(card).toBeInTheDocument();
    });

    it('should apply custom styles', () => {
      const { container } = render(
        <MetricCard {...defaultProps} style={{ border: '2px solid red' }} />,
        { wrapper }
      );

      const card = container.querySelector('.metric-card');
      expect(card).toHaveStyle({ border: '2px solid red' });
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty options array', () => {
      render(<MetricCard {...defaultProps} options={[]} />, { wrapper });

      expect(screen.queryByRole('radiogroup')).not.toBeInTheDocument();
    });

    it('should handle undefined value', () => {
      const { container } = render(
        <MetricCard {...defaultProps} value={undefined} options={[
          { value: 'N', label: 'Network', description: 'Network exploitable' }
        ]} />,
        { wrapper }
      );

      const selectedButton = container.querySelector('.selected');
      expect(selectedButton).toBeNull();
    });

    it('should handle null options gracefully', () => {
      render(<MetricCard {...defaultProps} options={null as any} />, { wrapper });

      expect(screen.getByText('Attack Vector')).toBeInTheDocument();
    });
  });
});
