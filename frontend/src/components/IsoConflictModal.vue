<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="cancel">
    <section class="conflict-modal" role="dialog" aria-modal="true" :aria-labelledby="titleId">
      <header class="modal-header">
        <div class="header-copy">
          <span class="header-mark" aria-hidden="true">!</span>
          <div>
            <p class="eyebrow">{{ t("iso.preflight_title") }}</p>
            <h2 :id="titleId">{{ t("iso.conflict_title") }}</h2>
          </div>
        </div>
        <button class="close-btn" type="button" :aria-label="t('iso.conflict_cancel')" @click="cancel">✕</button>
      </header>

      <div class="notice-strip">
        <span class="notice-icon" aria-hidden="true">↳</span>
        <p class="modal-desc">{{ t("iso.conflict_desc") }}</p>
      </div>

      <div class="conflict-list">
        <article v-for="group in conflictGroupsWithBadges" :key="group.key" class="conflict-item">
          <div class="conflict-file">
            <div class="file-title-row">
              <span class="file-type" aria-hidden="true">{{ group.kindBadge }}</span>
              <strong>{{ group.fileName }}</strong>
            </div>
            <span class="target-summary">
              <span class="target-dot" aria-hidden="true"></span>
              {{ group.targetDisks.join(" · ") }}
            </span>
            <small class="source-path" :title="group.sourcePath">{{ group.sourcePath }}</small>
          </div>
          <div class="choice-panel" :class="`choice-${group.action}`">
            <div v-if="group.hasSourceDuplicate" class="decision-block">
              <div class="choice-heading">
                <span class="choice-label-title">{{ fileExtLabel(group.sourcePath) }}×2</span>
                <span class="choice-current">{{
                  group.sourceAction === "rename" ? t("iso.conflict_keep_both") : t("iso.conflict_skip")
                }}</span>
              </div>
              <div class="action-options compact-options" role="group">
                <button
                  type="button"
                  class="action-option"
                  :class="{ selected: group.sourceAction === 'rename' }"
                  @click="updateSourceGroup(group, 'rename')"
                >
                  {{ t("iso.conflict_keep_both") }}
                </button>
                <button
                  type="button"
                  class="action-option"
                  :class="{ selected: group.sourceAction === 'skip' }"
                  @click="updateSourceGroup(group, 'skip')"
                >
                  {{ t("iso.conflict_skip") }}
                </button>
              </div>
            </div>
            <div v-if="group.hasTargetExists" class="decision-block">
              <div class="choice-heading">
                <span class="choice-label-title">USB</span>
                <span class="choice-current">{{ t(actionLabelKey(group.action)) }}</span>
              </div>
            </div>
            <div
              v-if="group.hasTargetExists"
              class="action-options"
              role="group"
              :aria-label="t('iso.conflict_action')"
            >
              <button
                v-for="option in targetActionOptions"
                :key="option.value"
                type="button"
                class="action-option"
                :class="{ selected: group.action === option.value }"
                :aria-pressed="group.action === option.value"
                @click="updateGroup(group, option.value)"
              >
                {{ t(option.label) }}
              </button>
            </div>
            <div
              v-else-if="!group.hasSourceDuplicate"
              class="action-options"
              role="group"
              :aria-label="t('iso.conflict_action')"
            >
              <button
                v-for="option in targetActionOptions"
                :key="option.value"
                type="button"
                class="action-option"
                :class="{ selected: group.action === option.value }"
                @click="updateGroup(group, option.value)"
              >
                {{ t(option.label) }}
              </button>
            </div>
          </div>
        </article>
      </div>

      <label class="apply-all">
        <input v-model="applyToAll" type="checkbox" @change="applyAllDecision" />
        <span
          ><strong>{{ t("iso.conflict_apply_all") }}</strong></span
        >
      </label>

      <footer class="modal-footer">
        <button class="btn-secondary" type="button" @click="cancel">{{ t("iso.conflict_cancel") }}</button>
        <button class="btn-primary" type="button" @click="confirm">{{ t("iso.conflict_confirm") }}</button>
      </footer>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { t } from "../i18n";

export interface IsoConflictItem {
  targetDisk: string;
  sourcePath: string;
  fileName: string;
  suggestedName: string;
  conflictType?: "source_duplicate" | "target_exists" | "source_duplicate_target_exists";
}

export interface IsoConflictDecision {
  targetDisk: string;
  sourcePath: string;
  targetName: string;
  action: "replace" | "skip" | "rename";
}

interface IsoConflictGroup {
  key: string;
  fileName: string;
  sourcePath: string;
  targetDisks: string[];
  conflictIndexes: number[];
  action: IsoConflictDecision["action"];
  conflictTypes: Set<NonNullable<IsoConflictItem["conflictType"]>>;
  hasSourceDuplicate: boolean;
  hasTargetExists: boolean;
  sourceAction: "rename" | "skip";
  duplicateOrder: number;
}

