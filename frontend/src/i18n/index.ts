import { ref, computed } from 'vue';
import type { TranslationDict } from './types';
export type { TranslationDict };

import { zhCn } from './locales/zh-CN';
import { enUs } from './locales/en-US';

export const SUPPORTED_LANGUAGES = [
  { code: 'zh-CN', name: '简体中文', nativeName: '简体中文', flag: '🇨🇳' },
  { code: 'en-US', name: 'English', nativeName: 'English', flag: '🇺🇸' },
  { code: 'zh-TW', name: '正體中文', nativeName: '正體中文', flag: '🇭🇰' },
  { code: 'ja-JP', name: '日本語', nativeName: '日本語', flag: '🇯🇵' },
  { code: 'ko-KR', name: '한국어', nativeName: '한국어', flag: '🇰🇷' },
  { code: 'de-DE', name: 'Deutsch', nativeName: 'Deutsch', flag: '🇩🇪' },
  { code: 'fr-FR', name: 'Français', nativeName: 'Français', flag: '🇫🇷' },
  { code: 'es-ES', name: 'Español (España)', nativeName: 'Español (España)', flag: '🇪🇸' },
  { code: 'es-LA', name: 'Español (Latinoamérica)', nativeName: 'Español (Latinoamérica)', flag: '🇲🇽' },
  { code: 'ru-RU', name: 'Русский', nativeName: 'Русский', flag: '🇷🇺' },
  { code: 'pt-BR', name: 'Português (Brasil)', nativeName: 'Português (Brasil)', flag: '🇧🇷' },
  { code: 'pt-PT', name: 'Português (Portugal)', nativeName: 'Português (Portugal)', flag: '🇵🇹' },
  { code: 'it-IT', name: 'Italiano', nativeName: 'Italiano', flag: '🇮🇹' },
  { code: 'tr-TR', name: 'Türkçe', nativeName: 'Türkçe', flag: '🇹🇷' },
  { code: 'pl-PL', name: 'Polski', nativeName: 'Polski', flag: '🇵🇱' },
  { code: 'vi-VN', name: 'Tiếng Việt', nativeName: 'Tiếng Việt', flag: '🇻🇳' },
  { code: 'ar-SA', name: 'العربية', nativeName: 'العربية', flag: '🇸🇦' },
  { code: 'ur-PK', name: 'اردو', nativeName: 'اردو', flag: '🇵🇰' },
  { code: 'az-AZ', name: 'Azərbaycanca', nativeName: 'Azərbaycanca', flag: '🇦🇿' },
  { code: 'da-DK', name: 'Dansk', nativeName: 'Dansk', flag: '🇩🇰' },
  { code: 'ka-GE', name: 'ქართული', nativeName: 'ქართული', flag: '🇬🇪' },
  { code: 'fa-IR', name: 'فارسی', nativeName: 'فارسی', flag: '🇮🇷' },
  { code: 'sl-SI', name: 'Slovenščina', nativeName: 'Slovenščina', flag: '🇸🇮' },
  { code: 'oc-FR', name: 'Occitan', nativeName: 'Occitan', flag: '🇫🇷' },
  { code: 'cs-CZ', name: 'Čeština', nativeName: 'Čeština', flag: '🇨🇿' },
  { code: 'sk-SK', name: 'Slovenčina', nativeName: 'Slovenčina', flag: '🇸🇰' },
  { code: 'bn-BD', name: 'বাংলা', nativeName: 'বাংলা', flag: '🇧🇩' },
  { code: 'hi-IN', name: 'हिन्दी', nativeName: 'हिन्दी', flag: '🇮🇳' },
  { code: 'nl-NL', name: 'Nederlands', nativeName: 'Nederlands', flag: '🇳🇱' },
  { code: 'ro-RO', name: 'Română', nativeName: 'Română', flag: '🇷🇴' },
  { code: 'hr-HR', name: 'Hrvatski', nativeName: 'Hrvatski', flag: '🇭🇷' },
  { code: 'hu-HU', name: 'Magyar', nativeName: 'Magyar', flag: '🇭🇺' },
  { code: 'sr-Latn', name: 'Srpski', nativeName: 'Srpski', flag: '🇷🇸' },
  { code: 'sr-Cyrl', name: 'Српски', nativeName: 'Српски', flag: '🇷🇸' },
  { code: 'th-TH', name: 'ไทย', nativeName: 'ไทย', flag: '🇹🇭' },
  { code: 'no-NO', name: 'Norsk', nativeName: 'Norsk', flag: '🇳🇴' },
  { code: 'lt-LT', name: 'Lietuvių', nativeName: 'Lietuvių', flag: '🇱🇹' },
  { code: 'mk-MK', name: 'Македонски', nativeName: 'Македонски', flag: '🇲🇰' },
  { code: 'he-IL', name: 'עברית', nativeName: 'עברית', flag: '🇮🇱' },
  { code: 'id-ID', name: 'Bahasa Indonesia', nativeName: 'Bahasa Indonesia', flag: '🇮🇩' },
  { code: 'nb-NO', name: 'Norsk Bokmål', nativeName: 'Norsk Bokmål', flag: '🇳🇴' },
  { code: 'uk-UA', name: 'Українська', nativeName: 'Українська', flag: '🇺🇦' },
  { code: 'el-GR', name: 'Ελληνικά', nativeName: 'Ελληνικά', flag: '🇬🇷' },
  { code: 'sv-SE', name: 'Svenska', nativeName: 'Svenska', flag: '🇸🇪' },
  { code: 'bg-BG', name: 'Български', nativeName: 'Български', flag: '🇧🇬' },
  { code: 'hy-AM', name: 'Հայերեն', nativeName: 'Հայերեն', flag: '🇦🇲' },
  { code: 'fi-FI', name: 'Suomi', nativeName: 'Suomi', flag: '🇫🇮' },
  { code: 'gl-ES', name: 'Galego', nativeName: 'Galego', flag: '🇪🇸' },
  { code: 'ca-ES', name: 'Català', nativeName: 'Català', flag: '🇪🇸' },
  { code: 'ta-IN', name: 'தமிழ்', nativeName: 'தமிழ்', flag: '🇮🇳' },
  { code: 'be-BY', name: 'Беларуская', nativeName: 'Беларуская', flag: '🇧🇾' },
  { code: 'ml-IN', name: 'മലയാളം', nativeName: 'മലയാളം', flag: '🇮🇳' },
  { code: 'et-EE', name: 'Eesti', nativeName: 'Eesti', flag: '🇪🇪' },
];


