<template>
  <section class="glass-card section-card">
    <h2>{{ t("deploy.title") }}</h2>
    <p class="section-desc" v-if="activeMode === 'cloud'">
      {{ t("deploy.desc_cloud") }}
    </p>
    <p class="section-desc" v-else>
      {{ t("deploy.desc_hybrid") }}
    </p>

    <!-- Filesystem Selection for Hybrid Mode & Cloud Mode (Hidden when upgrading an existing Ventoy/UniBoot drive) -->
    <div v-if="!isNonDestructive" class="fs-selector">
      <label class="fs-label">{{ t("settings.default_fs") }}</label>
      <CustomSelect
        :modelValue="selectedFsType"
        :disabled="!canChangeFs"
        @update:modelValue="(val) => emit('update:selectedFsType', val)"
        :options="[
          { value: 'exFAT', label: t('fs.exfat') },
          { value: 'NTFS', label: t('fs.ntfs') },
          { value: 'FAT32', label: t('fs.fat32') },
          { value: 'ext4', label: t('fs.ext4') },
        ]"
      />
    </div>

    <!-- Ventoy CLI Pre-flight Requirement Notice Banner (Hybrid Mode) -->
    <div v-if="activeMode === 'hybrid' && !isNonDestructive && !ventoyStatus.valid" class="ventoy-warning-card">
      <span class="warning-card-icon">⚠️</span>
      <div class="warning-card-body">
        <div class="warning-card-title">
          {{ isMacOs ? t("deploy.macos_alert_title") : t("deploy.no_ventoy_title") }}
        </div>
        <div class="warning-card-message">
          {{ isMacOs ? t("deploy.macos_alert_desc") : t("deploy.no_ventoy_desc") }}
        </div>
      </div>
      <button class="btn-secondary btn-sm" @click="emit('open-settings-ventoy')">
        <span class="btn-icon">⚙️</span>
        <span>{{ t("settings.title") }}</span>
      </button>
    </div>

    <!-- Safe Mode Notice Banner when upgrading an existing Ventoy/UniBoot drive -->
    <div v-if="isNonDestructive" class="safe-mode-notice">
      <span class="safe-notice-icon">🛡️</span>
      <div class="safe-notice-content">
        <div class="safe-notice-title">
          {{ activeMode === "cloud" ? t("safe.title_cloud") : t("safe.title_hybrid") }}
        </div>
        <div class="safe-notice-desc">
          {{ activeMode === "cloud" ? t("safe.desc_cloud") : t("safe.desc_hybrid") }}
        </div>
      </div>
    </div>

    <!-- Local ISO/IMG Image Source Selection Card (Hybrid Mode) -->
    <div
      v-if="activeMode === 'hybrid'"
      class="iso-card"
      :class="{ 'is-drag-over': isDragOver }"
      @dragenter="handleDragEnter"
      @dragover="handleDragOver"
      @dragleave="handleDragLeave"
      @drop="handleDrop"
    >
      <div class="iso-card-header">
        <div class="iso-title-group">
          <h3>
            <span class="section-icon">💿</span>
            <span>{{ t("iso.title") }}</span>
            <span class="optional-badge">{{ t("common.optional") }}</span>
          </h3>
          <span class="iso-subtitle">{{ t("iso.desc") }}</span>
        </div>
        <button class="btn-secondary add-iso-btn" :disabled="!canManageIso" @click="canManageIso && emit('select-iso')">
          <span class="btn-icon">➕</span>
          <span>{{ t("iso.add_btn") }}</span>
        </button>
      </div>

      <div class="iso-list-container">
        <!-- Overlay when dragging files over non-empty list -->
        <div v-if="isDragOver && selectedIsoFiles.length > 0" class="iso-drag-overlay">
          <span class="drag-icon">📥</span>
          <div class="drag-text">{{ t("iso.drag_drop_tip") }}</div>
        </div>

        <div
          v-if="selectedIsoFiles.length === 0"
          class="iso-empty-state"
          :class="{ 'drag-active': isDragOver, disabled: !canManageIso }"
          @click="canManageIso && emit('select-iso')"
        >
          <span class="empty-icon">📥</span>
          <div class="empty-text">{{ isDragOver ? t("iso.drag_drop_tip") : t("iso.empty_title") }}</div>
          <div class="empty-subtext">{{ t("iso.empty_sub") }}</div>
        </div>

        <div v-else class="iso-file-list">
          <div
            v-for="(file, index) in selectedIsoFiles"
            :key="index"
            class="iso-file-item"
            :class="{ 'is-checksum-target': selectedChecksumIsoIndex === index }"
          >
            <span class="iso-file-icon">{{ getFileIcon(file.name) }}</span>
            <div class="iso-file-info" @click="selectIsoForChecksum(index)">
              <div class="iso-file-name" :title="file.path">{{ file.name }}</div>
              <div class="iso-file-path">{{ file.path }}</div>
            </div>
            <div class="iso-item-actions">
              <!-- Checksum Status Badge (clickable to view single details) -->
              <span
                v-if="getIsoChecksumStatus(file.name)"
                class="iso-status-badge"
                :class="getIsoChecksumStatus(file.name)?.status"
                :title="t('checksum.inspect_single')"
                @click.stop="inspectSingleIso(index)"
              >
                <span class="badge-icon">{{ getIsoChecksumStatusIcon(file.name) }}</span>
                <span>{{ getIsoChecksumStatusText(file.name) }}</span>
              </span>
              <button
                class="iso-check-hash-btn"
                :class="{ active: checksumActiveTab === 'single' && selectedChecksumIsoIndex === index }"
                :title="t('checksum.inspect_single')"
                @click.stop="inspectSingleIso(index)"
              >
                <span class="btn-icon">🔒</span>
                <span>{{
                  selectedChecksumIsoIndex === index && checksumActiveTab === "single"
                    ? t("checksum.status_selected")
                    : t("checksum.status_verify")
                }}</span>
              </button>
              <button
                class="iso-remove-btn"
                :disabled="!canManageIso"
                :title="t('iso.remove')"
                @click.stop="canManageIso && emit('remove-iso', index)"
              >
                ✕
              </button>
            </div>
          </div>
        </div>

        <div v-if="selectedIsoFiles.length > 0" class="iso-footer">
          <!-- Left safe action group: summary count + toggle checksum button -->
          <div class="iso-footer-left">
            <span class="iso-count-summary">{{ t("iso.summary", { count: selectedIsoFiles.length }) }}</span>
            <button
              class="btn-toggle-checksum"
              :class="{ active: isChecksumPanelExpanded }"
              :title="isChecksumPanelExpanded ? t('checksum.collapse') : t('checksum.expand')"
              @click="isChecksumPanelExpanded = !isChecksumPanelExpanded"
            >
              <span class="btn-icon">🔍</span>
              <span>{{ isChecksumPanelExpanded ? t("checksum.collapse") : t("checksum.expand") }}</span>
              <span class="caret-icon">{{ isChecksumPanelExpanded ? "▴" : "▾" }}</span>
              <span v-if="hasAnyChecksumResult" class="checksum-mini-badge" :class="{ 'all-match': isAllBatchMatched }">
                {{ isAllBatchMatched ? "✅" : "⚠️" }}
              </span>
            </button>
          </div>

          <!-- Right isolated danger action: clear list button -->
          <button
            class="btn-clear-iso"
            :disabled="!canManageIso"
            :title="t('iso.clear')"
            @click="canManageIso && emit('clear-iso')"
          >
            <span class="btn-icon">🗑️</span>
            <span>{{ t("iso.clear") }}</span>
          </button>
        </div>

        <!-- Checksum Verification Card (Expandable Drawer) -->
        <div v-if="selectedIsoFiles.length > 0 && isChecksumPanelExpanded" class="checksum-container">
          <!-- Permanent file input for checksum sums files -->
          <input
            type="file"
            ref="sumsFileInputRef"
            style="display: none"
            accept=".txt,.sums,.checksum,.sha1,.sha1sum,.sha224,.sha256,.sha256sum,.sha256sums,.sha384,.sha512,.sha512sum,.sha512sums,.md5,.md5sum,.md5sums,*"
            multiple
            @change="handleSumsFileSelected"
          />

          <!-- Checksum Sub-Tabs Navigation (Visible when multiple ISOs selected) -->
          <div v-if="selectedIsoFiles.length > 1" class="checksum-subtabs-nav">
            <button
              class="sub-tab-btn"
              :class="{ active: checksumActiveTab === 'batch' }"
              @click="checksumActiveTab = 'batch'"
            >
              <span class="btn-icon">🔍</span>
              <span>{{ t("checksum.tab_batch") }}</span>
              <span class="tab-badge">{{ selectedIsoFiles.length }}</span>
            </button>
            <button
              class="sub-tab-btn"
              :class="{ active: checksumActiveTab === 'single' }"
              @click="checksumActiveTab = 'single'"
            >
              <span class="btn-icon">⚡</span>
              <span>{{ t("checksum.tab_single") }}</span>
            </button>
          </div>

          <!-- ================= TAB 1: 批量校验专属视图 ================= -->
          <div
            v-if="checksumActiveTab === 'batch' && selectedIsoFiles.length > 1"
            class="checksum-tab-panel batch-panel"
          >
            <div class="checksum-header">
              <div class="checksum-header-left">
                <span class="checksum-title">{{ t("checksum.tab_batch") }}</span>
                <span class="checksum-panel-desc">{{ t("checksum.batch_desc") }}</span>
              </div>

              <div class="checksum-header-actions">
                <div class="algo-selector">
                  <select v-model="selectedAlgo" class="algo-select">
                    <option value="sha256">SHA-256 {{ t("checksum.recommended") }}</option>
                    <option value="md5">MD5</option>
                    <option value="sha1">SHA-1</option>
                    <option value="sha384">SHA-384</option>
                    <option value="sha512">SHA-512</option>
                    <option value="crc32">CRC32</option>
                  </select>
                </div>

                <!-- 批量导入校验汇总文件按钮 -->
                <button
                  class="btn-secondary import-sums-header-btn"
                  :disabled="!canVerifyHash || isCalculatingHash || isBatchCalculating"
                  :title="t('checksum.import_file_title')"
                  @click="triggerSumsFilePick"
                >
                  <span class="btn-icon">📄</span>
                  <span>{{ t("checksum.import_file") }}</span>
                  <span
                    v-if="cachedHashCount > 0"
                    class="cached-count-pill"
                    :title="t('checksum.cache_loaded', { count: cachedHashCount })"
                  >
                    {{ cachedHashCount }}
                  </span>
                </button>

                <!-- 一键批量校验全部按钮 -->
                <button
                  class="btn-secondary batch-calc-btn"
                  :disabled="!canVerifyHash || isBatchCalculating || isCalculatingHash"
                  @click="handleBatchChecksum"
                >
                  <span class="btn-icon">{{ isBatchCalculating ? "⏳" : "🔍" }}</span>
                  <span>{{
                    isBatchCalculating
                      ? t("checksum.batch_verifying", {
                          current: batchProgress.current,
                          total: selectedIsoFiles.length,
                        })
                      : t("checksum.batch_verify_all")
                  }}</span>
                </button>
              </div>
            </div>

            <!-- Batch Checksum Summary Banner -->
            <div v-if="batchSummaryText" class="batch-summary-banner" :class="{ 'all-matched': isAllBatchMatched }">
              <span class="summary-icon">{{ isAllBatchMatched ? "✅" : "ℹ️" }}</span>
              <span>{{ batchSummaryText }}</span>
            </div>
          </div>

          <!-- ================= TAB 2: 单个校验专属视图 ================= -->
          <div v-else class="checksum-tab-panel single-panel">
            <!-- Target ISO Selector Bar -->
            <div class="checksum-target-bar">
              <span class="target-bar-label">🎯 {{ t("checksum.target_iso_label") }}:</span>
              <div v-if="selectedIsoFiles.length > 1" class="target-iso-selector">
                <select v-model="selectedChecksumIsoIndex" class="target-iso-select">
                  <option v-for="(file, idx) in selectedIsoFiles" :key="idx" :value="idx">
                    {{ idx + 1 }}. {{ file.name }}
                  </option>
                </select>
              </div>
              <div v-else class="target-iso-single-name" :title="selectedIsoFiles[0].path">
                {{ selectedIsoFiles[0].name }}
              </div>
            </div>

            <div class="checksum-header">
              <div class="checksum-header-left">
                <span class="checksum-title">{{ t("checksum.tab_single") }}</span>
                <span class="checksum-panel-desc">{{ t("checksum.single_desc") }}</span>
              </div>

              <div class="checksum-header-actions">
                <div class="algo-selector">
                  <select v-model="selectedAlgo" class="algo-select">
                    <option value="sha256">SHA-256 {{ t("checksum.recommended") }}</option>
                    <option value="md5">MD5</option>
                    <option value="sha1">SHA-1</option>
                    <option value="sha384">SHA-384</option>
                    <option value="sha512">SHA-512</option>
                    <option value="crc32">CRC32</option>
                  </select>
                </div>

                <!-- Single ISO calculate button -->
                <button
                  class="btn-secondary calc-hash-btn"
                  :disabled="!canVerifyHash || isCalculatingHash || isBatchCalculating"
                  @click="handleCalculateChecksum"
                >
                  <span class="btn-icon">{{ isCalculatingHash ? "⏳" : "⚡" }}</span>
                  <span>{{ isCalculatingHash ? t("checksum.calculating") : t("checksum.calc_btn") }}</span>
                </button>
              </div>
            </div>

            <!-- Single ISO Result & Compare Box -->
            <div class="checksum-result-box">
              <div v-if="calculatedHash" class="hash-code-row">
                <span class="hash-algo-badge">{{ currentChecksumAlgo.toUpperCase() }}</span>
                <code class="hash-code" :title="calculatedHash">{{ calculatedHash }}</code>
                <button
                  class="copy-hash-btn"
                  :class="{ copied: isHashCopied }"
                  :title="t('checksum.copy_hash')"
                  @click="copyHashToClipboard"
                >
                  {{ isHashCopied ? "✓" : "📋" }}
                </button>
              </div>
              <div v-else class="hash-empty-hint">
                <span class="hint-icon">⚡</span>
                <span>{{ t("checksum.single_desc") }}</span>
              </div>

              <div class="hash-compare-row">
                <input
                  v-model="expectedHashInput"
                  type="text"
                  class="hash-compare-input"
                  :placeholder="t('checksum.compare_placeholder')"
                />
                <button class="import-sums-btn" :title="t('checksum.import_file_title')" @click="triggerSumsFilePick">
                  <span class="btn-icon">📄</span>
                  <span>{{ t("checksum.import_file") }}</span>
                </button>
                <div
                  v-if="parsedExpectedHash && calculatedHash"
                  class="match-badge"
                  :class="isHashMatching ? 'match' : 'mismatch'"
                >
                  <span class="badge-icon">{{ isHashMatching ? "✅" : "❌" }}</span>
                  <span>{{ isHashMatching ? t("checksum.match_success") : t("checksum.match_mismatch") }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="deploy-box">
      <div class="selected-target">
        <span>{{ t("deploy.target_device") }}</span>
        <strong v-if="selectionMode === 'single'">
          {{ selectedDisk ? formatSelectedDiskTarget(selectedDisk) : t("disk.no_disk") }}
        </strong>
        <strong v-else>
          {{ selectedDevices.size > 0 ? t("deploy.batch_target", { count: selectedDevices.size }) : t("disk.no_disk") }}
        </strong>
      </div>

      <div v-if="isDeploying" class="deploy-active-container">
        <!-- Unified deployment stage & disk card -->
        <div class="batch-deploy-info">
          <div class="batch-current-disk" v-if="batchDeployInfo && batchDeployInfo.totalDisks > 1">
            {{ t("deploy.batch_current") }}: {{ batchDeployInfo.currentDiskIndex }}/{{ batchDeployInfo.totalDisks }}
          </div>
          <div class="batch-disk-name" v-if="batchDeployInfo?.currentDisk">
            {{ batchDeployInfo.currentDisk }}
          </div>
          <div class="batch-current-stage">
            {{ batchDeployInfo?.currentStage || t("deploy.stage_preparing") }}
          </div>
        </div>

        <ProgressBar
          :label="
            batchDeployInfo && batchDeployInfo.totalDisks > 1 ? t('deploy.batch_overall_progress') : t('deploy.writing')
          "
          :progress="deployProgress"
        />
        <div class="deploy-stats-row">
          <span class="stat-badge" v-if="(speedMBps || 0) > 0"
            >⚡ {{ t("deploy.stats_speed") }}: {{ speedMBps?.toFixed(1) }} MB/s</span
          >
          <span class="stat-badge" v-if="(elapsedSec || 0) > 0"
            >⏱️ {{ t("deploy.stats_elapsed") }}: {{ formatStatsTime(elapsedSec) }}</span
          >
          <span class="stat-badge" v-if="(etaSec || 0) > 0"
            >⌛ {{ t("deploy.stats_eta") }}: {{ formatStatsTime(etaSec) }}</span
          >
        </div>
        <button class="btn-cancel-deploy" @click="emit('cancel-deploy')">
          <span class="cancel-icon">🛑</span>
          <span>{{ t("deploy.btn_cancel") }}</span>
        </button>
      </div>

      <button
        v-else
        class="btn-primary deploy-btn"
        :class="{
          'safe-btn': isNonDestructive,
          'danger-disabled': activeMode === 'hybrid' && !isNonDestructive && !ventoyStatus.valid,
          'preflight-btn': isPreflight,
        }"
        :disabled="isDeployDisabled"
        :title="deployDisabledReason"
        @click="emit('deploy-click')"
      >
        <template v-if="isPreflight">
          <span class="deploy-icon preflight-spinner">⏳</span>
          <span>{{ t("deploy.checking") }}</span>
        </template>
        <template v-else>
          <span class="deploy-icon">{{ isNonDestructive ? "🛡️" : "🚀" }}</span>
          <span>{{ deployBtnText }}</span>
        </template>
      </button>

      <!-- Deploy success banner with Safely Eject button -->
      <div v-if="showDeploySuccessBanner" class="deploy-success-banner">
        <div class="deploy-success-icon">🎉</div>
        <div class="deploy-success-content">
          <div class="deploy-success-title">{{ t("deploy.success_banner_title") }}</div>
          <div class="deploy-success-desc">
            {{
              deploySuccessBanner.autoEjected
                ? t("deploy.toast_auto_ejected", { count: deploySuccessBanner.targets.length })
                : t("deploy.success_banner_desc")
            }}
          </div>
        </div>
        <div class="deploy-success-actions">
          <button class="btn-dismiss" @click="emit('dismiss-success-banner')" :title="t('common.close')">✕</button>
          <button
            v-if="!deploySuccessBanner.autoEjected"
            id="btn-safely-eject-after-deploy"
            class="btn-eject-success"
            @click="emit('safely-eject-success')"
          >
            <span class="btn-icon">⏏️</span>
            <span>{{ t("deploy.safely_eject_btn") }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Hypervisor Simulation Test Card -->
    <div class="vm-box">
      <div class="vm-header">
        <div class="vm-title-group">
          <h3>
            {{ t("vm.box_title") }}
            <span class="optional-badge">{{ t("common.optional") }}</span>
          </h3>
          <span class="badge success" v-if="hypervisorList.length > 0">
            {{ t("vm.installed") }}
          </span>
        </div>
      </div>

      <!-- VM Options Bar -->
      <div class="vm-options-bar">
        <!-- Boot Mode Selector -->
        <div class="vm-selector-container">
          <span class="vm-selector-label">{{ t("vm.boot_mode_label") }}</span>
          <div class="vm-select-wrapper">
            <select
              :value="selectedBootMode"
              :disabled="!canConfigureVm"
              @change="(e) => emit('update:selectedBootMode', (e.target as HTMLSelectElement).value)"
              class="vm-select boot-select"
            >
              <option value="uefi">{{ t("vm.boot_mode_uefi") }}</option>
              <option value="bios">{{ t("vm.boot_mode_bios") }}</option>
              <option value="auto">{{ t("vm.boot_mode_auto") }}</option>
            </select>
            <span class="select-arrow">▾</span>
          </div>
        </div>

        <!-- Single Hypervisor Badge -->
        <div v-if="hypervisorList.length <= 1" class="vm-selector-container">
          <span class="vm-selector-label">{{ t("vm.select_vm_label") }}</span>
          <span class="badge" :class="hypervisorList.length === 1 ? 'success' : 'muted'">
            {{ hypervisorList.length === 1 ? hypervisorList[0].name : t("vm.not_installed") }}
          </span>
        </div>

        <!-- Multiple Hypervisors Selector -->
        <div v-else class="vm-selector-container">
          <span class="vm-selector-label">{{ t("vm.select_vm_label") }}</span>
          <div class="vm-select-wrapper">
            <select
              :value="selectedVMType"
              :disabled="!canConfigureVm"
              @change="(e) => emit('update:selectedVMType', (e.target as HTMLSelectElement).value)"
              class="vm-select"
            >
              <option v-for="vm in hypervisorList" :key="vm.type" :value="vm.type">
                {{ vm.name }}
              </option>
            </select>
            <span class="select-arrow">▾</span>
          </div>
        </div>
      </div>

      <!-- Advanced Hardware Tuning Row -->
      <div class="vm-tuning-row">
        <!-- CPU Cores -->
        <div class="vm-tuning-item">
          <span class="vm-selector-label">{{ t("vm.cfg_cpu") }}</span>
          <div class="vm-select-wrapper sm">
            <select
              :value="vmCpuCores || 2"
              :disabled="!canConfigureVm"
              @change="(e) => emit('update:vmCpuCores', Number((e.target as HTMLSelectElement).value))"
              class="vm-select sm"
            >
              <option :value="1">1 {{ t("vm.cfg_core_singular") }}</option>
              <option :value="2">2 {{ t("vm.cfg_core_plural") }}</option>
              <option :value="4">4 {{ t("vm.cfg_core_plural") }}</option>
              <option :value="8">8 {{ t("vm.cfg_core_plural") }}</option>
            </select>
            <span class="select-arrow">▾</span>
          </div>
        </div>

        <!-- RAM Allocation -->
        <div class="vm-tuning-item">
          <span class="vm-selector-label">{{ t("vm.cfg_ram") }}</span>
          <div class="vm-select-wrapper sm">
            <select
              :value="vmMemoryMB || 2048"
              :disabled="!canConfigureVm"
              @change="(e) => emit('update:vmMemoryMB', Number((e.target as HTMLSelectElement).value))"
              class="vm-select sm"
            >
              <option :value="1024">1 GB</option>
              <option :value="2048">2 GB</option>
              <option :value="4096">4 GB</option>
              <option :value="8192">8 GB</option>
            </select>
            <span class="select-arrow">▾</span>
          </div>
        </div>

        <!-- Hardware Acceleration Toggle -->
        <label class="vm-checkbox-label">
          <input
            type="checkbox"
            :disabled="!canConfigureVm"
            :checked="vmDisplayAccel !== false"
            @change="(e) => emit('update:vmDisplayAccel', (e.target as HTMLInputElement).checked)"
          />
          {{ t("vm.cfg_accel") }}
        </label>
      </div>
      <p class="vm-desc">
        {{ t("vm.target") }}
        <strong v-if="activeVmTargetDevice" class="target-highlight">
          {{ formatVmTargetDisplay(activeVmTargetName, activeVmTargetDevice) }}
        </strong>
        <span v-else class="target-warn">
          {{ t("vm.no_disk_warn") }}
        </span>
      </p>
      <button
        class="vm-launch-btn"
        :class="{ 'vm-running': isVmRunning }"
        :disabled="isVmDisabled"
        :title="isVmRunning ? t('vm.tip_running') : vmDisabledReason"
        @click="isVmRunning ? emit('stop-vm') : !isVmDisabled && emit('launch-vm')"
      >
        <span class="btn-icon">{{ isLaunchingQemu ? "⏳" : isVmRunning ? "⏹" : "▶" }}</span>
        <span>{{ isLaunchingQemu ? t("vm.launching") : isVmRunning ? t("vm.running") : t("vm.run_test") }}</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import CustomSelect from "./CustomSelect.vue";
import ProgressBar from "./ProgressBar.vue";
import type { DiskInfo } from "./DiskPanel.vue";
import { t } from "../i18n";

const props = defineProps<{
  activeMode: "cloud" | "hybrid";
  selectionMode: "single" | "batch";
  selectedDisk: DiskInfo | null;
  selectedDevices: Set<string>;
  diskList: DiskInfo[];
  selectedFsType: string;
  isNonDestructive: boolean;
  isMacOs: boolean;
  ventoyStatus: { valid: boolean; version?: string; error?: string };
  selectedIsoFiles: { name: string; path: string }[];
  isDeploying: boolean;
  isPreflight?: boolean;
  deployProgress: number;
  batchDeployInfo?: {
    totalDisks: number;
    currentDiskIndex: number;
    currentDisk: string;
    currentStage: string;
    diskProgress: number;
    overallProgress: number;
    speedMBps: number;
    elapsedSec: number;
    etaSec: number;
  } | null;
  speedMBps?: number;
  elapsedSec?: number;
  etaSec?: number;
  deployBtnText: string;
  isDeployDisabled: boolean;
  deployDisabledReason: string;
  showDeploySuccessBanner: boolean;
  deploySuccessBanner: { autoEjected?: boolean; targets: string[] };
  hypervisorList: { type: string; name: string; installed: boolean }[];
  selectedBootMode: string;
  selectedVMType: string;
  vmCpuCores?: number;
  vmMemoryMB?: number;
  vmDisplayAccel?: boolean;
  isVmDisabled: boolean;
  vmDisabledReason: string;
  isLaunchingQemu: boolean;
  isVmRunning?: boolean;
  activeVmTargetName: string;
  activeVmTargetDevice: string;
  canChangeFs?: boolean;
  canManageIso?: boolean;
  canVerifyHash?: boolean;
  canConfigureVm?: boolean;
}>();

function formatStatsTime(seconds?: number): string {
  if (!seconds || seconds <= 0) return "00:00";
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

function formatSelectedDiskTarget(disk: DiskInfo): string {
  const name = disk.name || disk.device;
  const dev = disk.device;
  const mp = disk.mountPoint;
  if (mp && /^[A-Za-z]:[\\/]?$/.test(mp.trim())) {
    const cleanMp = mp.trim().endsWith("\\") ? mp.trim() : mp.trim() + "\\";
    return `${name} (${cleanMp} • ${dev})`;
  }
  if (mp && mp.trim() !== dev.trim()) {
    return `${name} (${mp.trim()} • ${dev})`;
  }
  return `${name} (${dev})`;
}

function formatVmTargetDisplay(name: string, dev: string): string {
  if (props.selectedDisk && props.selectedDisk.device === dev && props.selectedDisk.mountPoint) {
    const mp = props.selectedDisk.mountPoint;
    if (/^[A-Za-z]:[\\/]?$/.test(mp.trim())) {
      const cleanMp = mp.trim().endsWith("\\") ? mp.trim() : mp.trim() + "\\";
      return `${name} (${cleanMp} • ${dev})`;
    }
    if (mp.trim() !== dev.trim()) {
      return `${name} (${mp.trim()} • ${dev})`;
    }
  }
  return `${name} (${dev})`;
}

const emit = defineEmits<{
  (e: "update:selectedFsType", fs: string): void;
  (e: "update:selectedBootMode", mode: string): void;
  (e: "update:selectedVMType", type: string): void;
  (e: "update:vmCpuCores", cores: number): void;
  (e: "update:vmMemoryMB", ram: number): void;
  (e: "update:vmDisplayAccel", accel: boolean): void;
  (e: "open-settings-ventoy"): void;
  (e: "select-iso"): void;
  (e: "remove-iso", index: number): void;
  (e: "clear-iso"): void;
  (e: "drop-iso-paths", paths: string[]): void;
  (e: "deploy-click"): void;
  (e: "cancel-deploy"): void;
  (e: "dismiss-success-banner"): void;
  (e: "safely-eject-success"): void;
  (e: "launch-vm"): void;
  (e: "stop-vm"): void;
  (e: "update:isVerifying", verifying: boolean): void;
}>();

// Drag & Drop State & Handlers
const isDragOver = ref(false);
let dragCounter = 0;

function handleDragEnter(e: DragEvent) {
  e.preventDefault();
  dragCounter++;
  isDragOver.value = true;
}

function handleDragOver(e: DragEvent) {
  e.preventDefault();
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = "copy";
  }
  isDragOver.value = true;
}

function handleDragLeave(e: DragEvent) {
  e.preventDefault();
  dragCounter--;
  if (dragCounter <= 0) {
    dragCounter = 0;
    isDragOver.value = false;
  }
}

function handleDrop(e: DragEvent) {
  e.preventDefault();
  dragCounter = 0;
  isDragOver.value = false;
  if (!canManageIso.value) return;

  // In native Wails desktop runtime, OnFileDrop receives the system absolute paths.
  // Only fall back to HTML5 File API in pure browser demo mode.
  const isWailsDesktop = typeof (window as any).runtime?.OnFileDrop === "function";
  if (!isWailsDesktop && e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files.length > 0) {
    const paths: string[] = [];
    for (let i = 0; i < e.dataTransfer.files.length; i++) {
      const file = e.dataTransfer.files[i] as any;
      const p = file.path || file.name;
      if (p) paths.push(p);
    }
    if (paths.length > 0) {
      emit("drop-iso-paths", paths);
    }
  }
}

// Capabilities strictly driven by central state machine
const isDeployDisabled = computed(() => props.isDeployDisabled);
const canChangeFs = computed(() => props.canChangeFs ?? (!props.isDeploying && !props.isVmRunning));
const canManageIso = computed(() => props.canManageIso ?? (!props.isDeploying && !props.isVmRunning));
const canVerifyHash = computed(() => props.canVerifyHash ?? (!props.isDeploying && !props.isVmRunning));
const canConfigureVm = computed(() => props.canConfigureVm ?? (!props.isDeploying && !props.isVmRunning));

// Checksum State & Logic
const selectedChecksumIsoIndex = ref(0);
const selectedAlgo = ref("sha256");
const isCalculatingHash = ref(false);
const calculatedHash = ref("");
const currentChecksumAlgo = ref("sha256");
const expectedHashInput = ref("");
const isHashCopied = ref(false);
const sumsFileInputRef = ref<HTMLInputElement | null>(null);

// SHA 缓存与单个 ISO 的校验状态
interface ChecksumCache {
  [filename: string]: string;
}

export interface IsoChecksumStatus {
  calculated: string;
  expected: string;
  status: "idle" | "calculating" | "match" | "mismatch" | "no-expected";
  algo: string;
}

const checksumCache = ref<ChecksumCache>({});
const isoChecksumStatuses = ref<Record<string, IsoChecksumStatus>>({});
const loadedSumsFileCount = ref(0);
const isBatchCalculating = ref(false);
const batchProgress = ref({ current: 0, total: 0, matched: 0, mismatched: 0 });
const cachedHashCount = computed(() => Object.keys(checksumCache.value).length);
const checksumActiveTab = ref<"batch" | "single">("batch");
const isChecksumPanelExpanded = ref(false);
const hasAnyChecksumResult = computed(() => {
  return Object.values(isoChecksumStatuses.value).some((s) => s && (s.status === "match" || s.status === "mismatch"));
});

// 智能默认聚焦 Tab：当镜像数量从 <= 1 增加到多个时，自动聚焦「批量校验」；仅有 1 个或没有时聚焦「单个校验」
watch(
  () => props.selectedIsoFiles.length,
  (newCount, oldCount) => {
    if (newCount > 1 && (oldCount === undefined || oldCount <= 1)) {
      checksumActiveTab.value = "batch";
    } else if (newCount <= 1) {
      checksumActiveTab.value = "single";
    }
  },
  { immediate: true }
);

watch(
  () => props.selectedIsoFiles,
  (newFiles: any[]) => {
    if (selectedChecksumIsoIndex.value >= newFiles.length) {
      selectedChecksumIsoIndex.value = 0;
    }
    calculatedHash.value = "";

    // 当选择新 ISO 时，自动从缓存中查找期望值
    if (newFiles.length > 0 && selectedChecksumIsoIndex.value < newFiles.length) {
      const currentFile = newFiles[selectedChecksumIsoIndex.value];
      const cachedHash = checksumCache.value[currentFile.name];
      if (cachedHash && !expectedHashInput.value) {
        expectedHashInput.value = cachedHash;
      }
    }
  },
  { deep: true }
);

watch(
  () => isCalculatingHash.value || isBatchCalculating.value,
  (verifying) => {
    emit("update:isVerifying", verifying);
  }
);

function inspectSingleIso(index: number) {
  selectedChecksumIsoIndex.value = index;
  checksumActiveTab.value = "single";
  isChecksumPanelExpanded.value = true;
}

watch(selectedChecksumIsoIndex, () => {
  if (props.selectedIsoFiles.length > 0 && selectedChecksumIsoIndex.value < props.selectedIsoFiles.length) {
    const currentFile = props.selectedIsoFiles[selectedChecksumIsoIndex.value];
    const status = isoChecksumStatuses.value[currentFile.name];
    if (status && status.calculated) {
      calculatedHash.value = status.calculated;
      currentChecksumAlgo.value = status.algo || selectedAlgo.value;
      expectedHashInput.value = status.expected || checksumCache.value[currentFile.name] || "";
    } else {
      calculatedHash.value = "";
      const cachedHash = checksumCache.value[currentFile.name];
      expectedHashInput.value = cachedHash || "";
    }
  } else {
    calculatedHash.value = "";
    expectedHashInput.value = "";
  }
});

function selectIsoForChecksum(index: number) {
  selectedChecksumIsoIndex.value = index;
  calculatedHash.value = "";
}

function parseExpectedHashString(rawInput: string, currentFileName: string): string {
  if (!rawInput) return "";
  const trimmed = rawInput.trim();

  // If it's a multi-line checksum file content (e.g. SHA256SUMS file)
  if (trimmed.includes("\n")) {
    const lines = trimmed.split("\n");
    for (const line of lines) {
      const lineTrimmed = line.trim();
      if (!lineTrimmed || lineTrimmed.startsWith("#")) continue;
      // Line format: "hash_string  filename" or "hash_string *filename"
      if (currentFileName && lineTrimmed.toLowerCase().includes(currentFileName.toLowerCase())) {
        const parts = lineTrimmed.split(/\s+/);
        if (parts.length >= 1) return parts[0].toLowerCase();
      }
    }
  }

  // Single line or direct hash string
  const parts = trimmed.split(/\s+/);
  return parts[0].toLowerCase();
}

const parsedExpectedHash = computed(() => {
  const targetIdx = selectedChecksumIsoIndex.value < props.selectedIsoFiles.length ? selectedChecksumIsoIndex.value : 0;
  const currentFileName = props.selectedIsoFiles.length > 0 ? props.selectedIsoFiles[targetIdx].name : "";
  return parseExpectedHashString(expectedHashInput.value, currentFileName);
});

const isHashMatching = computed(() => {
  if (!parsedExpectedHash.value || !calculatedHash.value) return false;
  return parsedExpectedHash.value === calculatedHash.value.trim().toLowerCase();
});

function triggerSumsFilePick() {
  if (sumsFileInputRef.value) {
    sumsFileInputRef.value.value = "";
    sumsFileInputRef.value.click();
  }
}

// 解析校验文件内容，提取所有文件名->哈希值映射；支持纯哈希纯文本自动关联当前选中 ISO
function parseChecksumFileContent(content: string, currentSelectedFileName?: string): ChecksumCache {
  const cache: ChecksumCache = {};
  const lines = content.split("\n");

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;

    // 支持标准格式：hash  filename 或 hash *filename 或 hash  ./path/to/filename
    const match = trimmed.match(/^([a-fA-F0-9]+)\s+\*?(.+)$/);
    if (match) {
      const [, hash, filepath] = match;
      const filename = filepath.split("/").pop()?.trim() || filepath.trim();
      cache[filename] = hash.toLowerCase();
    } else if (currentSelectedFileName && /^[a-fA-F0-9]{32,128}$/.test(trimmed)) {
      // 兼容单文件纯哈希文件（无文件名）：自动关联给当前选中的 ISO
      cache[currentSelectedFileName] = trimmed.toLowerCase();
    }
  }

  return cache;
}