interface IsoConflictDisplayGroup extends IsoConflictGroup {
  kindBadge: string;
}

const props = defineProps<{
  isOpen: boolean;
  conflicts: IsoConflictItem[];
}>();

const emit = defineEmits<{
  cancel: [];
  confirm: [decisions: IsoConflictDecision[]];
}>();

const applyToAll = ref(false);
const decisions = ref<IsoConflictDecision[]>([]);
const titleId = computed(() => "iso-conflict-title");
const targetActionOptions = [
  { value: "rename", label: "iso.conflict_keep_both" },
  { value: "replace", label: "iso.conflict_replace" },
  { value: "skip", label: "iso.conflict_skip" },
] as const;
const conflictGroups = computed<IsoConflictGroup[]>(() => {
  const groups = new Map<string, IsoConflictGroup>();
  const counters = new Map<string, number>();
  props.conflicts.forEach((conflict, index) => {
    const key = `${conflict.sourcePath}\n${conflict.fileName}`;
    const group = groups.get(key);
    if (group) {
      group.targetDisks.push(conflict.targetDisk);
      group.conflictIndexes.push(index);
      group.conflictTypes.add(conflict.conflictType || "target_exists");
      group.hasSourceDuplicate = true;
      group.hasTargetExists = group.hasTargetExists || conflict.conflictType !== "source_duplicate";
      return;
    }
    const duplicateOrder = (counters.get(conflict.fileName) || 0) + 1;
    counters.set(conflict.fileName, duplicateOrder);
    const hasSourceDuplicate =
      conflict.conflictType === "source_duplicate" || conflict.conflictType === "source_duplicate_target_exists";
    const hasTargetExists =
      conflict.conflictType === "target_exists" || conflict.conflictType === "source_duplicate_target_exists";
    groups.set(key, {
      key,
      fileName: conflict.fileName,
      sourcePath: conflict.sourcePath,
      targetDisks: [conflict.targetDisk],
      conflictIndexes: [index],
      action: decisions.value[index]?.action || "rename",
      conflictTypes: new Set([conflict.conflictType || "target_exists"]),
      hasSourceDuplicate,
      hasTargetExists,
      sourceAction: "rename",
      duplicateOrder,
    });
    return;
  });
  return [...groups.values()];
});

/** Extracts the uppercased extension label (e.g. "ISO", "IMG", "WIM") from a source path. */
function fileExtLabel(sourcePath: string): string {
  const dotIndex = sourcePath.lastIndexOf(".");
  if (dotIndex === -1) return "IMG";
  return sourcePath.slice(dotIndex + 1).toUpperCase();
}

const conflictGroupsWithBadges = computed<IsoConflictDisplayGroup[]>(() =>
  conflictGroups.value.map((group) => {
    const ext = fileExtLabel(group.sourcePath);
    const kindBadge = group.conflictTypes.has("source_duplicate_target_exists")
      ? `${ext}×2 · USB`
      : group.conflictTypes.has("source_duplicate")
        ? `${ext}×2`
        : "USB";
    return { ...group, kindBadge };
  })
);

function defaultDecision(conflict: IsoConflictItem): IsoConflictDecision {
  return {
    targetDisk: conflict.targetDisk,
    sourcePath: conflict.sourcePath,
    targetName: conflict.suggestedName,
    action: "rename",
  };
}

watch(
  () => [props.isOpen, props.conflicts],
  () => {
    if (props.isOpen) {
      applyToAll.value = false;
      decisions.value = props.conflicts.map(defaultDecision);
    }
  },
  { immediate: true }
);

function updateGroup(group: IsoConflictGroup, action: IsoConflictDecision["action"]) {
  group.action = action;
  for (const index of group.conflictIndexes) {
    const conflict = props.conflicts[index];
    const decision = decisions.value[index];
    if (!conflict || !decision) continue;
    applyDecision(group, conflict, decision, action);
  }
  if (applyToAll.value) applyAllDecision();
}

function updateSourceGroup(group: IsoConflictGroup, action: "rename" | "skip") {
  group.sourceAction = action;
  for (const index of group.conflictIndexes) {
    const conflict = props.conflicts[index];
    const decision = decisions.value[index];
    if (!conflict || !decision) continue;
    applyDecision(group, conflict, decision, group.action);
  }
}

