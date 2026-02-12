# v2e Portal Implementation Plan

**Version:** 2.0.0
**Status:** Ready for Implementation
**Last Updated:** 2026-02-12
**Platform:** Desktop Only (1024px+)

---

## Overview

This implementation plan outlines the phased development of the v2e Portal - a macOS Desktop-Inspired Application Portal. Each phase includes detailed cost estimation (Lines of Code), acceptance criteria, change risk assessment, dependencies, and integration points.

**Total Estimated LoC:** ~9,500 (including tests)
**Total Estimated Duration:** 5 weeks
**Overall Risk Level:** Medium

---

## Phase 1: Core Desktop Infrastructure (Week 1)

**Risk Level:** MEDIUM
**Estimated LoC:** 1,800 (1,200 implementation + 600 tests)
**Estimated Time:** 5 business days

**Dependencies:** None (foundational phase)

### Tasks

| Task | Est. LoC | Acceptance Criteria | Risk |
|------|-----------|---------------------|------|
| 1.1 Create desktop directory structure | 50 | All directories per design.md exist | Low |
| 1.2 Set up desktop state management (Zustand) | 300 | Store created with persist middleware | Low |
| 1.3 Build basic desktop layout | 400 | Desktop renders at `/desktop` route | Medium |
| 1.4 Implement desktop icon component | 250 | Icons visible and selectable | Low |
| 1.5 Add drag-and-drop for icons | 200 | Icons draggable, snap to grid | Medium |

### Deliverables

- [ ] Desktop renders at `/desktop` (new root page)
- [ ] Desktop icons visible and selectable
- [ ] Dock renders with basic items
- [ ] Menu bar with search button
- [ ] Icon positions persist via Zustand persist
- [ ] Unit tests pass (`./build.sh -t`)
- [ ] TypeScript types defined in `types/desktop.ts`

### Integration Points

- **Integrates with:** Next.js App Router, Zustand state management
- **Consumes from:** `lib/portal/app-registry.ts` for app definitions
- **Provides to:** All subsequent phases (foundation layer)

### Change Risk Analysis

**Why Medium Risk:**
- Introduces new root route (`/desktop`) that changes application landing page
- Requires Zustand middleware configuration that must be tested thoroughly
- Grid snapping logic has edge cases (screen boundaries, overlapping icons)

**Mitigation Strategies:**
1. Feature flag the new `/desktop` route initially
2. Use parameterized tests for grid boundary conditions
3. Implement graceful degradation if localStorage is unavailable

---

## Phase 2: Window System (Week 2)

**Risk Level:** HIGH
**Estimated LoC:** 2,500 (1,700 implementation + 800 tests)
**Estimated Time:** 5 business days

**Dependencies:** Phase 1 complete

### Tasks

| Task | Est. LoC | Acceptance Criteria | Risk |
|------|-----------|---------------------|------|
| 2.1 Build window component with titlebar | 400 | Window renders with title and controls | Medium |
| 2.2 Implement window controls (close/min/max) | 350 | All buttons functional | Medium |
| 2.3 Add window dragging | 400 | Window moves smoothly on titlebar drag | High |
| 2.4 Add window resizing (8 directions) | 350 | Resize handles on all edges/corners | High |
| 2.5 Implement window layering (z-index) | 200 | Focus brings window to front | Low |
| 2.6 Add window animations (open/close/minimize) | 300 | Smooth animations for all operations | Medium |

### Deliverables

- [ ] Windows open when double-clicking desktop icons
- [ ] Windows are movable via titlebar drag
- [ ] Windows are resizable via edge/corner handles
- [ ] Window controls (close/min/max) work correctly
- [ ] Smooth animations (200ms open, 150ms close, 300ms minimize)
- [ ] Z-index management handles 10+ concurrent windows
- [ ] Minimizing windows creates dock thumbnail placeholder
- [ ] Unit and integration tests pass

### Integration Points

- **Integrates with:** Desktop state from Phase 1
- **Consumes from:** `lib/desktop/desktop-state.ts`, `APP_REGISTRY`
- **Provides to:** Phase 3 (dock integration), Phase 4 (app content)

### Change Risk Analysis

**Why High Risk:**
- Complex z-index management with edge cases (rapid clicking, minimize during drag)
- Resize logic has many edge cases (minimum size enforcement, bounds checking)
- Animation timing conflicts (close during open, minimize during drag)
- Browser-specific drag behaviors (especially Firefox)

**Mitigation Strategies:**
1. Use requestAnimationFrame for all drag/resize operations
2. Implement mutex locks to prevent conflicting state changes
3. Cross-browser testing in Chrome, Firefox, Safari, Edge
4. Resize observers for viewport boundary detection

---

## Phase 3: Dock & Quick Launch (Week 3)