// 辅助函数：读取文件为文本
function readFileAsText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      resolve((e.target?.result as string) || "");
    };
    reader.onerror = () => {
      reject(reader.error);
    };
    reader.readAsText(file);
  });
}

// 智能推断 Hash 算法：优先通过文件名/扩展名判断，次选通过内容中 Hash 字符串长度判断
function detectChecksumAlgorithm(fileNames: string[], hashes: string[]): string | null {
  // 1. 优先检查文件名与扩展名
  for (const name of fileNames) {
    const lower = (name || "").toLowerCase();
    if (
      lower.includes("sha256") ||
      lower.endsWith(".sha256") ||
      lower.endsWith(".sha256sum") ||
      lower.endsWith(".sha256sums")
    )
      return "sha256";
    if (
      lower.includes("sha512") ||
      lower.endsWith(".sha512") ||
      lower.endsWith(".sha512sum") ||
      lower.endsWith(".sha512sums")
    )
      return "sha512";
    if (lower.includes("sha384") || lower.endsWith(".sha384") || lower.endsWith(".sha384sum")) return "sha384";
    if (lower.includes("sha1") || lower.endsWith(".sha1") || lower.endsWith(".sha1sum")) return "sha1";
    if (lower.includes("md5") || lower.endsWith(".md5") || lower.endsWith(".md5sum") || lower.endsWith(".md5sums"))
      return "md5";
    if (lower.includes("crc32") || lower.endsWith(".crc") || lower.endsWith(".sfv")) return "crc32";
  }

  // 2. 次选：根据已提取到的哈希值长度进行密码学特征推断
  for (const h of hashes) {
    const clean = (h || "").trim();
    if (/^[a-fA-F0-9]{64}$/.test(clean)) return "sha256";
    if (/^[a-fA-F0-9]{128}$/.test(clean)) return "sha512";
    if (/^[a-fA-F0-9]{32}$/.test(clean)) return "md5";
    if (/^[a-fA-F0-9]{40}$/.test(clean)) return "sha1";
    if (/^[a-fA-F0-9]{96}$/.test(clean)) return "sha384";
    if (/^[a-fA-F0-9]{8}$/.test(clean)) return "crc32";
  }

  return null;
}

