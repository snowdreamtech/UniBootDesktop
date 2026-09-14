import { ref, computed } from 'vue';

export type LanguageCode = 'auto' | 'zh-CN' | 'en-US' | 'zh-TW';

export const currentLang = ref<LanguageCode>('auto');

export function detectSystemLanguage(): 'zh-CN' | 'en-US' | 'zh-TW' {
  if (typeof navigator !== 'undefined' && navigator.language) {
    const navLang = navigator.language.toLowerCase();
    if (navLang.startsWith('zh-tw') || navLang.startsWith('zh-hk') || navLang.startsWith('zh-hant')) {
      return 'zh-TW';
    }
    if (navLang.startsWith('zh')) {
      return 'zh-CN';
    }
  }
  return 'en-US';
}

export const resolvedLang = computed<'zh-CN' | 'en-US' | 'zh-TW'>(() => {
  if (currentLang.value === 'auto') {
    return detectSystemLanguage();
  }
  return currentLang.value;
});

const translations: Record<'zh-CN' | 'en-US' | 'zh-TW', Record<string, string>> = {
  'zh-CN': {
    'app.title': 'UniGoDesktop',
    'app.subtitle': 'UniBoot 桌面级软硬件引导制作引擎',
    'mode.cloud': '模式 B (1秒极速云安装盘)',
    'mode.hybrid': '模式 A (全能双模盘)',
    'disk.select_title': '选择目标 U 盘',
    'disk.select_desc': '仅自动扫描安全的可移动 U 盘，系统盘自动过滤保护。',
    'disk.single_mode': '单选模式',
    'disk.batch_mode': '批量多选模式',
    'disk.rescan': '重新扫描 U 盘',
    'disk.no_disk': '未选择 U 盘',
    'deploy.title': '启动盘制作与模拟测试',
    'deploy.start_create': '开始制作启动盘',
    'deploy.start_update': '🛡️ 开始无损更新 (保留数据)',
    'deploy.batch_create': '开始批量制作',
    'qemu.title': 'QEMU 引导模拟测试',
    'qemu.run_test': '▶ 运行 QEMU 启动测试',
    'settings.title': '首选项与偏好设置',
    'settings.tab_general': '基础设置',
    'settings.tab_network': '网络设置',
    'settings.tab_uniboot': 'UniBoot',
    'settings.tab_ventoy': 'Ventoy',
    'settings.language': '界面语言 (Language):',
    'settings.lang_auto': '🌐 自动识别 (Auto - 匹配操作系统)',
    'settings.lang_zh_cn': '🇨🇳 简体中文 (Simplified Chinese)',
    'settings.lang_en_us': '🇺🇸 English (英文)',
    'settings.lang_zh_tw': '🇭🇰 繁體中文 (Traditional Chinese)',
    'settings.theme': '界面主题与视觉风格 (Theme):',
    'settings.theme_dark': '🌙 深色极客风 (Dark Cyber Glow)',
    'settings.theme_light': '☀️ 浅色明亮风 (Light Crisp)',
  },
  'en-US': {
    'app.title': 'UniGoDesktop',
    'app.subtitle': 'UniBoot Desktop Bootable Drive Engine',
    'mode.cloud': 'Mode B (1-Sec Cloud Install Disk)',
    'mode.hybrid': 'Mode A (Full Hybrid Disk)',
    'disk.select_title': 'Select Target USB Drive',
    'disk.select_desc': 'Only scans safe removable USB drives. System drives are protected.',
    'disk.single_mode': 'Single Disk',
    'disk.batch_mode': 'Batch Select',
    'disk.rescan': 'Rescan USB Drives',
    'disk.no_disk': 'No USB Drive Selected',
    'deploy.title': 'Deployment & QEMU Test',
    'deploy.start_create': 'Start Deployment',
    'deploy.start_update': '🛡️ In-Place Upgrade (Data Safe)',
    'deploy.batch_create': 'Start Batch Deployment',
    'qemu.title': 'QEMU Boot Simulation',
    'qemu.run_test': '▶ Launch QEMU Test',
    'settings.title': 'Preferences & Settings',
    'settings.tab_general': 'General',
    'settings.tab_network': 'Network',
    'settings.tab_uniboot': 'UniBoot',
    'settings.tab_ventoy': 'Ventoy',
    'settings.language': 'App Language:',
    'settings.lang_auto': '🌐 Auto Detect (Match OS)',
    'settings.lang_zh_cn': '🇨🇳 简体中文 (Simplified Chinese)',
    'settings.lang_en_us': '🇺🇸 English',
    'settings.lang_zh_tw': '🇭🇰 繁體中文 (Traditional Chinese)',
    'settings.theme': 'UI Theme:',
    'settings.theme_dark': '🌙 Dark Cyber Glow',
    'settings.theme_light': '☀️ Light Crisp',
  },
  'zh-TW': {
    'app.title': 'UniGoDesktop',
    'app.subtitle': 'UniBoot 桌面級軟硬體引導製作引擎',
    'mode.cloud': '模式 B (1秒極速雲安裝盤)',
    'mode.hybrid': '模式 A (全能雙模盤)',
    'disk.select_title': '選擇目標 U 盤',
    'disk.select_desc': '僅自動掃描安全的可移動 U 盤，系統盤自動過濾保護。',
    'disk.single_mode': '單選模式',
    'disk.batch_mode': '批量多選模式',
    'disk.rescan': '重新掃描 U 盤',
    'disk.no_disk': '未選擇 U 盤',
    'deploy.title': '啟動盤製作與模擬測試',
    'deploy.start_create': '開始製作啟動盤',
    'deploy.start_update': '🛡️ 開始無損更新 (保留數據)',
    'deploy.batch_create': '開始批量製作',
    'qemu.title': 'QEMU 引導模擬測試',
    'qemu.run_test': '▶ 運行 QEMU 啟動測試',
    'settings.title': '首選項與偏好設定',
    'settings.tab_general': '基礎設定',
    'settings.tab_network': '網路設定',
    'settings.tab_uniboot': 'UniBoot',
    'settings.tab_ventoy': 'Ventoy',
    'settings.language': '介面語言 (Language):',
    'settings.lang_auto': '🌐 自動識別 (Auto - 匹配作業系統)',
    'settings.lang_zh_cn': '🇨🇳 簡體中文 (Simplified Chinese)',
    'settings.lang_en_us': '🇺🇸 English (英文)',
    'settings.lang_zh_tw': '🇭🇰 繁體中文 (Traditional Chinese)',
    'settings.theme': '介面主題與視覺風格 (Theme):',
    'settings.theme_dark': '🌙 深色極客風 (Dark Cyber Glow)',
    'settings.theme_light': '☀️ 淺色明亮風 (Light Crisp)',
  }
};

export function t(key: string): string {
  const dict = translations[resolvedLang.value] || translations['zh-CN'];
  return dict[key] || translations['zh-CN'][key] || key;
}

export function setLanguage(lang: string) {
  if (['auto', 'zh-CN', 'en-US', 'zh-TW'].includes(lang)) {
    currentLang.value = lang as LanguageCode;
  }
}