**Risk Level:** MEDIUM
**Estimated LoC:** 2,200 (1,500 implementation + 700 tests)
**Estimated Time:** 5 business days

**Dependencies:** Phase 1, Phase 2 complete

### Tasks

| Task | Est. LoC | Acceptance Criteria | Risk |
|------|-----------|---------------------|------|
| 3.1 Build dock component with magnification | 400 | Dock renders with hover effects | Low |
| 3.2 Implement dock item interactions | 300 | Click/right-click work | Medium |
| 3.3 Add quick launch modal (Cmd+K) | 350 | Modal opens with keyboard shortcut | Medium |
| 3.4 Implement search functionality | 250 | Search filters apps correctly | Low |
| 3.5 Add context menus for all elements | 400 | Context menus work on desktop/windows/dock | High |
| 3.6 Dock-thumbnail integration | 200 | Minimized windows show in dock | Medium |
| 3.7 Dock item reordering (drag) | 300 | Items draggable to reorder | Medium |

### Deliverables

- [ ] Dock is fully functional with magnification effect
- [ ] Quick launch opens with Cmd+K keyboard shortcut
- [ ] Search filters apps by name and tags
- [ ] Context menus work for desktop icons, windows, and dock items
- [ ] Right-click menus include: Open, Show Info, Remove, Add to Dock
- [ ] Dock indicators show running apps and window counts
- [ ] Minimized windows display as thumbnails in dock
- [ ] Dock items are draggable to reorder
- [ ] All interactions tested in unit and integration tests

### Integration Points

- **Integrates with:** Window system from Phase 2, Desktop state from Phase 1
- **Consumes from:** `lib/desktop/desktop-state.ts`, `lib/desktop/window-manager.ts`
- **Provides to:** Phase 4 (complete desktop environment)

### Change Risk Analysis

**Why Medium Risk:**
- Global keyboard shortcuts (Cmd+K) may conflict with browser/native OS
- Context menu positioning has edge cases (viewport boundaries)
- Dock magnification math can be jittery at edges
- Thumbnail rendering for minimized windows is complex

**Mitigation Strategies:**
1. Test keyboard shortcuts across all OS/browser combinations
2. Use viewport boundary detection for context menu positioning
3. CSS transforms for smooth magnification (GPU-accelerated)
4. HTML canvas for dock thumbnails (better performance than iframe screenshots)

---

## Phase 4: Content Integration (Week 4)

**Risk Level:** MEDIUM
**Estimated LoC:** 1,800 (1,200 implementation + 600 tests)
**Estimated Time:** 5 business days

**Dependencies:** Phase 1, Phase 2, Phase 3 complete

### Tasks

| Task | Est. LoC | Acceptance Criteria | Risk |
|------|-----------|---------------------|------|
| 4.1 Integrate existing apps as window content | 400 | All apps load in windows | Medium |
| 4.2 Implement iframe-based app loading | 300 | Iframes load correctly with isolation | Medium |
| 4.3 Add window state persistence | 300 | Window positions save/restore | Low |
| 4.4 Handle app-to-app communication | 250 | PostMessage bridge working | High |
| 4.5 Add desktop widgets (clock, system info) | 200 | Widgets display and update | Low |
| 4.6 Menu bar active app integration | 350 | Menu bar updates with focused window | Low |

### Deliverables

- [ ] All existing apps (CVE, CWE, CAPEC, ATT&CK, CVSS, GLC, Mcards, ETL, Bookmarks) open in windows
- [ ] Window positions persist across browser sessions
- [ ] Desktop layout (icon positions) saves between sessions
- [ ] App-to-app communication works via PostMessage bridge
- [ ] Menu bar shows active app name
- [ ] Desktop widgets (clock, date) display correctly
- [ ] Iframe isolation prevents CSS/JS conflicts
- [ ] Integration tests verify all apps load successfully

### Integration Points

- **Integrates with:** All existing v2e applications
- **Consumes from:** `APP_REGISTRY`, all existing app routes
- **Provides to:** Phase 5 (production optimization)

### Change Risk Analysis

**Why Medium Risk:**
- Iframe loading introduces CORS issues
- PostMessage communication has security implications
- Some existing apps may not work well in iframes (viewport expectations)
- State persistence must handle edge cases (corrupted storage, quota exceeded)

**Mitigation Strategies:**
1. Implement PostMessage whitelist for security
2. Add CSP headers for iframe isolation
3. Graceful fallback for apps that can't run in iframes
4. Storage quota monitoring and cleanup strategies
5. Migration path for state schema changes

---

## Phase 5: Polish & Production Launch (Week 5)

**Risk Level:** LOW
**Estimated LoC:** 1,200 (800 implementation + 400 tests)
**Estimated Time:** 5 business days

**Dependencies:** All previous phases complete

### Tasks

