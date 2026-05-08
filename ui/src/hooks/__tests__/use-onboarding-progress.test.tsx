import { renderHook, act } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { mockAuthContext, mockUserTokens } from "@/test/test-utils";

vi.mock("@/hooks/use-auth-provider", () => ({
  useAuthProvider: vi.fn(),
}));

import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useOnboardingProgress } from "../use-onboarding-progress";

const mockUseAuth = vi.mocked(useAuthProvider);

const ALL_FALSE = {
  saidHello: false,
  hasProject: false,
  visitedPricing: false,
  visitedRps: false,
  dismissed: false,
};

describe("useOnboardingProgress", () => {
  beforeEach(() => {
    localStorage.clear();
    mockUseAuth.mockReturnValue(mockAuthContext);
  });

  it("returns all-false defaults on first use", () => {
    const { result } = renderHook(() => useOnboardingProgress());
    expect(result.current.progress).toEqual(ALL_FALSE);
  });

  it("markStep sets the targeted step to true", () => {
    const { result } = renderHook(() => useOnboardingProgress());
    act(() => result.current.markStep("saidHello"));
    expect(result.current.progress.saidHello).toBe(true);
  });

  it("markStep leaves other steps untouched", () => {
    const { result } = renderHook(() => useOnboardingProgress());
    act(() => result.current.markStep("visitedPricing"));
    expect(result.current.progress.saidHello).toBe(false);
    expect(result.current.progress.visitedPricing).toBe(true);
    expect(result.current.progress.visitedRps).toBe(false);
  });

  it("markStep is a no-op when step is already true", () => {
    localStorage.setItem(
      `onboarding_${mockUserTokens.user.id}`,
      JSON.stringify({ ...ALL_FALSE, saidHello: true }),
    );
    const { result } = renderHook(() => useOnboardingProgress());
    const before = result.current.progress;
    act(() => result.current.markStep("saidHello"));
    // state reference unchanged — no re-render triggered
    expect(result.current.progress.saidHello).toBe(true);
    expect(result.current.progress).toEqual(before);
  });

  it("multiple steps can be marked independently", () => {
    const { result } = renderHook(() => useOnboardingProgress());
    act(() => result.current.markStep("saidHello"));
    act(() => result.current.markStep("visitedRps"));
    expect(result.current.progress.saidHello).toBe(true);
    expect(result.current.progress.visitedRps).toBe(true);
    expect(result.current.progress.visitedPricing).toBe(false);
  });

  it("dismiss sets dismissed to true", () => {
    const { result } = renderHook(() => useOnboardingProgress());
    act(() => result.current.dismiss());
    expect(result.current.progress.dismissed).toBe(true);
  });

  it("persists progress to localStorage under the user-scoped key", () => {
    const { result } = renderHook(() => useOnboardingProgress());
    act(() => result.current.markStep("visitedPricing"));
    const stored = JSON.parse(
      localStorage.getItem(`onboarding_${mockUserTokens.user.id}`)!,
    );
    expect(stored.visitedPricing).toBe(true);
  });

  it("reads existing progress from localStorage on mount", () => {
    localStorage.setItem(
      `onboarding_${mockUserTokens.user.id}`,
      JSON.stringify({ ...ALL_FALSE, hasProject: true, visitedRps: true }),
    );
    const { result } = renderHook(() => useOnboardingProgress());
    expect(result.current.progress.hasProject).toBe(true);
    expect(result.current.progress.visitedRps).toBe(true);
    expect(result.current.progress.saidHello).toBe(false);
  });

  it("uses 'anon' key when no user is logged in", () => {
    mockUseAuth.mockReturnValue({ ...mockAuthContext, user: null });
    const { result } = renderHook(() => useOnboardingProgress());
    act(() => result.current.markStep("saidHello"));
    const stored = JSON.parse(localStorage.getItem("onboarding_anon")!);
    expect(stored.saidHello).toBe(true);
    expect(localStorage.getItem(`onboarding_${mockUserTokens.user.id}`)).toBeNull();
  });
});
