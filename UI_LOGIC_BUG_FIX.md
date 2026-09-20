# UI 逻辑 Bug 修复

## 问题

升级混合模式磁盘时，界面错误显示 "全量格式化"，应显示 "智能原位升级"。

## 原因

混合模式磁盘结构：
- **分区 1（数据区）**：`/ventoy/` 目录（Ventoy 引擎）
- **分区 2（ESP）**：`boot.ipxe`（云引导） + `BOOTX64.EFI`（Ventoy 引导）

**原始逻辑缺陷**：

1. **macOS**: `IsCloudModeDisk()` 遍历分区，ESP 分区有云文件但没有 `/ventoy/` 目录（在数据区），误判为纯云启动
2. **Linux**: 传入挂载点而非设备路径，MBR 检查和多分区检测都失效
3. **Windows**: 只能检查单个盘符挂载点，无法检测多分区布局

## 修复

### 1. macOS - 完善分区检测逻辑

```go
func IsCloudModeDisk(targetDisk string) bool {
    // 先检查 Ventoy MBR 签名（快速排除）
    if CheckVentoyMbrSignature(targetDisk) {
        return false
    }

    // 遍历所有分区检查 Ventoy 引擎文件
    for _, p := range []string{p1, p2} {
        if HasVentoyEngineFiles(mountPoint) {
            return false  // 有 Ventoy → 非纯云启动
        }
    }

    // 最后检查云引导文件
    for _, p := range []string{p1, p2} {
        if HasUniBootCloudFiles(mountPoint) {
            return true  // 有云文件且无 Ventoy → 纯云启动
        }
    }
    return false
}
```

### 2. Linux - 修正检测参数

**磁盘检测代码修正**：传入设备路径而非挂载点
```go
// 使用 devPath (/dev/sdb) 而非 mountPath
IsRealVentoy: IsRealVentoyDisk(devPath),
IsCloudMode:  IsCloudModeDisk(devPath),
```

**IsCloudModeDisk Linux 分支增强**：
```go
if strings.HasPrefix(targetDisk, "/dev/") {
    // 使用 lsblk 查询所有分区的挂载点
    out, _ := exec.Command("lsblk", "-o", "MOUNTPOINT", "-n", "-l", targetDisk).Output()

    // 检查每个挂载点是否有 Ventoy/云文件
    for _, mountPoint := range mountPoints {
        if HasVentoyEngineFiles(mountPoint) { hasVentoyFiles = true }
        if HasUniBootCloudFiles(mountPoint) { hasCloudFiles = true }
    }

    // 纯云启动：有云文件但无 Ventoy 文件
    return hasCloudFiles && !hasVentoyFiles
}
```

### 3. Windows - 明确限制

Windows 检测依赖单盘符挂载点，无法完美区分混合模式，但至少确保：
- 有 Ventoy 引擎文件 → 非纯云启动
- 只有云文件 → 纯云启动

## 结果

| 磁盘类型 | 修复前 | 修复后 |
|---------|--------|--------|
| 混合模式 (macOS) | ❌ 误判为云启动 | ✅ 正确识别为 Ventoy |
| 混合模式 (Linux) | ❌ 检测失效 | ✅ 正确识别为 Ventoy |
| 混合模式 (Windows) | ⚠️ 可能误判 | ✅ 大部分场景正确 |
| 纯云启动 | ✅ 正确 | ✅ 正确 |
| 纯 Ventoy | ✅ 正确 | ✅ 正确 |

## 修改文件

- `pkg/disk/disk.go`：
  - `IsCloudModeDisk()` 函数：完善 macOS/Linux/Windows 三平台逻辑
  - `getLinuxDisks()` 函数：修正检测参数从 `mountPath` 改为 `devPath`