// 格式化算法展示名称
function formatAlgoDisplayName(algo: string): string {
  switch (algo.toLowerCase()) {
    case "sha256":
      return "SHA-256";
    case "sha512":
      return "SHA-512";
    case "sha384":
      return "SHA-384";
    case "sha1":
      return "SHA-1";
    case "md5":
      return "MD5";
    case "crc32":
      return "CRC32";
    default:
      return algo.toUpperCase();
  }
}

// 监听单文件手动输入期望 Hash，实时同步到缓存与状态中，确保单文件手动修改与批量校验无缝兼容
watch(expectedHashInput, (newVal) => {
  if (props.selectedIsoFiles.length > 0 && selectedChecksumIsoIndex.value < props.selectedIsoFiles.length) {
    const currentFile = props.selectedIsoFiles[selectedChecksumIsoIndex.value];
    const parsed = parseExpectedHashString(newVal, currentFile.name);
    if (parsed) {
      const autoAlgo = detectChecksumAlgorithm([], [parsed]);
      if (autoAlgo && autoAlgo !== selectedAlgo.value) {
        selectedAlgo.value = autoAlgo;
        currentChecksumAlgo.value = autoAlgo;
      }
      checksumCache.value[currentFile.name] = parsed;
      const existing = isoChecksumStatuses.value[currentFile.name];
      if (existing) {
        existing.expected = parsed;
        existing.algo = selectedAlgo.value;
        if (existing.calculated) {
          existing.status = existing.calculated.toLowerCase() === parsed.toLowerCase() ? "match" : "mismatch";
        }
      }
    }
  }
});

