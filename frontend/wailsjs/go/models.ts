export namespace config {
	
	export class AppConfig {
	    autoCheckUpdate: boolean;
	    theme: string;
	    language: string;
	    githubProxy: string;
	    proxyProtocol: string;
	    proxyHost: string;
	    proxyPort: number;
	    proxyUser: string;
	    proxyPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoCheckUpdate = source["autoCheckUpdate"];
	        this.theme = source["theme"];
	        this.language = source["language"];
	        this.githubProxy = source["githubProxy"];
	        this.proxyProtocol = source["proxyProtocol"];
	        this.proxyHost = source["proxyHost"];
	        this.proxyPort = source["proxyPort"];
	        this.proxyUser = source["proxyUser"];
	        this.proxyPassword = source["proxyPassword"];
	    }
	}

}

export namespace main {
	
	export class HelloInfo {
	    greeting: string;
	    os: string;
	    arch: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new HelloInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.greeting = source["greeting"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class NetworkTestResult {
	    connected: boolean;
	    latencyMs: number;
	    targetUrl: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.latencyMs = source["latencyMs"];
	        this.targetUrl = source["targetUrl"];
	        this.error = source["error"];
	    }
	}
	export class SystemInfo {
	    os: string;
	    arch: string;
	    goVersion: string;
	    appName: string;
	    version: string;
	    commit: string;
	    buildTime: string;
	    dataDir: string;
	    configDir: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.goVersion = source["goVersion"];
	        this.appName = source["appName"];
	        this.version = source["version"];
	        this.commit = source["commit"];
	        this.buildTime = source["buildTime"];
	        this.dataDir = source["dataDir"];
	        this.configDir = source["configDir"];
	    }
	}

}

export namespace updater {
	
	export class UpdateStatus {
	    hasUpdate: boolean;
	    currentTag: string;
	    latestTag: string;
	    downloadUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.currentTag = source["currentTag"];
	        this.latestTag = source["latestTag"];
	        this.downloadUrl = source["downloadUrl"];
	    }
	}

}

