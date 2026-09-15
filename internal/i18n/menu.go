// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package i18n

import (
	"strings"

	"github.com/snowdreamtech/unigodesktop/internal/env"
	"github.com/snowdreamtech/unigodesktop/pkg/config"
)

// MenuTranslations holds localized titles for OS native menu bar items.
type MenuTranslations struct {
	App        string
	About      string
	Hide       string
	ShowAll    string
	Quit       string
	Edit       string
	Undo       string
	Redo       string
	Cut        string
	Copy       string
	Paste      string
	SelectAll  string
	Window     string
	Minimize   string
	Zoom       string
	Help       string
}

// menuDataStores stores localized menu titles across all 53 supported locales.
var menuDataStores = map[string]MenuTranslations{
	"zh-CN": {
		App: "UniGoDesktop", About: "关于 UniGoDesktop", Hide: "隐藏 UniGoDesktop", ShowAll: "显示全部",
		Quit: "退出 UniGoDesktop", Edit: "编辑", Undo: "撤销", Redo: "重做",
		Cut: "剪切", Copy: "复制", Paste: "粘贴", SelectAll: "全选",
		Window: "窗口", Minimize: "最小化", Zoom: "缩放", Help: "帮助",
	},
	"zh-TW": {
		App: "UniGoDesktop", About: "關於 UniGoDesktop", Hide: "隱藏 UniGoDesktop", ShowAll: "顯示全部",
		Quit: "結束 UniGoDesktop", Edit: "編輯", Undo: "復原", Redo: "重做",
		Cut: "剪下", Copy: "複製", Paste: "貼上", SelectAll: "全選",
		Window: "視窗", Minimize: "縮到最小", Zoom: "縮放", Help: "說明",
	},
	"en-US": {
		App: "UniGoDesktop", About: "About UniGoDesktop", Hide: "Hide UniGoDesktop", ShowAll: "Show All",
		Quit: "Quit UniGoDesktop", Edit: "Edit", Undo: "Undo", Redo: "Redo",
		Cut: "Cut", Copy: "Copy", Paste: "Paste", SelectAll: "Select All",
		Window: "Window", Minimize: "Minimize", Zoom: "Zoom", Help: "Help",
	},
	"de-DE": {
		App: "UniGoDesktop", About: "Über UniGoDesktop", Hide: "UniGoDesktop ausblenden", ShowAll: "Alle einblenden",
		Quit: "UniGoDesktop beenden", Edit: "Bearbeiten", Undo: "Rückgängig", Redo: "Wiederholen",
		Cut: "Ausschneiden", Copy: "Kopieren", Paste: "Einfügen", SelectAll: "Alles auswählen",
		Window: "Fenster", Minimize: "Im Dock ablegen", Zoom: "Zoomen", Help: "Hilfe",
	},
	"fr-FR": {
		App: "UniGoDesktop", About: "À propos de UniGoDesktop", Hide: "Masquer UniGoDesktop", ShowAll: "Tout afficher",
		Quit: "Quitter UniGoDesktop", Edit: "Édition", Undo: "Annuler", Redo: "Rétablir",
		Cut: "Couper", Copy: "Copier", Paste: "Coller", SelectAll: "Tout sélectionner",
		Window: "Fenêtre", Minimize: "Réduire", Zoom: "Zoomer", Help: "Aide",
	},
	"es-ES": {
		App: "UniGoDesktop", About: "Acerca de UniGoDesktop", Hide: "Ocultar UniGoDesktop", ShowAll: "Mostrar todo",
		Quit: "Salir de UniGoDesktop", Edit: "Edición", Undo: "Deshacer", Redo: "Rehacer",
		Cut: "Cortar", Copy: "Copiar", Paste: "Pegar", SelectAll: "Seleccionar todo",
		Window: "Ventana", Minimize: "Minimizar", Zoom: "Zoom", Help: "Ayuda",
	},
	"es-LA": {
		App: "UniGoDesktop", About: "Acerca de UniGoDesktop", Hide: "Ocultar UniGoDesktop", ShowAll: "Mostrar todo",
		Quit: "Salir de UniGoDesktop", Edit: "Edición", Undo: "Deshacer", Redo: "Rehacer",
		Cut: "Cortar", Copy: "Copiar", Paste: "Pegar", SelectAll: "Seleccionar todo",
		Window: "Ventana", Minimize: "Minimizar", Zoom: "Zoom", Help: "Ayuda",
	},
	"it-IT": {
		App: "UniGoDesktop", About: "Informazioni su UniGoDesktop", Hide: "Nascondi UniGoDesktop", ShowAll: "Mostra tutte",
		Quit: "Esci da UniGoDesktop", Edit: "Composizione", Undo: "Annulla", Redo: "Ripristina",
		Cut: "Taglia", Copy: "Copia", Paste: "Incolla", SelectAll: "Seleziona tutto",
		Window: "Finestra", Minimize: "Contrai", Zoom: "Ridimensiona", Help: "Aiuto",
	},
	"ja-JP": {
		App: "UniGoDesktop", About: "UniGoDesktop について", Hide: "UniGoDesktop を非表示", ShowAll: "すべてを表示",
		Quit: "UniGoDesktop を終了", Edit: "編集", Undo: "元に戻す", Redo: "やり直す",
		Cut: "切り取り", Copy: "コピー", Paste: "貼り付け", SelectAll: "すべてを選択",
		Window: "ウィンドウ", Minimize: "最小化", Zoom: "拡大/縮小", Help: "ヘルプ",
	},
	"ko-KR": {
		App: "UniGoDesktop", About: "UniGoDesktop 정보", Hide: "UniGoDesktop 가리기", ShowAll: "모두 보기",
		Quit: "UniGoDesktop 종료", Edit: "편집", Undo: "실행 취소", Redo: "다시 실행",
		Cut: "오려두기", Copy: "복사", Paste: "붙여넣기", SelectAll: "전체 선택",
		Window: "윈도우", Minimize: "최소화", Zoom: "확대/축소", Help: "도움말",
	},
	"ru-RU": {
		App: "UniGoDesktop", About: "О программе UniGoDesktop", Hide: "Скрыть UniGoDesktop", ShowAll: "Показать все",
		Quit: "Завершить UniGoDesktop", Edit: "Правка", Undo: "Отменить", Redo: "Повторить",
		Cut: "Вырезать", Copy: "Скопировать", Paste: "Вставить", SelectAll: "Выделить все",
		Window: "Окно", Minimize: "Свернуть", Zoom: "Изменить масштаб", Help: "Справка",
	},
	"pt-BR": {
		App: "UniGoDesktop", About: "Sobre o UniGoDesktop", Hide: "Ocultar UniGoDesktop", ShowAll: "Mostrar Tudo",
		Quit: "Encerrar o UniGoDesktop", Edit: "Editar", Undo: "Desfazer", Redo: "Refazer",
		Cut: "Recortar", Copy: "Copiar", Paste: "Colar", SelectAll: "Selecionar Tudo",
		Window: "Janela", Minimize: "Minimizar", Zoom: "Deminuir/Aumentar Zoom", Help: "Ajuda",
	},
	"pt-PT": {
		App: "UniGoDesktop", About: "Sobre o UniGoDesktop", Hide: "Ocultar UniGoDesktop", ShowAll: "Mostrar Tudo",
		Quit: "Encerrar o UniGoDesktop", Edit: "Edição", Undo: "Desfazer", Redo: "Refazer",
		Cut: "Cortar", Copy: "Copiar", Paste: "Colar", SelectAll: "Selecionar Tudo",
		Window: "Janela", Minimize: "Minimizar", Zoom: "Redimensionar", Help: "Ajuda",
	},
	"ar-SA": {
		App: "UniGoDesktop", About: "حول UniGoDesktop", Hide: "إخفاء UniGoDesktop", ShowAll: "إظهار الكل",
		Quit: "إنهاء UniGoDesktop", Edit: "تعديل", Undo: "تراجع", Redo: "إعادة",
		Cut: "قص", Copy: "نسخ", Paste: "لصق", SelectAll: "تحديد الكل",
		Window: "نافذة", Minimize: "تصغير", Zoom: "تكبير/تصغير", Help: "مساعدة",
	},
	"vi-VN": {
		App: "UniGoDesktop", About: "Về UniGoDesktop", Hide: "Ẩn UniGoDesktop", ShowAll: "Hiện tất cả",
		Quit: "Thoát UniGoDesktop", Edit: "Chỉnh sửa", Undo: "Hoàn tác", Redo: "Làm lại",
		Cut: "Cắt", Copy: "Sao chép", Paste: "Dán", SelectAll: "Chọn tất cả",
		Window: "Cửa sổ", Minimize: "Thu nhỏ", Zoom: "Phóng to/Thu nhỏ", Help: "Trợ giúp",
	},
	"tr-TR": {
		App: "UniGoDesktop", About: "UniGoDesktop Hakkında", Hide: "UniGoDesktop'ı Gizle", ShowAll: "Tümünü Göster",
		Quit: "UniGoDesktop'tan Çık", Edit: "Düzenle", Undo: "Geri Al", Redo: "Yinele",
		Cut: "Kes", Copy: "Kopyala", Paste: "Yapıştır", SelectAll: "Tümünü Seç",
		Window: "Pencere", Minimize: "Simge Durumuna Küçült", Zoom: "Büyüt/Küçült", Help: "Yardım",
	},
	"pl-PL": {
		App: "UniGoDesktop", About: "O UniGoDesktop", Hide: "Ukryj UniGoDesktop", ShowAll: "Pokaż wszystkie",
		Quit: "Zakończ UniGoDesktop", Edit: "Edycja", Undo: "Cofnij", Redo: "Przywróć",
		Cut: "Wytnij", Copy: "Kopiuj", Paste: "Wklej", SelectAll: "Zaznacz wszystko",
		Window: "Okno", Minimize: "Zminimalizuj", Zoom: "Wypełnij", Help: "Pomoc",
	},
	"nl-NL": {
		App: "UniGoDesktop", About: "Over UniGoDesktop", Hide: "Verberg UniGoDesktop", ShowAll: "Toon alles",
		Quit: "UniGoDesktop sluiten", Edit: "Wijzig", Undo: "Herstel", Redo: "Opnieuw",
		Cut: "Knippen", Copy: "Kopiëren", Paste: "Plakken", SelectAll: "Selecteer alles",
		Window: "Venster", Minimize: "Minimiseer", Zoom: "Zoom", Help: "Help",
	},
	"sv-SE": {
		App: "UniGoDesktop", About: "Om UniGoDesktop", Hide: "Göm UniGoDesktop", ShowAll: "Visa alla",
		Quit: "Avsluta UniGoDesktop", Edit: "Redigera", Undo: "Ångra", Redo: "Gör om",
		Cut: "Klipp ut", Copy: "Kopiera", Paste: "Klistra in", SelectAll: "Markera allt",
		Window: "Fönster", Minimize: "Minimera", Zoom: "Zooma", Help: "Hjälp",
	},
	"da-DK": {
		App: "UniGoDesktop", About: "Om UniGoDesktop", Hide: "Skjul UniGoDesktop", ShowAll: "Vis alle",
		Quit: "Slut UniGoDesktop", Edit: "Rediger", Undo: "Fortryd", Redo: "Gentag",
		Cut: "Klip", Copy: "Kopier", Paste: "Sæt ind", SelectAll: "Vælg alt",
		Window: "Vindue", Minimize: "Minimer", Zoom: "Zoom", Help: "Hjælp",
	},
	"nb-NO": {
		App: "UniGoDesktop", About: "Om UniGoDesktop", Hide: "Skjul UniGoDesktop", ShowAll: "Vis alle",
		Quit: "Avslutt UniGoDesktop", Edit: "Rediger", Undo: "Angre", Redo: "Gjør om",
		Cut: "Klipp ut", Copy: "Kopier", Paste: "Lim inn", SelectAll: "Marker alt",
		Window: "Vindu", Minimize: "Minimer", Zoom: "Zoom", Help: "Hjelp",
	},
	"no-NO": {
		App: "UniGoDesktop", About: "Om UniGoDesktop", Hide: "Skjul UniGoDesktop", ShowAll: "Vis alle",
		Quit: "Avslutt UniGoDesktop", Edit: "Rediger", Undo: "Angre", Redo: "Gjør om",
		Cut: "Klipp ut", Copy: "Kopier", Paste: "Lim inn", SelectAll: "Marker alt",
		Window: "Vindu", Minimize: "Minimer", Zoom: "Zoom", Help: "Hjelp",
	},
	"fi-FI": {
		App: "UniGoDesktop", About: "Tietoja UniGoDesktopista", Hide: "Kätke UniGoDesktop", ShowAll: "Näytä kaikki",
		Quit: "Lopeta UniGoDesktop", Edit: "Muokkaa", Undo: "Peru", Redo: "Tee uudelleen",
		Cut: "Leikkaa", Copy: "Kopioi", Paste: "Sijoita", SelectAll: "Valitse kaikki",
		Window: "Ikkuna", Minimize: "Pienennä", Zoom: "Zoomaa", Help: "Ohje",
	},
	"cs-CZ": {
		App: "UniGoDesktop", About: "O aplikaci UniGoDesktop", Hide: "Skrýt UniGoDesktop", ShowAll: "Zobrazit vše",
		Quit: "Ukončit UniGoDesktop", Edit: "Úpravy", Undo: "Zpět", Redo: "Znovu",
		Cut: "Vyjmout", Copy: "Kopírovat", Paste: "Vložit", SelectAll: "Vybrat vše",
		Window: "Okno", Minimize: "Minimalizovat", Zoom: "Přiblížit", Help: "Nápověda",
	},
	"sk-SK": {
		App: "UniGoDesktop", About: "O aplikácii UniGoDesktop", Hide: "Skryť UniGoDesktop", ShowAll: "Zobraziť všetky",
		Quit: "Ukončiť UniGoDesktop", Edit: "Upraviť", Undo: "Späť", Redo: "Znovu",
		Cut: "Vystrihnúť", Copy: "Kopírovať", Paste: "Prilepiť", SelectAll: "Vybrať všetko",
		Window: "Okno", Minimize: "Minimalizovať", Zoom: "Zväčšiť", Help: "Pomocník",
	},
	"hu-HU": {
		App: "UniGoDesktop", About: "A UniGoDesktop névjegye", Hide: "A UniGoDesktop elrejtése", ShowAll: "Összes megjelenítése",
		Quit: "Kilépés a UniGoDesktopból", Edit: "Szerkesztés", Undo: "Visszavonás", Redo: "Ismétlés",
		Cut: "Kivágás", Copy: "Másolás", Paste: "Beillesztés", SelectAll: "Összes kijelölése",
		Window: "Ablak", Minimize: "Kis méret", Zoom: "Nagyítás", Help: "Súgó",
	},
	"ro-RO": {
		App: "UniGoDesktop", About: "Despre UniGoDesktop", Hide: "Ascunde UniGoDesktop", ShowAll: "Afișează toate",
		Quit: "Închide UniGoDesktop", Edit: "Editare", Undo: "Anulează", Redo: "Refă",
		Cut: "Tunde", Copy: "Copiază", Paste: "Lipește", SelectAll: "Selectează tot",
		Window: "Fereastră", Minimize: "Minimizează", Zoom: "Zoom", Help: "Ajutor",
	},
	"bg-BG": {
		App: "UniGoDesktop", About: "Относно UniGoDesktop", Hide: "Скриване на UniGoDesktop", ShowAll: "Показване на всички",
		Quit: "Изход от UniGoDesktop", Edit: "Редактиране", Undo: "Отмяна", Redo: "Повторение",
		Cut: "Изрязване", Copy: "Копиране", Paste: "Поставяне", SelectAll: "Избиране на всички",
		Window: "Прозорец", Minimize: "Минимизиране", Zoom: "Мащабиране", Help: "Помощ",
	},
	"uk-UA": {
		App: "UniGoDesktop", About: "Про програму UniGoDesktop", Hide: "Сховати UniGoDesktop", ShowAll: "Показати всі",
		Quit: "Завершити UniGoDesktop", Edit: "Редагування", Undo: "Скасувати", Redo: "Повторити",
		Cut: "Вирізати", Copy: "Копіювати", Paste: "Вставити", SelectAll: "Виділити все",
		Window: "Вікно", Minimize: "Згорнути", Zoom: "Масштаб", Help: "Довідка",
	},
	"el-GR": {
		App: "UniGoDesktop", About: "Σχετικά με το UniGoDesktop", Hide: "Απόκρυψη UniGoDesktop", ShowAll: "Εμφάνιση όλων",
		Quit: "Έξοδος από UniGoDesktop", Edit: "Επεξεργασία", Undo: "Αναίρεση", Redo: "Επανάληψη",
		Cut: "Αποκοπή", Copy: "Αντιγραφή", Paste: "Επικόλληση", SelectAll: "Επιλογή όλων",
		Window: "Παράθυρο", Minimize: "Ελαχιστοποίηση", Zoom: "Εστίαση", Help: "Βοήθεια",
	},
	"he-IL": {
		App: "UniGoDesktop", About: "על אודות UniGoDesktop", Hide: "הסתר את UniGoDesktop", ShowAll: "הצג הכל",
		Quit: "סיום UniGoDesktop", Edit: "עריכה", Undo: "בטל", Redo: "בצע שוב",
		Cut: "גזור", Copy: "העתק", Paste: "הדבק", SelectAll: "בחר הכל",
		Window: "חלון", Minimize: "מזער", Zoom: "זום", Help: "עזרה",
	},
	"fa-IR": {
		App: "UniGoDesktop", About: "درباره UniGoDesktop", Hide: "پنهان کردن UniGoDesktop", ShowAll: "نمایش همه",
		Quit: "خروج از UniGoDesktop", Edit: "ویرایش", Undo: "واکشی", Redo: "انجام دوباره",
		Cut: "برش", Copy: "کپی", Paste: "جای‌گذاری", SelectAll: "انتخاب همه",
		Window: "پنجره", Minimize: "کمینه کردن", Zoom: "بزرگ‌نمایی", Help: "راهنما",
	},
	"hi-IN": {
		App: "UniGoDesktop", About: "UniGoDesktop के बारे में", Hide: "UniGoDesktop छिपाएं", ShowAll: "सभी दिखाएं",
		Quit: "UniGoDesktop से बाहर निकलें", Edit: "संपादित करें", Undo: "पूर्ववत करें", Redo: "फिर से करें",
		Cut: "कट करें", Copy: "कॉपी करें", Paste: "पेस्ट करें", SelectAll: "सभी चुनें",
		Window: "विंडो", Minimize: "छोटा करें", Zoom: "ज़ूम", Help: "सहायता",
	},
	"bn-BD": {
		App: "UniGoDesktop", About: "UniGoDesktop সম্পর্কে", Hide: "UniGoDesktop লুকান", ShowAll: "সব দেখান",
		Quit: "UniGoDesktop বন্ধ করুন", Edit: "সম্পাদনা", Undo: "পূর্বাবস্থায় ফেরান", Redo: "পুনরায় করুন",
		Cut: "কাট", Copy: "কপি", Paste: "পেস্ট", SelectAll: "সব নির্বাচন করুন",
		Window: "উইন্ডো", Minimize: "ছোট করুন", Zoom: "জুম", Help: "সহায়তা",
	},
	"ta-IN": {
		App: "UniGoDesktop", About: "UniGoDesktop பற்றி", Hide: "UniGoDesktop மறை", ShowAll: "அனைத்தையும் காட்டு",
		Quit: "UniGoDesktop வெளியேறு", Edit: "தொகு", Undo: "செயல் நீக்கு", Redo: "மீண்டும் செய்",
		Cut: "வெட்டு", Copy: "நகலெடு", Paste: "ஒட்டு", SelectAll: "அனைத்தையும் தேர்ந்தெடு",
		Window: "சாளரம்", Minimize: "சிறிதாக்கு", Zoom: "பெரிதாக்கு", Help: "உதவி",
	},
	"ml-IN": {
		App: "UniGoDesktop", About: "UniGoDesktop നെ കുറിച്ച്", Hide: "UniGoDesktop മറയ്ക്കുക", ShowAll: "എല്ലാം കാണിക്കുക",
		Quit: "UniGoDesktop പുറത്തുകടക്കുക", Edit: "എഡിറ്റ് ചെയ്യുക", Undo: "തിരുത്തുക", Redo: "വീണ്ടും ചെയ്യുക",
		Cut: "മുറിക്കുക", Copy: "പകർപ്പുക", Paste: "ഒട്ടിക്കുക", SelectAll: "എല്ലാം തിരഞ്ഞെടുക്കുക",
		Window: "വിൻഡോ", Minimize: "ചെറുതാക്കുക", Zoom: "വലുതാക്കുക", Help: "സഹായം",
	},
	"th-TH": {
		App: "UniGoDesktop", About: "เกี่ยวกับ UniGoDesktop", Hide: "ซ่อน UniGoDesktop", ShowAll: "แสดงทั้งหมด",
		Quit: "ออกจาก UniGoDesktop", Edit: "แก้ไข", Undo: "เลิกทำ", Redo: "ทำซ้ำ",
		Cut: "ตัด", Copy: "คัดลอก", Paste: "วาง", SelectAll: "เลือกทั้งหมด",
		Window: "หน้าต่าง", Minimize: "ย่อหน้าต่าง", Zoom: "ซูม", Help: "ช่วยเหลือ",
	},
	"id-ID": {
		App: "UniGoDesktop", About: "Tentang UniGoDesktop", Hide: "Sembunyikan UniGoDesktop", ShowAll: "Tampilkan Semua",
		Quit: "Keluar UniGoDesktop", Edit: "Edit", Undo: "Batalkan", Redo: "Ulangi",
		Cut: "Potong", Copy: "Salin", Paste: "Tempel", SelectAll: "Pilih Semua",
		Window: "Jendela", Minimize: "Minimalkan", Zoom: "Perbesar", Help: "Bantuan",
	},
	"ca-ES": {
		App: "UniGoDesktop", About: "Quant a UniGoDesktop", Hide: "Amaga UniGoDesktop", ShowAll: "Mostra-ho tot",
		Quit: "Surt de UniGoDesktop", Edit: "Edició", Undo: "Desfés", Redo: "Refés",
		Cut: "Retalla", Copy: "Copia", Paste: "Enganxa", SelectAll: "Selecciona-ho tot",
		Window: "Finestra", Minimize: "Minimitza", Zoom: "Amplia", Help: "Ajuda",
	},
	"gl-ES": {
		App: "UniGoDesktop", About: "Sobre UniGoDesktop", Hide: "Ocultar UniGoDesktop", ShowAll: "Amosar todo",
		Quit: "Saír de UniGoDesktop", Edit: "Editar", Undo: "Desfacer", Redo: "Refacer",
		Cut: "Cortar", Copy: "Copiar", Paste: "Pegar", SelectAll: "Seleccionar todo",
		Window: "Xanela", Minimize: "Minimizar", Zoom: "Ampliar", Help: "Axuda",
	},
	"et-EE": {
		App: "UniGoDesktop", About: "UniGoDesktop teave", Hide: "Peida UniGoDesktop", ShowAll: "Kuva kõik",
		Quit: "Välju UniGoDesktopist", Edit: "Redigeeri", Undo: "Võta tagasi", Redo: "Tee uuesti",
		Cut: "Lõika", Copy: "Kopeeri", Paste: "Aseta", SelectAll: "Vali kõik",
		Window: "Aken", Minimize: "Minimeeri", Zoom: "Suurenda", Help: "Abi",
	},
	"lt-LT": {
		App: "UniGoDesktop", About: "Apie „UniGoDesktop“", Hide: "Slėpti „UniGoDesktop“", ShowAll: "Rodyti visus",
		Quit: "Baigti „UniGoDesktop“", Edit: "Redaguoti", Undo: "Atšaukti", Redo: "Grąžinti",
		Cut: "Iškirpti", Copy: "Kopijuoti", Paste: "Įklijuoti", SelectAll: "Žymėti viską",
		Window: "Langas", Minimize: "Sumažinti", Zoom: "Mastelis", Help: "Pagalba",
	},
	"be-BY": {
		App: "UniGoDesktop", About: "Пра праграму UniGoDesktop", Hide: "Схаваць UniGoDesktop", ShowAll: "Паказаць усё",
		Quit: "Скончыць UniGoDesktop", Edit: "Праўка", Undo: "Адмяніць", Redo: "Паўтарыць",
		Cut: "Выразаць", Copy: "Капіраваць", Paste: "Уставіць", SelectAll: "Вылучыць усё",
		Window: "Акно", Minimize: "Згарнуць", Zoom: "Маштаб", Help: "Даведка",
	},
	"sr-Cyrl": {
		App: "UniGoDesktop", About: "О апликацији UniGoDesktop", Hide: "Сакриј UniGoDesktop", ShowAll: "Прикажи све",
		Quit: "Напусти UniGoDesktop", Edit: "Уређивање", Undo: "Поништи", Redo: "Понови",
		Cut: "Исеци", Copy: "Копирај", Paste: "Налепи", SelectAll: "Изабери све",
		Window: "Прозор", Minimize: "Минимализуј", Zoom: "Увећај", Help: "Помоћ",
	},
	"sr-Latn": {
		App: "UniGoDesktop", About: "O aplikaciji UniGoDesktop", Hide: "Sakrij UniGoDesktop", ShowAll: "Prikaži sve",
		Quit: "Napusti UniGoDesktop", Edit: "Uređivanje", Undo: "Poništi", Redo: "Ponovi",
		Cut: "Iseci", Copy: "Kopiraj", Paste: "Nalepi", SelectAll: "Izaberi sve",
		Window: "Prozor", Minimize: "Minimalizuj", Zoom: "Uvećaj", Help: "Pomoć",
	},
	"hr-HR": {
		App: "UniGoDesktop", About: "O aplikaciji UniGoDesktop", Hide: "Sakrij UniGoDesktop", ShowAll: "Prikaži sve",
		Quit: "Zatvori UniGoDesktop", Edit: "Uredi", Undo: "Poništi", Redo: "Ponovi",
		Cut: "Izreži", Copy: "Kopiraj", Paste: "Zalijepi", SelectAll: "Odaberi sve",
		Window: "Prozor", Minimize: "Minimiziraj", Zoom: "Zumiraj", Help: "Pomoć",
	},
	"sl-SI": {
		App: "UniGoDesktop", About: "O programu UniGoDesktop", Hide: "Skrij UniGoDesktop", ShowAll: "Prikaži vse",
		Quit: "Zapri UniGoDesktop", Edit: "Uredi", Undo: "Razveljavi", Redo: "Uveljavi",
		Cut: "Izreži", Copy: "Kopiraj", Paste: "Prilepi", SelectAll: "Izberi vse",
		Window: "Okno", Minimize: "Minimiziraj", Zoom: "Povečaj", Help: "Pomoč",
	},
	"mk-MK": {
		App: "UniGoDesktop", About: "За UniGoDesktop", Hide: "Скриј го UniGoDesktop", ShowAll: "Прикажи ги сите",
		Quit: "Напушти го UniGoDesktop", Edit: "Уреди", Undo: "Врати", Redo: "Повтори",
		Cut: "Исечи", Copy: "Копирај", Paste: "Залепи", SelectAll: "Избери сѐ",
		Window: "Прозорец", Minimize: "Минимизирај", Zoom: "Зумирај", Help: "Помош",
	},
	"az-AZ": {
		App: "UniGoDesktop", About: "UniGoDesktop haqqında", Hide: "UniGoDesktop Gizlət", ShowAll: "Hamısını Göstər",
		Quit: "UniGoDesktop-dan Çıx", Edit: "Düzəliş et", Undo: "Ləğv et", Redo: "Təkrar et",
		Cut: "Kəs", Copy: "Kopyala", Paste: "Yapışdır", SelectAll: "Hamısını seç",
		Window: "Pəncərə", Minimize: "Kiçilt", Zoom: "Miqyas", Help: "Kömək",
	},
	"hy-AM": {
		App: "UniGoDesktop", About: "UniGoDesktop-ի մասին", Hide: "Թաքցնել UniGoDesktop", ShowAll: "Ցուցադրել բոլորը",
		Quit: "Փակել UniGoDesktop", Edit: "Խմբագրել", Undo: "Հետարկել", Redo: "Կրկնել",
		Cut: "Կտրել", Copy: "Պատճենել", Paste: "Տեղադրել", SelectAll: "Ընտրել բոլորը",
		Window: "Պատուհան", Minimize: "Փոքրացնել", Zoom: "Մեծացնել", Help: "Օգնություն",
	},
	"ka-GE": {
		App: "UniGoDesktop", About: "UniGoDesktop-ის შესახებ", Hide: "UniGoDesktop-ის დამალვა", ShowAll: "ყველას ჩვენება",
		Quit: "UniGoDesktop-იდან გამოსვლা", Edit: "რედაქტირება", Undo: "დაბრუნება", Redo: "გამეორება",
		Cut: "ამოჭრა", Copy: "კოპირება", Paste: "ჩასმა", SelectAll: "ყველაფრის მონიშვნാ",
		Window: "ფანჯარა", Minimize: "ჩაკეცილი", Zoom: "მასშტაბირება", Help: "დახმარება",
	},
	"ur-PK": {
		App: "UniGoDesktop", About: "UniGoDesktop کے بارے میں", Hide: "UniGoDesktop چھپائیں", ShowAll: "سب دکھائیں",
		Quit: "UniGoDesktop بند کریں", Edit: "ترمیم", Undo: "منسوخ کریں", Redo: "دوبارہ کریں",
		Cut: "کٹ کریں", Copy: "کپی کریں", Paste: "پیسٹ کریں", SelectAll: "تمام منتخب کریں",
		Window: "ونڈو", Minimize: "چھوٹا کریں", Zoom: "زوم", Help: "مدد",
	},
	"oc-FR": {
		App: "UniGoDesktop", About: "A propòs de UniGoDesktop", Hide: "Amagar UniGoDesktop", ShowAll: "Mstrar tot",
		Quit: "Quitar UniGoDesktop", Edit: "Edicion", Undo: "Anullar", Redo: "Tornar far",
		Cut: "Copar", Copy: "Copiar", Paste: "Pgar", SelectAll: "Seleccionar tot",
		Window: "Fenèstra", Minimize: "Reduire", Zoom: "Agrandir", Help: "Ajuda",
	},
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
		langCode = env.Get("LANG")
		if strings.Contains(strings.ToLower(langCode), "zh_tw") || strings.Contains(strings.ToLower(langCode), "zh_hk") {
			langCode = "zh-TW"
		} else if strings.Contains(strings.ToLower(langCode), "zh") {
			langCode = "zh-CN"
		} else if strings.Contains(strings.ToLower(langCode), "de") {
			langCode = "de-DE"
		} else if strings.Contains(strings.ToLower(langCode), "fr") {
			langCode = "fr-FR"
		} else if strings.Contains(strings.ToLower(langCode), "es") {
			langCode = "es-ES"
		} else if strings.Contains(strings.ToLower(langCode), "ja") {
			langCode = "ja-JP"
		} else if strings.Contains(strings.ToLower(langCode), "ko") {
			langCode = "ko-KR"
		} else if strings.Contains(strings.ToLower(langCode), "ru") {
			langCode = "ru-RU"
		} else {
			langCode = "en-US"
		}
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
