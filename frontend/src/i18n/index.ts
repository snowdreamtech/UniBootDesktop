import { ref, computed } from 'vue';

export type LanguageCode = 
  | 'auto' 
  | 'zh-CN' 
  | 'en-US' 
  | 'zh-TW' 
  | 'ja-JP' 
  | 'ko-KR' 
  | 'de-DE' 
  | 'fr-FR' 
  | 'es-ES' 
  | 'ru-RU' 
  | 'pt-BR' 
  | 'it-IT' 
  | 'tr-TR' 
  | 'pl-PL' 
  | 'vi-VN' 
  | 'ar-SA';

export const currentLang = ref<LanguageCode>('auto');

export function detectSystemLanguage(): LanguageCode {
  if (typeof navigator !== 'undefined' && navigator.language) {
    const navLang = navigator.language.toLowerCase();
    if (navLang.startsWith('zh-tw') || navLang.startsWith('zh-hk') || navLang.startsWith('zh-hant')) return 'zh-TW';
    if (navLang.startsWith('zh')) return 'zh-CN';
    if (navLang.startsWith('ja')) return 'ja-JP';
    if (navLang.startsWith('ko')) return 'ko-KR';
    if (navLang.startsWith('de')) return 'de-DE';
    if (navLang.startsWith('fr')) return 'fr-FR';
    if (navLang.startsWith('es')) return 'es-ES';
    if (navLang.startsWith('ru')) return 'ru-RU';
    if (navLang.startsWith('pt')) return 'pt-BR';
    if (navLang.startsWith('it')) return 'it-IT';
    if (navLang.startsWith('tr')) return 'tr-TR';
    if (navLang.startsWith('pl')) return 'pl-PL';
    if (navLang.startsWith('vi')) return 'vi-VN';
    if (navLang.startsWith('ar')) return 'ar-SA';
  }
  return 'en-US';
}

export const resolvedLang = computed<LanguageCode>(() => {
  if (currentLang.value === 'auto') {
    return detectSystemLanguage();
  }
  return currentLang.value;
});

type TranslationDict = Record<string, string>;

