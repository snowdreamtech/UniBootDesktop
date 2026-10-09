/// <reference types="vite/client" />

declare global {
  interface DiskInfo {
    device: string;
    name: string;
    size: number;
    formatted: string;
    freeSpace?: number;
    freeFormatted?: string;
    isRemovable: boolean;
    isSystem: boolean;
    usbVersion?: string;
    usbSpeed?: string;
    vendor?: string;
    fileSystem?: string;
    partitionScheme?: string;
    writable?: boolean;
    serialNumber?: string;
    vendorId?: string;
    productId?: string;
    smartStatus?: string;
    busPower?: string;
    busPowerUsed?: string;
    sectorSize?: string;
    transportProtocol?: string;
    bootStatus?: string;
    bootStatusCode?: string;
    controllerVendor?: string;
    isFakeUsb3?: boolean;
    protocolCode?: string;
    isRealVentoy?: boolean;
    isCloudMode?: boolean;
    isGenericBoot?: boolean;
    thirdPartyBootType?: string;
    thirdPartyBootCode?: string;
    unibootVersion?: string;
    unibootMode?: string;
    mountPoint?: string;
  }

  interface Window {
    runtime?: {
      EventsOn(eventName: string, callback: (data: any) => void): void;
      EventsOff(eventName: string, ...additionalEvents: string[]): void;
      EventsOnce(eventName: string, callback: (data: any) => void): void;
      EventsEmit(eventName: string, ...optionalData: any[]): void;
      BrowserOpenURL(url: string): void;
      OnFileDrop?(callback: (x: number, y: number, paths: string[]) => void, useDropTarget?: boolean): void;
      OnFileDropOff?(): void;
    };
    go?: {
      main?: {
        App?: {
          Greet(name: string): Promise<string>;
          GetHelloInfo(): Promise<any>;
          GetSystemInfo(): Promise<any>;
          TestNetwork(targetUrl: string): Promise<any>;
          CheckUpdate(): Promise<any>;
          GetConfig(): Promise<any>;
          SaveConfig(cfg: any): Promise<any>;
          SetTheme?(theme: string): Promise<void>;
          IsSystemDarkTheme?(): Promise<boolean>;
          OpenURL(url: string): Promise<void>;
          GetDiskList(): Promise<any[]>;
          SelectIsoFiles(title?: string, ventoyFilter?: string, allFilter?: string): Promise<string[]>;
          LogAction?(level: string, message: string, details?: string): Promise<void>;
          DeployHybridMode(targetDisk: string, fsType?: string, isoPaths?: string[], expected?: DiskInfo): Promise<any>;
          DeployHybridModeBatch(
            targetDisks: string[],
            fsType?: string,
            isoPaths?: string[],
            expected?: DiskInfo[]
          ): Promise<any[]>;
          DeployHybridModeBatchWithPlans?(
            targetDisks: string[],
            fsType?: string,
            isoPaths?: string[],
            plans?: any[],
            expected?: DiskInfo[]
          ): Promise<any[]>;
          PreflightIsoCopy?(targetDisks: string[], isoPaths: string[]): Promise<any[]>;
          DeployCloudMode(targetDisk: string, fsType?: string, expected?: DiskInfo): Promise<any>;
          DeployCloudModeBatch(targetDisks: string[], fsType?: string, expected?: DiskInfo[]): Promise<any[]>;
          CheckQEMU(): Promise<any>;
          LaunchQEMU(targetDisk: string): Promise<void>;
          DetectHypervisors?(): Promise<any[]>;
          DetectBestHypervisor?(): Promise<any>;
          LaunchVM?(targetDisk: string, vmType?: string, bootMode?: string): Promise<void>;
          GetFirmwareList(): Promise<any[]>;
          GetUniBootReleaseInfo(): Promise<any>;
          SyncUniBootFirmware(): Promise<any>;
          GetVentoyReleaseInfo?(): Promise<any>;
          DownloadVentoyRelease?(): Promise<any>;
          SelectDirectory?(title?: string): Promise<string>;
          ValidateVentoyCli(ventoyPath: string): Promise<any>;
          EjectDisk(targetDisk: string): Promise<void>;
          BatchEjectDisks?(targetDisks: string[]): Promise<{ success: string[]; failed: Record<string, string> }>;
          ReloadAppMenu(lang: string): Promise<void>;
          RequestPrivilegeElevation?(customPrompt?: string): Promise<boolean>;
        };
      };
    };
  }
}

export {};
