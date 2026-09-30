// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export function logUserAction(level: string, message: string, details: string = ""): void {
  if (window.go && window.go.main && window.go.main.App && (window.go.main.App as any).LogAction) {
    (window.go.main.App as any).LogAction(level, message, details);
  }
}
