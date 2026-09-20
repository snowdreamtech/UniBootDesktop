# UI 逻辑 Bug 修复

## 问题

升级混合模式磁盘时，界面错误显示 "全量格式化"，应显示 "智能原位升级"。

## 原因

混合模式磁盘结构：
- **分区 1（数据区）**：`/ventoy/` 目录（Ventoy 引擎）
- **分区 2（ESP）**：`boot.ipxe`（云引导） + `BOOTX64.EFI`（Ventoy 引导）

`IsCloudModeDisk()` 原逻辑：遍历分区，检查 `HasUniBootCloudFiles() && !HasVentoyEngineFiles()`

❌ **错误**：ESP 分区有云文件但没有 `/ventoy/` 目录（在数据区），导致函数返回 `true`，误判为纯云启动盘。

## 修复

重写检测逻辑，从单分区检查改为整盘检查：

1. 先检查 Ventoy MBR 签名 → 有则非纯云启动
2. 遍历所有分区检查 Ventoy 引擎文件 → 有则非纯云启动
3. 最后检查云引导文件 → 有且无 Ventoy 才是纯云启动

## 结果

| 磁盘类型 | 修复前 | 修复后 |
|---------|--------|--------|
| 混合模式 | ❌ 误判为云启动 | ✅ 正确识别为 Ventoy |
| 纯云启动 | ✅ 正确 | ✅ 正确 |
| 纯 Ventoy | ✅ 正确 | ✅ 正确 |

## 文件

- `pkg/disk/disk.go`：修复 `IsCloudModeDisk()` 函数
