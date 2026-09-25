import { ref, computed } from 'vue';
import type { TranslationDict } from './types';
export type { TranslationDict };

import { zhCn } from './locales/zh-CN';
import { enUs } from './locales/en-US';

export const SUPPORTED_LANGUAGES = [
  { code: 'zh-CN', name: '简体中文', nativeName: '简体中文' },
  { code: 'en-US', name: 'English', nativeName: 'English' },
  { code: 'zh-TW', name: '正體中文', nativeName: '正體中文' },
  { code: 'ja-JP', name: '日本語', nativeName: '日本語' },
  { code: 'ko-KR', name: '한국어', nativeName: '한국어' },
  { code: 'de-DE', name: 'Deutsch', nativeName: 'Deutsch' },
  { code: 'fr-FR', name: 'Français', nativeName: 'Français' },
  { code: 'es-ES', name: 'Español (España)', nativeName: 'Español (España)' },
  { code: 'es-LA', name: 'Español (Latinoamérica)', nativeName: 'Español (Latinoamérica)' },
  { code: 'ru-RU', name: 'Русский', nativeName: 'Русский' },
  { code: 'pt-BR', name: 'Português (Brasil)', nativeName: 'Português (Brasil)' },
  { code: 'pt-PT', name: 'Português (Portugal)', nativeName: 'Português (Portugal)' },
  { code: 'it-IT', name: 'Italiano', nativeName: 'Italiano' },
  { code: 'tr-TR', name: 'Türkçe', nativeName: 'Türkçe' },
  { code: 'pl-PL', name: 'Polski', nativeName: 'Polski' },
  { code: 'vi-VN', name: 'Tiếng Việt', nativeName: 'Tiếng Việt' },
  { code: 'ar-SA', name: 'العربية', nativeName: 'العربية' },
  { code: 'ur-PK', name: 'اردو', nativeName: 'اردو' },
  { code: 'az-AZ', name: 'Azərbaycanca', nativeName: 'Azərbaycanca' },
  { code: 'da-DK', name: 'Dansk', nativeName: 'Dansk' },
  { code: 'ka-GE', name: 'ქართული', nativeName: 'ქართული' },
  { code: 'fa-IR', name: 'فارسی', nativeName: 'فارسی' },
  { code: 'sl-SI', name: 'Slovenščina', nativeName: 'Slovenščina' },
  { code: 'oc-FR', name: 'Occitan', nativeName: 'Occitan' },
  { code: 'cs-CZ', name: 'Čeština', nativeName: 'Čeština' },
  { code: 'sk-SK', name: 'Slovenčina', nativeName: 'Slovenčina' },
  { code: 'bn-BD', name: 'বাংলা', nativeName: 'বাংলা' },
  { code: 'hi-IN', name: 'हिन्दी', nativeName: 'हिन्दी' },
  { code: 'nl-NL', name: 'Nederlands', nativeName: 'Nederlands' },
  { code: 'ro-RO', name: 'Română', nativeName: 'Română' },
  { code: 'hr-HR', name: 'Hrvatski', nativeName: 'Hrvatski' },
  { code: 'hu-HU', name: 'Magyar', nativeName: 'Magyar' },
  { code: 'sr-Latn', name: 'Srpski', nativeName: 'Srpski' },
  { code: 'sr-Cyrl', name: 'Српски', nativeName: 'Српски' },
  { code: 'th-TH', name: 'ไทย', nativeName: 'ไทย' },
  { code: 'no-NO', name: 'Norsk', nativeName: 'Norsk' },
  { code: 'lt-LT', name: 'Lietuvių', nativeName: 'Lietuvių' },
  { code: 'mk-MK', name: 'Македонски', nativeName: 'Македонски' },
  { code: 'he-IL', name: 'עברית', nativeName: 'עברית' },
  { code: 'id-ID', name: 'Bahasa Indonesia', nativeName: 'Bahasa Indonesia' },
  { code: 'nb-NO', name: 'Norsk Bokmål', nativeName: 'Norsk Bokmål' },
  { code: 'uk-UA', name: 'Українська', nativeName: 'Українська' },
  { code: 'el-GR', name: 'Ελληνικά', nativeName: 'Ελληνικά' },
  { code: 'sv-SE', name: 'Svenska', nativeName: 'Svenska' },
  { code: 'bg-BG', name: 'Български', nativeName: 'Български' },
  { code: 'hy-AM', name: 'Հայերեն', nativeName: 'Հայերեն' },
  { code: 'fi-FI', name: 'Suomi', nativeName: 'Suomi' },
  { code: 'gl-ES', name: 'Galego', nativeName: 'Galego' },
  { code: 'ca-ES', name: 'Català', nativeName: 'Català' },
  { code: 'ta-IN', name: 'தமிழ்', nativeName: 'தமிழ்' },
  { code: 'be-BY', name: 'Беларуская', nativeName: 'Беларуская' },
  { code: 'ml-IN', name: 'മലയാളം', nativeName: 'മലയാളം' },
  { code: 'et-EE', name: 'Eesti', nativeName: 'Eesti' },
];


