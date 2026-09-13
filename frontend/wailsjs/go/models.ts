export namespace config {
	
	export class AppConfig {
	    mode: string;
	    autoCheckUpdate: boolean;
	    theme: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.autoCheckUpdate = source["autoCheckUpdate"];
	        this.theme = source["theme"];
	    }
	}

}

export namespace disk {
	
	export class DiskInfo {
	    device: string;
	    name: string;
	    size: number;
	    formatted: string;
	    isRemovable: boolean;
	    isSystem: boolean;
	    usbVersion: string;
	    usbSpeed: string;
	    vendor: string;
	    fileSystem: string;
	    partitionScheme: string;
	    writable: boolean;
	    isFakeUsb3: boolean;
	    protocolCode: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device = source["device"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.formatted = source["formatted"];
	        this.isRemovable = source["isRemovable"];
	        this.isSystem = source["isSystem"];
	        this.usbVersion = source["usbVersion"];
	        this.usbSpeed = source["usbSpeed"];
	        this.vendor = source["vendor"];
	        this.fileSystem = source["fileSystem"];
	        this.partitionScheme = source["partitionScheme"];
	        this.writable = source["writable"];
	        this.isFakeUsb3 = source["isFakeUsb3"];
	        this.protocolCode = source["protocolCode"];
	    }
	}

}

export namespace installer {
	
	export class DeployResult {
	    success: boolean;
	    mode: string;
	    target: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.mode = source["mode"];
	        this.target = source["target"];
	        this.message = source["message"];
	    }
	}

}

export namespace qemu {
	
	export class QEMUStatus {
	    installed: boolean;
	    path: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new QEMUStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.path = source["path"];
	        this.version = source["version"];
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