function applyDecision(
  group: IsoConflictGroup,
  conflict: IsoConflictItem,
  decision: IsoConflictDecision,
  targetAction: IsoConflictDecision["action"]
) {
  if (group.hasSourceDuplicate && group.sourceAction === "skip") {
    decision.action = "skip";
    decision.targetName = conflict.fileName;
    return;
  }

  const mustKeepSeparate = group.hasSourceDuplicate && group.duplicateOrder > 1 && targetAction === "replace";
  decision.action = mustKeepSeparate ? "rename" : targetAction;
  decision.targetName = targetAction === "rename" || mustKeepSeparate ? conflict.suggestedName : conflict.fileName;
}

function actionLabelKey(action: IsoConflictDecision["action"]) {
  if (action === "replace") return "iso.conflict_replace" as const;
  if (action === "skip") return "iso.conflict_skip" as const;
  return "iso.conflict_keep_both" as const;
}

function applyAllDecision() {
  if (!applyToAll.value || decisions.value.length === 0) return;
  const source = decisions.value[0];
  if (!source) return;
  decisions.value = decisions.value.map((decision, index) => {
    const conflict = props.conflicts[index];
    const group = conflictGroups.value.find((candidate) => candidate.conflictIndexes.includes(index));
    if (!group) return decision;
    const nextDecision = { ...decision };
    applyDecision(group, conflict, nextDecision, group.action);
    return {
      ...nextDecision,
      action: source.action === "skip" ? "skip" : nextDecision.action,
    };
  });
}

function cancel() {
  emit("cancel");
}

function confirm() {
  emit("confirm", decisions.value);
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background: rgba(12, 18, 32, 0.72);
  backdrop-filter: blur(12px);
}

.conflict-modal {
  width: min(780px, 100%);
  max-height: min(800px, 92vh);
  overflow: auto;
  border: 1px solid rgba(148, 163, 184, 0.36);
  border-radius: 16px;
  background: var(--modal-bg, var(--card-bg));
  color: var(--text-main);
  box-shadow: 0 28px 90px rgba(0, 0, 0, 0.36);
  overflow-wrap: anywhere;
}

.modal-header,
.modal-footer,
.conflict-item,
.apply-all,
.header-copy,
.file-title-row,
.target-summary,
.choice-heading {
  display: flex;
  align-items: center;
}

.decision-block + .decision-block {
  margin-top: 0.55rem;
  padding-top: 0.55rem;
  border-top: 1px solid var(--card-border);
}

.compact-options {
  grid-template-columns: repeat(2, 1fr);
}

.decision-block + .action-options {
  margin-top: 0.4rem;
}

.modal-header {
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--card-border);
  background: var(--modal-header-bg, var(--section-bg));
}

.header-copy {
  gap: 0.8rem;
}

.header-mark {
  display: grid;
  flex: 0 0 2.25rem;
  width: 2.25rem;
  height: 2.25rem;
  place-items: center;
  border: 1px solid rgba(245, 158, 11, 0.55);
  border-radius: 9px;
  background: rgba(245, 158, 11, 0.14);
  color: #f59e0b;
  font-size: 1.25rem;
  font-weight: 800;
}

.eyebrow {
  margin: 0 0 0.25rem;
  color: var(--text-muted);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

h2 {
  margin: 0;
  color: var(--text-main);
  font-size: 1.45rem;
  font-weight: 800;
}

.notice-strip {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  margin: 1.25rem 1.5rem 1rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--alert-warning-border, rgba(245, 158, 11, 0.35));
  border-radius: 10px;
  background: var(--alert-warning-bg, rgba(245, 158, 11, 0.1));
}

.notice-icon {
  display: grid;
  flex: 0 0 1.25rem;
  width: 1.25rem;
  height: 1.25rem;
  place-items: center;
  border-radius: 50%;
  background: #f59e0b;
  color: #1f2937;
  font-size: 0.8rem;
  font-weight: 800;
}

.modal-desc {
  margin: 0;
  color: var(--text-muted);
  font-size: 0.88rem;
  line-height: 1.55;
}

.close-btn {
  border: 0;
  background: transparent;
  color: var(--text-muted);
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.35rem 0.5rem;
  border-radius: 6px;
}

.close-btn:hover {
  background: var(--section-bg);
  color: var(--text-main);
}

.conflict-list {
  display: grid;
  gap: 0.85rem;
  padding: 0 1.5rem;
}

.conflict-item {
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid var(--card-border);
  border-radius: 11px;
  background: var(--section-bg);
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.conflict-item:hover {
  border-color: rgba(59, 130, 246, 0.55);
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.1);
}

.conflict-file {
  display: grid;
  flex: 1 1 auto;
  min-width: 0;
  gap: 0.45rem;
  padding-top: 0.1rem;
}

.file-title-row {
  min-width: 0;
  gap: 0.55rem;
}

