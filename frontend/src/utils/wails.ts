/**
 * Wails environment detection and utility helpers.
 * Copyright (c) 2026 SnowdreamTech. All rights reserved.
 * Licensed under the MIT License.
 */

/**
 * Checks whether the application is running inside the Wails desktop webview container.
 */
export const isWails = (): boolean => {
  return typeof window !== "undefined" && Boolean((window as any)?.go?.main?.App);
};

/**
 * Checks whether Wails runtime events and window/browser APIs are available.
 */
export const isWailsRuntime = (): boolean => {
  return typeof window !== "undefined" && Boolean((window as any)?.runtime);
};
