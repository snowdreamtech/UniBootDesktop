# 批量校验功能设计文档

## 📋 需求分析

### 当前问题

- 校验功能只针对单个 ISO 设计
- 每次只能校验一个文件
- 导入 SHA 文件后只提取当前选中文件的哈希值
- 无法高效校验多个 ISO 文件

### 目标功能

1. **批量导入**：支持导入包含多个文件哈希的 SHA256SUMS 文件
2. **智能缓存**：缓存所有文件名到哈希值的映射
3. **批量计算**：一键计算所有已选 ISO 的哈希值
4. **状态追踪**：每个 ISO 独立显示校验状态
5. **兼容单文件**：保持现有单文件校验功能

---

## 🎯 核心设计

### 1. 数据结构

```typescript
// SHA 缓存：文件名 -> 哈希值映射
interface ChecksumCache {
  [filename: string]: string;
}

// 单个 ISO 的校验状态
interface IsoChecksumStatus {
  calculated: string;      // 计算出的哈希值
  expected: string;        // 期望的哈希值（来自缓存）
  status: 'idle' | 'calculating' | 'match' | 'mismatch' | 'no-expected';
  algo: string;            // 使用的算法 (sha256/md5)
}

// 状态存储
const checksumCache = ref<ChecksumCache>({});
const isoChecksumStatuses = ref<Map<string, IsoChecksumStatus>>(new Map());
```

### 2. 工作流程

#### 导入 SHA 文件

```
用户点击"导入校验文件"
  ↓
读取文件内容（如 SHA256SUMS）
  ↓
解析所有行：hash filename
  ↓
存入 checksumCache: { "ubuntu.iso": "abc123...", "debian.iso": "def456..." }
  ↓
自动为当前所有 ISO 匹配期望哈希值
```

#### 批量计算

```
用户点击"批量校验所有文件"
  ↓
遍历所有已选 ISO
  ↓
逐个调用 CalculateFileChecksum()
  ↓
对比 calculated vs expected
  ↓
更新每个文件的状态为 match/mismatch/no-expected
```

#### 单文件计算（保持向后兼容）

```
用户选择单个 ISO 并点击"计算Hash"
  ↓
计算当前选中文件的哈希
  ↓
自动从 cache 查找期望值
  ↓
显示匹配结果
```

---

## 🖼️ UI 设计

### 批量校验模式（新增）

```
┌─────────────────────────────────────────────────┐
│ 📄 校验 Hash 值                                   │
├─────────────────────────────────────────────────┤
│ 已添加 3 个系统镜像文件                            │
│                                                   │
│ ☑ ubuntu-24.04.4-live-server-amd64.iso           │
│    ✓ 已校验 - 匹配  [SHA-256]                     │
│                                                   │
│ ☑ debian-12.9.0-amd64-netinst.iso                │
│    ✗ 不匹配  [SHA-256]                           │
│                                                   │
│ ☑ archlinux-2024.12.01-x86_64.iso                │
│    - 未校验  [等待计算]                           │
│                                                   │
│ [📄 导入 SHA 文件] [🔍 批量校验全部]              │
│                                                   │
│ 进度：2/3 已校验 | 1 匹配 | 1 不匹配              │
└─────────────────────────────────────────────────┘
```

### 单文件校验模式（现有 + 增强）

```
┌─────────────────────────────────────────────────┐
│ 选择文件：                                        │
│ [ubuntu-24.04 ▼] [算法: SHA-256 ▼] [计算Hash]   │
│                                                   │
│ SHA256: e9bf2d22eecc... [📋 复制]                │
│                                                   │
│ 期望Hash: e9bf2d22eecc... [📄 导入文件]          │
│ ✅ Hash 校验匹配一致                              │
│                                                   │
│ 💡 提示：已从缓存自动匹配期望值                    │
└─────────────────────────────────────────────────┘
```

---

## 🔧 实现要点

### 1. 导入文件解析增强

```typescript
function parseChecksumFile(fileContent: string): ChecksumCache {
  const cache: ChecksumCache = {};
  const lines = fileContent.split('\n');

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;

    // 支持格式：
    // hash  filename
    // hash *filename
    // hash  ./path/to/filename
    const match = trimmed.match(/^([a-fA-F0-9]+)\s+\*?(.+)$/);
    if (match) {
      const [, hash, filepath] = match;
      const filename = filepath.split('/').pop() || filepath;
      cache[filename.trim()] = hash.toLowerCase();
    }
  }

  return cache;
}
```

### 2. 批量计算函数

