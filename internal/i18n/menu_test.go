// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package i18n

import (
	"testing"
)

func TestGetMenuTranslations(t *testing.T) {
	tests := []struct {
		langCode string
		wantApp  string
		wantEdit string
	}{
		{"zh-CN", "UniBootDesktop", "编辑"},
		{"zh-TW", "UniBootDesktop", "編輯"},
		{"en-US", "UniBootDesktop", "Edit"},
		{"de-DE", "UniBootDesktop", "Bearbeiten"},
		{"fr-FR", "UniBootDesktop", "Édition"},
		{"es-ES", "UniBootDesktop", "Edición"},
		{"ja-JP", "UniBootDesktop", "編集"},
		{"ko-KR", "UniBootDesktop", "편집"},
		{"ru-RU", "UniBootDesktop", "Правка"},
		{"ar-SA", "UniBootDesktop", "تعديل"},
		{"unknown-locale", "UniBootDesktop", "Edit"},
	}

	for _, tt := range tests {
		got := GetMenuTranslations(tt.langCode)
		if got.App != tt.wantApp {
			t.Errorf("GetMenuTranslations(%q).App = %q; want %q", tt.langCode, got.App, tt.wantApp)
		}
		if got.Edit != tt.wantEdit {
			t.Errorf("GetMenuTranslations(%q).Edit = %q; want %q", tt.langCode, got.Edit, tt.wantEdit)
		}
	}
}