// Vite glob import for dynamic lazy-loading locale chunks (excluding pre-bundled zh-CN and en-US)
const localeLoaders = import.meta.glob<Record<string, any>>(['./locales/*.ts', '!./locales/zh-CN.ts', '!./locales/en-US.ts']);

// Reactive map of loaded locale dictionaries
const loadedDictionaries = ref<Record<string, TranslationDict>>({
  'zh-CN': zhCn,
  'en-US': enUs,
});

const DEFAULT_LOCALE = 'zh-CN';

function getInitialLocale(): string {
  const saved = localStorage.getItem('unigo_locale');
  if (saved && SUPPORTED_LANGUAGES.some(l => l.code === saved)) {
    return saved;
  }
  const navLang = navigator.language;
  const matched = SUPPORTED_LANGUAGES.find(l => l.code === navLang || l.code.startsWith(navLang.split('-')[0]));
  if (matched) {
    return matched.code;
  }
  return DEFAULT_LOCALE;
}

export const currentLocale = ref<string>(getInitialLocale());
export const currentLang = currentLocale;

export async function setLocale(locale: string) {
  if (!SUPPORTED_LANGUAGES.some(l => l.code === locale)) return;

  if (!loadedDictionaries.value[locale]) {
    const loader = localeLoaders[`./locales/${locale}.ts`];
    if (loader) {
      try {
        const mod = await loader();
        const exportKey = Object.keys(mod).find(k => k !== 'default') || Object.keys(mod)[0];
        if (exportKey && mod[exportKey]) {
          loadedDictionaries.value[locale] = mod[exportKey];
        }
      } catch (e) {
        console.error(`Failed to load locale chunk for ${locale}:`, e);
      }
    }
  }

  currentLocale.value = locale;
  localStorage.setItem('unigo_locale', locale);
  updateDocumentDir();
}

export const setLanguage = setLocale;

// Asynchronously load initial locale if not pre-bundled
const initLoc = getInitialLocale();
if (initLoc !== 'zh-CN' && initLoc !== 'en-US') {
  setLocale(initLoc);
}

export const isRtl = computed(() => {
  return ['ar-SA', 'he-IL', 'fa-IR', 'ur-PK'].includes(currentLocale.value);
});

export function updateDocumentDir() {
  if (typeof document !== 'undefined') {
    if (isRtl.value) {
      document.documentElement.setAttribute('dir', 'rtl');
    } else {
      document.documentElement.removeAttribute('dir');
    }
  }
}

updateDocumentDir();

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
