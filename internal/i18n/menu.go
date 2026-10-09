// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package i18n

import (
	"strings"

	"github.com/snowdreamtech/unibootdesktop/internal/env"
	"github.com/snowdreamtech/unibootdesktop/pkg/config"
)

// MenuTranslations holds localized titles for OS native menu bar items.
type MenuTranslations struct {
	App       string
	About     string
	Hide      string
	ShowAll   string
	Quit      string
	Edit      string
	Undo      string
	Redo      string
	Cut       string
	Copy      string
	Paste     string
	SelectAll string
	Window    string
	Minimize  string
	Zoom      string
	Help      string
}

// menuDataStores stores localized menu titles across all 53 supported locales.
var menuDataStores = map[string]MenuTranslations{
	"zh-CN": {
		App: "UniBootDesktop", About: "关于 UniBootDesktop", Hide: "隐藏 UniBootDesktop", ShowAll: "显示全部",
		Quit: "退出 UniBootDesktop", Edit: "编辑", Undo: "撤销", Redo: "重做",
		Cut: "剪切", Copy: "复制", Paste: "粘贴", SelectAll: "全选",
		Window: "窗口", Minimize: "最小化", Zoom: "缩放", Help: "帮助",
	},
	"zh-TW": {
		App: "UniBootDesktop", About: "關於 UniBootDesktop", Hide: "隱藏 UniBootDesktop", ShowAll: "顯示全部",
		Quit: "結束 UniBootDesktop", Edit: "編輯", Undo: "復原", Redo: "重做",
		Cut: "剪下", Copy: "複製", Paste: "貼上", SelectAll: "全選",
		Window: "視窗", Minimize: "縮到最小", Zoom: "縮放", Help: "說明",
	},
	"en-US": {
		App: "UniBootDesktop", About: "About UniBootDesktop", Hide: "Hide UniBootDesktop", ShowAll: "Show All",
		Quit: "Quit UniBootDesktop", Edit: "Edit", Undo: "Undo", Redo: "Redo",
		Cut: "Cut", Copy: "Copy", Paste: "Paste", SelectAll: "Select All",
		Window: "Window", Minimize: "Minimize", Zoom: "Zoom", Help: "Help",
	},
	"de-DE": {
		App: "UniBootDesktop", About: "Über UniBootDesktop", Hide: "UniBootDesktop ausblenden", ShowAll: "Alle einblenden",
		Quit: "UniBootDesktop beenden", Edit: "Bearbeiten", Undo: "Rückgängig", Redo: "Wiederholen",
		Cut: "Ausschneiden", Copy: "Kopieren", Paste: "Einfügen", SelectAll: "Alles auswählen",
		Window: "Fenster", Minimize: "Im Dock ablegen", Zoom: "Zoomen", Help: "Hilfe",
	},
	"fr-FR": {
		App: "UniBootDesktop", About: "À propos de UniBootDesktop", Hide: "Masquer UniBootDesktop", ShowAll: "Tout afficher",
		Quit: "Quitter UniBootDesktop", Edit: "Édition", Undo: "Annuler", Redo: "Rétablir",
		Cut: "Couper", Copy: "Copier", Paste: "Coller", SelectAll: "Tout sélectionner",
		Window: "Fenêtre", Minimize: "Réduire", Zoom: "Zoomer", Help: "Aide",
	},
	"es-ES": {
		App: "UniBootDesktop", About: "Acerca de UniBootDesktop", Hide: "Ocultar UniBootDesktop", ShowAll: "Mostrar todo",
		Quit: "Salir de UniBootDesktop", Edit: "Edición", Undo: "Deshacer", Redo: "Rehacer",
		Cut: "Cortar", Copy: "Copiar", Paste: "Pegar", SelectAll: "Seleccionar todo",
		Window: "Ventana", Minimize: "Minimizar", Zoom: "Zoom", Help: "Ayuda",
	},
	"es-LA": {
		App: "UniBootDesktop", About: "Acerca de UniBootDesktop", Hide: "Ocultar UniBootDesktop", ShowAll: "Mostrar todo",
		Quit: "Salir de UniBootDesktop", Edit: "Edición", Undo: "Deshacer", Redo: "Rehacer",
		Cut: "Cortar", Copy: "Copiar", Paste: "Pegar", SelectAll: "Seleccionar todo",
		Window: "Ventana", Minimize: "Minimizar", Zoom: "Zoom", Help: "Ayuda",
	},
	"it-IT": {
		App: "UniBootDesktop", About: "Informazioni su UniBootDesktop", Hide: "Nascondi UniBootDesktop", ShowAll: "Mostra tutte",
		Quit: "Esci da UniBootDesktop", Edit: "Composizione", Undo: "Annulla", Redo: "Ripristina",
		Cut: "Taglia", Copy: "Copia", Paste: "Incolla", SelectAll: "Seleziona tutto",
		Window: "Finestra", Minimize: "Contrai", Zoom: "Ridimensiona", Help: "Aiuto",
	},
	"ja-JP": {
		App: "UniBootDesktop", About: "UniBootDesktop について", Hide: "UniBootDesktop を非表示", ShowAll: "すべてを表示",
		Quit: "UniBootDesktop を終了", Edit: "編集", Undo: "元に戻す", Redo: "やり直す",
		Cut: "切り取り", Copy: "コピー", Paste: "貼り付け", SelectAll: "すべてを選択",
		Window: "ウィンドウ", Minimize: "最小化", Zoom: "拡大/縮小", Help: "ヘルプ",
	},
	"ko-KR": {
		App: "UniBootDesktop", About: "UniBootDesktop 정보", Hide: "UniBootDesktop 가리기", ShowAll: "모두 보기",
		Quit: "UniBootDesktop 종료", Edit: "편집", Undo: "실행 취소", Redo: "다시 실행",
		Cut: "오려두기", Copy: "복사", Paste: "붙여넣기", SelectAll: "전체 선택",
		Window: "윈도우", Minimize: "최소화", Zoom: "확대/축소", Help: "도움말",
	},
	"ru-RU": {
		App: "UniBootDesktop", About: "О программе UniBootDesktop", Hide: "Скрыть UniBootDesktop", ShowAll: "Показать все",
		Quit: "Завершить UniBootDesktop", Edit: "Правка", Undo: "Отменить", Redo: "Повторить",
		Cut: "Вырезать", Copy: "Скопировать", Paste: "Вставить", SelectAll: "Выделить все",
		Window: "Окно", Minimize: "Свернуть", Zoom: "Изменить масштаб", Help: "Справка",
	},
	"pt-BR": {
		App: "UniBootDesktop", About: "Sobre o UniBootDesktop", Hide: "Ocultar UniBootDesktop", ShowAll: "Mostrar Tudo",
		Quit: "Encerrar o UniBootDesktop", Edit: "Editar", Undo: "Desfazer", Redo: "Refazer",
		Cut: "Recortar", Copy: "Copiar", Paste: "Colar", SelectAll: "Selecionar Tudo",
		Window: "Janela", Minimize: "Minimizar", Zoom: "Deminuir/Aumentar Zoom", Help: "Ajuda",
	},
	"pt-PT": {
		App: "UniBootDesktop", About: "Sobre o UniBootDesktop", Hide: "Ocultar UniBootDesktop", ShowAll: "Mostrar Tudo",
		Quit: "Encerrar o UniBootDesktop", Edit: "Edição", Undo: "Desfazer", Redo: "Refazer",
		Cut: "Cortar", Copy: "Copiar", Paste: "Colar", SelectAll: "Selecionar Tudo",
		Window: "Janela", Minimize: "Minimizar", Zoom: "Redimensionar", Help: "Ajuda",
	},
	"ar-SA": {
		App: "UniBootDesktop", About: "حول UniBootDesktop", Hide: "إخفاء UniBootDesktop", ShowAll: "إظهار الكل",
		Quit: "إنهاء UniBootDesktop", Edit: "تعديل", Undo: "تراجع", Redo: "إعادة",
		Cut: "قص", Copy: "نسخ", Paste: "لصق", SelectAll: "تحديد الكل",
		Window: "نافذة", Minimize: "تصغير", Zoom: "تكبير/تصغير", Help: "مساعدة",
	},
	"vi-VN": {
		App: "UniBootDesktop", About: "Về UniBootDesktop", Hide: "Ẩn UniBootDesktop", ShowAll: "Hiện tất cả",
		Quit: "Thoát UniBootDesktop", Edit: "Chỉnh sửa", Undo: "Hoàn tác", Redo: "Làm lại",
		Cut: "Cắt", Copy: "Sao chép", Paste: "Dán", SelectAll: "Chọn tất cả",
		Window: "Cửa sổ", Minimize: "Thu nhỏ", Zoom: "Phóng to/Thu nhỏ", Help: "Trợ giúp",
	},
	"tr-TR": {
		App: "UniBootDesktop", About: "UniBootDesktop Hakkında", Hide: "UniBootDesktop'ı Gizle", ShowAll: "Tümünü Göster",
		Quit: "UniBootDesktop'tan Çık", Edit: "Düzenle", Undo: "Geri Al", Redo: "Yinele",
		Cut: "Kes", Copy: "Kopyala", Paste: "Yapıştır", SelectAll: "Tümünü Seç",
		Window: "Pencere", Minimize: "Simge Durumuna Küçült", Zoom: "Büyüt/Küçült", Help: "Yardım",
	},
	"pl-PL": {
		App: "UniBootDesktop", About: "O UniBootDesktop", Hide: "Ukryj UniBootDesktop", ShowAll: "Pokaż wszystkie",
		Quit: "Zakończ UniBootDesktop", Edit: "Edycja", Undo: "Cofnij", Redo: "Przywróć",
		Cut: "Wytnij", Copy: "Kopiuj", Paste: "Wklej", SelectAll: "Zaznacz wszystko",
		Window: "Okno", Minimize: "Zminimalizuj", Zoom: "Wypełnij", Help: "Pomoc",
	},
	"nl-NL": {
		App: "UniBootDesktop", About: "Over UniBootDesktop", Hide: "Verberg UniBootDesktop", ShowAll: "Toon alles",
		Quit: "UniBootDesktop sluiten", Edit: "Wijzig", Undo: "Herstel", Redo: "Opnieuw",
		Cut: "Knippen", Copy: "Kopiëren", Paste: "Plakken", SelectAll: "Selecteer alles",
		Window: "Venster", Minimize: "Minimiseer", Zoom: "Zoom", Help: "Help",
	},
	"sv-SE": {
		App: "UniBootDesktop", About: "Om UniBootDesktop", Hide: "Göm UniBootDesktop", ShowAll: "Visa alla",
		Quit: "Avsluta UniBootDesktop", Edit: "Redigera", Undo: "Ångra", Redo: "Gör om",
		Cut: "Klipp ut", Copy: "Kopiera", Paste: "Klistra in", SelectAll: "Markera allt",
		Window: "Fönster", Minimize: "Minimera", Zoom: "Zooma", Help: "Hjälp",
	},
	"da-DK": {
		App: "UniBootDesktop", About: "Om UniBootDesktop", Hide: "Skjul UniBootDesktop", ShowAll: "Vis alle",
		Quit: "Slut UniBootDesktop", Edit: "Rediger", Undo: "Fortryd", Redo: "Gentag",
		Cut: "Klip", Copy: "Kopier", Paste: "Sæt ind", SelectAll: "Vælg alt",
		Window: "Vindue", Minimize: "Minimer", Zoom: "Zoom", Help: "Hjælp",
	},
	"nb-NO": {
		App: "UniBootDesktop", About: "Om UniBootDesktop", Hide: "Skjul UniBootDesktop", ShowAll: "Vis alle",
		Quit: "Avslutt UniBootDesktop", Edit: "Rediger", Undo: "Angre", Redo: "Gjør om",
		Cut: "Klipp ut", Copy: "Kopier", Paste: "Lim inn", SelectAll: "Marker alt",
		Window: "Vindu", Minimize: "Minimer", Zoom: "Zoom", Help: "Hjelp",
	},
	"no-NO": {
		App: "UniBootDesktop", About: "Om UniBootDesktop", Hide: "Skjul UniBootDesktop", ShowAll: "Vis alle",
		Quit: "Avslutt UniBootDesktop", Edit: "Rediger", Undo: "Angre", Redo: "Gjør om",
		Cut: "Klipp ut", Copy: "Kopier", Paste: "Lim inn", SelectAll: "Marker alt",
		Window: "Vindu", Minimize: "Minimer", Zoom: "Zoom", Help: "Hjelp",
	},
	"fi-FI": {
		App: "UniBootDesktop", About: "Tietoja UniBootDesktopista", Hide: "Kätke UniBootDesktop", ShowAll: "Näytä kaikki",
		Quit: "Lopeta UniBootDesktop", Edit: "Muokkaa", Undo: "Peru", Redo: "Tee uudelleen",
		Cut: "Leikkaa", Copy: "Kopioi", Paste: "Sijoita", SelectAll: "Valitse kaikki",
		Window: "Ikkuna", Minimize: "Pienennä", Zoom: "Zoomaa", Help: "Ohje",
	},
	"cs-CZ": {
		App: "UniBootDesktop", About: "O aplikaci UniBootDesktop", Hide: "Skrýt UniBootDesktop", ShowAll: "Zobrazit vše",
		Quit: "Ukončit UniBootDesktop", Edit: "Úpravy", Undo: "Zpět", Redo: "Znovu",
		Cut: "Vyjmout", Copy: "Kopírovat", Paste: "Vložit", SelectAll: "Vybrat vše",
		Window: "Okno", Minimize: "Minimalizovat", Zoom: "Přiblížit", Help: "Nápověda",
	},
	"sk-SK": {
		App: "UniBootDesktop", About: "O aplikácii UniBootDesktop", Hide: "Skryť UniBootDesktop", ShowAll: "Zobraziť všetky",
		Quit: "Ukončiť UniBootDesktop", Edit: "Upraviť", Undo: "Späť", Redo: "Znovu",
		Cut: "Vystrihnúť", Copy: "Kopírovať", Paste: "Prilepiť", SelectAll: "Vybrať všetko",
		Window: "Okno", Minimize: "Minimalizovať", Zoom: "Zväčšiť", Help: "Pomocník",
	},
	"hu-HU": {
		App: "UniBootDesktop", About: "A UniBootDesktop névjegye", Hide: "A UniBootDesktop elrejtése", ShowAll: "Összes megjelenítése",
		Quit: "Kilépés a UniBootDesktopból", Edit: "Szerkesztés", Undo: "Visszavonás", Redo: "Ismétlés",
		Cut: "Kivágás", Copy: "Másolás", Paste: "Beillesztés", SelectAll: "Összes kijelölése",
		Window: "Ablak", Minimize: "Kis méret", Zoom: "Nagyítás", Help: "Súgó",
	},
	"ro-RO": {
		App: "UniBootDesktop", About: "Despre UniBootDesktop", Hide: "Ascunde UniBootDesktop", ShowAll: "Afișează toate",
		Quit: "Închide UniBootDesktop", Edit: "Editare", Undo: "Anulează", Redo: "Refă",
		Cut: "Tunde", Copy: "Copiază", Paste: "Lipește", SelectAll: "Selectează tot",
		Window: "Fereastră", Minimize: "Minimizează", Zoom: "Zoom", Help: "Ajutor",
	},
	"bg-BG": {
		App: "UniBootDesktop", About: "Относно UniBootDesktop", Hide: "Скриване на UniBootDesktop", ShowAll: "Показване на всички",
		Quit: "Изход от UniBootDesktop", Edit: "Редактиране", Undo: "Отмяна", Redo: "Повторение",
		Cut: "Изрязване", Copy: "Копиране", Paste: "Поставяне", SelectAll: "Избиране на всички",
		Window: "Прозорец", Minimize: "Минимизиране", Zoom: "Мащабиране", Help: "Помощ",
	},
	"uk-UA": {
		App: "UniBootDesktop", About: "Про програму UniBootDesktop", Hide: "Сховати UniBootDesktop", ShowAll: "Показати всі",
		Quit: "Завершити UniBootDesktop", Edit: "Редагування", Undo: "Скасувати", Redo: "Повторити",
		Cut: "Вирізати", Copy: "Копіювати", Paste: "Вставити", SelectAll: "Виділити все",
		Window: "Вікно", Minimize: "Згорнути", Zoom: "Масштаб", Help: "Довідка",
	},
	"el-GR": {
		App: "UniBootDesktop", About: "Σχετικά με το UniBootDesktop", Hide: "Απόκρυψη UniBootDesktop", ShowAll: "Εμφάνιση όλων",
		Quit: "Έξοδος από UniBootDesktop", Edit: "Επεξεργασία", Undo: "Αναίρεση", Redo: "Επανάληψη",
		Cut: "Αποκοπή", Copy: "Αντιγραφή", Paste: "Επικόλληση", SelectAll: "Επιλογή όλων",
		Window: "Παράθυρο", Minimize: "Ελαχιστοποίηση", Zoom: "Εστίαση", Help: "Βοήθεια",
	},
	"he-IL": {
		App: "UniBootDesktop", About: "על אודות UniBootDesktop", Hide: "הסתר את UniBootDesktop", ShowAll: "הצג הכל",
		Quit: "סיום UniBootDesktop", Edit: "עריכה", Undo: "בטל", Redo: "בצע שוב",
		Cut: "גזור", Copy: "העתק", Paste: "הדבק", SelectAll: "בחר הכל",
		Window: "חלון", Minimize: "מזער", Zoom: "זום", Help: "עזרה",
	},
	"fa-IR": {
		App: "UniBootDesktop", About: "درباره UniBootDesktop", Hide: "پنهان کردن UniBootDesktop", ShowAll: "نمایش همه",
		Quit: "خروج از UniBootDesktop", Edit: "ویرایش", Undo: "واکشی", Redo: "انجام دوباره",
		Cut: "برش", Copy: "کپی", Paste: "جای‌گذاری", SelectAll: "انتخاب همه",
		Window: "پنجره", Minimize: "کمینه کردن", Zoom: "بزرگ‌نمایی", Help: "راهنما",
	},
	"hi-IN": {
		App: "UniBootDesktop", About: "UniBootDesktop के बारे में", Hide: "UniBootDesktop छिपाएं", ShowAll: "सभी दिखाएं",
		Quit: "UniBootDesktop से बाहर निकलें", Edit: "संपादित करें", Undo: "पूर्ववत करें", Redo: "फिर से करें",
		Cut: "कट करें", Copy: "कॉपी करें", Paste: "पेस्ट करें", SelectAll: "सभी चुनें",
		Window: "विंडो", Minimize: "छोटा करें", Zoom: "ज़ूम", Help: "सहायता",
	},
	"bn-BD": {
		App: "UniBootDesktop", About: "UniBootDesktop সম্পর্কে", Hide: "UniBootDesktop লুকান", ShowAll: "সব দেখান",
		Quit: "UniBootDesktop বন্ধ করুন", Edit: "সম্পাদনা", Undo: "পূর্বাবস্থায় ফেরান", Redo: "পুনরায় করুন",
		Cut: "কাট", Copy: "কপি", Paste: "পেস্ট", SelectAll: "সব নির্বাচন করুন",
		Window: "উইন্ডো", Minimize: "ছোট করুন", Zoom: "জুম", Help: "সহায়তা",
	},
	"ta-IN": {
		App: "UniBootDesktop", About: "UniBootDesktop பற்றி", Hide: "UniBootDesktop மறை", ShowAll: "அனைத்தையும் காட்டு",
		Quit: "UniBootDesktop வெளியேறு", Edit: "தொகு", Undo: "செயல் நீக்கு", Redo: "மீண்டும் செய்",
		Cut: "வெட்டு", Copy: "நகலெடு", Paste: "ஒட்டு", SelectAll: "அனைத்தையும் தேர்ந்தெடு",
		Window: "சாளரம்", Minimize: "சிறிதாக்கு", Zoom: "பெரிதாக்கு", Help: "உதவி",
	},
	"ml-IN": {
		App: "UniBootDesktop", About: "UniBootDesktop നെ കുറിച്ച്", Hide: "UniBootDesktop മറയ്ക്കുക", ShowAll: "എല്ലാം കാണിക്കുക",
		Quit: "UniBootDesktop പുറത്തുകടക്കുക", Edit: "എഡിറ്റ് ചെയ്യുക", Undo: "തിരുത്തുക", Redo: "വീണ്ടും ചെയ്യുക",
		Cut: "മുറിക്കുക", Copy: "പകർപ്പുക", Paste: "ഒട്ടിക്കുക", SelectAll: "എല്ലാം തിരഞ്ഞെടുക്കുക",
		Window: "വിൻഡോ", Minimize: "ചെറുതാക്കുക", Zoom: "വലുതാക്കുക", Help: "സഹായം",
	},
	"th-TH": {
		App: "UniBootDesktop", About: "เกี่ยวกับ UniBootDesktop", Hide: "ซ่อน UniBootDesktop", ShowAll: "แสดงทั้งหมด",
		Quit: "ออกจาก UniBootDesktop", Edit: "แก้ไข", Undo: "เลิกทำ", Redo: "ทำซ้ำ",
		Cut: "ตัด", Copy: "คัดลอก", Paste: "วาง", SelectAll: "เลือกทั้งหมด",
		Window: "หน้าต่าง", Minimize: "ย่อหน้าต่าง", Zoom: "ซูม", Help: "ช่วยเหลือ",
	},
	"id-ID": {
		App: "UniBootDesktop", About: "Tentang UniBootDesktop", Hide: "Sembunyikan UniBootDesktop", ShowAll: "Tampilkan Semua",
		Quit: "Keluar UniBootDesktop", Edit: "Edit", Undo: "Batalkan", Redo: "Ulangi",
		Cut: "Potong", Copy: "Salin", Paste: "Tempel", SelectAll: "Pilih Semua",
		Window: "Jendela", Minimize: "Minimalkan", Zoom: "Perbesar", Help: "Bantuan",
	},
	"ca-ES": {
		App: "UniBootDesktop", About: "Quant a UniBootDesktop", Hide: "Amaga UniBootDesktop", ShowAll: "Mostra-ho tot",
		Quit: "Surt de UniBootDesktop", Edit: "Edició", Undo: "Desfés", Redo: "Refés",
		Cut: "Retalla", Copy: "Copia", Paste: "Enganxa", SelectAll: "Selecciona-ho tot",
		Window: "Finestra", Minimize: "Minimitza", Zoom: "Amplia", Help: "Ajuda",
	},
	"gl-ES": {
		App: "UniBootDesktop", About: "Sobre UniBootDesktop", Hide: "Ocultar UniBootDesktop", ShowAll: "Amosar todo",
		Quit: "Saír de UniBootDesktop", Edit: "Editar", Undo: "Desfacer", Redo: "Refacer",
		Cut: "Cortar", Copy: "Copiar", Paste: "Pegar", SelectAll: "Seleccionar todo",
		Window: "Xanela", Minimize: "Minimizar", Zoom: "Ampliar", Help: "Axuda",
	},
	"et-EE": {
		App: "UniBootDesktop", About: "UniBootDesktop teave", Hide: "Peida UniBootDesktop", ShowAll: "Kuva kõik",
		Quit: "Välju UniBootDesktopist", Edit: "Redigeeri", Undo: "Võta tagasi", Redo: "Tee uuesti",
		Cut: "Lõika", Copy: "Kopeeri", Paste: "Aseta", SelectAll: "Vali kõik",
		Window: "Aken", Minimize: "Minimeeri", Zoom: "Suurenda", Help: "Abi",
	},
	"lt-LT": {
		App: "UniBootDesktop", About: "Apie „UniBootDesktop“", Hide: "Slėpti „UniBootDesktop“", ShowAll: "Rodyti visus",
		Quit: "Baigti „UniBootDesktop“", Edit: "Redaguoti", Undo: "Atšaukti", Redo: "Grąžinti",
		Cut: "Iškirpti", Copy: "Kopijuoti", Paste: "Įklijuoti", SelectAll: "Žymėti viską",
		Window: "Langas", Minimize: "Sumažinti", Zoom: "Mastelis", Help: "Pagalba",
	},
	"be-BY": {
		App: "UniBootDesktop", About: "Пра праграму UniBootDesktop", Hide: "Схаваць UniBootDesktop", ShowAll: "Паказаць усё",
		Quit: "Скончыць UniBootDesktop", Edit: "Праўка", Undo: "Адмяніць", Redo: "Паўтарыць",
		Cut: "Выразаць", Copy: "Капіраваць", Paste: "Уставіць", SelectAll: "Вылучыць усё",
		Window: "Акно", Minimize: "Згарнуць", Zoom: "Маштаб", Help: "Даведка",
	},
	"sr-Cyrl": {
		App: "UniBootDesktop", About: "О апликацији UniBootDesktop", Hide: "Сакриј UniBootDesktop", ShowAll: "Прикажи све",
		Quit: "Напусти UniBootDesktop", Edit: "Уређивање", Undo: "Поништи", Redo: "Понови",
		Cut: "Исеци", Copy: "Копирај", Paste: "Налепи", SelectAll: "Изабери све",
		Window: "Прозор", Minimize: "Минимализуј", Zoom: "Увећај", Help: "Помоћ",
	},
	"sr-Latn": {
		App: "UniBootDesktop", About: "O aplikaciji UniBootDesktop", Hide: "Sakrij UniBootDesktop", ShowAll: "Prikaži sve",
		Quit: "Napusti UniBootDesktop", Edit: "Uređivanje", Undo: "Poništi", Redo: "Ponovi",
		Cut: "Iseci", Copy: "Kopiraj", Paste: "Nalepi", SelectAll: "Izaberi sve",
		Window: "Prozor", Minimize: "Minimalizuj", Zoom: "Uvećaj", Help: "Pomoć",
	},
	"hr-HR": {
		App: "UniBootDesktop", About: "O aplikaciji UniBootDesktop", Hide: "Sakrij UniBootDesktop", ShowAll: "Prikaži sve",
		Quit: "Zatvori UniBootDesktop", Edit: "Uredi", Undo: "Poništi", Redo: "Ponovi",
		Cut: "Izreži", Copy: "Kopiraj", Paste: "Zalijepi", SelectAll: "Odaberi sve",
		Window: "Prozor", Minimize: "Minimiziraj", Zoom: "Zumiraj", Help: "Pomoć",
	},
	"sl-SI": {
		App: "UniBootDesktop", About: "O programu UniBootDesktop", Hide: "Skrij UniBootDesktop", ShowAll: "Prikaži vse",
		Quit: "Zapri UniBootDesktop", Edit: "Uredi", Undo: "Razveljavi", Redo: "Uveljavi",
		Cut: "Izreži", Copy: "Kopiraj", Paste: "Prilepi", SelectAll: "Izberi vse",
		Window: "Okno", Minimize: "Minimiziraj", Zoom: "Povečaj", Help: "Pomoč",
	},
	"mk-MK": {
		App: "UniBootDesktop", About: "За UniBootDesktop", Hide: "Скриј го UniBootDesktop", ShowAll: "Прикажи ги сите",
		Quit: "Напушти го UniBootDesktop", Edit: "Уреди", Undo: "Врати", Redo: "Повтори",
		Cut: "Исечи", Copy: "Копирај", Paste: "Залепи", SelectAll: "Избери сѐ",
		Window: "Прозорец", Minimize: "Минимизирај", Zoom: "Зумирај", Help: "Помош",
	},
	"az-AZ": {
		App: "UniBootDesktop", About: "UniBootDesktop haqqında", Hide: "UniBootDesktop Gizlət", ShowAll: "Hamısını Göstər",
		Quit: "UniBootDesktop-dan Çıx", Edit: "Düzəliş et", Undo: "Ləğv et", Redo: "Təkrar et",
		Cut: "Kəs", Copy: "Kopyala", Paste: "Yapışdır", SelectAll: "Hamısını seç",
		Window: "Pəncərə", Minimize: "Kiçilt", Zoom: "Miqyas", Help: "Kömək",
	},
	"hy-AM": {
		App: "UniBootDesktop", About: "UniBootDesktop-ի մասին", Hide: "Թաքցնել UniBootDesktop", ShowAll: "Ցուցադրել բոլորը",
		Quit: "Փակել UniBootDesktop", Edit: "Խմբագրել", Undo: "Հետարկել", Redo: "Կրկնել",
		Cut: "Կտրել", Copy: "Պատճենել", Paste: "Տեղադրել", SelectAll: "Ընտրել բոլորը",
		Window: "Պատուհան", Minimize: "Փոքրացնել", Zoom: "Մեծացնել", Help: "Օգնություն",
	},
	"ka-GE": {
		App: "UniBootDesktop", About: "UniBootDesktop-ის შესახებ", Hide: "UniBootDesktop-ის დამალვა", ShowAll: "ყველას ჩვენება",
		Quit: "UniBootDesktop-იდან გამოსვლা", Edit: "რედაქტირება", Undo: "დაბრუნება", Redo: "გამეორება",
		Cut: "ამოჭრა", Copy: "კოპირება", Paste: "ჩასმა", SelectAll: "ყველაფრის მონიშვნാ",
		Window: "ფანჯარა", Minimize: "ჩაკეცილი", Zoom: "მასშტაბირება", Help: "დახმარება",
	},
	"ur-PK": {
		App: "UniBootDesktop", About: "UniBootDesktop کے بارے میں", Hide: "UniBootDesktop چھپائیں", ShowAll: "سب دکھائیں",
		Quit: "UniBootDesktop بند کریں", Edit: "ترمیم", Undo: "منسوخ کریں", Redo: "دوبارہ کریں",
		Cut: "کٹ کریں", Copy: "کپی کریں", Paste: "پیسٹ کریں", SelectAll: "تمام منتخب کریں",
		Window: "ونڈو", Minimize: "چھوٹا کریں", Zoom: "زوم", Help: "مدد",
	},
	"oc-FR": {
		App: "UniBootDesktop", About: "A propòs de UniBootDesktop", Hide: "Amagar UniBootDesktop", ShowAll: "Mstrar tot",
		Quit: "Quitar UniBootDesktop", Edit: "Edicion", Undo: "Anullar", Redo: "Tornar far",
		Cut: "Copar", Copy: "Copiar", Paste: "Pgar", SelectAll: "Seleccionar tot",
		Window: "Fenèstra", Minimize: "Reduire", Zoom: "Agrandir", Help: "Ajuda",
	},
}

