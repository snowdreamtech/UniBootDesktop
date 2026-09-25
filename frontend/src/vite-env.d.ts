/// <reference types="vite/client" />

declare global {
  interface Window {
    runtime?: {
      EventsOn(eventName: string, callback: (data: any) => void): void;
      EventsOff(eventName: string, ...additionalEvents: string[]): void;
      EventsOnce(eventName: string, callback: (data: any) => void): void;
      EventsEmit(eventName: string, ...optionalData: any[]): void;
      BrowserOpenURL(url: string): void;
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
          OpenURL(url: string): Promise<void>;
        };
      };
    };
  }
}

export {};
