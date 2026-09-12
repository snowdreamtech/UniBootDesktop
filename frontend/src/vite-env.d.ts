/// <reference types="vite/client" />

interface Window {
  go?: {
    main?: {
      App?: {
        GetDiskList(): Promise<any[]>;
        DeployModeA(targetDisk: string): Promise<any>;
        DeployModeB(targetDisk: string): Promise<any>;
        CheckQEMU(): Promise<any>;
        CheckUpdate(): Promise<any>;
        GetConfig(): Promise<any>;
      };
    };
  };
}