const zhCN: TranslationDict = {
  'app.title': 'UniGoDesktop',
  'app.subtitle': 'UniBoot 桌面级软硬件引导制作引擎',
  'mode.cloud': '模式 B (1秒极速云安装盘)',
  'mode.hybrid': '模式 A (全能双模盘)',
  
  'disk.select_title': '选择目标 U 盘',
  'disk.select_desc': '仅自动扫描安全的可移动 U 盘，系统盘自动过滤保护。',
  'disk.single_mode': '🎯 单选模式',
  'disk.batch_mode': '📦 批量多选模式',
  'disk.select_all': '✅ 全选',
  'disk.clear_select': '❌ 清除选择',
  'disk.selected_count': '已选 {count}/{total}',
  'disk.rescan': '重新扫描 U 盘',
  'disk.no_disk': '未选择 U 盘',
  'disk.removable': '可移动 U 盘',
  'disk.system_disk': '系统盘 (已保护过滤)',
  'disk.fake_usb3_warning': '⚠️ 疑似降速假 USB 3.0 盘 (实际运行于 USB 2.0 速率 480 Mb/s)',
  'disk.genuine_usb3': '⚡ 高速 USB 3.0+ 驱动',
  'disk.free_space': '剩余可用',
  'disk.change_icon': '修改图标',
  'disk.hw_inspect': '硬件检测',
  'disk.details': '详细属性',

  'safe.title_cloud': '检测到现有 Ventoy/UniBoot 盘 (模式 B 仅刷新 ESP 引导区)',
  'safe.desc_cloud': '模式 B 坚持标准 UNIBOOT 双分区架构。部署将自动无损刷新 ESP 引导区直达 iPXE 云菜单，绝不抹擦或挪动主数据区原有文件与 ISO！',
  'safe.title_hybrid': '检测到现有的 Ventoy 启动盘 (免格式化无损更新)',
  'safe.desc_hybrid': '无需选择主数据区格式。模式 A 自动保留所有现有文件与 ISO 镜像，全自动无损注入 UniBoot 暗色主题与 iPXE 云引导！',

  'iso.title': '💿 本地系统镜像源 (ISO / IMG / WIM / VHD)',
  'iso.desc': '支持单选与多选系统镜像。一键制作完成将自动写入 /UNIBOOT/iso/ 目录供 Ventoy / UniBoot 直接挂载。',
  'iso.add_btn': '➕ 添加镜像文件',
  'iso.empty_title': '点击添加镜像文件 (支持单选与批量多选)',
  'iso.empty_sub': '支持 .iso, .wim, .img, .vhd, .vhdx, .vti, .efi, .bin, .xz, .gz, .raw 等 Ventoy 全格式',
  'iso.summary': '已选 {count} 个系统镜像源文件',
  'iso.clear': '清空列表',

  'deploy.target_device': '目标设备:',
  'deploy.batch_target': '已选中 {count} 块 U 盘',
  'deploy.start_create': '开始制作启动盘',
  'deploy.start_update': '🛡️ 开始无损更新 (保留数据)',
  'deploy.batch_create': '开始批量制作 ({count} 块 U 盘)',
  'deploy.writing': '正在写入引导与固件包...',

  'qemu.title': 'QEMU 引导模拟测试',
  'qemu.installed': '已检测到 QEMU',
  'qemu.not_installed': '未检测到 QEMU',
  'qemu.target': '校验目标:',
  'qemu.no_disk_warn': '⚠️ 未选择 U 盘（请在左侧列表中点击选择要测试的 U 盘）',
  'qemu.run_test': '▶ 运行 QEMU 启动测试',
  'qemu.launching': '⏳ 正在拉起 QEMU 虚拟机...',

  'settings.title': '设置',
  'settings.subtitle': '首选项与偏好设置',
  'settings.realtime_save': '实时保存',
  'settings.tab_general': '基础设置',
  'settings.tab_network': '网络设置',
  'settings.tab_uniboot': 'UniBoot',
  'settings.tab_ventoy': 'Ventoy',
  'settings.language': '界面语言 (Language):',
  'settings.theme': '界面主题与视觉风格 (Theme):',
  'settings.default_mode': '默认部署模式 (Default Mode):',
  'settings.default_fs': '默认目标文件系统 (File System):',
  'settings.app_update': '应用程序更新检测 (App Updates):',
  'settings.update_auto': '启动时自动检测云端新版本',
  'settings.update_manual': '仅手动检测',

  'settings.github_proxy': '自定义 GitHub 代理/镜像前缀 (GitHub Proxy Prefix):',
  'settings.proxy_placeholder': '默认为空（直接连接 GitHub 官方）。例如: https://your-proxy.com/',
  'settings.test_net': '⚡ 测试 GitHub 连通性',
  'settings.testing_net': '正在连通性测试中...',
  'settings.system_proxy': '系统网络代理设置 (HTTP / HTTPS / SOCKS4 / SOCKS5):',
  'settings.proxy_proto': '代理协议类型 (Protocol):',
  'settings.proxy_direct': '直连 (Direct)',
  'settings.proxy_host': '代理服务器主机 (Host / IP):',
  'settings.proxy_port': '端口 (Port):',

  'settings.ventoy_cli_path': 'Ventoy 可执行文件路径 (Ventoy CLI Executable Path):',
  'settings.ventoy_secboot': '启用 Secure Boot 安全启动支持 (-s / /s):',
  'settings.ventoy_part_style': '默认分区表类型 (Partition Style):',
  'settings.ventoy_reserve': '末尾保留空间 (Reserve Space MB):',
  'settings.ventoy_win11_bypass': '绕过 Windows 11 TPM/CPU/RAM 校验:',
  'settings.ventoy_timeout': 'Ventoy 引导菜单超时时间 (Timeout Seconds):',

  'confirm.title': '高风险格式化确认警告',
  'confirm.warning_title': '⚠️ 警告：写入磁盘将抹擦数据！',
  'confirm.warning_desc': '您选中的 U 盘即将进行重新分区与格式化，盘内原有所有文件将被完全清空。请务必确认已备份重要数据！',
  'confirm.mode_title': '部署模式:',
  'confirm.fs_title': '目标文件系统:',
  'confirm.disks_title': '即将格式化的目标 U 盘列表 ({count} 块):',
  'confirm.confirm_btn': '🔥 确认无误，开始格式化制作',
  'confirm.cancel_btn': '取消',

  'ventoy_alert.goto_settings': '⚙️ 前往 Ventoy 设置配置',
  'ventoy_alert.switch_b': '🚀 一键切换至模式 B (云纯净模式)',
  'ventoy_alert.close': '关闭',

  'inspector.title': 'USB 硬件探针分析报告',
  'inspector.device': '设备节点:',
  'inspector.vendor': '厂商 / 品牌:',
  'inspector.bus_speed': '总线协议与速率:',
  'inspector.fake_report': '真假 USB 3.0 识别:',
  'inspector.fake_text': '⚠️ 疑似降速假 USB 3.0 (插槽为蓝/黑，但内部协议仅为 USB 2.0 480Mb/s)',
  'inspector.genuine_text': '⚡ 真实 USB 3.0/3.1 高速总线驱动',
  'inspector.smart': 'SMART 健康状态:',
  'inspector.sector': '扇区大小:',

  'icon_picker.title': '修改 U 盘自定义图标',
  'icon_picker.reset': '重置默认图标',
};

