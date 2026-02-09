import { StateCreator } from 'zustand';

interface UIState {
  theme: 'light' | 'dark';
  sidebarOpen: boolean;
  selectedTab: 'nodes' | 'edges' | 'properties';
  showNodeDetails: boolean;
  showEdgeDetails: boolean;
  showMiniMap: boolean;
  showControls: boolean;
  setTheme: (theme: 'light' | 'dark') => void;
  setSidebarOpen: (open: boolean) => void;
  setSelectedTab: (tab: 'nodes' | 'edges' | 'properties') => void;
  setShowNodeDetails: (show: boolean) => void;
  setShowEdgeDetails: (show: boolean) => void;
  setShowMiniMap: (show: boolean) => void;
  setShowControls: (show: boolean) => void;
  toggleSidebar: () => void;
  toggleNodeDetails: () => void;
  toggleEdgeDetails: () => void;
  toggleMiniMap: () => void;
  resetUI: () => void;
}

export const createUISlice: StateCreator<UIState> = (set, get) => ({
  theme: 'dark',
  sidebarOpen: true,
  selectedTab: 'nodes',
  showNodeDetails: false,
  showEdgeDetails: false,
  showMiniMap: true,
  showControls: true,
  setTheme: (theme) => set({ theme }),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  setSelectedTab: (tab) => set({ selectedTab: tab }),
  setShowNodeDetails: (show) => set({ showNodeDetails: show }),
  setShowEdgeDetails: (show) => set({ showEdgeDetails: show }),
  setShowMiniMap: (show) => set({ showMiniMap: show }),
  setShowControls: (show) => set({ showControls: show }),
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  toggleNodeDetails: () => set((state) => ({ showNodeDetails: !state.showNodeDetails })),
  toggleEdgeDetails: () => set((state) => ({ showEdgeDetails: !state.showEdgeDetails })),
  toggleMiniMap: () => set((state) => ({ showMiniMap: !state.showMiniMap })),
  resetUI: () => set({
    theme: 'dark',
    sidebarOpen: true,
    selectedTab: 'nodes',
    showNodeDetails: false,
    showEdgeDetails: false,
    showMiniMap: true,
    showControls: true,
  }),
});

export default createUISlice;
