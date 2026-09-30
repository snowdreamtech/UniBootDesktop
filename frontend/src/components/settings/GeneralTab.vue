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

        <!-- System Tray & Window Close Action -->
        <div class="form-group span-full">
          <label class="form-label">📌 {{ t("settings.tray_section") || "System Tray & Window Behavior" }}</label>
          <div class="tray-setting-card">
            <div class="switch-row">
              <div class="switch-label-group">
                <span class="switch-title">{{ t("settings.enable_tray") || "Enable System Tray" }}</span>
                <span class="switch-desc">{{ t("settings.enable_tray_desc") || "Keep app alive in system tray / menu bar (disabled by default)" }}</span>
              </div>
              <label class="toggle-switch">
                <input
                  type="checkbox"
                  :checked="enableTray === true"
                  @change="onEnableTrayChange(($event.target as HTMLInputElement).checked)"
                />
                <span class="slider"></span>
              </label>
            </div>

            <div v-if="enableTray" class="close-action-group">
              <label class="form-label sub-label">{{ t("settings.close_action") || "When clicking window close button:" }}</label>
              <div class="radio-group horizontal">
                <label class="radio-label">
                  <input
                    type="radio"
                    value="quit"
                    :checked="closeAction === 'quit'"
                    @change="onCloseActionChange('quit')"
                  />
                  <span>{{ t("settings.close_action_quit") || "Quit application" }}</span>
                </label>
                <label class="radio-label">
                  <input
                    type="radio"
                    value="minimize_to_tray"
                    :checked="closeAction === 'minimize_to_tray'"
                    @change="onCloseActionChange('minimize_to_tray')"
                  />
                  <span>{{ t("settings.close_action_minimize") || "Minimize to system tray" }}</span>
                </label>
              </div>
            </div>
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
  enableTray: boolean;
  closeAction: string;
  languageOptions: Array<{ value: string; label: string }>;
  themeOptions: Array<{ value: string; label: string }>;
}>();

const emit = defineEmits<{
  (e: "update:language", val: string): void;
  (e: "update:theme", val: string): void;
  (e: "update:autoCheckUpdate", val: boolean): void;
  (e: "update:enableTray", val: boolean): void;
  (e: "update:closeAction", val: string): void;
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

function onEnableTrayChange(val: boolean) {
  emit("update:enableTray", val);
  emit("change");
}

function onCloseActionChange(val: string) {
  emit("update:closeAction", val);
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

.sub-label {
  font-size: 0.78rem;
  margin-bottom: 0.25rem;
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

/* Tray setting card */
.tray-setting-card {
  background: rgba(0, 0, 0, 0.15);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  padding: 0.85rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.switch-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.switch-label-group {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.switch-title {
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--text-main);
}

.switch-desc {
  font-size: 0.72rem;
  color: var(--text-muted);
}

.close-action-group {
  padding-top: 0.65rem;
  border-top: 1px dashed var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

/* Modern Toggle Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
  flex-shrink: 0;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(148, 163, 184, 0.3);
  transition: 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 22px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--accent-cyan);
}

input:checked + .slider:before {
  transform: translateX(18px);
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

