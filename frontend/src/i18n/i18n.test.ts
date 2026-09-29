import { describe, it, expect, beforeEach } from "vitest";
import { t, setLanguage, currentLocale, isRtl, SUPPORTED_LANGUAGES } from "./index";

describe("i18n module", () => {
  beforeEach(async () => {
    await setLanguage("en-US");
  });

  it("translates pre-bundled en-US string correctly", () => {
    expect(currentLocale.value).toBe("en-US");
    const appTitle = t("app.title");
    expect(appTitle).toBe("UniGoDesktop");
  });

  it("interpolates parameters correctly with {key}", async () => {
    await setLanguage("en-US");
    const notice = t("about.newVersionNotice", { tag: "v1.2.3" });
    expect(notice).toContain("v1.2.3");
  });

  it("switches to zh-CN and translates", async () => {
    await setLanguage("zh-CN");
    expect(currentLocale.value).toBe("zh-CN");
    expect(t("common.close")).toBe("关闭");
  });

  it("identifies RTL languages correctly", async () => {
    await setLanguage("ar-SA");
    expect(isRtl.value).toBe(true);

    await setLanguage("en-US");
    expect(isRtl.value).toBe(false);
  });

  it("contains 53 supported languages", () => {
    expect(SUPPORTED_LANGUAGES.length).toBe(53);
  });
});
