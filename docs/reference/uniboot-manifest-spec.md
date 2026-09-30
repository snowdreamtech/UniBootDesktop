# UniBoot 启动盘元数据清单规范 (uniboot.json Specification)

> **版本**：1.0.0  
> **更新时间**：2026-09-21  
> **状态**：正式规范 (Active Standard)

---

## 1. 概述与设计背景

UniBoot 启动盘支持 **“1秒极速纯云盘 (Cloud Mode)”** 与 **“本地+云端双引导混合盘 (Hybrid Mode)”** 两种部署形态。

为了解决以往依赖不可靠的“磁盘卷标（Volume Label）”导致的同名盘识别冲突、错误挂载点路由以及模糊匹配等安全隐患，UniBoot 规范正式确立 **`ipxe/uniboot.json`** 作为启动介质自描述、身份识别与固件版本协同的**法定唯一信源（Single Source of Truth, SSOT）**。

---

## 2. 存放路径规范 (Canonical Location)

无论属于哪种部署模式，该清单文件**恒定**收敛放置在 UniBoot 固件所在分区的 `ipxe/` 子目录下：

| 部署模式 | 挂载宿主分区 | 完整绝对相对路径 | 分区特性 |
| :--- | :--- | :--- | :--- |
| **纯云盘 (Cloud Mode)** | **ESP 独立引导分区** (Partition 2) | `<ESP_MOUNT>/ipxe/uniboot.json` | 64MB FAT32 隐藏分区；主数据分区 100% 纯净无残留 |
| **混合盘 (Hybrid Mode)** | **主数据存储分区** (Partition 1) | `<DATA_MOUNT>/ipxe/uniboot.json` | 用户可见数据分区；与 `/ventoy/` 核心目录并列独立 |

> [!IMPORTANT]
> **绝对禁止多分区重复写入**：纯云盘绝不污染主数据分区，混合盘绝不篡改 Ventoy VTOYEFI 分区。每个模式仅有唯一合法的存储位置。

---

## 3. JSON 数据格式定义 (Schema)

### 3.1 完整示例

```json
{
  "magic": "UNIBOOT_DISK",
  "version": "1.0.0",
  "mode": "cloud",
  "arch": [
    "x86_64",
    "arm64",
    "ia32"
  ],
  "engine": {
    "name": "ipxe",
    "version": "1.21.1"
  },
  "created_at": 1726910000,
  "uuid": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 3.2 字段说明表

| 字段名 | 类型 | 必需 | 约束与有效值 | 说明 |
| :--- | :--- | :---: | :--- | :--- |
| `magic` | `string` | **是** | 固定为 `"UNIBOOT_DISK"` | **物理魔数指纹**。用于防伪与绝对防碰撞；魔数不符直接视为非 UniBoot 介质。 |
| `version` | `string` | **是** | 符合 SemVer 规范 (如 `"1.0.0"`) | **固件与配置规范版本**。用于桌面端向后兼容判定及免格盘 OTA 升级。 |
| `mode` | `string` | **是** | `"cloud"` 或 `"hybrid"` | **部署模式自描述**。<br>• `cloud`：1秒极速纯云盘<br>• `hybrid`：Ventoy+UniBoot 混合引导盘 |
| `arch` | `[]string` | 否 | `["x86_64", "arm64", "ia32"]` | 该介质内置的架构固件支持列表。 |
| `engine.name` | `string` | 否 | `"ipxe"` 或 `"ventoy+ipxe"` | 底层引导引擎分类。 |
| `engine.version` | `string` | 否 | 字符串，如 `"1.21.1"` | 引导引擎的具体编译版本号。 |
| `created_at` | `int64` | **是** | Unix 秒级时间戳 | 磁盘初次制作或最近一次固件升级时间。 |
| `uuid` | `string` | 否 | 标准 UUIDv4 字符串 | 本次安装部署的全球唯一实例 ID，用于批量日志审计与跟踪。 |

---

## 4. 识别与校验逻辑 (Validation Logic)

系统读取目标磁盘时，执行以下严格校验步骤：

1. **路径直读**：直接定位目标挂载点下的 `ipxe/uniboot.json`。
2. **魔数匹配**：读取并反序列化 JSON，核验 `manifest.magic == "UNIBOOT_DISK"`。
3. **模式确认**：
   - 若 `mode == "cloud"`，且磁盘 MBR 无 Ventoy 特征，确认为 **UniBoot 纯云盘**；
   - 若 `mode == "hybrid"`，且磁盘 Sector 0 具备 Ventoy 引导签名，确认为 **UniBoot 混合引导盘**。
4. **向后兼容性 (Fallback)**：
   - 对于规范发布前制作的历史磁盘（尚无 `uniboot.json`），系统向后兼容探测 `ipxe/uniboot.ipxe` 或 `boot.ipxe`，并在下一次更新时自动补全规范的 `uniboot.json` 清单。

---

## 5. 安全准则

1. **零卷标依赖 (Zero Label Dependency)**：不得根据卷标（Volume Name）来推测或替代 `uniboot.json` 的判定结果。
2. **幂等升级 (Idempotent Upgrade)**：桌面客户端执行“智能升级（Smart Upgrade）”时，仅更新固件和 `uniboot.json` 的 `version` 与 `created_at`，绝不擦除数据分区用户已有文件。
