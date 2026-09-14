import { ref, computed } from 'vue';
import type { TranslationDict } from './types';
export type { TranslationDict };

import { zhCn } from './locales/zh-CN';
import { enUs } from './locales/en-US';
import { zhTw } from './locales/zh-TW';
import { jaJp } from './locales/ja-JP';
import { koKr } from './locales/ko-KR';
import { deDe } from './locales/de-DE';
import { frFr } from './locales/fr-FR';
import { esEs } from './locales/es-ES';
import { esLa } from './locales/es-LA';
import { ruRu } from './locales/ru-RU';
import { ptBr } from './locales/pt-BR';
import { ptPt } from './locales/pt-PT';
import { itIt } from './locales/it-IT';
import { trTr } from './locales/tr-TR';
import { plPl } from './locales/pl-PL';
import { viVn } from './locales/vi-VN';
import { arSa } from './locales/ar-SA';
import { urPk } from './locales/ur-PK';
import { azAz } from './locales/az-AZ';
import { daDk } from './locales/da-DK';
import { kaGe } from './locales/ka-GE';
import { faIr } from './locales/fa-IR';
import { slSi } from './locales/sl-SI';
import { ocFr } from './locales/oc-FR';
import { csCz } from './locales/cs-CZ';
import { skSk } from './locales/sk-SK';
import { bnBd } from './locales/bn-BD';
import { hiIn } from './locales/hi-IN';
import { nlNl } from './locales/nl-NL';
import { roRo } from './locales/ro-RO';
import { hrHr } from './locales/hr-HR';
import { huHu } from './locales/hu-HU';
import { srLatn } from './locales/sr-Latn';
import { srCyrl } from './locales/sr-Cyrl';
import { thTh } from './locales/th-TH';
import { noNo } from './locales/no-NO';
import { ltLt } from './locales/lt-LT';
import { mkMk } from './locales/mk-MK';
import { heIl } from './locales/he-IL';
import { idId } from './locales/id-ID';
import { nbNo } from './locales/nb-NO';
import { ukUa } from './locales/uk-UA';
import { elGr } from './locales/el-GR';
import { svSe } from './locales/sv-SE';
import { bgBg } from './locales/bg-BG';
import { hyAm } from './locales/hy-AM';
import { fiFi } from './locales/fi-FI';
import { glEs } from './locales/gl-ES';
import { caEs } from './locales/ca-ES';
import { taIn } from './locales/ta-IN';
import { beBy } from './locales/be-BY';
import { mlIn } from './locales/ml-IN';
import { etEe } from './locales/et-EE';


export const SUPPORTED_LANGUAGES = [
  { code: 'zh-CN', name: '简体中文', flag: '🇨🇳' },
  { code: 'en-US', name: 'English', flag: '🇺🇸' },
  { code: 'zh-TW', name: '正體中文', flag: '🇭🇰' },
  { code: 'ja-JP', name: '日本語', flag: '🇯🇵' },
  { code: 'ko-KR', name: '한국어', flag: '🇰🇷' },
  { code: 'de-DE', name: 'Deutsch', flag: '🇩🇪' },
  { code: 'fr-FR', name: 'Français', flag: '🇫🇷' },
  { code: 'es-ES', name: 'Español (España)', flag: '🇪🇸' },
  { code: 'es-LA', name: 'Español (Latinoamérica)', flag: '🇲🇽' },
  { code: 'ru-RU', name: 'Русский', flag: '🇷🇺' },
  { code: 'pt-BR', name: 'Português (Brasil)', flag: '🇧🇷' },
  { code: 'pt-PT', name: 'Português (Portugal)', flag: '🇵🇹' },
  { code: 'it-IT', name: 'Italiano', flag: '🇮🇹' },
  { code: 'tr-TR', name: 'Türkçe', flag: '🇹🇷' },
  { code: 'pl-PL', name: 'Polski', flag: '🇵🇱' },
  { code: 'vi-VN', name: 'Tiếng Việt', flag: '🇻🇳' },
  { code: 'ar-SA', name: 'العربية', flag: '🇸🇦' },
  { code: 'ur-PK', name: 'اردو', flag: '🇵🇰' },
  { code: 'az-AZ', name: 'Azərbaycanca', flag: '🇦🇿' },
  { code: 'da-DK', name: 'Dansk', flag: '🇩🇰' },
  { code: 'ka-GE', name: 'ქართული', flag: '🇬🇪' },
  { code: 'fa-IR', name: 'فارسی', flag: '🇮🇷' },
  { code: 'sl-SI', name: 'Slovenščina', flag: '🇸🇮' },
  { code: 'oc-FR', name: 'Occitan', flag: '🇫🇷' },
  { code: 'cs-CZ', name: 'Čeština', flag: '🇨🇿' },
  { code: 'sk-SK', name: 'Slovenčina', flag: '🇸🇰' },
  { code: 'bn-BD', name: 'বাংলা', flag: '🇧🇩' },
  { code: 'hi-IN', name: 'हिन्दी', flag: '🇮🇳' },
  { code: 'nl-NL', name: 'Nederlands', flag: '🇳🇱' },
  { code: 'ro-RO', name: 'Română', flag: '🇷🇴' },
  { code: 'hr-HR', name: 'Hrvatski', flag: '🇭🇷' },
  { code: 'hu-HU', name: 'Magyar', flag: '🇭🇺' },
  { code: 'sr-Latn', name: 'Srpski', flag: '🇷🇸' },
  { code: 'sr-Cyrl', name: 'Српски', flag: '🇷🇸' },
  { code: 'th-TH', name: 'ไทย', flag: '🇹🇭' },
  { code: 'no-NO', name: 'Norsk', flag: '🇳🇴' },
  { code: 'lt-LT', name: 'Lietuvių', flag: '🇱🇹' },
  { code: 'mk-MK', name: 'Македонски', flag: '🇲🇰' },
  { code: 'he-IL', name: 'עברית', flag: '🇮🇱' },
  { code: 'id-ID', name: 'Bahasa Indonesia', flag: '🇮🇩' },
  { code: 'nb-NO', name: 'Norsk Bokmål', flag: '🇳🇴' },
  { code: 'uk-UA', name: 'Українська', flag: '🇺🇦' },
  { code: 'el-GR', name: 'Ελληνικά', flag: '🇬🇷' },
  { code: 'sv-SE', name: 'Svenska', flag: '🇸🇪' },
  { code: 'bg-BG', name: 'Български', flag: '🇧🇬' },
  { code: 'hy-AM', name: 'Հայերեն', flag: '🇦🇲' },
  { code: 'fi-FI', name: 'Suomi', flag: '🇫🇮' },
  { code: 'gl-ES', name: 'Galego', flag: '🇪🇸' },
  { code: 'ca-ES', name: 'Català', flag: '🇪🇸' },
  { code: 'ta-IN', name: 'தமிழ்', flag: '🇮🇳' },
  { code: 'be-BY', name: 'Беларуская', flag: '🇧🇾' },
  { code: 'ml-IN', name: 'മലയാളം', flag: '🇮🇳' },
  { code: 'et-EE', name: 'Eesti', flag: '🇪🇪' },
];


