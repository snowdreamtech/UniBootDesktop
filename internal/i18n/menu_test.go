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
		{"zh-CN", "UniGoDesktop", "编辑"},
		{"zh-TW", "UniGoDesktop", "編輯"},
		{"en-US", "UniGoDesktop", "Edit"},
		{"de-DE", "UniGoDesktop", "Bearbeiten"},
		{"fr-FR", "UniGoDesktop", "Édition"},
		{"es-ES", "UniGoDesktop", "Edición"},
		{"ja-JP", "UniGoDesktop", "編集"},
		{"ko-KR", "UniGoDesktop", "편집"},
		{"ru-RU", "UniGoDesktop", "Правка"},
		{"ar-SA", "UniGoDesktop", "تعديل"},
		{"unknown-locale", "UniGoDesktop", "Edit"},
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
