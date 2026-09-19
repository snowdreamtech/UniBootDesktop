// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref } from 'vue';
import { t } from '../i18n';
import { SelectIsoFiles } from '../../wailsjs/go/main/App';
import { logUserAction } from '../utils/logger';

export interface IsoFileItem {
  name: string;
  path: string;
}

export const SUPPORTED_IMAGE_EXTS = [
  '.iso', '.img', '.wim', '.vhd', '.vhdx', '.vti', '.efi', '.bin', '.xz', '.gz', '.raw'
];

export function useIsoManager(showToast: (msg: string, type: 'info' | 'warning' | 'error' | 'success') => void) {
  const selectedIsoFiles = ref<IsoFileItem[]>([]);
  const isoCopyStatus = ref<string>('');

  function isSupportedImage(filePath: string): boolean {
    const lower = (filePath || '').toLowerCase();
    return SUPPORTED_IMAGE_EXTS.some(ext => lower.endsWith(ext));
  }

  function addIsoFilesByPaths(paths: string[]): number {
    if (!paths || paths.length === 0) return 0;
    let added = 0;
    let invalidCount = 0;

    for (const p of paths) {
      if (!p) continue;
      if (!isSupportedImage(p)) {
        invalidCount++;
        continue;
      }
      const name = p.split(/[/\\]/).pop() || p;
      const existingIndex = selectedIsoFiles.value.findIndex(
        f => f.path === p || f.name === name
      );

      if (existingIndex >= 0) {
        // If the existing entry only has the filename, upgrade it to the full absolute path
        if ((p.includes('/') || p.includes('\\')) && !selectedIsoFiles.value[existingIndex].path.includes('/') && !selectedIsoFiles.value[existingIndex].path.includes('\\')) {
          selectedIsoFiles.value[existingIndex].path = p;
        }
        continue;
      }

      selectedIsoFiles.value.push({ name, path: p });
      added++;
      logUserAction('INFO', 'Added image source file', `${name} (${p})`);
    }

    if (added > 0) {
      showToast(t('deploy.toast_added_iso', { count: added }), 'success');
    } else if (invalidCount > 0 && selectedIsoFiles.value.length === 0) {
      showToast(t('iso.drag_unsupported'), 'warning');
    }

    return added;
  }

  async function handleSelectIsoFiles() {
    if (typeof SelectIsoFiles === 'function') {
      try {
        const paths: string[] = await SelectIsoFiles(
          t('dialog.selectIsoTitle'),
          t('dialog.ventoyFilter'),
          t('dialog.allFilesFilter')
        );
        if (paths && paths.length > 0) {
          addIsoFilesByPaths(paths);
        }
      } catch (err: any) {
        console.error('SelectIsoFiles error:', err);
      }
    } else {
      // Mock for browser demo
      const mockFiles = [
        { name: 'ubuntu-24.04-desktop-amd64.iso', path: '/Users/demo/Downloads/ubuntu-24.04-desktop-amd64.iso' },
        { name: 'Windows11_23H2_Chinese_Simplified_x64.iso', path: '/Users/demo/Downloads/Windows11_23H2_Chinese_Simplified_x64.iso' }
      ];
      for (const m of mockFiles) {
        if (!selectedIsoFiles.value.some(f => f.path === m.path)) {
          selectedIsoFiles.value.push(m);
        }
      }
      showToast(t('deploy.toast_added_demo_iso'), 'info');
    }
  }

  function removeIsoFile(index: number) {
    const item = selectedIsoFiles.value[index];
    const name = item ? (typeof item === 'string' ? item : item.name || item.path) : '';
    selectedIsoFiles.value.splice(index, 1);
    if (name) {
      logUserAction('INFO', 'User removed ISO source file from selection list', name);
    }
  }

  function clearIsoFiles() {
    selectedIsoFiles.value = [];
    logUserAction('INFO', 'User cleared all ISO source files from selection list');
  }

  return {
    selectedIsoFiles,
    isoCopyStatus,
    isSupportedImage,
    addIsoFilesByPaths,
    handleSelectIsoFiles,
    removeIsoFile,
    clearIsoFiles,
  };
}