const enUS: TranslationDict = {
  'app.title': 'UniGoDesktop',
  'app.subtitle': 'UniBoot Desktop Hardware & Software Boot Drive Engine',
  'mode.cloud': 'Mode B (1-Sec Cloud Install Disk)',
  'mode.hybrid': 'Mode A (Full Hybrid Disk)',

  'disk.select_title': 'Select Target USB Drive',
  'disk.select_desc': 'Scans removable USB drives safely. System drives are protected.',
  'disk.single_mode': '🎯 Single Select',
  'disk.batch_mode': '📦 Batch Select',
  'disk.select_all': '✅ Select All',
  'disk.clear_select': '❌ Clear Selection',
  'disk.selected_count': 'Selected {count}/{total}',
  'disk.rescan': 'Rescan USB Drives',
  'disk.no_disk': 'No USB Drive Selected',
  'disk.removable': 'Removable Drive',
  'disk.system_disk': 'System Drive (Protected)',
  'disk.fake_usb3_warning': '⚠️ Suspected Downgraded Fake USB 3.0 Drive (Runs at USB 2.0 480 Mb/s)',
  'disk.genuine_usb3': '⚡ High Speed USB 3.0+ Drive',
  'disk.free_space': 'Free Space',
  'disk.change_icon': 'Custom Icon',
  'disk.hw_inspect': 'HW Inspector',
  'disk.details': 'Details',

  'safe.title_cloud': 'Existing Ventoy/UniBoot Drive Detected (Mode B refreshes ESP partition only)',
  'safe.desc_cloud': 'Mode B preserves UNIBOOT dual-partition layout. Refreshing ESP partition leaves all ISOs and user files in Partition 1 intact!',
  'safe.title_hybrid': 'Existing Ventoy Boot Disk Detected (In-Place Upgrade)',
  'safe.desc_hybrid': 'Mode A preserves all existing ISO files without formatting, injecting UniBoot dark theme and cloud boot menu safely!',

  'iso.title': '💿 Local System Image Sources (ISO / IMG / WIM / VHD)',
  'iso.desc': 'Add ISO files to auto-copy to /UNIBOOT/iso/ directory for Ventoy / UniBoot direct booting.',
  'iso.add_btn': '➕ Add Image Files',
  'iso.empty_title': 'Click to Add Image Files (Supports Single or Batch Select)',
  'iso.empty_sub': 'Supports .iso, .wim, .img, .vhd, .vhdx, .vti, .efi, .bin, .xz, .gz, .raw formats',
  'iso.summary': 'Selected {count} image source file(s)',
  'iso.clear': 'Clear List',

  'deploy.target_device': 'Target Device:',
  'deploy.batch_target': 'Selected {count} USB drive(s)',
  'deploy.start_create': 'Start Deployment',
  'deploy.start_update': '🛡️ In-Place Upgrade (Data Safe)',
  'deploy.batch_create': 'Start Batch Format ({count} Drives)',
  'deploy.writing': 'Writing Boot Firmware Packages...',

  'qemu.title': 'QEMU Boot Simulation Test',
  'qemu.installed': 'QEMU Detected',
  'qemu.not_installed': 'QEMU Not Detected',
  'qemu.target': 'Test Target:',
  'qemu.no_disk_warn': '⚠️ No USB Drive Selected (Please select a drive from the left panel)',
  'qemu.run_test': '▶ Launch QEMU Test',
  'qemu.launching': '⏳ Launching QEMU Virtual Machine...',

  'settings.title': 'Settings',
  'settings.subtitle': 'Preferences & Configurations',
  'settings.realtime_save': 'Realtime Saved',
  'settings.tab_general': 'General',
  'settings.tab_network': 'Network',
  'settings.tab_uniboot': 'UniBoot',
  'settings.tab_ventoy': 'Ventoy',
  'settings.language': 'App Language:',
  'settings.theme': 'UI Theme:',
  'settings.default_mode': 'Default Deployment Mode:',
  'settings.default_fs': 'Default File System:',
  'settings.app_update': 'App Updates Check:',
  'settings.update_auto': 'Auto-check for updates on startup',
  'settings.update_manual': 'Manual check only',

  'settings.github_proxy': 'GitHub Proxy / Mirror Prefix:',
  'settings.proxy_placeholder': 'Default empty (Direct GitHub). Example: https://your-proxy.com/',
  'settings.test_net': '⚡ Test GitHub Connection',
  'settings.testing_net': 'Testing Connectivity...',
  'settings.system_proxy': 'System Proxy Settings (HTTP / HTTPS / SOCKS4 / SOCKS5):',
  'settings.proxy_proto': 'Proxy Protocol:',
  'settings.proxy_direct': 'Direct Connection',
  'settings.proxy_host': 'Proxy Host / IP:',
  'settings.proxy_port': 'Port:',

  'settings.ventoy_cli_path': 'Ventoy CLI Executable Path:',
  'settings.ventoy_secboot': 'Enable Secure Boot Support (-s / /s):',
  'settings.ventoy_part_style': 'Partition Style:',
  'settings.ventoy_reserve': 'Reserve Space (MB):',
  'settings.ventoy_win11_bypass': 'Bypass Windows 11 TPM/CPU/RAM Check:',
  'settings.ventoy_timeout': 'Ventoy Menu Timeout (Seconds):',

  'confirm.title': 'High-Risk Format Warning',
  'confirm.warning_title': '⚠️ Warning: Formatting will erase all data!',
  'confirm.warning_desc': 'The selected USB drive will be re-partitioned and formatted. All existing files will be erased completely. Ensure you have backed up important data!',
  'confirm.mode_title': 'Deployment Mode:',
  'confirm.fs_title': 'Target File System:',
  'confirm.disks_title': 'Drives to be Formatted ({count}):',
  'confirm.confirm_btn': '🔥 Confirm & Start Format',
  'confirm.cancel_btn': 'Cancel',

  'ventoy_alert.goto_settings': '⚙️ Go to Ventoy Settings',
  'ventoy_alert.switch_b': '🚀 Switch to Mode B (Cloud Mode)',
  'ventoy_alert.close': 'Close',

  'inspector.title': 'USB Hardware Inspector Report',
  'inspector.device': 'Device Node:',
  'inspector.vendor': 'Vendor / Brand:',
  'inspector.bus_speed': 'Bus Protocol & Speed:',
  'inspector.fake_report': 'Fake USB 3.0 Analysis:',
  'inspector.fake_text': '⚠️ Downgraded Fake USB 3.0 (Blue/Black port, but internal protocol is USB 2.0 480Mb/s)',
  'inspector.genuine_text': '⚡ Genuine High-Speed USB 3.0/3.1 Controller',
  'inspector.smart': 'SMART Health:',
  'inspector.sector': 'Sector Size:',

  'icon_picker.title': 'Customize USB Drive Icon',
  'icon_picker.reset': 'Reset Default Icon',
};