// DetectSystemLocale returns the detected host operating system language locale.
func DetectSystemLocale() string {
	for _, envKey := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG", "LC_CTYPE"} {
		val := strings.TrimSpace(env.Get(envKey))
		if val != "" && val != "C" && val != "POSIX" {
			norm := strings.ToLower(val)
			if strings.Contains(norm, "zh_tw") || strings.Contains(norm, "zh_hk") || strings.Contains(norm, "zh-hant") {
				return "zh-TW"
			} else if strings.Contains(norm, "zh") {
				return "zh-CN"
			} else if strings.Contains(norm, "de") {
				return "de-DE"
			} else if strings.Contains(norm, "fr") {
				return "fr-FR"
			} else if strings.Contains(norm, "es") {
				return "es-ES"
			} else if strings.Contains(norm, "ja") {
				return "ja-JP"
			} else if strings.Contains(norm, "ko") {
				return "ko-KR"
			} else if strings.Contains(norm, "ru") {
				return "ru-RU"
			} else if strings.Contains(norm, "en") {
				return "en-US"
			}
		}
	}
	return "zh-CN"
}

// GetMenuTranslations returns localized menu strings based on target language code or system language fallback.
func GetMenuTranslations(langCode string) MenuTranslations {
	langCode = strings.TrimSpace(langCode)
	if langCode == "" || langCode == "auto" {
		if cfg, err := config.Load(); err == nil && cfg != nil && cfg.Language != "" && cfg.Language != "auto" {
			langCode = cfg.Language
		}
	}
	if langCode == "" || langCode == "auto" {
		langCode = DetectSystemLocale()
	}

	// Normalize locale string (e.g., zh_CN -> zh-CN)
	normalized := strings.ReplaceAll(langCode, "_", "-")

	if t, exists := menuDataStores[normalized]; exists {
		return t
	}

	// Prefix match fallback (e.g. en-GB -> en-US, zh-HK -> zh-TW)
	prefix := strings.Split(normalized, "-")[0]
	for k, t := range menuDataStores {
		if strings.HasPrefix(k, prefix) {
			return t
		}
	}

	return menuDataStores["en-US"]
}