async function handleSumsFileSelected(event: Event) {
  const target = event.target as HTMLInputElement;
  if (!target.files || target.files.length === 0) return;

  const files = Array.from(target.files);
  let totalLoaded = 0;
  let totalHashes = 0;

  const currentFile =
    props.selectedIsoFiles.length > 0 && selectedChecksumIsoIndex.value < props.selectedIsoFiles.length
      ? props.selectedIsoFiles[selectedChecksumIsoIndex.value]
      : undefined;

  // 逐个读取所有选中的文件
  for (const file of files) {
    try {
      const text = await readFileAsText(file);
      if (text) {
        const parsed = parseChecksumFileContent(text, currentFile?.name);
        const hashCount = Object.keys(parsed).length;

        if (hashCount > 0) {
          // 合并到缓存中
          Object.assign(checksumCache.value, parsed);
          totalLoaded++;
          totalHashes += hashCount;
        }
      }
    } catch (err) {
      console.error(`Failed to read file ${file.name}:`, err);
    }
  }

  loadedSumsFileCount.value = totalLoaded;

  // 显示加载结果并同步所有 ISO 的校验状态
  if (totalLoaded > 0) {
    // 自动检测校验文件对应的 Hash 算法并联动切换
    const fileNames = files.map((f) => f.name);
    const allHashes = Object.values(checksumCache.value);
    const detectedAlgo = detectChecksumAlgorithm(fileNames, allHashes);
    let algoSwitchedMsg = "";

    if (detectedAlgo && detectedAlgo !== selectedAlgo.value) {
      selectedAlgo.value = detectedAlgo;
      currentChecksumAlgo.value = detectedAlgo;
      const displayAlgo = formatAlgoDisplayName(detectedAlgo);
      algoSwitchedMsg = `\n${t("checksum.auto_algo_switched", { algo: displayAlgo })}`;
    }

    // 自动为当前所有文件匹配期望值
    props.selectedIsoFiles.forEach((file) => {
      const cached = checksumCache.value[file.name];
      if (cached) {
        const existing = isoChecksumStatuses.value[file.name];
        if (existing && existing.calculated) {
          existing.expected = cached;
          existing.status = existing.calculated.toLowerCase() === cached.toLowerCase() ? "match" : "mismatch";
          existing.algo = selectedAlgo.value;
        } else if (!existing) {
          isoChecksumStatuses.value[file.name] = {
            calculated: "",
            expected: cached,
            status: "idle",
            algo: selectedAlgo.value,
          };
        }
      }
    });

    if (props.selectedIsoFiles.length > 0 && selectedChecksumIsoIndex.value < props.selectedIsoFiles.length) {
      const currentFile = props.selectedIsoFiles[selectedChecksumIsoIndex.value];
      const cachedHash = checksumCache.value[currentFile.name];
      expectedHashInput.value = cachedHash || "";
    } else {
      expectedHashInput.value = "";
    }

    alert(t("checksum.cache_loaded", { count: totalHashes }) + algoSwitchedMsg);
  } else {
    expectedHashInput.value = "";
    alert(t("checksum.no_valid_hashes"));
  }
}

