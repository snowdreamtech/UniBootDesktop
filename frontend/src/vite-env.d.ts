/// <reference types="vite/client" />

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          GetDiskList(): Promise<any[]>;
          DeployModeA(targetDisk: string, fsType?: string): Promise<any>;
          DeployModeABatch(targetDisks: string[], fsType?: string): Promise<any[]>;
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

