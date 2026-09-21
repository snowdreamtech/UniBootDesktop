import { t } from "../i18n";

export interface VentoyValidation {
  valid: boolean;
  version: string;
  code: string;
  detail?: string;
  message: string;
  executablePath: string;
}

export function getVentoyValidationTitle(validation: VentoyValidation): string {
  if (validation.valid) {
    return t("settings.ventoyToolchain");
  }

  switch (validation.code) {
    case "macos_unsupported":
      return t("deploy.macos_alert_title");
    case "path_empty":
    case "path_too_long":
    case "path_missing":
    case "executable_missing":
    case "os_mismatch":
    case "execution_failed":
    default:
      return t("deploy.no_ventoy_title");
  }
}

export function getVentoyValidationMessage(validation: VentoyValidation): string {
  if (validation.valid) {
    return t("settings.ventoyToolchain");
  }

  switch (validation.code) {
    case "macos_unsupported":
      return t("deploy.macos_alert_desc");
    case "path_empty":
    case "path_too_long":
    case "path_missing":
    case "executable_missing":
    case "os_mismatch":
    case "execution_failed":
    default:
      return t("deploy.no_ventoy_desc");
  }
}