async function handleCalculateChecksum() {
  if (
    props.selectedIsoFiles.length === 0 ||
    !canVerifyHash.value ||
    isCalculatingHash.value ||
    isBatchCalculating.value
  )
    return;
  const targetIdx = selectedChecksumIsoIndex.value < props.selectedIsoFiles.length ? selectedChecksumIsoIndex.value : 0;
  const fileToVerify = props.selectedIsoFiles[targetIdx];
  isCalculatingHash.value = true;
  isHashCopied.value = false;
  try {
    const w = window as any;
    if (w.go && w.go.main && w.go.main.App && typeof w.go.main.App.CalculateFileChecksum === "function") {
      const res = await w.go.main.App.CalculateFileChecksum(fileToVerify.path, selectedAlgo.value);
      if (res && res.hash) {
        calculatedHash.value = res.hash.toLowerCase();
        currentChecksumAlgo.value = res.algorithm || selectedAlgo.value;
      }
    } else {
      // Standalone preview mock hash calculation
      await new Promise((resolve) => setTimeout(resolve, 500));
      calculatedHash.value =
        selectedAlgo.value === "md5"
          ? "e10adc3949ba59abbe56e057f20f883e"
          : "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824";
      currentChecksumAlgo.value = selectedAlgo.value;
    }

    // 同步更新单个文件的状态
    if (calculatedHash.value) {
      const exp = parseExpectedHashString(expectedHashInput.value, fileToVerify.name);
      let statusType: "match" | "mismatch" | "no-expected" = "no-expected";
      if (exp) {
        statusType = exp.toLowerCase() === calculatedHash.value.toLowerCase() ? "match" : "mismatch";
      }
      isoChecksumStatuses.value[fileToVerify.name] = {
        calculated: calculatedHash.value,
        expected: exp,
        status: statusType,
        algo: currentChecksumAlgo.value,
      };
    }
  } catch (e) {
    console.error("Checksum calculation error:", e);
  } finally {
    isCalculatingHash.value = false;
  }
}

async function handleBatchChecksum() {
  if (
    !props.selectedIsoFiles ||
    props.selectedIsoFiles.length === 0 ||
    !canVerifyHash.value ||
    isBatchCalculating.value ||
    isCalculatingHash.value
  )
    return;

  isBatchCalculating.value = true;
  const total = props.selectedIsoFiles.length;
  let matched = 0;
  let mismatched = 0;

  batchProgress.value = { current: 0, total, matched: 0, mismatched: 0 };

  const w = window as any;
  for (let i = 0; i < total; i++) {
    const file = props.selectedIsoFiles[i];
    batchProgress.value.current = i + 1;

    const expected = checksumCache.value[file.name] || "";
    isoChecksumStatuses.value[file.name] = {
      calculated: "",
      expected,
      status: "calculating",
      algo: selectedAlgo.value,
    };

    let computedHash = "";
    try {
      if (w.go && w.go.main && w.go.main.App && typeof w.go.main.App.CalculateFileChecksum === "function") {
        const res = await w.go.main.App.CalculateFileChecksum(file.path, selectedAlgo.value);
        if (res && res.hash) {
          computedHash = res.hash.toLowerCase();
        }
      } else {
        await new Promise((resolve) => setTimeout(resolve, 300));
        computedHash =
          selectedAlgo.value === "md5"
            ? "e10adc3949ba59abbe56e057f20f883e"
            : "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824";
      }
    } catch (err) {
      console.error(`Batch checksum failed for ${file.name}:`, err);
    }

    if (computedHash) {
      const isMatch = expected && computedHash === expected.toLowerCase();
      let statusType: "match" | "mismatch" | "no-expected" = "no-expected";
      if (expected) {
        if (isMatch) {
          statusType = "match";
          matched++;
        } else {
          statusType = "mismatch";
          mismatched++;
        }
      }
      isoChecksumStatuses.value[file.name] = {
        calculated: computedHash,
        expected,
        status: statusType,
        algo: selectedAlgo.value,
      };
    } else {
      isoChecksumStatuses.value[file.name] = {
        calculated: "",
        expected,
        status: "mismatch",
        algo: selectedAlgo.value,
      };
      mismatched++;
    }

    batchProgress.value.matched = matched;
    batchProgress.value.mismatched = mismatched;

    // 若当前项正是单文件视图当前选中的项，联动更新下方视图
    if (i === selectedChecksumIsoIndex.value) {
      calculatedHash.value = computedHash;
      currentChecksumAlgo.value = selectedAlgo.value;
      if (expected) {
        expectedHashInput.value = expected;
      }
    }
  }

  isBatchCalculating.value = false;
}

function getIsoChecksumStatus(fileName: string): IsoChecksumStatus | undefined {
  return isoChecksumStatuses.value[fileName];
}

function getIsoChecksumStatusIcon(fileName: string): string {
  const s = isoChecksumStatuses.value[fileName];
  if (!s) return "";
  switch (s.status) {
    case "calculating":
      return "⏳";
    case "match":
      return "✓";
    case "mismatch":
      return "✕";
    case "no-expected":
      return "❓";
    default:
      return "";
  }
}

function getIsoChecksumStatusText(fileName: string): string {
  const s = isoChecksumStatuses.value[fileName];
  if (!s) return "";
  switch (s.status) {
    case "calculating":
      return t("checksum.status_calculating");
    case "match":
      return t("checksum.status_match");
    case "mismatch":
      return t("checksum.status_mismatch");
    case "no-expected":
      return t("checksum.status_no_expected");
    default:
      return t("checksum.status_idle");
  }
}