// Vite glob import for dynamic lazy-loading locale chunks (excluding pre-bundled zh-CN and en-US)
const localeLoaders = import.meta.glob<Record<string, any>>(['./locales/*.ts', '!./locales/zh-CN.ts', '!./locales/en-US.ts']);

// Reactive map of loaded locale dictionaries
const loadedDictionaries = ref<Record<string, TranslationDict>>({
  'zh-CN': zhCn,
  'en-US': enUs,
});

const DEFAULT_LOCALE = 'zh-CN';

export function detectSystemLocale(): string {
  if (typeof navigator === 'undefined') return DEFAULT_LOCALE;
  const navLang = navigator.language;
  const matched = SUPPORTED_LANGUAGES.find(l => l.code === navLang || l.code.startsWith(navLang.split('-')[0]));
  return matched ? matched.code : DEFAULT_LOCALE;
}

function getInitialLocale(): string {
  const saved = typeof localStorage !== 'undefined' ? localStorage.getItem('unigo_locale') : null;
  if (saved && saved !== 'auto' && SUPPORTED_LANGUAGES.some(l => l.code === saved)) {
    return saved;
  }
  return detectSystemLocale();
}

export const selectedLangSetting = ref<string>(
  (typeof localStorage !== 'undefined' && localStorage.getItem('unigo_locale')) || 'auto'
);
export const currentLocale = ref<string>(getInitialLocale());
export const currentLang = currentLocale;

export async function setLocale(locale: string) {
  selectedLangSetting.value = locale;
  let targetLocale = locale;

  if (locale === 'auto') {
    targetLocale = detectSystemLocale();
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('unigo_locale', 'auto');
    }
  } else {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('unigo_locale', locale);
    }
  }

  if (!SUPPORTED_LANGUAGES.some(l => l.code === targetLocale)) {
    targetLocale = DEFAULT_LOCALE;
  }

  if (!loadedDictionaries.value[targetLocale]) {
    const loader = localeLoaders[`./locales/${targetLocale}.ts`];
    if (loader) {
      try {
        const mod = await loader();
        const exportKey = Object.keys(mod).find(k => k !== 'default') || Object.keys(mod)[0];
        if (exportKey && mod[exportKey]) {
          loadedDictionaries.value[targetLocale] = mod[exportKey];
        }
      } catch (e) {
        console.error(`Failed to load locale chunk for ${targetLocale}:`, e);
      }
    }
  }

  currentLocale.value = targetLocale;
  updateDocumentDir();
}

export const setLanguage = setLocale;

export const isRtl = computed(() => {
  return ['ar-SA', 'he-IL', 'fa-IR', 'ur-PK'].includes(currentLocale.value);
});

export function updateDocumentDir() {
  if (typeof document !== 'undefined') {
    const dir = isRtl.value ? 'rtl' : 'ltr';
    document.documentElement.setAttribute('dir', dir);
    document.documentElement.dir = dir;
  }
}

// Asynchronously load initial locale if not pre-bundled
const initLoc = getInitialLocale();
if (initLoc !== 'zh-CN' && initLoc !== 'en-US') {
  setLocale(selectedLangSetting.value === 'auto' ? 'auto' : initLoc);
} else {
  updateDocumentDir();
}

export function t(key: keyof TranslationDict, params?: Record<string, string | number>): string {
  const dict = loadedDictionaries.value[currentLocale.value] || loadedDictionaries.value[DEFAULT_LOCALE] || zhCn;
  let text = dict[key] || loadedDictionaries.value[DEFAULT_LOCALE]?.[key] || zhCn[key] || key;

  if (params) {
    Object.keys(params).forEach(pKey => {
      text = text.replace(new RegExp(`{\s*${pKey}\s*}`, 'g'), String(params[pKey]));
    });
  }

  return text;
}
