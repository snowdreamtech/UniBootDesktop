// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/**
 * Checks whether a mouse event occurred on a scrollbar (window level or container level).
 * Accurately handles both LTR and RTL orientations, classic and overlay scrollbars.
 * This prevents native window dragging from hijacking scrollbar clicks.
 */
export function isClickOnScrollbar(e: MouseEvent): boolean {
  if (typeof window === "undefined" || !e) return false;

  const doc = document.documentElement;
  const winInnerWidth = window.innerWidth;
  const winInnerHeight = window.innerHeight;
  const isDocRtl = doc.dir === "rtl" || (window.getComputedStyle && window.getComputedStyle(doc).direction === "rtl");

  // 1. Root / Document viewport vertical scrollbar
  if (doc.scrollHeight > winInnerHeight || (document.body && document.body.scrollHeight > winInnerHeight)) {
    const rootScrollbarWidth = winInnerWidth - doc.clientWidth;
    const hitWidth = Math.max(rootScrollbarWidth, 16);
    if (isDocRtl) {
      if (e.clientX <= hitWidth) return true;
    } else {
      if (e.clientX >= winInnerWidth - hitWidth) return true;
    }
  }

  // 2. Root / Document viewport horizontal scrollbar
  if (doc.scrollWidth > winInnerWidth || (document.body && document.body.scrollWidth > winInnerWidth)) {
    const rootScrollbarHeight = winInnerHeight - doc.clientHeight;
    const hitHeight = Math.max(rootScrollbarHeight, 16);
    if (e.clientY >= winInnerHeight - hitHeight) {
      return true;
    }
  }

  // 3. Element-level scrollbars across the event path
  const path = typeof e.composedPath === "function" ? e.composedPath() : [];
  const elementsToCheck: HTMLElement[] = [];
  if (path.length > 0) {
    for (const node of path) {
      if (node && typeof (node as any).getBoundingClientRect === "function") {
        elementsToCheck.push(node as HTMLElement);
      }
    }
  } else {
    let curr = e.target as HTMLElement | null;
    while (curr) {
      elementsToCheck.push(curr);
      curr = curr.parentElement;
    }
  }

  for (const el of elementsToCheck) {
    if (el === document.documentElement || el === document.body) {
      continue;
    }

    const style = window.getComputedStyle ? window.getComputedStyle(el) : ({} as CSSStyleDeclaration);
    const overflowY = style.overflowY;
    const overflowX = style.overflowX;
    const isElRtl = style.direction === "rtl";

    const hasVerticalScroll =
      (overflowY === "auto" || overflowY === "scroll" || overflowY === "overlay") && el.scrollHeight > el.clientHeight;
    const hasHorizontalScroll =
      (overflowX === "auto" || overflowX === "scroll" || overflowX === "overlay") && el.scrollWidth > el.clientWidth;

    if (!hasVerticalScroll && !hasHorizontalScroll) {
      continue;
    }

    const rect = el.getBoundingClientRect();
    if (e.clientX < rect.left || e.clientX > rect.right || e.clientY < rect.top || e.clientY > rect.bottom) {
      continue;
    }

    if (hasVerticalScroll) {
      const scrollbarWidth = el.offsetWidth - el.clientLeft - el.clientWidth;
      const hitWidth = Math.max(scrollbarWidth, 16);
      if (isElRtl) {
        if (e.clientX <= rect.left + hitWidth) {
          return true;
        }
      } else {
        if (e.clientX >= rect.right - hitWidth) {
          return true;
        }
      }
    }

    if (hasHorizontalScroll) {
      const scrollbarHeight = el.offsetHeight - el.clientTop - el.clientHeight;
      const hitHeight = Math.max(scrollbarHeight, 16);
      if (e.clientY >= rect.bottom - hitHeight) {
        return true;
      }
    }
  }

  return false;
}

/**
 * Triggers native window drag if supported by the Wails runtime.
 */
export function triggerNativeDrag(): void {
  const w = window as any;
  if (typeof w.WailsInvoke === "function") {
    w.WailsInvoke("drag");
  } else if (w.runtime && typeof w.runtime.WindowStartDrag === "function") {
    w.runtime.WindowStartDrag();
  }
}
