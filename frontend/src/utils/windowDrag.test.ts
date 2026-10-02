import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { isClickOnScrollbar, triggerNativeDrag } from "./windowDrag";

describe("windowDrag utility", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    delete (globalThis as any).window;
    delete (globalThis as any).document;
  });

  it("returns false if window or document is undefined", () => {
    delete (globalThis as any).window;
    delete (globalThis as any).document;
    const evt = { clientX: 100, clientY: 100 } as MouseEvent;
    expect(isClickOnScrollbar(evt)).toBe(false);
  });

  it("detects viewport vertical scrollbar clicks", () => {
    const mockDoc = {
      scrollHeight: 1200,
      clientHeight: 600,
      clientWidth: 985,
      scrollWidth: 1000,
    };
    (globalThis as any).document = {
      documentElement: mockDoc,
      body: { scrollHeight: 1200, scrollWidth: 1000 },
    };
    (globalThis as any).window = {
      innerWidth: 1000,
      innerHeight: 600,
    };

    // Click inside content area
    const contentClick = { clientX: 500, clientY: 300 } as MouseEvent;
    expect(isClickOnScrollbar(contentClick)).toBe(false);

    // Click on vertical scrollbar (clientX >= 985)
    const scrollbarClick = { clientX: 990, clientY: 300 } as MouseEvent;
    expect(isClickOnScrollbar(scrollbarClick)).toBe(true);
  });

  it("detects element scrollbar clicks", () => {
    const mockContainer = {
      scrollHeight: 500,
      clientHeight: 200,
      scrollWidth: 300,
      clientWidth: 285,
      offsetWidth: 300,
      offsetHeight: 200,
      clientLeft: 0,
      clientTop: 0,
      getBoundingClientRect: () => ({
        left: 100,
        right: 400,
        top: 50,
        bottom: 250,
      }),
      parentElement: null,
    };

    (globalThis as any).document = {
      documentElement: { scrollHeight: 600, clientHeight: 600, scrollWidth: 1000, clientWidth: 1000 },
      body: { scrollHeight: 600, scrollWidth: 1000 },
    };
    (globalThis as any).window = {
      innerWidth: 1000,
      innerHeight: 600,
      getComputedStyle: () => ({
        overflowY: "auto",
        overflowX: "hidden",
      }),
    };

    const eventOnScrollbar = {
      clientX: 390,
      clientY: 100,
      composedPath: () => [mockContainer],
      target: mockContainer,
    } as unknown as MouseEvent;

    expect(isClickOnScrollbar(eventOnScrollbar)).toBe(true);

    const eventInsideContent = {
      clientX: 200,
      clientY: 100,
      composedPath: () => [mockContainer],
      target: mockContainer,
    } as unknown as MouseEvent;

    expect(isClickOnScrollbar(eventInsideContent)).toBe(false);
  });

  it("triggers Wails drag via WailsInvoke or WindowStartDrag", () => {
    const mockWindow = {
      WailsInvoke: vi.fn(),
    };
    (globalThis as any).window = mockWindow;

    triggerNativeDrag();
    expect(mockWindow.WailsInvoke).toHaveBeenCalledWith("drag");

    delete (mockWindow as any).WailsInvoke;
    (mockWindow as any).runtime = { WindowStartDrag: vi.fn() };
    triggerNativeDrag();
    expect((mockWindow as any).runtime.WindowStartDrag).toHaveBeenCalled();
  });
});