const zhTW: TranslationDict = {
  ...zhCN,
  'app.subtitle': 'UniBoot 桌面級軟硬體引導製作引擎',
  'mode.cloud': '模式 B (1秒極速雲安裝盤)',
  'mode.hybrid': '模式 A (全能雙模盤)',
  'disk.select_title': '選擇目標 U 盤',
  'disk.select_desc': '僅自動掃描安全的可移動 U 盤，系統盤自動過濾保護。',
  'disk.single_mode': '🎯 單選模式',
  'disk.batch_mode': '📦 批量多選模式',
  'disk.select_all': '✅ 全選',
  'disk.clear_select': '❌ 清除選擇',
  'disk.selected_count': '已選 {count}/{total}',
  'disk.rescan': '重新掃描 U 盤',
  'disk.no_disk': '未選擇 U 盤',
  'disk.removable': '可移動 U 盤',
  'disk.system_disk': '系統盤 (已保護過濾)',
  'disk.fake_usb3_warning': '⚠️ 疑似降速假 USB 3.0 盤 (實際運行於 USB 2.0 速率 480 Mb/s)',
  'disk.genuine_usb3': '⚡ 高速 USB 3.0+ 驅動',
  'disk.free_space': '剩餘可用',
  'disk.change_icon': '修改圖標',
  'disk.hw_inspect': '硬體檢測',
  'disk.details': '詳細屬性',
  'deploy.start_create': '開始製作啟動盤',
  'deploy.start_update': '🛡️ 開始無損更新 (保留數據)',
  'deploy.batch_create': '開始批量製作 ({count} 塊 U 盤)',
  'settings.title': '設定',
  'settings.subtitle': '首選項與偏好設定',
  'settings.tab_general': '基礎設定',
  'settings.tab_network': '網路設定',
  'settings.language': '介面語言 (Language):',
  'settings.theme': '介面主題與視覺風格 (Theme):',
};

