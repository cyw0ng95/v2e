/**
 * Canvas Toolbar Component Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CanvasToolbar } from '@/components/glc/toolbar/canvas-toolbar';
import { D3FEND_PRESET } from '@/lib/glc/presets';
import { ReactFlowProvider } from '@xyflow/react';

describe('CanvasToolbar', () => {
  const renderWithProvider = (ui: React.ReactElement) => {
    return render(
      <ReactFlowProvider>
        {ui}
      </ReactFlowProvider>
    );
  };

  const mockHandlers = {
    onShowShortcuts: vi.fn(),
    onShowExport: vi.fn(),
    onShowShare: vi.fn(),
    onShowInferences: vi.fn(),
    onShowSTIXImport: vi.fn(),
  };

  it('should render toolbar buttons', () => {
    renderWithProvider(
      <CanvasToolbar
        preset={D3FEND_PRESET}
        graphName="Test Graph"
        {...mockHandlers}
      />
    );

    expect(screen.getByTitle(/undo/i)).toBeInTheDocument();
    expect(screen.getByTitle(/redo/i)).toBeInTheDocument();
    expect(screen.getByTitle(/zoom in/i)).toBeInTheDocument();
    expect(screen.getByTitle(/zoom out/i)).toBeInTheDocument();
  });

  it('should display graph name', () => {
    renderWithProvider(
      <CanvasToolbar
        preset={D3FEND_PRESET}
        graphName="My Test Graph"
        {...mockHandlers}
      />
    );

    expect(screen.getByText('My Test Graph')).toBeInTheDocument();
  });

  it('should call onShowShortcuts when help clicked', () => {
    renderWithProvider(
      <CanvasToolbar
        preset={D3FEND_PRESET}
        graphName="Test"
        {...mockHandlers}
      />
    );

    const helpButton = screen.getByTitle(/help/i);
    fireEvent.click(helpButton);

    expect(mockHandlers.onShowShortcuts).toHaveBeenCalledTimes(1);
  });

  it('should call onShowExport when export clicked', () => {
    renderWithProvider(
      <CanvasToolbar
        preset={D3FEND_PRESET}
        graphName="Test"
        {...mockHandlers}
      />
    );

    const exportButton = screen.getByTitle(/export/i);
    fireEvent.click(exportButton);

    expect(mockHandlers.onShowExport).toHaveBeenCalledTimes(1);
  });

  it('should call onShowShare when share clicked', () => {
    renderWithProvider(
      <CanvasToolbar
        preset={D3FEND_PRESET}
        graphName="Test"
        {...mockHandlers}
      />
    );

    const shareButton = screen.getByTitle(/share/i);
    fireEvent.click(shareButton);

    expect(mockHandlers.onShowShare).toHaveBeenCalledTimes(1);
  });

  it('should show D3FEND inference button for D3FEND preset', () => {
    renderWithProvider(
      <CanvasToolbar
        preset={D3FEND_PRESET}
        graphName="Test"
        {...mockHandlers}
      />
    );

    expect(screen.getByTitle(/inference/i)).toBeInTheDocument();
  });

  it('should call onShowInferences when inference clicked', () => {
    renderWithProvider(
      <CanvasToolbar
        preset={D3FEND_PRESET}
        graphName="Test"
        {...mockHandlers}
      />
    );

    const inferenceButton = screen.getByTitle(/inference/i);
    fireEvent.click(inferenceButton);

    expect(mockHandlers.onShowInferences).toHaveBeenCalledTimes(1);
  });
});
