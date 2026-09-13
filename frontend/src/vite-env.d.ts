/// <reference types="vite/client" />

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          GetDiskList(): Promise<any[]>;
          DeployModeA(targetDisk: string, fsType?: string): Promise<any>;
          DeployModeABatch(targetDisks: string[], fsType?: string): Promise<any[]>;
          DeployModeB(targetDisk: string): Promise<any>;
          DeployModeBBatch(targetDisks: string[]): Promise<any[]>;
          CheckQEMU(): Promise<any>;
          CheckUpdate(): Promise<any>;
          GetConfig(): Promise<any>;
        };
      };
    };
  }
}

export {};
