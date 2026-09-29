import { describe, it, expect, afterEach } from "vitest";
import { isWails, isWailsRuntime } from "./wails";

describe("wails utils", () => {
  afterEach(() => {
    delete (globalThis as any).window;
  });

  it("isWails returns false when window is undefined", () => {
    delete (globalThis as any).window;
    expect(isWails()).toBe(false);
  });

  it("isWails returns false when window.go is undefined", () => {
    (globalThis as any).window = {};
    expect(isWails()).toBe(false);
  });

  it("isWails returns true when window.go.main.App is defined", () => {
    (globalThis as any).window = {
      go: {
        main: {
          App: {},
        },
      },
    };
    expect(isWails()).toBe(true);
  });

  it("isWailsRuntime returns false when window.runtime is undefined", () => {
    (globalThis as any).window = {};
    expect(isWailsRuntime()).toBe(false);
  });

  it("isWailsRuntime returns true when window.runtime is defined", () => {
    (globalThis as any).window = {
      runtime: {
        EventsOn: () => {},
      },
    };
    expect(isWailsRuntime()).toBe(true);
  });
});
