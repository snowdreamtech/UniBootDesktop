/// <reference types="vite/client" />

declare global {
  interface Window {
    WailsInvoke?: (message: string) => void;
    runtime?: {
      EventsOn(eventName: string, callback: (data: any) => void): void;
      EventsOff(eventName: string, ...additionalEvents: string[]): void;
      EventsOnce(eventName: string, callback: (data: any) => void): void;
      EventsEmit(eventName: string, ...optionalData: any[]): void;
      BrowserOpenURL(url: string): void;
      WindowStartDrag?(): void;
      WindowToggleMaximise?(): void;
      WindowMinimise?(): void;
      WindowShow?(): void;
      WindowHide?(): void;
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
          OpenURL(url: string): Promise<void>;
          ReloadAppMenu?(lang: string): Promise<void>;
          GetRecentLogs?(): Promise<any>;
          ClearLogs?(): Promise<void>;
          LogAction?(level: string, message: string, details: string): Promise<void>;
          ExportLogs?(
            content: string,
            title: string,
            logFilter: string,
            textFilter: string,
            allFilter: string
          ): Promise<string>;
          GetDiskList?(): Promise<any>;
          EjectDisk?(targetDisk: string): Promise<void>;
          DetectHypervisors?(): Promise<any>;
          DetectBestHypervisor?(): Promise<any>;
          GetDefaultVMConfig?(): Promise<any>;
          LaunchVM?(targetDisk: string, vmType: string, bootMode: string): Promise<void>;
          LaunchVMWithConfig?(targetDisk: string, vmType: string, cfg: any): Promise<void>;
          StopVM?(): Promise<void>;
          IsPrivileged?(): Promise<boolean>;
          RequestPrivilegeElevation?(customPrompt?: string): Promise<boolean>;
          CheckConfigHealth?(): Promise<any>;
          CalculateFileChecksum?(filePath: string, algo?: string): Promise<any>;
        };
      };
    };
  }
}

export {};