const dictionaries: Record<string, TranslationDict> = {
  'zh-CN': zhCn,
  'en-US': enUs,
  'zh-TW': zhTw,
  'ja-JP': jaJp,
  'ko-KR': koKr,
  'de-DE': deDe,
  'fr-FR': frFr,
  'es-ES': esEs,
  'es-LA': esLa,
  'ru-RU': ruRu,
  'pt-BR': ptBr,
  'pt-PT': ptPt,
  'it-IT': itIt,
  'tr-TR': trTr,
  'pl-PL': plPl,
  'vi-VN': viVn,
  'ar-SA': arSa,
  'ur-PK': urPk,
  'az-AZ': azAz,
  'da-DK': daDk,
  'ka-GE': kaGe,
  'fa-IR': faIr,
  'sl-SI': slSi,
  'oc-FR': ocFr,
  'cs-CZ': csCz,
  'sk-SK': skSk,
  'bn-BD': bnBd,
  'hi-IN': hiIn,
  'nl-NL': nlNl,
  'ro-RO': roRo,
  'hr-HR': hrHr,
  'hu-HU': huHu,
  'sr-Latn': srLatn,
  'sr-Cyrl': srCyrl,
  'th-TH': thTh,
  'no-NO': noNo,
  'lt-LT': ltLt,
  'mk-MK': mkMk,
  'he-IL': heIl,
  'id-ID': idId,
  'nb-NO': nbNo,
  'uk-UA': ukUa,
  'el-GR': elGr,
  'sv-SE': svSe,
  'bg-BG': bgBg,
  'hy-AM': hyAm,
  'fi-FI': fiFi,
  'gl-ES': glEs,
  'ca-ES': caEs,
  'ta-IN': taIn,
  'be-BY': beBy,
  'ml-IN': mlIn,
  'et-EE': etEe,
};


const DEFAULT_LOCALE = 'zh-CN';

function getInitialLocale(): string {
  const saved = localStorage.getItem('unigo_locale');
  if (saved && dictionaries[saved]) {
    return saved;
  }
  const navLang = navigator.language;
  if (dictionaries[navLang]) {
    return navLang;
  }
  const shortLang = navLang.split('-')[0];
  const matched = SUPPORTED_LANGUAGES.find(l => l.code.startsWith(shortLang));
  if (matched) {
    return matched.code;
  }
  return DEFAULT_LOCALE;
}

export const currentLocale = ref<string>(getInitialLocale());

export function setLocale(locale: string) {
  if (dictionaries[locale]) {
    currentLocale.value = locale;
    localStorage.setItem('unigo_locale', locale);
    updateDocumentDir();
  }
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
  const dict = dictionaries[currentLocale.value] || dictionaries[DEFAULT_LOCALE];
  let text = dict[key] || dictionaries[DEFAULT_LOCALE][key] || key;

  if (params) {
    Object.keys(params).forEach(pKey => {
      text = text.replace(new RegExp(`{\s*${pKey}\s*}`, 'g'), String(params[pKey]));
    });
  }

  return text;
}