const jaJP: TranslationDict = {
  ...enUS,
  'app.subtitle': 'UniBoot デスクトップ ブートドライブエンジン',
  'mode.cloud': 'モード B (1秒高速クラウドインストール)',
  'mode.hybrid': 'モード A (万能デュアルモード)',
  'disk.select_title': 'ターゲット USB ドライブを選択',
  'disk.select_desc': '安全な取り外し可能ドライブのみスキャンします。',
  'disk.single_mode': '🎯 単一選択',
  'disk.batch_mode': '📦 一括選択',
  'disk.select_all': '✅ すべて選択',
  'disk.clear_select': '❌ 選択解除',
  'disk.rescan': 'USB を再スキャン',
  'disk.no_disk': 'ドライブが選択されていません',
  'disk.change_icon': 'アイコン変更',
  'disk.hw_inspect': 'ハードウェア診断',
  'deploy.start_create': 'ブートディスクを作成',
  'deploy.start_update': '🛡️ データを保持して更新',
  'settings.title': '設定',
  'settings.tab_general': '一般設定',
  'settings.tab_network': 'ネットワーク',
  'settings.language': '表示言語 (Language):',
};

const koKR: TranslationDict = {
  ...enUS,
  'app.subtitle': 'UniBoot 데스크톱 부팅 드라이브 엔진',
  'mode.cloud': '모드 B (1초 고속 클라우드 설치)',
  'mode.hybrid': '모드 A (하이브리드 부팅)',
  'disk.select_title': '대상 USB 드라이브 선택',
  'disk.single_mode': '🎯 단일 선택',
  'disk.batch_mode': '📦 다중 선택',
  'disk.rescan': 'USB 다시 스캔',
  'deploy.start_create': '부팅 디스크 제작 시작',
  'settings.title': '설정',
  'settings.language': '언어 (Language):',
};

