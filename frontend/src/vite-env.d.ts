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
    controllerVendor?: string;
    isFakeUsb3?: boolean;
    protocolCode?: string;
    isRealVentoy?: boolean;
    isModeB?: boolean;
  }

  interface Window {
    runtime?: {
      EventsOn(eventName: string, callback: (data: any) => void): void;
      EventsOff(eventName: string, ...additionalEvents: string[]): void;
      EventsOnce(eventName: string, callback: (data: any) => void): void;
      EventsEmit(eventName: string, ...optionalData: any[]): void;
    };
    go?: {
      main?: {
        App?: {
          GetDiskList(): Promise<any[]>;
          SelectIsoFiles(): Promise<string[]>;
          DeployModeA(targetDisk: string, fsType?: string, isoPaths?: string[]): Promise<any>;
          DeployModeABatch(targetDisks: string[], fsType?: string, isoPaths?: string[]): Promise<any[]>;
          DeployModeB(targetDisk: string, fsType?: string): Promise<any>;
          DeployModeBBatch(targetDisks: string[], fsType?: string): Promise<any[]>;
          CheckQEMU(): Promise<any>;
          LaunchQEMU(targetDisk: string): Promise<void>;
          CheckUpdate(): Promise<any>;
          GetConfig(): Promise<any>;
          SaveConfig(cfg: any): Promise<any>;
          GetFirmwareList(): Promise<any[]>;
          GetUniBootReleaseInfo(): Promise<any>;
          SyncUniBootFirmware(): Promise<any>;
          ValidateVentoyCli(ventoyPath: string): Promise<any>;
        };
      };
    };
  }
}

export {};