const batchSummaryText = computed(() => {
  if (batchProgress.value.total === 0) return "";
  if (isBatchCalculating.value) {
    return t("checksum.batch_verifying", { current: batchProgress.value.current, total: batchProgress.value.total });
  }
  return t("checksum.batch_result", {
    matched: batchProgress.value.matched,
    mismatched: batchProgress.value.mismatched,
  });
});

const isAllBatchMatched = computed(() => {
  return (
    batchProgress.value.total > 0 &&
    !isBatchCalculating.value &&
    batchProgress.value.matched === batchProgress.value.total
  );
});

function copyHashToClipboard() {
  if (!calculatedHash.value) return;
  navigator.clipboard.writeText(calculatedHash.value);
  isHashCopied.value = true;
  setTimeout(() => {
    isHashCopied.value = false;
  }, 2000);
}

function getFileIcon(filename: string): string {
  const ext = filename.split(".").pop()?.toLowerCase();
  switch (ext) {
    case "iso":
      return "💿";
    case "wim":
      return "📦";
    case "img":
    case "raw":
      return "💾";
    case "vhd":
    case "vhdx":
    case "vti":
      return "💽";
    case "efi":
    case "bin":
      return "⚙️";
    case "xz":
    case "gz":
      return "🗜️";
    default:
      return "📄";
  }
}
</script>

<style scoped>
.section-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 16px;
  padding: 1.5rem;
  backdrop-filter: blur(16px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
}

.section-card h2 {
  font-size: 1.2rem;
  font-weight: 700;
  margin: 0 0 0.25rem 0;
  color: var(--text-main);
}

.section-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0 0 1rem 0;
}

.fs-selector {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
  margin-bottom: 1rem;
}

.fs-label {
  flex: 0 0 auto;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
}

.fs-selector :deep(.custom-select-container) {
  flex: 1 1 auto;
  width: auto;
  min-width: 0;
}

.safe-mode-notice {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: var(--alert-success-bg);
  border: 1px solid var(--alert-success-border);
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.safe-notice-icon {
  font-size: 22px;
  flex-shrink: 0;
}

.safe-notice-content {
  flex: 1;
}

.safe-notice-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--alert-success-title);
  margin-bottom: 4px;
}

.safe-notice-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
}

.safe-notice-desc b {
  color: var(--alert-success-title);
}

.ventoy-warning-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.warning-card-icon {
  font-size: 22px;
  flex-shrink: 0;
  margin-top: 2px;
}

.warning-card-body {
  flex: 1;
  min-width: 0;
}

.warning-card-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--alert-danger-title);
  margin-bottom: 4px;
}

.warning-card-message {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
}

.ventoy-warning-card .btn-secondary,
.btn-secondary.btn-sm,
.btn-sm {
  width: auto !important;
  white-space: nowrap !important;
  flex-shrink: 0 !important;
  padding: 0.45rem 0.85rem !important;
  font-size: 0.8rem !important;
  margin-top: 2px;
}

.iso-card {
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1rem;
  margin-bottom: 1rem;
  background: var(--subtab-container-bg);
}

.iso-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.iso-title-group {
  flex: 1;
  min-width: 0;
}

.iso-title-group h3 {
  font-size: 0.95rem;
  font-weight: 700;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--text-main);
}