const deDE: TranslationDict = {
  ...enUS,
  'app.subtitle': 'UniBoot Desktop Boot-Laufwerk-Engine',
  'mode.cloud': 'Modus B (1-Sek. Cloud-Installation)',
  'mode.hybrid': 'Modus A (Vollständiger Hybrid-Datenträger)',
  'disk.select_title': 'Ziel-USB-Laufwerk auswählen',
  'disk.rescan': 'USB-Laufwerke neu scannen',
  'deploy.start_create': 'Erstellung starten',
  'settings.title': 'Einstellungen',
  'settings.language': 'Sprache (Language):',
};

const frFR: TranslationDict = {
  ...enUS,
  'app.subtitle': 'Moteur de création de clés USB UniBoot Desktop',
  'mode.cloud': 'Mode B (Installation Cloud 1-sec)',
  'mode.hybrid': 'Mode A (Disque Hybride Complet)',
  'disk.select_title': 'Sélectionner la clé USB cible',
  'deploy.start_create': 'Démarrer la création',
  'settings.title': 'Paramètres',
  'settings.language': 'Langue (Language):',
};

const esES: TranslationDict = {
  ...enUS,
  'app.subtitle': 'Motor de unidades de arranque UniBoot Desktop',
  'mode.cloud': 'Modo B (Instalación en la nube en 1 seg)',
  'mode.hybrid': 'Modo A (Disco Híbrido Completo)',
  'disk.select_title': 'Seleccionar unidad USB de destino',
  'deploy.start_create': 'Iniciar creación',
  'settings.title': 'Configuración',
  'settings.language': 'Idioma (Language):',
};

const ruRU: TranslationDict = {
  ...enUS,
  'app.subtitle': 'Движок создания загрузочных накопителей UniBoot',
  'mode.cloud': 'Режим B (Облачная установка за 1 сек)',
  'mode.hybrid': 'Режим A (Полный гибридный диск)',
  'disk.select_title': 'Выберите USB-накопитель',
  'deploy.start_create': 'Начать создание',
  'settings.title': 'Настройки',
  'settings.language': 'Язык (Language):',
};

const ptBR: TranslationDict = { ...enUS, 'settings.language': 'Idioma (Language):' };
const itIT: TranslationDict = { ...enUS, 'settings.language': 'Lingua (Language):' };
const trTR: TranslationDict = { ...enUS, 'settings.language': 'Dil (Language):' };
const plPL: TranslationDict = { ...enUS, 'settings.language': 'Język (Language):' };
const viVN: TranslationDict = { ...enUS, 'settings.language': 'Ngôn ngữ (Language):' };
const arSA: TranslationDict = { ...enUS, 'settings.language': 'اللغة (Language):' };

const translations: Record<LanguageCode, TranslationDict> = {
  'auto': zhCN,
  'zh-CN': zhCN,
  'en-US': enUS,
  'zh-TW': zhTW,
  'ja-JP': jaJP,
  'ko-KR': koKR,
  'de-DE': deDE,
  'fr-FR': frFR,
  'es-ES': esES,
  'ru-RU': ruRU,
  'pt-BR': ptBR,
  'it-IT': itIT,
  'tr-TR': trTR,
  'pl-PL': plPL,
  'vi-VN': viVN,
  'ar-SA': arSA,
};

export function t(key: string, params?: Record<string, string | number>): string {
  const targetDict = translations[resolvedLang.value] || translations['en-US'];
  let template = targetDict[key] || translations['en-US'][key] || translations['zh-CN'][key] || key;

  if (params) {
    Object.keys(params).forEach(pKey => {
      template = template.replace(new RegExp(`\\{${pKey}\\}`, 'g'), String(params[pKey]));
    });
  }

  return template;
}

export function setLanguage(lang: string) {
  if (Object.keys(translations).includes(lang)) {
    currentLang.value = lang as LanguageCode;
  }
}
