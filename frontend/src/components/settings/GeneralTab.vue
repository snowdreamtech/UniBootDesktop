<template>
  <div class="tab-content">
    <div class="settings-section">
      <h4 class="section-title">
        <span>⚙️ {{ t("settings.tab_general") || "General Preferences" }}</span>
        <span class="badge info">{{ t("settings.realtime_save") || "Auto-Saved" }}</span>
      </h4>

      <div class="grid-form">
        <!-- Language Selection -->
        <div class="form-group highlight-form-group">
          <label class="form-label highlight-label">🌐 {{ t("settings.language") || "Language" }}</label>
          <CustomSelect :model-value="language" :options="languageOptions" @change="onLanguageChange" />
        </div>

        <!-- Theme Selection -->
        <div class="form-group">
          <label class="form-label">🎨 {{ t("settings.theme") || "Appearance Theme" }}</label>
          <CustomSelect :model-value="theme" :options="themeOptions" @change="onThemeChange" />
        </div>

        <!-- Auto Check Updates -->
        <div class="form-group span-full">
          <label class="form-label">🔄 {{ t("settings.app_update") || "Software Updates" }}</label>
          <div class="radio-group horizontal">
            <label class="radio-label">
              <input
                type="radio"
                :value="true"
                :checked="autoCheckUpdate === true"
                @change="onAutoCheckUpdateChange(true)"
              />
              <span>{{ t("settings.update_auto") || "Automatically check on startup" }}</span>
            </label>
            <label class="radio-label">
              <input
                type="radio"
                :value="false"
                :checked="autoCheckUpdate === false"
                @change="onAutoCheckUpdateChange(false)"
              />
              <span>{{ t("settings.update_manual") || "Manual check only" }}</span>
            </label>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from "../../i18n";
import CustomSelect from "../CustomSelect.vue";

defineProps<{
  language: string;
  theme: string;
  autoCheckUpdate: boolean;
  languageOptions: Array<{ value: string; label: string }>;
  themeOptions: Array<{ value: string; label: string }>;
}>();

const emit = defineEmits<{
  (e: "update:language", val: string): void;
  (e: "update:theme", val: string): void;
  (e: "update:autoCheckUpdate", val: boolean): void;
  (e: "languageChange", val: string): void;
  (e: "themeChange", val: string): void;
  (e: "change"): void;
}>();

function onLanguageChange(val: string) {
  emit("update:language", val);
  emit("languageChange", val);
}

function onThemeChange(val: string) {
  emit("update:theme", val);
  emit("themeChange", val);
}

function onAutoCheckUpdateChange(val: boolean) {
  emit("update:autoCheckUpdate", val);
  emit("change");
}
</script>

<style scoped>
.tab-content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.settings-section {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1.25rem;
}

.section-title {
  font-size: 0.95rem;
  font-weight: 700;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--accent-cyan);
}

.badge {
  font-size: 0.725rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-weight: 600;
}

.badge.info {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
}

.grid-form {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.span-full {
  grid-column: span 2;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-group.highlight-form-group {
  position: relative;
  z-index: 20;
}

.form-label {
  font-size: 0.825rem;
  color: var(--text-muted);
  font-weight: 600;
}

.highlight-label {
  color: var(--accent-cyan);
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-top: 0.2rem;
}

.radio-group.horizontal {
  flex-direction: row;
  gap: 1.5rem;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--text-main);
  cursor: pointer;
}

@media (max-width: 600px) {
  .grid-form {
    grid-template-columns: 1fr;
  }
  .span-full {
    grid-column: span 1;
  }
}
</style>