.optional-badge {
  font-size: 0.72rem;
  background: var(--badge-optional-bg, rgba(56, 189, 248, 0.16));
  border: 1px solid var(--badge-optional-border, rgba(56, 189, 248, 0.35));
  color: var(--badge-optional-text, #38bdf8);
  padding: 0.12rem 0.5rem;
  border-radius: 9999px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  line-height: 1.2;
  letter-spacing: 0.02em;
  user-select: none;
  transition: all 0.2s ease;
}

.iso-subtitle {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.add-iso-btn {
  padding: 0.35rem 0.85rem !important;
  font-size: 0.75rem !important;
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
  width: auto !important;
  border-radius: 6px;
  background: linear-gradient(180deg, rgba(0, 229, 255, 0.2) 0%, rgba(0, 229, 255, 0.08) 100%);
  border: 1px solid rgba(0, 229, 255, 0.35);
  color: var(--accent-cyan);
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.2),
    0 1px 4px rgba(0, 229, 255, 0.15);
}

.add-iso-btn:hover {
  background: linear-gradient(180deg, rgba(0, 229, 255, 0.3) 0%, rgba(0, 229, 255, 0.15) 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.3),
    0 3px 10px rgba(0, 229, 255, 0.3);
}

[data-theme="light"] .add-iso-btn {
  background: linear-gradient(180deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #38bdf8;
  color: #0284c7;
  box-shadow:
    inset 0 1px 0 #ffffff,
    0 1px 3px rgba(2, 132, 199, 0.1);
}

[data-theme="light"] .add-iso-btn:hover {
  background: linear-gradient(180deg, #e0f2fe 0%, #bae6fd 100%);
  border-color: #0284c7;
  color: #0369a1;
  box-shadow:
    inset 0 1px 0 #ffffff,
    0 3px 8px rgba(2, 132, 199, 0.2);
}

.deploy-btn {
  width: 100%;
  padding: 0.8rem;
  font-size: 0.95rem;
  font-weight: 700;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: linear-gradient(135deg, #00f0ff 0%, #0077ff 100%);
  color: #070a12;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.4),
    0 4px 15px rgba(0, 229, 255, 0.35);
}

.deploy-btn:hover:not(:disabled) {
  transform: translateY(-1.5px);
  background: linear-gradient(135deg, #38f9ff 0%, #1a8cff 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.5),
    0 6px 20px rgba(0, 229, 255, 0.5);
}

[data-theme="light"] .deploy-btn {
  background: linear-gradient(135deg, #0284c7 0%, #1d4ed8 100%);
  color: #ffffff;
  border: 1px solid rgba(2, 132, 199, 0.3);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.4),
    0 4px 14px rgba(2, 132, 199, 0.35);
}

[data-theme="light"] .deploy-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #0369a1 0%, #1e40af 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.5),
    0 6px 20px rgba(2, 132, 199, 0.45);
}

.iso-card {
  transition:
    border-color 0.25s ease,
    box-shadow 0.25s ease;
}

.iso-card.is-drag-over {
  border-color: #38bdf8 !important;
  box-shadow: 0 0 24px rgba(56, 189, 248, 0.35) !important;
}

.iso-list-container {
  position: relative;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.75rem;
  transition: border-color 0.2s ease;
}

.iso-drag-overlay {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.86);
  backdrop-filter: blur(4px);
  border: 2px dashed #38bdf8;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 50;
  pointer-events: none;
}

.iso-drag-overlay .drag-icon {
  font-size: 2.2rem;
  margin-bottom: 0.4rem;
  animation: dragBounce 0.7s infinite alternate ease-in-out;
}

.iso-drag-overlay .drag-text {
  font-size: 0.95rem;
  font-weight: 700;
  color: #38bdf8;
  letter-spacing: 0.5px;
}

@keyframes dragBounce {
  from {
    transform: translateY(0);
  }
  to {
    transform: translateY(-5px);
  }
}

.iso-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.2rem;
  border: 2px dashed var(--card-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.iso-empty-state:hover {
  background: var(--btn-sec-hover-bg);
}

.iso-empty-state.disabled {
  opacity: 0.6;
  cursor: not-allowed;
  pointer-events: none;
}

.iso-empty-state.drag-active {
  border-color: #38bdf8 !important;
  background: rgba(56, 189, 248, 0.12) !important;
  transform: scale(1.01);
  box-shadow: 0 0 20px rgba(56, 189, 248, 0.28);
}

.iso-empty-state.drag-active .empty-icon {
  animation: dragBounce 0.7s infinite alternate ease-in-out;
}

.iso-empty-state.drag-active .empty-text {
  color: #38bdf8;
}

.empty-icon {
  font-size: 1.5rem;
  margin-bottom: 0.3rem;
  transition: transform 0.2s ease;
}

.empty-text {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
  transition: color 0.2s ease;
}

.empty-subtext {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.iso-file-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  max-height: 140px;
  overflow-y: auto;
}

.iso-file-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.45rem 0.65rem;
  background: var(--btn-sec-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  font-size: 0.82rem;
  transition: all 0.2s ease;
}

.iso-file-item.is-checksum-target {
  border-color: var(--accent-cyan, #38bdf8);
  background: rgba(56, 189, 248, 0.08);
}

.iso-file-info {
  flex: 1;
  overflow: hidden;
  cursor: pointer;
}

.iso-item-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.iso-check-hash-btn {
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-muted);
  border-radius: 4px;
  padding: 0.15rem 0.45rem;
  font-size: 0.72rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.iso-check-hash-btn:hover {
  border-color: var(--accent-cyan, #38bdf8);
  color: var(--accent-cyan, #38bdf8);
}

.iso-check-hash-btn.active {
  background: rgba(56, 189, 248, 0.2);
  color: var(--accent-cyan, #38bdf8);
  border-color: var(--accent-cyan, #38bdf8);
  font-weight: 600;
}

.iso-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  font-size: 0.68rem;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
  border: 1px solid transparent;
}

.iso-status-badge .badge-icon {
  font-size: 0.72rem;
  line-height: 1;
}

.iso-status-badge.calculating {
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  border-color: rgba(56, 189, 248, 0.35);
}

.iso-status-badge.match {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
  border-color: rgba(16, 185, 129, 0.35);
}

.iso-status-badge.mismatch {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
  border-color: rgba(239, 68, 68, 0.35);
}

.iso-status-badge.no-expected {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border-color: rgba(245, 158, 11, 0.35);
}

.iso-status-badge.idle {
  background: rgba(148, 163, 184, 0.15);
  color: var(--text-muted);
  border-color: rgba(148, 163, 184, 0.25);
}

.checksum-target-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0.55rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  font-size: 0.78rem;
}

.target-bar-label {
  font-weight: 600;
  color: var(--text-muted);
  white-space: nowrap;
}

.target-iso-selector {
  flex: 1;
  overflow: hidden;
}

.target-iso-select {
  width: 100%;
  padding: 0.2rem 0.4rem;
  font-size: 0.76rem;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-main);
  text-overflow: ellipsis;
}

.target-iso-single-name {
  flex: 1;
  font-weight: 600;
  color: var(--accent-cyan, #38bdf8);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.iso-file-info {
  flex: 1;
  overflow: hidden;
}

.iso-file-name {
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.iso-file-path {
  font-size: 0.72rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.iso-remove-btn {
  border: none;
  background: transparent;
  color: #ef4444;
  cursor: pointer;
  font-size: 0.85rem;
}

.iso-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.5rem;
  padding-top: 0.4rem;
  border-top: 1px dashed var(--card-border);
}

.iso-footer-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.iso-count-summary {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.btn-toggle-checksum {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.22rem 0.65rem;
  font-size: 0.74rem;
  font-weight: 600;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-toggle-checksum .btn-icon {
  font-size: 0.72rem;
}

.btn-toggle-checksum .caret-icon {
  font-size: 0.68rem;
  color: var(--text-muted);
  transition: transform 0.2s ease;
}

.btn-toggle-checksum:hover {
  background: rgba(56, 189, 248, 0.1);
  border-color: #38bdf8;
  color: #38bdf8;
}

.btn-toggle-checksum.active {
  background: rgba(56, 189, 248, 0.15);
  border-color: #38bdf8;
  color: #38bdf8;
}

.checksum-mini-badge {
  font-size: 0.68rem;
  line-height: 1;
  margin-left: 0.1rem;
}

.btn-clear-iso {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.55rem;
  font-size: 0.74rem;
  font-weight: 600;
  border-radius: 6px;
  border: 1px solid rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.08);
  color: #ef4444;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-clear-iso .btn-icon {
  font-size: 0.72rem;
  line-height: 1;
}

.btn-clear-iso:hover {
  background: rgba(239, 68, 68, 0.18);
  border-color: #ef4444;
  color: #f87171;
}

/* Checksum Verification Card Styles */
.checksum-container {
  margin-top: 0.75rem;
  padding-top: 0.65rem;
  border-top: 1px dashed var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.checksum-subtabs-nav {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem;
  background: var(--subtab-container-bg);
  border-radius: 8px;
  border: 1px solid var(--card-border);
  margin-bottom: 0.25rem;
}

.checksum-subtabs-nav .sub-tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.8rem;
  font-size: 0.78rem;
  font-weight: 600;
  border-radius: 6px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--subtab-btn-text);
  cursor: pointer;
  transition: all 0.15s ease;
}

.checksum-subtabs-nav .sub-tab-btn:hover {
  color: var(--text-main);
}

.checksum-subtabs-nav .sub-tab-btn.active {
  background: var(--subtab-btn-active-bg);
  color: var(--subtab-btn-active-text);
  border-color: var(--subtab-btn-active-border);
}

.tab-badge {
  background: rgba(56, 189, 248, 0.2);
  color: var(--accent-cyan);
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0.05rem 0.35rem;
  border-radius: 9999px;
  line-height: 1.2;
}

.checksum-tab-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.checksum-panel-desc {
  font-size: 0.72rem;
  color: var(--text-muted);
}

.hash-empty-hint {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.6rem;
  font-size: 0.74rem;
  color: var(--text-muted);
  background: var(--input-bg);
  border-radius: 6px;
  border: 1px dashed var(--card-border);
}

.hash-empty-hint .hint-icon {
  color: var(--accent-cyan);
}

.checksum-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.checksum-header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.checksum-header-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.checksum-title {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
}

.algo-selector {
  display: flex;
  align-items: center;
}

.algo-select {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-main);
}

.import-sums-header-btn {
  padding: 0.3rem 0.65rem !important;
  font-size: 0.78rem !important;
  border-radius: 6px !important;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.cached-count-pill {
  background: rgba(56, 189, 248, 0.25);
  color: var(--accent-cyan, #38bdf8);
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0.05rem 0.35rem;
  border-radius: 9999px;
  line-height: 1.2;
}

[data-theme="light"] .cached-count-pill {
  background: #bae6fd;
  color: #0284c7;
}

.calc-hash-btn {
  padding: 0.3rem 0.75rem !important;
  font-size: 0.78rem !important;
  border-radius: 6px !important;
}

.batch-calc-btn {
  padding: 0.3rem 0.75rem !important;
  font-size: 0.78rem !important;
  border-radius: 6px !important;
  background: var(--btn-sec-bg);
  border: 1px solid var(--accent-cyan, #38bdf8) !important;
  color: var(--accent-cyan, #38bdf8) !important;
  font-weight: 600;
  transition: all 0.15s ease;
}

.batch-calc-btn:hover:not(:disabled) {
  background: rgba(56, 189, 248, 0.18) !important;
}

.batch-summary-banner {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  background: rgba(56, 189, 248, 0.1);
  border: 1px solid rgba(56, 189, 248, 0.25);
  color: var(--text-main);
}

.batch-summary-banner.all-matched {
  background: rgba(16, 185, 129, 0.12);
  border-color: rgba(16, 185, 129, 0.3);
  color: #34d399;
}

.checksum-result-box {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  padding: 0.55rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
}

.hash-code-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  overflow: hidden;
}

.hash-algo-badge {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  background: rgba(56, 189, 248, 0.15);
  color: var(--accent-cyan);
}

.hash-code {
  flex: 1;
  font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
  font-size: 0.72rem;
  color: var(--text-main);
  background: var(--input-bg);
  padding: 0.25rem 0.4rem;
  border-radius: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.copy-hash-btn {
  background: var(--btn-sec-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  border-radius: 4px;
  padding: 0;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  min-width: 32px;
  height: 24px;
  flex-shrink: 0;
}

.copy-hash-btn:hover {
  background: var(--btn-sec-hover-bg);
  color: var(--accent-cyan);
  border-color: var(--accent-cyan);
}

.copy-hash-btn.copied {
  background: rgba(16, 185, 129, 0.18);
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.4);
}

.hash-compare-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.hash-compare-input {
  flex: 1;
  padding: 0.3rem 0.5rem;
  font-size: 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-main);
}

.import-sums-btn {
  background: var(--btn-sec-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.25rem 0.55rem;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s ease;
}

.import-sums-btn:hover {
  background: var(--btn-sec-hover-bg);
  color: var(--accent-cyan);
  border-color: var(--btn-sec-hover-border);
}

.match-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  white-space: nowrap;
}

.match-badge.match {
  background: var(--badge-success-bg);
  color: var(--badge-success-text);
  border: 1px solid var(--alert-success-border);
}

.match-badge.mismatch {
  background: var(--badge-danger-bg);
  color: var(--badge-danger-text);
  border: 1px solid var(--alert-danger-border);
}

.deploy-box {
  background: var(--subtab-container-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.selected-target {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
  color: var(--text-main);
}

.deploy-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  box-shadow: none !important;
  transform: none !important;
}

.deploy-btn.preflight-btn {
  background: linear-gradient(135deg, #f59e0b 0%, #b45309 100%);
  color: #ffffff;
  border: 1px solid rgba(245, 158, 11, 0.4);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.3),
    0 4px 15px rgba(245, 158, 11, 0.3);
  opacity: 0.85;
  cursor: wait;
  animation: preflight-pulse 1.2s ease-in-out infinite;
}

@keyframes preflight-pulse {
  0%,
  100% {
    opacity: 0.85;
  }
  50% {
    opacity: 0.65;
  }
}

.preflight-spinner {
  display: inline-block;
  animation: spinner-spin 1s linear infinite;
}

@keyframes spinner-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.deploy-btn.safe-btn {
  background: linear-gradient(135deg, #10b981 0%, #047857 100%);
  color: #ffffff;
  border: 1px solid rgba(16, 185, 129, 0.4);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.4),
    0 4px 15px rgba(16, 185, 129, 0.3);
}

.deploy-btn.safe-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #34d399 0%, #059669 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.5),
    0 6px 20px rgba(16, 185, 129, 0.45);
}

.deploy-success-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--alert-success-bg);
  border: 1px solid var(--alert-success-border);
  border-radius: 8px;
  margin-top: 0.75rem;
}

.deploy-success-icon {
  font-size: 1.5rem;
}

.deploy-success-content {
  flex: 1;
}

.deploy-success-title {
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--alert-success-title);
}

.deploy-success-desc {
  font-size: 0.78rem;
  color: var(--alert-success-text);
}

.deploy-success-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.btn-dismiss {
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.9rem;
}

.btn-eject-success {
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  border: 1px solid var(--alert-success-border);
  background: var(--badge-success-bg);
  color: var(--badge-success-text);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
}

.vm-box {
  border: 1px solid var(--card-border);
  border-radius: 12px;
  padding: 1rem;
  background: var(--subtab-container-bg);
}

.vm-header {
  margin-bottom: 0.75rem;
}

.vm-header h3 {
  font-size: 0.95rem;
  font-weight: 700;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--text-main);
}

.vm-title-group {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.vm-options-bar {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.vm-selector-container {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.vm-selector-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.vm-select-wrapper {
  position: relative;
}

.vm-select {
  width: 100%;
  padding: 0.4rem 1.6rem 0.4rem 0.6rem;
  border-radius: 8px;
  border: 1px solid var(--card-border);
  font-size: 0.82rem;
  appearance: none;
  background: var(--input-bg);
  color: var(--text-main);
}

.select-arrow {
  position: absolute;
  right: 0.6rem;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.badge {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  display: inline-block;
  font-weight: 600;
}

.badge.success {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
}

.badge.muted {
  background: var(--btn-sec-bg);
  color: var(--text-muted);
}

.vm-desc {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin: 0.4rem 0 0.75rem 0;
}

.target-highlight {
  color: var(--accent-cyan);
}

.target-warn {
  color: #fbbf24;
}

.vm-launch-btn {
  width: 100%;
  padding: 0.75rem;
  font-size: 0.9rem;
  font-weight: 700;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: linear-gradient(135deg, #00f0ff 0%, #0077ff 100%);
  color: #070a12;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.45),
    0 4px 16px rgba(0, 229, 255, 0.35);
}

.vm-launch-btn:hover:not(:disabled) {
  transform: translateY(-1.5px);
  background: linear-gradient(135deg, #38f9ff 0%, #1a8cff 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.6),
    0 6px 22px rgba(0, 229, 255, 0.5);
}

.vm-launch-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  pointer-events: none;
  box-shadow: none !important;
  transform: none !important;
}

[data-theme="light"] .vm-launch-btn {
  background: linear-gradient(135deg, #0396e6 0%, #0284c7 45%, #2563eb 100%);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.35);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.45),
    inset 0 -1px 0 rgba(0, 0, 0, 0.12),
    0 4px 16px rgba(2, 132, 199, 0.35);
}

[data-theme="light"] .vm-launch-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #38bdf8 0%, #0284c7 45%, #1d4ed8 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.6),
    0 6px 20px rgba(2, 132, 199, 0.45);
}

.btn-secondary {
  width: 100%;
  padding: 0.65rem;
  border-radius: 8px;
  border: 1px solid var(--btn-sec-border);
  background: var(--btn-sec-bg);
  color: var(--btn-sec-text);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
}

.btn-secondary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-secondary:hover:not(:disabled) {
  background: var(--btn-sec-hover-bg);
  border-color: var(--btn-sec-hover-border);
  color: var(--btn-sec-hover-text);
}

/* Light Theme Overrides for DeployPanel */
[data-theme="light"] .section-card {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 8px 32px rgba(15, 23, 42, 0.06);
}

[data-theme="light"] .section-card h2 {
  color: #0f172a;
}

[data-theme="light"] .section-desc {
  color: #64748b;
}

[data-theme="light"] .fs-label {
  color: #0f172a;
}

[data-theme="light"] .safe-mode-notice {
  background: #f0f9ff;
  border-color: #7dd3fc;
}

[data-theme="light"] .safe-notice-title {
  color: #0369a1;
}

[data-theme="light"] .safe-notice-desc {
  color: #0c4a6e;
}

[data-theme="light"] .safe-notice-desc b {
  color: #0284c7;
}

[data-theme="light"] .ventoy-warning-card {
  background: #fef2f2;
  border-color: #fca5a5;
}

[data-theme="light"] .warning-card-title {
  color: #dc2626;
}

[data-theme="light"] .warning-card-message {
  color: #991b1b;
}

[data-theme="light"] .deploy-box {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
}

[data-theme="light"] .selected-target {
  color: #334155;
}

[data-theme="light"] .iso-card {
  background: #f8fafc;
  border-color: #cbd5e1;
}

[data-theme="light"] .iso-card.is-drag-over {
  border-color: #0284c7 !important;
  box-shadow: 0 0 24px rgba(2, 132, 199, 0.3) !important;
}

[data-theme="light"] .iso-drag-overlay {
  background: rgba(248, 250, 252, 0.9);
  border-color: #0284c7;
}

[data-theme="light"] .iso-drag-overlay .drag-text {
  color: #0284c7;
}

[data-theme="light"] .iso-title-group h3 {
  color: #0f172a;
}

[data-theme="light"] .iso-list-container {
  background: #ffffff;
  border-color: #cbd5e1;
}

[data-theme="light"] .iso-empty-state {
  border-color: #cbd5e1;
}

[data-theme="light"] .iso-empty-state.drag-active {
  border-color: #0284c7 !important;
  background: rgba(2, 132, 199, 0.08) !important;
  box-shadow: 0 0 20px rgba(2, 132, 199, 0.22);
}

[data-theme="light"] .iso-empty-state.drag-active .empty-text {
  color: #0284c7;
}

[data-theme="light"] .empty-text {
  color: #1e293b;
}

[data-theme="light"] .iso-file-item {
  background: #f8fafc;
  border-color: #cbd5e1;
}

[data-theme="light"] .iso-file-name {
  color: #0f172a;
}

[data-theme="light"] .iso-file-path {
  color: #475569;
}

[data-theme="light"] .iso-footer {
  border-top-color: #cbd5e1;
}

[data-theme="light"] .btn-toggle-checksum {
  border-color: #cbd5e1;
  background: #f1f5f9;
  color: #1e293b;
}

[data-theme="light"] .btn-toggle-checksum:hover {
  background: #e2e8f0;
  border-color: #0284c7;
  color: #0284c7;
}

[data-theme="light"] .btn-toggle-checksum.active {
  background: rgba(2, 132, 199, 0.1);
  border-color: #0284c7;
  color: #0284c7;
}

[data-theme="light"] .vm-box {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.04);
}

[data-theme="light"] .vm-header h3 {
  color: #0f172a;
}

[data-theme="light"] .badge.success {
  background: #dcfce7;
  color: #15803d;
}

[data-theme="light"] .badge.muted {
  background: #f1f5f9;
  color: #64748b;
}

[data-theme="light"] .optional-badge {
  background: #e0f2fe;
  border-color: #bae6fd;
  color: #0284c7;
}

[data-theme="light"] .vm-select {
  background: #ffffff;
  color: #0f172a;
  border-color: #cbd5e1;
}

[data-theme="light"] .vm-desc {
  color: #64748b;
}

[data-theme="light"] .btn-secondary {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.05);
}

[data-theme="light"] .btn-secondary:hover:not(:disabled) {
  background: #f0f9ff;
  border-color: #38bdf8;
  color: #0284c7;
}

[data-theme="light"] .target-highlight {
  color: #0284c7;
}

[data-theme="light"] .target-warn {
  color: #d97706;
}

/* Light theme overrides for Checksum & Batch Checksum */
[data-theme="light"] .checksum-subtabs-nav {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

[data-theme="light"] .checksum-subtabs-nav .sub-tab-btn {
  color: #475569;
}

[data-theme="light"] .checksum-subtabs-nav .sub-tab-btn:hover {
  color: #0f172a;
}

[data-theme="light"] .checksum-subtabs-nav .sub-tab-btn.active {
  background: #ffffff;
  color: #0284c7;
  border-color: #0284c7;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

[data-theme="light"] .tab-badge {
  background: #e0f2fe;
  color: #0284c7;
}

[data-theme="light"] .hash-empty-hint {
  background: #f8fafc;
  color: #64748b;
  border-color: #cbd5e1;
}

[data-theme="light"] .btn-clear-iso {
  background: #fef2f2;
  border-color: #fecaca;
  color: #dc2626;
}

[data-theme="light"] .btn-clear-iso:hover {
  background: #fee2e2;
  border-color: #ef4444;
  color: #b91c1c;
}

[data-theme="light"] .iso-status-badge.calculating {
  background: #e0f2fe;
  color: #0369a1;
  border-color: #bae6fd;
}

[data-theme="light"] .iso-status-badge.match {
  background: #dcfce7;
  color: #15803d;
  border-color: #86efac;
}

[data-theme="light"] .iso-status-badge.mismatch {
  background: #fee2e2;
  color: #b91c1c;
  border-color: #fca5a5;
}

[data-theme="light"] .iso-status-badge.no-expected {
  background: #fef3c7;
  color: #b45309;
  border-color: #fde68a;
}

[data-theme="light"] .iso-status-badge.idle {
  background: #f1f5f9;
  color: #475569;
  border-color: #cbd5e1;
}

[data-theme="light"] .batch-calc-btn {
  background: #f0f9ff !important;
  color: #0284c7 !important;
  border-color: #0284c7 !important;
}

[data-theme="light"] .batch-calc-btn:hover:not(:disabled) {
  background: #e0f2fe !important;
}

[data-theme="light"] .batch-summary-banner {
  background: #f0f9ff;
  border-color: #bae6fd;
  color: #0369a1;
}

[data-theme="light"] .batch-summary-banner.all-matched {
  background: #f0fdf4;
  border-color: #86efac;
  color: #15803d;
}

.deploy-active-container {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  margin-bottom: 0.85rem;
}

.batch-deploy-info {
  background: var(--subtab-container-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.6rem 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.batch-current-disk {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--primary);
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.batch-disk-name {
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-family: "SF Mono", Monaco, "Courier New", monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.batch-current-stage {
  font-size: 0.78rem;
  color: var(--text-main);
  font-weight: 500;
  margin-top: 0.2rem;
  font-style: italic;
}

.deploy-stats-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
  font-size: 0.78rem;
  font-weight: 600;
}

.stat-badge {
  background: var(--subtab-container-bg);
  color: var(--text-main);
  padding: 0.25rem 0.55rem;
  border-radius: 6px;
  border: 1px solid var(--card-border);
}

.btn-cancel-deploy {
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  color: var(--alert-danger-title);
  font-size: 0.825rem;
  font-weight: 700;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  width: 100%;
}

.btn-cancel-deploy:hover {
  background: var(--badge-danger-bg);
  border-color: var(--danger);
  color: var(--danger);
}

.vm-tuning-row {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  margin-top: 0.65rem;
  margin-bottom: 0.75rem;
  flex-wrap: wrap;
}

.vm-tuning-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.vm-select-wrapper.sm {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.vm-select.sm {
  padding: 0.25rem 1.4rem 0.25rem 0.55rem;
  font-size: 0.78rem;
  border-radius: 6px;
}

.vm-checkbox-label {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-main);
  cursor: pointer;
  user-select: none;
}

.vm-checkbox-label input[type="checkbox"] {
  accent-color: var(--accent-cyan);
  width: 14px;
  height: 14px;
  cursor: pointer;
}
</style>
