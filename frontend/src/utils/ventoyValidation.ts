import { t } from "../i18n";

export interface VentoyValidation {
  valid: boolean;
  version: string;
  code?: string;
  detail?: string;
  message: string;
  executablePath: string;
}

export function getVentoyValidationTitle(validation: VentoyValidation): string {
  if (validation.valid) {
    return t("settings.ventoyToolchain");
  }
  if (validation.code === "macos_unsupported") {
    return t("deploy.macos_alert_title");
  }
  return t("deploy.no_ventoy_title");
}

export function getVentoyValidationMessage(validation: VentoyValidation): string {
  if (validation.code === "macos_unsupported") {
    return t("deploy.macos_alert_desc");
  }
  if (validation.valid) {
    return t("settings.ventoyToolchain");
  }
  return t("deploy.no_ventoy_desc");
}