| Task | Est. LoC | Acceptance Criteria | Risk |
|------|-----------|---------------------|------|
| 5.1 Add wallpaper selector | 150 | Users can choose wallpapers | Low |
| 5.2 Implement theme variations (5 themes) | 200 | Light/dark/custom themes work | Low |
| 5.3 Performance optimization | 200 | Lighthouse score > 90 | Medium |
| 5.4 Cross-browser testing suite | 100 | All browsers pass tests | Low |
| 5.5 Accessibility audit and fixes | 150 | WCAG 2.1 AA compliant | Medium |
| 5.6 Documentation and user guide | 150 | README and user docs complete | Low |
| 5.7 Production deployment | 100 | Deployed to production | Low |
| 5.8 Bug fixes from testing | 150 | All critical bugs resolved | Low |

### Deliverables

- [ ] Wallpaper selector with 5 preset options
- [ ] 5 theme variations (light, dark, high contrast)
- [ ] Lighthouse performance score > 90
- [ ] Lighthouse accessibility score > 90
- [ ] Lighthouse best practices score > 90
- [ ] WCAG 2.1 AA compliance verified
- [ ] Smooth 60fps animations (verified via Chrome DevTools)
- [ ] Cross-browser compatibility: Chrome 120+, Firefox 121+, Safari 17+, Edge 120+
- [ ] User documentation in README
- [ ] Production deployment completed
- [ ] All integration tests passing
- [ ] No console errors on production build

### Integration Points

- **Integrates with:** All previous phases
- **Consumes from:** All components and state
- **Provides to:** Production users, maintenance team

### Change Risk Analysis

**Why Low Risk:**
- Mostly additive features (themes, wallpapers)
- No breaking changes to existing functionality
- Bug fixes only address reported issues
- Deployment uses existing CI/CD pipeline

**Mitigation Strategies:**
1. A/B test theme changes with small user group
2. Performance budgets in CI pipeline (fail build if budgets exceeded)
3. Automated accessibility testing in CI (axe-core)
4. Canary deployment before full rollout

---

## Summary

### Total Project Metrics

| Metric | Value |
|---------|--------|
| **Total Estimated LoC** | 9,500 (6,400 impl + 3,100 tests) |
| **Total Duration** | 5 weeks (25 business days) |
| **High Risk Phases** | 1 (Phase 2: Window System) |
| **Medium Risk Phases** | 3 (Phases 1, 3, 4) |
| **Low Risk Phases** | 1 (Phase 5) |
| **Critical Path** | Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 |
| **Parallelizable Tasks** | Dock widgets, theme variations, documentation |

### Risk Heatmap

| Phase | Risk | Impact | Likelihood | Severity |
|--------|-------|---------|------------|----------|
| Phase 1 | Medium | Low | Medium | 6/10 |
| Phase 2 | High | High | Medium | 8/10 |
| Phase 3 | Medium | Medium | Low | 5/10 |
| Phase 4 | Medium | High | Low | 6/10 |
| Phase 5 | Low | Low | Low | 2/10 |

### Dependencies Graph

```
Phase 1 (Foundation)
    │
    ├── Phase 2 (Window System)
    │       │
    │       └── Phase 3 (Dock & Quick Launch)
    │               │
    │               └── Phase 4 (Content Integration)
    │                       │
    │                       └── Phase 5 (Polish & Launch)
    │
    └── Cannot proceed to any phase without Phase 1
```

### Acceptance Gates

Each phase must meet the following gates before proceeding:

1. **Code Review**: All changes reviewed by at least one team member
2. **Test Coverage**: Unit tests > 80%, integration tests covering critical paths
3. **Build Success**: `./build.sh -t` passes without errors
4. **Manual Testing**: Smoke test on Chrome, Firefox, Safari
5. **Documentation**: RPC APIs documented in `service.md` (if applicable)
6. **Performance**: No performance regressions from previous phase

### Rollback Strategy

If a phase introduces critical issues:

1. **Phase 1**: Feature flag `/desktop` route, revert to old landing page
2. **Phase 2**: Revert window state changes, keep Phase 1
3. **Phase 3**: Disable dock via config, use desktop-only navigation
4. **Phase 4**: Fallback to direct app links (bypass windows)
5. **Phase 5**: Revert to previous theme/wallpaper defaults

### Success Criteria

The entire implementation is successful when:

- [ ] All 5 phases completed
- [ ] All acceptance criteria met
- [ ] Lighthouse scores > 90 (Performance, Accessibility, Best Practices)
- [ ] WCAG 2.1 AA compliance verified
- [ ] Cross-browser compatibility achieved
- [ ] Production deployment stable for 7 days
- [ ] User feedback positive (> 4.0/5.0 rating)
- [ ] No critical bugs in issue tracker

---

**Plan Status:** Ready for Execution
**Next Steps:** Begin Phase 1 development
**Owner:** v2e Development Team
**Review Date:** Weekly during implementation