.file-title-row strong {
  overflow: hidden;
  color: var(--text-main);
  font-size: 0.95rem;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-type {
  flex: 0 0 auto;
  padding: 0.18rem 0.35rem;
  border: 1px solid rgba(59, 130, 246, 0.4);
  border-radius: 4px;
  color: var(--accent-cyan, #38bdf8);
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.file-title-row:has(.file-type) .file-type {
  background: rgba(59, 130, 246, 0.08);
}

.target-summary {
  gap: 0.4rem;
  color: var(--accent-cyan, #38bdf8);
  font-size: 0.78rem;
  font-weight: 650;
}

.target-dot {
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 50%;
  background: var(--accent-cyan, #38bdf8);
}

.source-path {
  display: block;
  overflow: hidden;
  overflow-wrap: anywhere;
  color: var(--text-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.7rem;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.choice-panel {
  flex: 0 0 260px;
  padding: 0.7rem;
  border: 1px solid var(--card-border);
  border-radius: 9px;
  background: var(--card-bg);
}

.choice-panel.choice-rename {
  border-color: rgba(16, 185, 129, 0.42);
}

.choice-panel.choice-replace {
  border-color: rgba(245, 158, 11, 0.58);
}

.choice-panel.choice-skip {
  border-color: rgba(148, 163, 184, 0.5);
}

.choice-heading {
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.55rem;
}

.choice-label-title {
  color: var(--accent-color, #60a5fa);
  font-size: 0.8rem;
  font-weight: 800;
}

.choice-current {
  color: var(--text-muted);
  font-size: 0.72rem;
  font-weight: 700;
}

.action-options {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.35rem;
}

.action-option {
  min-height: 2.2rem;
  padding: 0.35rem 0.3rem;
  border: 1px solid var(--card-border);
  border-radius: 6px;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.78rem;
  font-weight: 700;
  white-space: normal;
}

.action-option:hover,
.action-option:focus-visible {
  border-color: var(--accent-color, #60a5fa);
  color: var(--text-main);
}

.action-option.selected {
  color: #fff;
}

.choice-rename .action-option.selected {
  border-color: #10b981;
  background: #10b981;
}

.choice-replace .action-option.selected {
  border-color: #f59e0b;
  background: #f59e0b;
  color: #1f2937;
}

.choice-skip .action-option.selected {
  border-color: #64748b;
  background: #64748b;
}

.apply-all {
  gap: 0.65rem;
  margin: 1rem 1.5rem 0;
  padding: 0.8rem 0.9rem;
  border: 1px solid var(--card-border);
  border-radius: 9px;
  background: var(--section-bg);
  color: var(--text-muted);
  font-size: 0.82rem;
}

.apply-all strong {
  color: var(--text-main);
  font-weight: 700;
}

.apply-all input {
  flex: 0 0 auto;
  accent-color: var(--accent-color, #3b82f6);
}

.modal-footer {
  justify-content: flex-end;
  gap: 0.65rem;
  margin-top: 1.25rem;
  padding: 1rem 1.5rem 1.25rem;
  border-top: 1px solid var(--card-border);
  flex-wrap: wrap;
}

.btn-secondary,
.btn-primary {
  flex: 0 1 12rem;
  min-height: 2.65rem;
  padding: 0.55rem 1rem;
  border-radius: 7px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 750;
  white-space: normal;
}

.btn-secondary {
  border: 1px solid var(--card-border);
  background: transparent;
  color: var(--text-muted);
}

.btn-secondary:hover {
  border-color: var(--text-muted);
  color: var(--text-main);
}

.btn-primary {
  border: 1px solid var(--accent-color, #3b82f6);
  background: var(--accent-color, #3b82f6);
  color: #fff;
  box-shadow: 0 5px 14px rgba(37, 99, 235, 0.24);
}

.btn-primary:hover {
  filter: brightness(1.08);
}

/* Keep focus visible for keyboard users without adding visual noise to the card. */
.action-option:focus-visible,
.btn-primary:focus-visible,
.btn-secondary:focus-visible,
.close-btn:focus-visible {
  outline: 2px solid var(--accent-color, #60a5fa);
  outline-offset: 2px;
}

/* Narrow layouts stack evidence above the decision controls. */
@media (max-width: 640px) {
  .modal-overlay {
    padding: 0.75rem;
  }

  .modal-header {
    padding: 1rem;
  }

  .notice-strip,
  .conflict-list {
    margin-inline: 1rem;
  }

  .conflict-list {
    padding: 0;
  }

  .conflict-item {
    display: grid;
  }

  .choice-panel {
    width: 100%;
  }

  .apply-all {
    margin-inline: 1rem;
  }

  .modal-footer {
    align-items: stretch;
    flex-direction: column-reverse;
    padding-inline: 1rem;
  }

  .btn-secondary,
  .btn-primary {
    flex-basis: auto;
    width: 100%;
  }
}
</style>