```typescript
async function handleBatchChecksum() {
  if (props.selectedIsoFiles.length === 0) return;

  isBatchCalculating.value = true;
  let successCount = 0;
  let failCount = 0;

  for (const [index, file] of props.selectedIsoFiles.entries()) {
    // 更新状态为计算中
    const status: IsoChecksumStatus = {
      calculated: '',
      expected: checksumCache.value[file.name] || '',
      status: 'calculating',
      algo: selectedAlgo.value
    };
    isoChecksumStatuses.value.set(file.name, status);

    try {
      // 调用后端计算
      const w = window as any;
      const res = await w.go.main.App.CalculateFileChecksum(file.path, selectedAlgo.value);

      if (res && res.hash) {
        status.calculated = res.hash.toLowerCase();

        // 判断匹配状态
        if (!status.expected) {
          status.status = 'no-expected';
        } else if (status.calculated === status.expected) {
          status.status = 'match';
          successCount++;
        } else {
          status.status = 'mismatch';
          failCount++;
        }
      }
    } catch (err) {
      status.status = 'no-expected';
      failCount++;
    }

    isoChecksumStatuses.value.set(file.name, status);
  }

  isBatchCalculating.value = false;

  // 显示结果提示
  if (failCount === 0) {
    alert(`✅ 全部校验成功！${successCount} 个文件哈希匹配`);
  } else {
    alert(`⚠️ 校验完成：${successCount} 个匹配，${failCount} 个不匹配或无期望值`);
  }
}
```

### 3. UI 状态指示器组件

```vue
<template>
  <div class="iso-checksum-status" :class="status">
    <span class="status-icon">{{ icon }}</span>
    <span class="status-text">{{ text }}</span>
  </div>
</template>

<script setup>
const props = defineProps<{
  status: 'idle' | 'calculating' | 'match' | 'mismatch' | 'no-expected';
}>();

const icon = computed(() => {
  switch (props.status) {
    case 'match': return '✓';
    case 'mismatch': return '✗';
    case 'calculating': return '⏳';
    default: return '-';
  }
});

const text = computed(() => {
  switch (props.status) {
    case 'match': return '已校验 - 匹配';
    case 'mismatch': return '不匹配';
    case 'calculating': return '计算中...';
    case 'no-expected': return '未找到期望值';
    default: return '未校验';
  }
});
</script>
```

---

## 📱 国际化文本

```typescript
{
  "checksum": {
    "batch_verify_all": "批量校验全部",
    "batch_verifying": "正在校验 {current}/{total}",
    "batch_result": "{matched} 匹配，{mismatched} 不匹配",
    "status_idle": "未校验",
    "status_calculating": "计算中...",
    "status_match": "已校验 - 匹配",
    "status_mismatch": "不匹配",
    "status_no_expected": "未找到期望值",
    "cache_loaded": "已加载 {count} 个文件的校验值",
    "auto_matched": "已从缓存自动匹配期望值"
  }
}
```

---

## ✅ 实现步骤

1. **Phase 1：数据层**
   - [ ] 添加 `checksumCache` 和 `isoChecksumStatuses` 状态
   - [ ] 实现 `parseChecksumFile()` 函数
   - [ ] 修改 `handleSumsFileSelected()` 解析并缓存所有哈希

2. **Phase 2：逻辑层**
   - [ ] 实现 `handleBatchChecksum()` 批量计算函数
   - [ ] 修改 `handleCalculateChecksum()` 支持从缓存自动匹配
   - [ ] 添加状态同步逻辑

3. **Phase 3：UI 层**
   - [ ] 在 ISO 列表中为每个文件添加状态指示器
   - [ ] 添加"批量校验全部"按钮
   - [ ] 添加批量进度显示
   - [ ] 显示缓存统计信息

4. **Phase 4：优化**
   - [ ] 添加批量计算的取消功能
   - [ ] 支持并发计算（可选）
   - [ ] 添加校验结果导出功能

---

## 🎬 用户使用场景

### 场景 1：批量下载后验证

```
1. 用户下载了 ubuntu.iso, debian.iso, arch.iso
2. 同时下载了官方的 SHA256SUMS 文件
3. 在应用中添加所有 ISO 文件
4. 点击"导入校验文件"，选择 SHA256SUMS
5. 点击"批量校验全部"
6. 等待计算完成，查看每个文件的匹配状态
```

### 场景 2：单文件快速验证（向后兼容）

```
1. 用户只添加一个 ISO 文件
2. 手动输入或导入期望 SHA 值
3. 点击"计算Hash"
4. 立即显示匹配结果
```

---

## 🚀 未来扩展

- 支持多种算法批量校验（SHA256 + MD5 + SHA1）
- 保存校验历史记录
- 支持拖拽导入多个 SHA 文件
- 自动从 ISO 元数据中查找官方 SHA 源
- 导出校验报告（PDF/HTML）

---

**创建时间**：2026-09-19
**作者**：Kiro AI Assistant
**状态**：设计阶段 - 待实现
