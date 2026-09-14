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

  'deploy.title': '启动盘制作与模拟测试',
  'deploy.desc_cloud': '纯净 iPXE 云引导 • 极速初始化双分区并写入多架构 iPXE 网络引导固件。',
  'deploy.desc_hybrid': '集成 Ventoy 核心 + UniBoot 专属暗色主题 & iPXE 网络扩展，支持放置数 GB 大 ISO 镜像。',
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
  'confirm.title_safe': '无损更新确认：无需格式化磁盘',
  'confirm.title_mixed': '⚡ 智能混合制作确认',
  'confirm.title_danger': '格式化确认：即将初始化目标磁盘',
  'confirm.safe_banner_title': '💡 无损增量更新提示：无需格式化',
  'confirm.safe_banner_desc': '检测到目标 U 盘已存在 Ventoy / UniBoot 引导结构。系统将自动采用增量注入技术，跳过擦盘与格式化，直接更新引导固件与主题。您 U 盘中的所有文件、ISO 镜像均 100% 原样保留！',
  'confirm.mixed_banner_title': '🛡️ 智能混合模式：Ventoy 盘无损更新，空白盘初始化',
  'confirm.mixed_banner_desc': '已选中 {ventoyCount} 块引导盘（自动免格式化/保留数据）与 {blankCount} 块普通 U 盘（初始化格式化）。系统将进行差异化精准处理！',
  'confirm.danger_banner_title': '💥 警告：格式化过程不可逆！',
  'confirm.danger_banner_desc': '写入将对目标设备进行底层重新分区与格式化，改写主引导记录 (MBR/GPT)。所选 U 盘上的全部现有数据与资料将被清空。',
  'confirm.summary_title': '即将写入引导的目标磁盘：',
  'confirm.smart_safe_tag': '🛡️ 智能免格式化',
  'confirm.ventoy_group_title': '🛡️ 无损增量更新盘（保留所有 ISO 镜像）：',
  'confirm.blank_group_title': '⚠️ 重新格式化写入盘：',
  'confirm.batch_summary_title': '已选中 {count} 块 U 盘并行写入：',
  'confirm.fs_format': '{fs} 格式',

  'ventoy_alert.goto_settings': '⚙️ 前往 Ventoy 设置配置',
  'ventoy_alert.switch_b': '🚀 一键切换至模式 B (云纯净模式)',
  'ventoy_alert.close': '关闭',
  'ventoy_alert.default_title': '无法制作模式 A 启动盘',
  'ventoy_alert.banner_title': '💡 前置条件缺失与引导提示',

  'inspector.title': 'USB 硬件探针分析报告',
  'inspector.subtitle': '物理 PHY 协议识别、底层传输速率与硬件数据',
  'inspector.device': '设备节点:',
  'inspector.vendor': '厂商 / 品牌:',
  'inspector.bus_speed': '总线协议与速率:',
  'inspector.fake_report': '真假 USB 3.0 识别:',
  'inspector.fake_text': '⚠️ 疑似降速假 USB 3.0 (插槽为蓝/黑，但内部协议仅为 USB 2.0 480Mb/s)',
  'inspector.genuine_text': '⚡ 真实 USB 3.0/3.1 高速总线驱动',
  'inspector.smart': 'SMART 健康状态:',
  'inspector.sector': '扇区大小:',
  'inspector.fake_title': '伪造 USB 3.0 预警！(Fake USB 3.0 Alert)',
  'inspector.fake_desc': '设备宣传名称带有 USB 3.0 / 3.1 标识，但操作系统物理层实际协商速率仅为 {speed} (USB 2.0 High-Speed PHY)。此 U 盘疑似被刷固件或使用虚假蓝色接口。',
  'inspector.genuine_title': '物理硬件校验通过 (Genuine USB 3.0+ Device)',
  'inspector.genuine_desc': '硬件物理层已成功建立 SuperSpeed/SuperSpeed+ 高速通道，实测协商速率为 {speed}。',
  'inspector.usb2_title': '标准 USB 2.0 传输接口',
  'inspector.usb2_desc': '设备硬件版本为 USB 2.0，理论物理最高速率 480 Mb/s (High-Speed)。',
  'inspector.section_basic': '📊 设备基础信息 (Basic Device Info)',
  'inspector.section_hw': '🛠️ 设备硬件详情 (Hardware Details)',
  'inspector.lbl_name': '设备名称 (Device Name)',
  'inspector.lbl_vendor': '厂商/制造商 (Vendor)',
  'inspector.lbl_size': '设备容量 (Storage Size)',
  'inspector.lbl_free': '可用剩余空间 (Free Space)',
  'inspector.lbl_protocol': 'USB 协议版本 (Protocol Version)',
  'inspector.lbl_speed': '物理 PHY 速率 (Negotiated Speed)',
  'inspector.lbl_fs': '文件系统格式 (File System)',
  'inspector.lbl_scheme': '分区表架构 (Partition Scheme)',
  'inspector.lbl_mount': '系统挂载路径 (Mount Path)',
  'inspector.lbl_perm': '读写权限 (Disk Permission)',
  'inspector.val_rw': '✅ 可读写 (Read-Write)',
  'inspector.val_ro': '🔒 写保护/只读 (Read-Only)',
  'inspector.lbl_boot_status': '引导状态 (Boot Status)',
  'inspector.val_data_disk': '📁 数据存储盘 (未检出系统引导)',
  'inspector.lbl_smart': 'S.M.A.R.T. 健康状态 (SMART Status)',
  'inspector.val_smart_good': '✅ 健康 (Verified)',
  'inspector.lbl_bus_power': '接口总线供电 (USB Bus Power)',
  'inspector.val_power_limit': '(端口上限: {limit})',
  'inspector.lbl_transport': '底层传输协议 (Transport Protocol)',
  'inspector.lbl_sector': '扇区物理/逻辑大小 (Sector Size)',
  'inspector.lbl_controller': '主控芯片厂商 (Controller Vendor)',

  'icon_picker.title': '修改 U 盘自定义图标',
  'icon_picker.reset': '重置默认图标',
  'icon_picker.subtitle': '为目标设备 {name} 挑选满意的形象外观：',
  'icon_picker.usb_desc': '经典 SuperSpeed 5 Gbps 盘',
  'icon_picker.usb2_desc': '经典 High-Speed 基础盘',
  'icon_picker.usb3_1_desc': 'SuperSpeed+ 10 Gbps 高速盘',
  'icon_picker.usb3_2_desc': 'Gen 2x2 20 Gbps 极速双通道',
  'icon_picker.usb4_desc': '旗舰 40 Gbps 协议盘',
  'icon_picker.boot_desc': '带有闪电标志的引导盘',
  'icon_picker.ssd_desc': '高速拉丝铝盒固态盘',
  'icon_picker.typec_desc': '适配手机与 Mac 的 Type-C 盘',
  'icon_picker.secure_desc': '带物理密码锁的加密硬件',
  'icon_picker.reader_desc': '插入式多功能内存读卡器',
  'icon_picker.hdd_desc': '2.5 寸高容量机械移动盘',
  'icon_picker.key_desc': '物理密钥 FIDO2 安全盘',

  'disk.tag_boot': 'BOOT U盘',
  'disk.tag_ssd': '移动固态',
  'disk.tag_typec': 'Type-C 盘',
  'disk.tag_secure': '加密 U盘',
  'disk.tag_reader': '读卡器',
  'disk.tag_hdd': '移动硬盘',
  'disk.tag_key': '安全钥匙',
  'disk.tag_cdrom': '虚拟光驱',

  'inspector.std_controller': '通用 Standard Controller',
  'inspector.lbl_vid_pid': '硬件设备标识 (USB VID / PID)',
  'inspector.lbl_serial': '设备物理序列号 (Serial Number)',
  'inspector.ext_protocols': 'USB 协议支持库扩展 (Extensible Protocol Standards)',
  'inspector.close': '确定',

  'deploy.toast_switched_b': '已切换至原生支持的【模式 B (1秒极速云引导盘)】！',
  'deploy.toast_added_iso': '已成功添加 {count} 个镜像源文件',
  'deploy.toast_added_demo_iso': '已添加 2 个示例镜像源文件 (浏览器演示)',
  'deploy.toast_select_target': '⚠️ 请先在左侧磁盘列表中选择目标 U 盘',
  'deploy.toast_select_batch': '⚠️ 请先勾选要批量制作的目标 U 盘',
  'deploy.toast_no_disks': '⚠️ 当前未检测到任何可用的 U 盘设备！请插入 U 盘后再试。',

  'deploy.tip_writing': '正在写入引导固件...',
  'deploy.tip_select_single': '请先选择要制作的目标 U 盘',
  'deploy.tip_select_batch': '请先勾选要批量制作的目标 U 盘',
  'deploy.tip_macos_unsupported': '❌ macOS 平台暂不支持全新格式化制作 Mode A 盘 (请使用模式 B)',
  'deploy.tip_need_ventoy': '全新制作模式 A 需依赖 Ventoy CLI 环境',

  'deploy.macos_alert_title': 'macOS 暂不支持 Ventoy CLI 全新格式化',
  'deploy.macos_alert_desc': '官方 Ventoy 暂不支持在 macOS 上直接运行格式化程序。制作【模式 A】全新盘需依赖 Ventoy CLI；建议直接选择原生支持的【模式 B (1秒极速云引导盘)】！如需使用模式 A，请先在 Win/Linux 上完成 Ventoy 盘初始化后插入 macOS 无损升级。',
  'deploy.no_ventoy_title': '未检测到 Ventoy CLI 执行文件',
  'deploy.no_ventoy_desc': '全新制作【模式 A (Ventoy 双模盘)】需依赖本地 Ventoy CLI 程序 (Ventoy2Disk)。请先前往设置配置 Ventoy 可执行文件路径，或直接一键切换至不需要 Ventoy CLI 的【模式 B (1秒极速云引导)】！',

  'deploy.result_batch_success': '成功完成 {count} 块 U 盘的极速云安装盘部署！',
  'deploy.result_success': '成功部署模式 {mode} 到 {targets}',
  'deploy.alert_success': '🎉 部署成功！\n\n{msg}',
  'deploy.alert_fail': '❌ 部署失败：\n\n{msg}',

  'qemu.tip_launching': 'QEMU 模拟器正在拉起启动中...',
  'qemu.tip_deploying': '烧录部署中，请等待部署完成后再测试',
  'qemu.tip_not_installed': '未检测到 QEMU 模拟器，请先安装 QEMU (brew/port install qemu)',
  'qemu.tip_select_target': '请先在左侧列表点击选择要测试的目标 U 盘',
  'qemu.tip_ready': '点击在当前桌面拉起 QEMU 虚拟机校验 U 盘引导',
  'qemu.toast_select_first': '⚠️ 请先在左侧磁盘列表中点击选择要测试的目标 U 盘！',
  'qemu.toast_not_installed': '❌ 未检测到 QEMU 模拟器！请先安装 QEMU (brew install qemu 或 port install qemu)',
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

  'deploy.title': 'Bootable Drive Creation & QEMU Test',
  'deploy.desc_cloud': 'Pure iPXE Cloud Boot • Ultra-fast dual-partition setup with multi-arch iPXE network firmware.',
  'deploy.desc_hybrid': 'Integrated Ventoy core + UniBoot dark theme & iPXE cloud extension, supports multi-GB ISO images.',
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
  'confirm.title_safe': 'In-Place Upgrade: No Format Required',
  'confirm.title_mixed': '⚡ Smart Mixed Mode Confirmation',
  'confirm.title_danger': 'Format Warning: Disk Initialization',
  'confirm.safe_banner_title': '💡 In-Place Incremental Update Notice (Data Safe)',
  'confirm.safe_banner_desc': 'Ventoy / UniBoot boot structure detected on target drive. System will perform an incremental update skipping format. All existing files & ISOs are 100% preserved!',
  'confirm.mixed_banner_title': '🛡️ Smart Mixed Mode: In-Place Update for Boot Drives, Format for Blank Drives',
  'confirm.mixed_banner_desc': 'Selected {ventoyCount} boot drive(s) (In-place update) and {blankCount} blank drive(s) (Full format).',
  'confirm.danger_banner_title': '💥 Warning: Formatting is irreversible!',
  'confirm.danger_banner_desc': 'Writing will re-partition and format the target device (MBR/GPT). All existing files on selected drive(s) will be completely erased!',
  'confirm.summary_title': 'Target Drives for Boot Deployment:',
  'confirm.smart_safe_tag': '🛡️ Smart Format Skip',
  'confirm.ventoy_group_title': '🛡️ In-Place Upgrade Drives (All ISOs Preserved):',
  'confirm.blank_group_title': '⚠️ Drives to be Formatted:',
  'confirm.batch_summary_title': 'Selected {count} USB drive(s) for Parallel Format:',
  'confirm.fs_format': '{fs} Format',

  'ventoy_alert.goto_settings': '⚙️ Go to Ventoy Settings',
  'ventoy_alert.switch_b': '🚀 Switch to Mode B (Cloud Mode)',
  'ventoy_alert.close': 'Close',
  'ventoy_alert.default_title': 'Cannot Create Mode A Disk',
  'ventoy_alert.banner_title': '💡 Missing Pre-requisites & Guidance',

  'inspector.title': 'USB Hardware Inspector Report',
  'inspector.subtitle': 'PHY Protocol, Speed & Hardware Analysis',
  'inspector.device': 'Device Node:',
  'inspector.vendor': 'Vendor / Brand:',
  'inspector.bus_speed': 'Bus Protocol & Speed:',
  'inspector.fake_report': 'Fake USB 3.0 Analysis:',
  'inspector.fake_text': '⚠️ Downgraded Fake USB 3.0 (Blue/Black port, but internal protocol is USB 2.0 480Mb/s)',
  'inspector.genuine_text': '⚡ Genuine High-Speed USB 3.0/3.1 Controller',
  'inspector.smart': 'SMART Health:',
  'inspector.sector': 'Sector Size:',
  'inspector.fake_title': 'Fake USB 3.0 Warning Alert!',
  'inspector.fake_desc': 'Device advertises USB 3.0/3.1, but actual physical layer speed is negotiated at only {speed} (USB 2.0 High-Speed PHY). This drive likely has spoofed firmware or a fake blue port.',
  'inspector.genuine_title': 'Physical Hardware Verification Passed (Genuine USB 3.0+ Device)',
  'inspector.genuine_desc': 'Physical PHY layer established SuperSpeed/SuperSpeed+ link with measured speed of {speed}.',
  'inspector.usb2_title': 'Standard USB 2.0 Interface',
  'inspector.usb2_desc': 'Device hardware is USB 2.0, theoretical physical max speed 480 Mb/s.',
  'inspector.section_basic': '📊 Basic Device Info',
  'inspector.section_hw': '🛠️ Hardware Details',
  'inspector.lbl_name': 'Device Name',
  'inspector.lbl_vendor': 'Vendor / Brand',
  'inspector.lbl_size': 'Storage Size',
  'inspector.lbl_free': 'Free Space',
  'inspector.lbl_protocol': 'USB Protocol Version',
  'inspector.lbl_speed': 'Negotiated Speed (PHY)',
  'inspector.lbl_fs': 'File System',
  'inspector.lbl_scheme': 'Partition Scheme',
  'inspector.lbl_mount': 'Mount Path',
  'inspector.lbl_perm': 'Disk Permission',
  'inspector.val_rw': '✅ Read-Write',
  'inspector.val_ro': '🔒 Read-Only (Write Protected)',
  'inspector.lbl_boot_status': 'Boot Status',
  'inspector.val_data_disk': '📁 Data Drive (No bootloader detected)',
  'inspector.lbl_smart': 'SMART Health Status',
  'inspector.val_smart_good': '✅ Healthy (Verified)',
  'inspector.lbl_bus_power': 'USB Bus Power',
  'inspector.val_power_limit': '(Port Max: {limit})',
  'inspector.lbl_transport': 'Transport Protocol',
  'inspector.lbl_sector': 'Sector Size (Physical / Logical)',
  'inspector.lbl_controller': 'Controller Chip Vendor',

  'icon_picker.title': 'Customize USB Drive Icon',
  'icon_picker.reset': 'Reset Default Icon',
  'icon_picker.subtitle': 'Choose custom appearance for target drive {name}:',
  'icon_picker.usb_desc': 'Classic SuperSpeed 5 Gbps Drive',
  'icon_picker.usb2_desc': 'Classic High-Speed 480 Mbps Drive',
  'icon_picker.usb3_1_desc': 'SuperSpeed+ 10 Gbps High Speed Drive',
  'icon_picker.usb3_2_desc': 'Gen 2x2 20 Gbps Dual-Channel Drive',
  'icon_picker.usb4_desc': 'Flagship 40 Gbps USB4/Thunderbolt Drive',
  'icon_picker.boot_desc': 'Bootable Drive with Lightning Emblem',
  'icon_picker.ssd_desc': 'Portable Solid State Drive (PSSD)',
  'icon_picker.typec_desc': 'Dual Type-C Drive for Phone & Mac',
  'icon_picker.secure_desc': 'Encrypted Hardware with Physical Keypad',
  'icon_picker.reader_desc': 'SD / MicroSD Card Reader',
  'icon_picker.hdd_desc': '2.5" Portable Mechanical Hard Drive (HDD)',
  'icon_picker.key_desc': 'FIDO2 Security Key Hardware',

  'disk.tag_boot': 'Boot Drive',
  'disk.tag_ssd': 'Portable SSD',
  'disk.tag_typec': 'Type-C Drive',
  'disk.tag_secure': 'Encrypted Drive',
  'disk.tag_reader': 'Card Reader',
  'disk.tag_hdd': 'Portable HDD',
  'disk.tag_key': 'Security Key',
  'disk.tag_cdrom': 'Virtual CD-ROM',

  'inspector.std_controller': 'Standard Controller',
  'inspector.lbl_vid_pid': 'Hardware ID (USB VID / PID)',
  'inspector.lbl_serial': 'Physical Serial Number',
  'inspector.ext_protocols': 'Extensible Protocol Standards',
  'inspector.close': 'Close',

  'deploy.toast_switched_b': 'Switched to native Mode B (1-Sec Cloud Install Disk)!',
  'deploy.toast_added_iso': 'Successfully added {count} image source file(s)',
  'deploy.toast_added_demo_iso': 'Added 2 demo image source files (Browser Demo)',
  'deploy.toast_select_target': '⚠️ Please select a target USB drive from the left panel first',
  'deploy.toast_select_batch': '⚠️ Please check target USB drives for batch format first',
  'deploy.toast_no_disks': '⚠️ No USB drive detected! Please insert a USB drive and try again.',

  'deploy.tip_writing': 'Writing boot firmware...',
  'deploy.tip_select_single': 'Please select a target USB drive first',
  'deploy.tip_select_batch': 'Please check target USB drives for batch formatting',
  'deploy.tip_macos_unsupported': '❌ macOS does not support fresh formatting for Mode A drive (Please use Mode B)',
  'deploy.tip_need_ventoy': 'Fresh Mode A deployment requires Ventoy CLI environment',

  'deploy.macos_alert_title': 'macOS Ventoy CLI Fresh Formatting Unsupported',
  'deploy.macos_alert_desc': 'Official Ventoy CLI does not support running direct disk formatting on macOS. Mode A fresh formatting requires Ventoy CLI; we recommend natively supported Mode B (1-Sec Cloud Disk)! For Mode A, initialize Ventoy drive on Windows/Linux first then perform in-place upgrade on macOS.',
  'deploy.no_ventoy_title': 'Ventoy CLI Executable Not Detected',
  'deploy.no_ventoy_desc': 'Fresh Mode A (Ventoy Hybrid Disk) requires local Ventoy CLI executable (Ventoy2Disk). Please configure Ventoy CLI path in Settings or switch to Mode B (Cloud Boot Disk) which requires no Ventoy CLI!',

  'deploy.result_batch_success': 'Successfully deployed 1-Sec Cloud Install Disk to {count} USB drive(s)!',
  'deploy.result_success': 'Successfully deployed Mode {mode} to {targets}',
  'deploy.alert_success': '🎉 Deployment Successful!\n\n{msg}',
  'deploy.alert_fail': '❌ Deployment Failed:\n\n{msg}',

  'qemu.tip_launching': 'Launching QEMU emulator...',
  'qemu.tip_deploying': 'Deploying boot files, please wait until finished',
  'qemu.tip_not_installed': 'QEMU emulator not found. Please install QEMU first (brew/port install qemu)',
  'qemu.tip_select_target': 'Please select a target USB drive from the left panel first',
  'qemu.tip_ready': 'Click to launch QEMU VM to verify USB bootloader on current desktop',
  'qemu.toast_select_first': '⚠️ Please click to select a target USB drive from the left panel first!',
  'qemu.toast_not_installed': '❌ QEMU emulator not found! Please install QEMU (brew install qemu or port install qemu)',
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
