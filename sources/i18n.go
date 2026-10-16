package main

// Traductions de l'interface, des messages et des fichiers produits.
// Ordre des colonnes : fr, en, es, de, it, pt, nl, zh, ja.
// La base Job Manager produite est identique quelle que soit la langue.

import "fmt"

type Lang int

const (
	LFR Lang = iota
	LEN
	LES
	LDE
	LIT
	LPT
	LNL
	LZH
	LJA
	nLang
)

var langCodes = [nLang]string{"fr", "en", "es", "de", "it", "pt", "nl", "zh", "ja"}
var langNames = [nLang]string{"Français", "English", "Español", "Deutsch", "Italiano", "Português", "Nederlands", "中文", "日本語"}

// langue courante
var cur = LFR

func langFromCode(c string) (Lang, bool) {
	for i, x := range langCodes {
		if x == c {
			return Lang(i), true
		}
	}
	return LFR, false
}

// T : texte traduit dans la langue courante (repli : anglais, puis français).
func T(k string) string {
	v, ok := tr[k]
	if !ok {
		return k
	}
	if s := v[cur]; s != "" {
		return s
	}
	if v[LEN] != "" {
		return v[LEN]
	}
	return v[LFR]
}

func Tf(k string, a ...any) string { return fmt.Sprintf(T(k), a...) }

// Séparateur et décimale du récapitulatif CSV (réglages Excel par pays).
func csvSep() string {
	if cur == LEN || cur == LZH || cur == LJA {
		return ","
	}
	return ";"
}

func csvDec() string {
	if cur == LEN || cur == LZH || cur == LJA {
		return "."
	}
	return ","
}

var tr = map[string][nLang]string{
	// ---------------------------------------------------------------- fenêtre
	"app.title": {
		"VAP – Convertisseur .DAT → JobManager",
		"VAP – .DAT → JobManager Converter",
		"VAP – Convertidor .DAT → JobManager",
		"VAP – Konverter .DAT → JobManager",
		"VAP – Convertitore .DAT → JobManager",
		"VAP – Conversor .DAT → JobManager",
		"VAP – Converter .DAT → JobManager",
		"VAP – .DAT → JobManager 转换器",
		"VAP – .DAT → JobManager 変換ツール",
	},
	"hdr.title": {
		"Convertisseur .DAT → JobManager",
		".DAT → JobManager Converter",
		"Convertidor .DAT → JobManager",
		"Konverter .DAT → JobManager",
		"Convertitore .DAT → JobManager",
		"Conversor .DAT → JobManager",
		"Converter .DAT → JobManager",
		".DAT → JobManager 转换器",
		".DAT → JobManager 変換ツール",
	},
	"hdr.sub": {
		"(Paramètres du travail)",
		"(Job Settings)",
		"(Ajustes del trabajo)",
		"(Auftragseinstellungen)",
		"(Impostazioni del lavoro)",
		"(Configurações do trabalho)",
		"(Taakinstellingen)",
		"（作业设置）",
		"（ジョブ設定）",
	},
	"sec1": {
		"Anciens réglages Epilog (.DAT)",
		"Old Epilog settings (.DAT)",
		"Ajustes antiguos de Epilog (.DAT)",
		"Alte Epilog-Einstellungen (.DAT)",
		"Vecchie impostazioni Epilog (.DAT)",
		"Configurações antigas da Epilog (.DAT)",
		"Oude Epilog-instellingen (.DAT)",
		"旧版 Epilog 设置 (.DAT)",
		"旧 Epilog 設定 (.DAT)",
	},
	"sec2": {
		"Sauvegarde Job Manager (.epidb)",
		"Job Manager backup (.epidb)",
		"Copia de seguridad de Job Manager (.epidb)",
		"Job Manager-Sicherung (.epidb)",
		"Backup di Job Manager (.epidb)",
		"Backup do Job Manager (.epidb)",
		"Job Manager-back-up (.epidb)",
		"Job Manager 备份 (.epidb)",
		"Job Manager バックアップ (.epidb)",
	},
	"sec3": {
		"Machine de destination",
		"Target machine",
		"Máquina de destino",
		"Zielmaschine",
		"Macchina di destinazione",
		"Máquina de destino",
		"Doelmachine",
		"目标机器",
		"変換先のマシン",
	},
	"sec4": {
		"Options de conversion",
		"Conversion options",
		"Opciones de conversión",
		"Konvertierungsoptionen",
		"Opzioni di conversione",
		"Opções de conversão",
		"Conversieopties",
		"转换选项",
		"変換オプション",
	},
	"btn.browse": {
		"Parcourir…", "Browse…", "Examinar…", "Durchsuchen…", "Sfoglia…", "Procurar…", "Bladeren…", "浏览…", "参照…",
	},
	"chk.recursive": {
		"Inclure les sous-dossiers", "Include subfolders", "Incluir subcarpetas", "Unterordner einbeziehen",
		"Includi sottocartelle", "Incluir subpastas", "Submappen meenemen", "包含子文件夹", "サブフォルダーを含める",
	},
	"lbl.machine": {
		"Machine", "Machine", "Máquina", "Maschine", "Macchina", "Máquina", "Machine", "机器", "マシン",
	},
	"lbl.folder": {
		"Nom du dossier", "Folder name", "Nombre de carpeta", "Ordnername", "Nome cartella", "Nome da pasta", "Mapnaam", "文件夹名称", "フォルダー名",
	},
	"chk.air": {
		"Assistance d'air sur Marche pour les découpes",
		"Air assist On for cuts",
		"Asistencia de aire activada en los cortes",
		"Air Assist beim Schneiden einschalten",
		"Assistenza aria attiva per i tagli",
		"Assistência de ar ligada nos cortes",
		"Luchtondersteuning aan bij snijden",
		"切割时开启空气辅助",
		"カット時にエアアシストをオン",
	},
	"chk.top": {
		"Placer le dossier en tête de liste",
		"Place the folder at the top of the list",
		"Colocar la carpeta al principio de la lista",
		"Ordner an den Anfang der Liste setzen",
		"Metti la cartella in cima all'elenco",
		"Colocar a pasta no topo da lista",
		"Map bovenaan de lijst plaatsen",
		"将文件夹置于列表顶部",
		"フォルダーを一覧の先頭に置く",
	},
	"lbl.sort": {
		"Tri des vecteurs", "Vector sorting", "Orden de vectores", "Vektorsortierung", "Ordinamento vettori", "Ordem dos vetores", "Vectorsortering", "矢量排序", "ベクターの並べ替え",
	},
	"sort.none": {
		"Aucun", "None", "Ninguno", "Keine", "Nessuno", "Nenhum", "Geen", "无", "なし",
	},
	"sort.inout": {
		"Intérieur vers extérieur", "Inside to outside", "De dentro hacia fuera", "Von innen nach außen",
		"Dall'interno verso l'esterno", "De dentro para fora", "Van binnen naar buiten", "由内向外", "内側から外側へ",
	},
	"sort.opt": {
		"Optimisé", "Optimized", "Optimizado", "Optimiert", "Ottimizzato", "Otimizado", "Geoptimaliseerd", "优化", "最適化",
	},
	"sort.dat": {
		"Comme dans le .DAT", "As in the .DAT", "Como en el .DAT", "Wie in der .DAT", "Come nel .DAT", "Como no .DAT", "Zoals in het .DAT", "与 .DAT 相同", ".DAT と同じ",
	},
	"lbl.laser": {
		"Source laser", "Laser source", "Fuente láser", "Laserquelle", "Sorgente laser", "Fonte laser", "Laserbron", "激光源", "レーザー光源",
	},
	"laser.co2": {
		"CO2", "CO2", "CO2", "CO2", "CO2", "CO2", "CO2", "CO2", "CO2",
	},
	"laser.fiber": {
		"Fibre", "Fiber", "Fibra", "Faser", "Fibra", "Fibra", "Fiber", "光纤", "ファイバー",
	},
	"btn.quit": {
		"Quitter", "Quit", "Salir", "Beenden", "Esci", "Sair", "Afsluiten", "退出", "終了",
	},
	"btn.help": {
		"Notice / procédure", "User guide", "Guía / procedimiento", "Anleitung", "Guida / procedura", "Guia / procedimento", "Handleiding", "使用说明", "操作手順",
	},
	"btn.donate": {
		"Faire un don", "Donate", "Donar", "Spenden", "Fai una donazione", "Fazer uma doação", "Doneren", "捐赠", "寄付する",
	},
	"btn.suggest": {
		"Suggestion / bug", "Feedback / bug", "Sugerencia / error", "Vorschlag / Fehler", "Suggerimento / bug", "Sugestão / erro", "Suggestie / bug", "建议 / 问题", "ご意見 / 不具合",
	},
	"mail.q1": {
		"Que s'est-il passé, ou quelle est votre idée ?", "What happened, or what is your idea?", "¿Qué ha pasado, o cuál es su idea?", "Was ist passiert, oder was ist Ihre Idee?",
		"Cosa è successo, o qual è la sua idea?", "O que aconteceu, ou qual é a sua ideia?", "Wat is er gebeurd, of wat is uw idee?", "发生了什么，或者您有什么建议？", "何が起きましたか？ またはご提案は？",
	},
	"mail.q2": {
		"Quelle machine ? (ex. Mini 24 40 W)", "Which machine? (e.g. Mini 24 40 W)", "¿Qué máquina? (p. ej. Mini 24 40 W)", "Welche Maschine? (z. B. Mini 24 40 W)",
		"Quale macchina? (es. Mini 24 40 W)", "Qual máquina? (ex. Mini 24 40 W)", "Welke machine? (bijv. Mini 24 40 W)", "哪台机器？（例如 Mini 24 40 W）", "どのマシンですか？（例：Mini 24 40 W）",
	},
	"mail.q3": {
		"Joignez si possible le fichier .DAT concerné.", "If possible, attach the .DAT file concerned.", "Si es posible, adjunte el archivo .DAT correspondiente.", "Bitte hängen Sie, wenn möglich, die betreffende .DAT-Datei an.",
		"Se possibile, alleghi il file .DAT interessato.", "Se possível, anexe o arquivo .DAT em questão.", "Voeg indien mogelijk het betreffende .DAT-bestand toe.", "如可能，请附上相关的 .DAT 文件。", "可能であれば、該当する .DAT ファイルを添付してください。",
	},
	"mail.gh": {
		"Vous avez un compte GitHub ? Vous pouvez aussi écrire ici :", "Have a GitHub account? You can also post here:", "¿Tiene cuenta de GitHub? También puede escribir aquí:", "Sie haben ein GitHub-Konto? Sie können auch hier schreiben:",
		"Ha un account GitHub? Può anche scrivere qui:", "Tem conta no GitHub? Também pode escrever aqui:", "Heeft u een GitHub-account? U kunt ook hier schrijven:", "有 GitHub 账号？也可以在这里提交：", "GitHub アカウントをお持ちの方はこちらにも投稿できます：",
	},
	"btn.convert": {
		"Convertir", "Convert", "Convertir", "Konvertieren", "Converti", "Converter", "Converteren", "转换", "変換",
	},
	"btn.open": {
		"Ouvrir le dossier", "Open folder", "Abrir carpeta", "Ordner öffnen", "Apri cartella", "Abrir pasta", "Map openen", "打开文件夹", "フォルダーを開く",
	},
	"btn.close": {
		"Fermer", "Close", "Cerrar", "Schließen", "Chiudi", "Fechar", "Sluiten", "关闭", "閉じる",
	},

	// ---------------------------------------------------------------- accueil
	"welcome.0": {
		"Bienvenue !", "Welcome!", "¡Bienvenido!", "Willkommen!", "Benvenuto!", "Bem-vindo!", "Welkom!", "欢迎！", "ようこそ！",
	},
	"welcome.1": {
		"1) Choisissez le dossier des anciens fichiers .DAT.",
		"1) Choose the folder containing the old .DAT files.",
		"1) Elija la carpeta de los antiguos archivos .DAT.",
		"1) Wählen Sie den Ordner mit den alten .DAT-Dateien.",
		"1) Scegli la cartella dei vecchi file .DAT.",
		"1) Escolha a pasta dos antigos arquivos .DAT.",
		"1) Kies de map met de oude .DAT-bestanden.",
		"1) 选择存放旧 .DAT 文件的文件夹。",
		"1) 旧 .DAT ファイルのフォルダーを選択します。",
	},
	"welcome.2": {
		"2) Choisissez la sauvegarde Job Manager (.epidb) contenant la machine.",
		"2) Choose the Job Manager backup (.epidb) that contains the machine.",
		"2) Elija la copia de seguridad de Job Manager (.epidb) que contiene la máquina.",
		"2) Wählen Sie die Job Manager-Sicherung (.epidb), die die Maschine enthält.",
		"2) Scegli il backup di Job Manager (.epidb) che contiene la macchina.",
		"2) Escolha o backup do Job Manager (.epidb) que contém a máquina.",
		"2) Kies de Job Manager-back-up (.epidb) met de machine.",
		"2) 选择包含该机器的 Job Manager 备份 (.epidb)。",
		"2) マシンを含む Job Manager バックアップ (.epidb) を選択します。",
	},
	"welcome.3": {
		"3) Vérifiez la machine de destination et le nom du dossier.",
		"3) Check the target machine and the folder name.",
		"3) Compruebe la máquina de destino y el nombre de la carpeta.",
		"3) Prüfen Sie die Zielmaschine und den Ordnernamen.",
		"3) Verifica la macchina di destinazione e il nome della cartella.",
		"3) Verifique a máquina de destino e o nome da pasta.",
		"3) Controleer de doelmachine en de mapnaam.",
		"3) 检查目标机器和文件夹名称。",
		"3) 変換先のマシンとフォルダー名を確認します。",
	},
	"welcome.4": {
		"4) Réglez les options si nécessaire, puis cliquez sur « Convertir ».",
		"4) Adjust the options if needed, then click \"Convert\".",
		"4) Ajuste las opciones si es necesario y haga clic en «Convertir».",
		"4) Passen Sie bei Bedarf die Optionen an und klicken Sie auf „Konvertieren“.",
		"4) Regola le opzioni se necessario, poi fai clic su «Converti».",
		"4) Ajuste as opções se necessário e clique em \"Converter\".",
		"4) Pas de opties zo nodig aan en klik op 'Converteren'.",
		"4) 如有需要调整选项，然后点击“转换”。",
		"4) 必要に応じてオプションを設定し、「変換」をクリックします。",
	},
	"welcome.5": {
		"Vos fichiers d'origine ne sont jamais modifiés : un nouveau dossier est créé.",
		"Your original files are never modified: a new folder is created.",
		"Sus archivos originales nunca se modifican: se crea una carpeta nueva.",
		"Ihre Originaldateien werden nie verändert: Es wird ein neuer Ordner angelegt.",
		"I file originali non vengono mai modificati: viene creata una nuova cartella.",
		"Os arquivos originais nunca são modificados: uma nova pasta é criada.",
		"Uw originele bestanden worden nooit gewijzigd: er wordt een nieuwe map gemaakt.",
		"原始文件不会被修改：程序会新建一个文件夹。",
		"元のファイルは変更されません。新しいフォルダーが作成されます。",
	},

	// ---------------------------------------------------------------- messages interface
	"msg.title": {
		"VAP Convertisseur", "VAP Converter", "VAP Convertidor", "VAP Konverter", "VAP Convertitore", "VAP Conversor", "VAP Converter", "VAP 转换器", "VAP 変換ツール",
	},
	"info.unreadable": {
		"Dossier illisible", "Folder cannot be read", "Carpeta ilegible", "Ordner nicht lesbar", "Cartella illeggibile", "Pasta ilegível", "Map onleesbaar", "无法读取文件夹", "フォルダーを読み込めません",
	},
	"info.count": {
		"%d fichier(s) .DAT trouvé(s)", "%d .DAT file(s) found", "%d archivo(s) .DAT encontrado(s)", "%d .DAT-Datei(en) gefunden",
		"%d file .DAT trovati", "%d arquivo(s) .DAT encontrado(s)", "%d .DAT-bestand(en) gevonden", "找到 %d 个 .DAT 文件", ".DAT ファイルが %d 件見つかりました",
	},
	"err.openBackup": {
		"Impossible d'ouvrir cette sauvegarde :\n",
		"Cannot open this backup:\n",
		"No se puede abrir esta copia de seguridad:\n",
		"Diese Sicherung kann nicht geöffnet werden:\n",
		"Impossibile aprire questo backup:\n",
		"Não é possível abrir este backup:\n",
		"Kan deze back-up niet openen:\n",
		"无法打开此备份：\n",
		"このバックアップを開けません：\n",
	},
	"err.notBackup": {
		"Ce fichier ne ressemble pas à une sauvegarde Job Manager :\n",
		"This file does not look like a Job Manager backup:\n",
		"Este archivo no parece una copia de seguridad de Job Manager:\n",
		"Diese Datei ist offenbar keine Job Manager-Sicherung:\n",
		"Questo file non sembra un backup di Job Manager:\n",
		"Este arquivo não parece ser um backup do Job Manager:\n",
		"Dit bestand lijkt geen Job Manager-back-up te zijn:\n",
		"此文件似乎不是 Job Manager 备份：\n",
		"このファイルは Job Manager のバックアップではないようです：\n",
	},
	"warn.noMachine": {
		"Aucune machine dans cette sauvegarde.\nCréez d'abord la machine dans Job Manager, puis refaites une sauvegarde.",
		"No machine in this backup.\nFirst add the machine in Job Manager, then make a new backup.",
		"No hay ninguna máquina en esta copia de seguridad.\nPrimero cree la máquina en Job Manager y vuelva a hacer una copia de seguridad.",
		"Diese Sicherung enthält keine Maschine.\nLegen Sie die Maschine zuerst in Job Manager an und erstellen Sie dann eine neue Sicherung.",
		"Nessuna macchina in questo backup.\nCrea prima la macchina in Job Manager, poi esegui un nuovo backup.",
		"Nenhuma máquina neste backup.\nPrimeiro crie a máquina no Job Manager e depois faça um novo backup.",
		"Geen machine in deze back-up.\nVoeg eerst de machine toe in Job Manager en maak daarna een nieuwe back-up.",
		"此备份中没有机器。\n请先在 Job Manager 中添加机器，然后重新备份。",
		"このバックアップにはマシンがありません。\n先に Job Manager でマシンを追加し、もう一度バックアップしてください。",
	},
	"log.backupLoaded": {
		"Sauvegarde chargée : %d machine(s) trouvée(s).",
		"Backup loaded: %d machine(s) found.",
		"Copia de seguridad cargada: %d máquina(s) encontrada(s).",
		"Sicherung geladen: %d Maschine(n) gefunden.",
		"Backup caricato: %d macchina/e trovata/e.",
		"Backup carregado: %d máquina(s) encontrada(s).",
		"Back-up geladen: %d machine(s) gevonden.",
		"备份已加载：找到 %d 台机器。",
		"バックアップを読み込みました：マシン %d 台。",
	},
	"warn.chooseDat": {
		"Choisissez le dossier qui contient les anciens fichiers .DAT.",
		"Choose the folder that contains the old .DAT files.",
		"Elija la carpeta que contiene los antiguos archivos .DAT.",
		"Wählen Sie den Ordner mit den alten .DAT-Dateien.",
		"Scegli la cartella che contiene i vecchi file .DAT.",
		"Escolha a pasta que contém os antigos arquivos .DAT.",
		"Kies de map met de oude .DAT-bestanden.",
		"请选择存放旧 .DAT 文件的文件夹。",
		"旧 .DAT ファイルがあるフォルダーを選択してください。",
	},
	"warn.chooseBackup": {
		"Choisissez la sauvegarde Job Manager (.epidb).",
		"Choose the Job Manager backup (.epidb).",
		"Elija la copia de seguridad de Job Manager (.epidb).",
		"Wählen Sie die Job Manager-Sicherung (.epidb).",
		"Scegli il backup di Job Manager (.epidb).",
		"Escolha o backup do Job Manager (.epidb).",
		"Kies de Job Manager-back-up (.epidb).",
		"请选择 Job Manager 备份 (.epidb)。",
		"Job Manager バックアップ (.epidb) を選択してください。",
	},
	"warn.chooseMachine": {
		"Choisissez la machine de destination.",
		"Choose the target machine.",
		"Elija la máquina de destino.",
		"Wählen Sie die Zielmaschine.",
		"Scegli la macchina di destinazione.",
		"Escolha a máquina de destino.",
		"Kies de doelmachine.",
		"请选择目标机器。",
		"変換先のマシンを選択してください。",
	},
	"ask.replace": {
		"Le dossier « %s » existe déjà dans cette sauvegarde.\n\nLe remplacer par les réglages convertis ?",
		"The folder \"%s\" already exists in this backup.\n\nReplace it with the converted settings?",
		"La carpeta «%s» ya existe en esta copia de seguridad.\n\n¿Reemplazarla por los ajustes convertidos?",
		"Der Ordner „%s“ ist in dieser Sicherung bereits vorhanden.\n\nDurch die konvertierten Einstellungen ersetzen?",
		"La cartella «%s» esiste già in questo backup.\n\nSostituirla con le impostazioni convertite?",
		"A pasta \"%s\" já existe neste backup.\n\nSubstituí-la pelas configurações convertidas?",
		"De map '%s' bestaat al in deze back-up.\n\nVervangen door de geconverteerde instellingen?",
		"此备份中已存在文件夹“%s”。\n\n是否用转换后的设置替换？",
		"フォルダー「%s」はこのバックアップに既に存在します。\n\n変換した設定で置き換えますか？",
	},
	"log.convStart": {
		"── Conversion vers « %s » (%s) ──",
		"── Converting to \"%s\" (%s) ──",
		"── Conversión hacia «%s» (%s) ──",
		"── Konvertierung nach „%s“ (%s) ──",
		"── Conversione verso «%s» (%s) ──",
		"── Conversão para \"%s\" (%s) ──",
		"── Conversie naar '%s' (%s) ──",
		"── 转换到“%s”（%s）──",
		"── 「%s」へ変換（%s）──",
	},
	"log.fail": {
		"✖ Échec : ", "✖ Failed: ", "✖ Error: ", "✖ Fehler: ", "✖ Errore: ", "✖ Falha: ", "✖ Mislukt: ", "✖ 失败：", "✖ 失敗：",
	},
	"err.convFailed": {
		"La conversion a échoué :\n\n",
		"The conversion failed:\n\n",
		"La conversión ha fallado:\n\n",
		"Die Konvertierung ist fehlgeschlagen:\n\n",
		"La conversione non è riuscita:\n\n",
		"A conversão falhou:\n\n",
		"De conversie is mislukt:\n\n",
		"转换失败：\n\n",
		"変換に失敗しました：\n\n",
	},
	"ask.sendReport": {
		"\n\nUn rapport d'erreur a été enregistré :\n%s\n\nL'envoyer à VAP par e-mail ? (le fichier sera sélectionné pour le joindre)",
		"\n\nAn error report has been saved:\n%s\n\nSend it to VAP by e-mail? (the file will be selected so you can attach it)",
		"\n\nSe ha guardado un informe de error:\n%s\n\n¿Enviarlo a VAP por correo? (el archivo quedará seleccionado para adjuntarlo)",
		"\n\nEin Fehlerbericht wurde gespeichert:\n%s\n\nPer E-Mail an VAP senden? (die Datei wird zum Anhängen ausgewählt)",
		"\n\nÈ stato salvato un rapporto di errore:\n%s\n\nInviarlo a VAP per e-mail? (il file verrà selezionato per allegarlo)",
		"\n\nUm relatório de erro foi salvo:\n%s\n\nEnviá-lo à VAP por e-mail? (o arquivo será selecionado para anexá-lo)",
		"\n\nEr is een foutrapport opgeslagen:\n%s\n\nPer e-mail naar VAP sturen? (het bestand wordt geselecteerd om toe te voegen)",
		"\n\n已保存错误报告：\n%s\n\n通过电子邮件发送给 VAP？（将选中该文件以便添加为附件）",
		"\n\nエラーレポートを保存しました：\n%s\n\nVAP にメールで送信しますか？（添付できるようにファイルが選択されます）",
	},
	"log.report": {
		"Rapport d'erreur : ", "Error report: ", "Informe de error: ", "Fehlerbericht: ", "Rapporto di errore: ",
		"Relatório de erro: ", "Foutrapport: ", "错误报告：", "エラーレポート：",
	},
	"file.report": {
		"rapport d'erreur.txt", "error report.txt", "informe de error.txt", "Fehlerbericht.txt", "rapporto di errore.txt",
		"relatório de erro.txt", "foutrapport.txt", "错误报告.txt", "エラーレポート.txt",
	},
	"msg.sendMail": {
		"Écrivez à %s depuis votre messagerie habituelle (Gmail, Outlook…).\nL'adresse a été copiée : collez-la avec Ctrl+V dans le destinataire.",
		"Write to %s from your usual e-mail (Gmail, Outlook…).\nThe address has been copied: paste it with Ctrl+V into the recipient field.",
		"Escriba a %s desde su correo habitual (Gmail, Outlook…).\nLa dirección se ha copiado: péguela con Ctrl+V en el destinatario.",
		"Schreiben Sie an %s aus Ihrem gewohnten E-Mail-Programm (Gmail, Outlook…).\nDie Adresse wurde kopiert: mit Strg+V als Empfänger einfügen.",
		"Scriva a %s dalla sua posta abituale (Gmail, Outlook…).\nL'indirizzo è stato copiato: lo incolli con Ctrl+V nel destinatario.",
		"Escreva para %s a partir do seu e-mail habitual (Gmail, Outlook…).\nO endereço foi copiado: cole-o com Ctrl+V no destinatário.",
		"Schrijf naar %s vanuit uw gebruikelijke e-mail (Gmail, Outlook…).\nHet adres is gekopieerd: plak het met Ctrl+V bij de ontvanger.",
		"请通过您常用的邮箱（Gmail、Outlook…）写信至 %s。\n地址已复制：在收件人栏按 Ctrl+V 粘贴。",
		"いつものメール（Gmail、Outlook など）から %s に送信してください。\nアドレスはコピー済みです：宛先に Ctrl+V で貼り付けてください。",
	},
	"msg.attachReport": {
		"Merci de joindre le rapport d'erreur « %s » (il vient d'être sélectionné dans l'explorateur).",
		"Please attach the error report \"%s\" (it has just been selected in Explorer).",
		"Adjunte el informe de error «%s» (acaba de seleccionarse en el Explorador).",
		"Bitte hängen Sie den Fehlerbericht „%s“ an (er wurde soeben im Explorer ausgewählt).",
		"Alleghi il rapporto di errore «%s» (è appena stato selezionato in Esplora file).",
		"Anexe o relatório de erro «%s» (acabou de ser selecionado no Explorador).",
		"Voeg het foutrapport „%s” toe (het is zojuist geselecteerd in de Verkenner).",
		"请附上错误报告“%s”（已在资源管理器中选中）。",
		"エラーレポート「%s」を添付してください（エクスプローラーで選択済みです）。",
	},
	"log.done": {
		"✔ %d réglage(s) converti(s), %d ignoré(s).",
		"✔ %d setting(s) converted, %d skipped.",
		"✔ %d ajuste(s) convertido(s), %d omitido(s).",
		"✔ %d Einstellung(en) konvertiert, %d übersprungen.",
		"✔ %d impostazione/i convertita/e, %d ignorata/e.",
		"✔ %d configuração(ões) convertida(s), %d ignorada(s).",
		"✔ %d instelling(en) geconverteerd, %d overgeslagen.",
		"✔ 已转换 %d 项设置，跳过 %d 项。",
		"✔ 設定 %d 件を変換、%d 件をスキップ。",
	},
	"log.outDir": {
		"Dossier créé : ", "Folder created: ", "Carpeta creada: ", "Ordner erstellt: ", "Cartella creata: ", "Pasta criada: ", "Map aangemaakt: ", "已创建文件夹：", "作成したフォルダー：",
	},
	"msg.done": {
		"Conversion terminée !\n\n%d réglages ajoutés dans le dossier « %s ».\n\nPour l'installer : Job Manager > Restaurer la base de données, puis choisir le fichier .epidb du dossier créé.\n\n⚠ Avant toute production, vérifiez chaque réglage sur la machine et le matériau (essai sur une chute).\n\nOuvrir le dossier maintenant ?",
		"Conversion complete!\n\n%d settings added to the folder \"%s\".\n\nTo install: Job Manager > Restore Database, then choose the .epidb file in the new folder.\n\n⚠ Before any production, check each setting on the machine and the material (test on scrap).\n\nOpen the folder now?",
		"¡Conversión terminada!\n\n%d ajustes añadidos a la carpeta «%s».\n\nPara instalarlos: Job Manager > Restaurar base de datos (Restore Database) y elija el archivo .epidb de la carpeta creada.\n\n⚠ Antes de producir, compruebe cada ajuste en la máquina y el material (prueba en un retal).\n\n¿Abrir la carpeta ahora?",
		"Konvertierung abgeschlossen!\n\n%d Einstellungen im Ordner „%s“ hinzugefügt.\n\nInstallation: Job Manager > Datenbank wiederherstellen (Restore Database), dann die .epidb-Datei aus dem neuen Ordner wählen.\n\n⚠ Vor der Produktion jede Einstellung an Maschine und Material prüfen (Test an einem Reststück).\n\nOrdner jetzt öffnen?",
		"Conversione completata!\n\n%d impostazioni aggiunte nella cartella «%s».\n\nPer installarle: Job Manager > Ripristina database (Restore Database), poi scegli il file .epidb della cartella creata.\n\n⚠ Prima della produzione, controllare ogni impostazione sulla macchina e sul materiale (prova su uno scarto).\n\nAprire la cartella adesso?",
		"Conversão concluída!\n\n%d configurações adicionadas à pasta \"%s\".\n\nPara instalar: Job Manager > Restaurar banco de dados (Restore Database) e escolha o arquivo .epidb da pasta criada.\n\n⚠ Antes de produzir, verifique cada configuração na máquina e no material (teste numa sobra).\n\nAbrir a pasta agora?",
		"Conversie voltooid!\n\n%d instellingen toegevoegd aan de map '%s'.\n\nInstalleren: Job Manager > Database herstellen (Restore Database) en kies het .epidb-bestand in de nieuwe map.\n\n⚠ Controleer vóór productie elke instelling op de machine en het materiaal (test op een restje).\n\nDe map nu openen?",
		"转换完成！\n\n已向文件夹“%[2]s”添加 %[1]d 项设置。\n\n安装方法：Job Manager > 恢复数据库 (Restore Database)，然后选择新文件夹中的 .epidb 文件。\n\n⚠ 正式生产前，请在机器和材料上检查每项设置（先用废料试切）。\n\n现在打开文件夹吗？",
		"変換が完了しました！\n\nフォルダー「%[2]s」に設定を %[1]d 件追加しました。\n\nインストール方法：Job Manager > データベースの復元 (Restore Database) で、作成されたフォルダーの .epidb ファイルを選択します。\n\n⚠ 本番加工の前に、機械と材料で各設定を確認してください（端材で試し加工）。\n\n今すぐフォルダーを開きますか？",
	},
	"dlg.datTitle": {
		"Dossier des anciens réglages .DAT", "Folder of old .DAT settings", "Carpeta de los antiguos ajustes .DAT", "Ordner der alten .DAT-Einstellungen",
		"Cartella delle vecchie impostazioni .DAT", "Pasta das configurações antigas .DAT", "Map met oude .DAT-instellingen", "旧 .DAT 设置文件夹", "旧 .DAT 設定のフォルダー",
	},
	"dlg.bkTitle": {
		"Sauvegarde Job Manager", "Job Manager backup", "Copia de seguridad de Job Manager", "Job Manager-Sicherung", "Backup di Job Manager", "Backup do Job Manager", "Job Manager-back-up", "Job Manager 备份", "Job Manager バックアップ",
	},
	"dlg.bkFilter": {
		"Sauvegarde Job Manager (*.epidb)", "Job Manager backup (*.epidb)", "Copia de seguridad de Job Manager (*.epidb)", "Job Manager-Sicherung (*.epidb)",
		"Backup di Job Manager (*.epidb)", "Backup do Job Manager (*.epidb)", "Job Manager-back-up (*.epidb)", "Job Manager 备份 (*.epidb)", "Job Manager バックアップ (*.epidb)",
	},
	"dlg.allFilter": {
		"Tous les fichiers (*.*)", "All files (*.*)", "Todos los archivos (*.*)", "Alle Dateien (*.*)", "Tutti i file (*.*)", "Todos os arquivos (*.*)", "Alle bestanden (*.*)", "所有文件 (*.*)", "すべてのファイル (*.*)",
	},
	"info.modernUnavail": {
		"(info) boîte moderne indisponible (0x%08X), boîte classique utilisée",
		"(info) modern dialog unavailable (0x%08X), classic dialog used",
		"", "", "", "", "", "", "",
	},
	"info.modernErr": {
		"(info) boîte moderne en erreur (0x%08X), boîte classique utilisée",
		"(info) modern dialog error (0x%08X), classic dialog used",
		"", "", "", "", "", "", "",
	},
	"err.window": {
		"Impossible de créer la fenêtre.", "Cannot create the window.", "No se puede crear la ventana.", "Das Fenster kann nicht erstellt werden.",
		"Impossibile creare la finestra.", "Não é possível criar a janela.", "Kan het venster niet maken.", "无法创建窗口。", "ウィンドウを作成できません。",
	},
	"help.title": {
		"Notice — Conversion et importation", "Guide — Conversion and import", "Guía — Conversión e importación", "Anleitung — Konvertierung und Import",
		"Guida — Conversione e importazione", "Guia — Conversão e importação", "Handleiding — Conversie en import", "使用说明 — 转换与导入", "操作手順 — 変換とインポート",
	},
	"help.caption": {
		"VAP – Notice / procédure", "VAP – User guide", "VAP – Guía / procedimiento", "VAP – Anleitung", "VAP – Guida / procedura", "VAP – Guia / procedimento", "VAP – Handleiding", "VAP – 使用说明", "VAP – 操作手順",
	},

	// ---------------------------------------------------------------- moteur
	"mode.cut": {
		"Découpe", "Cut", "Corte", "Schneiden", "Taglio", "Corte", "Snijden", "切割", "カット",
	},
	"mode.engrave": {
		"Gravure", "Engrave", "Grabado", "Gravur", "Incisione", "Gravação", "Graveren", "雕刻", "彫刻",
	},
	"mode.combined": {
		"Combiné", "Combined", "Combinado", "Kombiniert", "Combinato", "Combinado", "Gecombineerd", "雕刻+切割", "彫刻＋カット",
	},
	"err.tooSmall": {
		"fichier trop petit", "file too small", "archivo demasiado pequeño", "Datei zu klein", "file troppo piccolo", "arquivo pequeno demais", "bestand te klein", "文件太小", "ファイルが小さすぎます",
	},
	"err.notEpilog": {
		"pas un fichier de réglages Epilog", "not an Epilog settings file", "no es un archivo de ajustes Epilog", "keine Epilog-Einstellungsdatei",
		"non è un file di impostazioni Epilog", "não é um arquivo de configurações Epilog", "geen Epilog-instellingenbestand", "不是 Epilog 设置文件", "Epilog 設定ファイルではありません",
	},
	"err.unknownFormat": {
		"format de réglages inconnu", "unknown settings format", "formato de ajustes desconocido", "unbekanntes Einstellungsformat",
		"formato di impostazioni sconosciuto", "formato de configurações desconhecido", "onbekende instellingenindeling", "未知的设置格式", "不明な設定形式",
	},
	"err.unknownJob": {
		"type de job inconnu (%d)", "unknown job type (%d)", "tipo de trabajo desconocido (%d)", "unbekannter Auftragstyp (%d)",
		"tipo di lavoro sconosciuto (%d)", "tipo de trabalho desconhecido (%d)", "onbekend taaktype (%d)", "未知的作业类型 (%d)", "不明なジョブの種類 (%d)",
	},
	"err.outOfRange": {
		"valeurs vitesse/puissance hors limites", "speed/power values out of range", "valores de velocidad/potencia fuera de rango",
		"Geschwindigkeits-/Leistungswerte außerhalb des Bereichs", "valori di velocità/potenza fuori limite", "valores de velocidade/potência fora dos limites",
		"snelheid-/vermogenswaarden buiten bereik", "速度/功率值超出范围", "速度/出力の値が範囲外です",
	},
	"err.readDir": {
		"lecture du dossier : %w", "reading the folder: %w", "lectura de la carpeta: %w", "Lesen des Ordners: %w",
		"lettura della cartella: %w", "leitura da pasta: %w", "map lezen: %w", "读取文件夹：%w", "フォルダーの読み込み：%w",
	},
	"err.noDat": {
		"aucun fichier .DAT trouvé dans ce dossier", "no .DAT file found in this folder", "no se encontró ningún archivo .DAT en esta carpeta",
		"keine .DAT-Datei in diesem Ordner gefunden", "nessun file .DAT trovato in questa cartella", "nenhum arquivo .DAT encontrado nesta pasta",
		"geen .DAT-bestand gevonden in deze map", "此文件夹中未找到 .DAT 文件", "このフォルダーに .DAT ファイルがありません",
	},
	"skip.fmt": {
		"%s : %s", "%s: %s", "%s: %s", "%s: %s", "%s: %s", "%s: %s", "%s: %s", "%s：%s", "%s：%s",
	},
	"log.skipped": {
		"  ⚠ Ignoré : %s (%s)", "  ⚠ Skipped: %s (%s)", "  ⚠ Omitido: %s (%s)", "  ⚠ Übersprungen: %s (%s)",
		"  ⚠ Ignorato: %s (%s)", "  ⚠ Ignorado: %s (%s)", "  ⚠ Overgeslagen: %s (%s)", "  ⚠ 已跳过：%s（%s）", "  ⚠ スキップ：%s（%s）",
	},
	"skip.testGrid": {
		"grille de test (color mapping)", "test grid (color mapping)", "cuadrícula de prueba (color mapping)", "Testraster (Color Mapping)",
		"griglia di prova (color mapping)", "grade de teste (color mapping)", "testraster (color mapping)", "测试网格（颜色映射）", "テストグリッド（カラーマッピング）",
	},
	"skip.testGridLog": {
		"grille de test", "test grid", "cuadrícula de prueba", "Testraster", "griglia di prova", "grade de teste", "testraster", "测试网格", "テストグリッド",
	},
	"skip.dup": {
		"doublon (sous-dossier)", "duplicate (subfolder)", "duplicado (subcarpeta)", "Duplikat (Unterordner)", "duplicato (sottocartella)",
		"duplicado (subpasta)", "dubbel (submap)", "重复（子文件夹）", "重複（サブフォルダー）",
	},
	"err.noUsable": {
		"aucun réglage exploitable dans ce dossier", "no usable setting in this folder", "ningún ajuste utilizable en esta carpeta",
		"keine verwendbare Einstellung in diesem Ordner", "nessuna impostazione utilizzabile in questa cartella", "nenhuma configuração utilizável nesta pasta",
		"geen bruikbare instelling in deze map", "此文件夹中没有可用的设置", "このフォルダーに使用できる設定がありません",
	},
	"log.read": {
		"%d fichiers .DAT lus.", "%d .DAT files read.", "%d archivos .DAT leídos.", "%d .DAT-Dateien gelesen.", "%d file .DAT letti.",
		"%d arquivos .DAT lidos.", "%d .DAT-bestanden gelezen.", "已读取 %d 个 .DAT 文件。", ".DAT ファイルを %d 件読み込みました。",
	},
	"log.replaced": {
		"Dossier existant « %s » vidé et remplacé.", "Existing folder \"%s\" emptied and replaced.", "Carpeta existente «%s» vaciada y reemplazada.",
		"Vorhandener Ordner „%s“ geleert und ersetzt.", "Cartella esistente «%s» svuotata e sostituita.", "Pasta existente \"%s\" esvaziada e substituída.",
		"Bestaande map '%s' geleegd en vervangen.", "已清空并替换现有文件夹“%s”。", "既存のフォルダー「%s」を空にして置き換えました。",
	},
	"err.createFolder": {
		"création du dossier : %w", "creating the folder: %w", "creación de la carpeta: %w", "Anlegen des Ordners: %w",
		"creazione della cartella: %w", "criação da pasta: %w", "map aanmaken: %w", "创建文件夹：%w", "フォルダーの作成：%w",
	},
	"err.addSetting": {
		"ajout de « %s » : %w", "adding \"%s\": %w", "añadiendo «%s»: %w", "Hinzufügen von „%s“: %w", "aggiunta di «%s»: %w",
		"adição de \"%s\": %w", "toevoegen van '%s': %w", "添加“%s”：%w", "「%s」の追加：%w",
	},
	"err.integrity": {
		"le contrôle d'intégrité de la base a échoué", "the database integrity check failed", "la comprobación de integridad de la base de datos ha fallado",
		"die Integritätsprüfung der Datenbank ist fehlgeschlagen", "il controllo di integrità del database non è riuscito", "a verificação de integridade do banco de dados falhou",
		"de integriteitscontrole van de database is mislukt", "数据库完整性检查失败", "データベースの整合性チェックに失敗しました",
	},
	"csv.cols": {
		"Sous-dossier|Matériau|Type|Résolution (DPI)|Gravure vitesse %|Gravure puissance %|Découpe vitesse %|Découpe puissance %|Découpe fréquence %|Épaisseur (mm)|Remarque",
		"Subfolder|Material|Type|Resolution (DPI)|Engrave speed %|Engrave power %|Cut speed %|Cut power %|Cut frequency %|Thickness (mm)|Note",
		"Subcarpeta|Material|Tipo|Resolución (DPI)|Grabado velocidad %|Grabado potencia %|Corte velocidad %|Corte potencia %|Corte frecuencia %|Espesor (mm)|Observación",
		"Unterordner|Material|Typ|Auflösung (DPI)|Gravur Geschwindigkeit %|Gravur Leistung %|Schneiden Geschwindigkeit %|Schneiden Leistung %|Schneiden Frequenz %|Dicke (mm)|Hinweis",
		"Sottocartella|Materiale|Tipo|Risoluzione (DPI)|Incisione velocità %|Incisione potenza %|Taglio velocità %|Taglio potenza %|Taglio frequenza %|Spessore (mm)|Nota",
		"Subpasta|Material|Tipo|Resolução (DPI)|Gravação velocidade %|Gravação potência %|Corte velocidade %|Corte potência %|Corte frequência %|Espessura (mm)|Observação",
		"Submap|Materiaal|Type|Resolutie (DPI)|Graveren snelheid %|Graveren vermogen %|Snijden snelheid %|Snijden vermogen %|Snijden frequentie %|Dikte (mm)|Opmerking",
		"子文件夹|材料|类型|分辨率 (DPI)|雕刻速度 %|雕刻功率 %|切割速度 %|切割功率 %|切割频率 %|厚度 (mm)|备注",
		"サブフォルダー|材料|種類|解像度 (DPI)|彫刻速度 %|彫刻出力 %|カット速度 %|カット出力 %|カット周波数 %|厚さ (mm)|備考",
	},
	"sub.root": {
		"(racine)", "(root)", "(raíz)", "(Stammordner)", "(radice)", "(raiz)", "(hoofdmap)", "（根目录）", "（ルート）",
	},
	"sub.metal": {
		"Métal", "Metal", "Metal", "Metall", "Metallo", "Metal", "Metaal", "金属", "金属",
	},
	"sub.wood": {
		"Bois", "Wood", "Madera", "Holz", "Legno", "Madeira", "Hout", "木材", "木材",
	},
	"sub.plastic": {
		"Plastique", "Plastic", "Plástico", "Kunststoff", "Plastica", "Plástico", "Kunststof", "塑料", "プラスチック",
	},
	"rem.photo": {
		"Trame photo (Stucki) à régler", "Photo dithering (Stucki) to set", "Tramado fotográfico (Stucki) por ajustar", "Foto-Rasterung (Stucki) einstellen",
		"Retinatura foto (Stucki) da regolare", "Pontilhado de foto (Stucki) a ajustar", "Foto-dithering (Stucki) instellen", "需设置照片抖动（Stucki）", "写真用ディザ（Stucki）の設定が必要",
	},
	"rem.stamp": {
		"Mode tampon à régler", "Stamp mode to set", "Modo sello por ajustar", "Stempelmodus einstellen", "Modalità timbro da regolare",
		"Modo carimbo a ajustar", "Stempelmodus instellen", "需设置印章模式", "スタンプモードの設定が必要",
	},
	"rem.passes": {
		"%d passes (lu dans le nom)", "%d passes (read from the name)", "%d pasadas (leído del nombre)", "%d Durchgänge (aus dem Namen)",
		"%d passate (dal nome)", "%d passagens (lido do nome)", "%d doorgangen (uit de naam)", "%d 遍（取自名称）", "%d パス（名前から読み取り）",
	},
	"rem.colors": {
		"Gestion des couleurs : %d couleurs", "Color mapping: %d colors", "Gestión de colores: %d colores", "Farbzuordnung: %d Farben",
		"Gestione colori: %d colori", "Gestão de cores: %d cores", "Kleurbeheer: %d kleuren", "颜色管理：%d 种颜色", "カラー管理：%d 色",
	},
	"csv.skipped": {
		"Ignoré", "Skipped", "Omitido", "Übersprungen", "Ignorato", "Ignorado", "Overgeslagen", "已跳过", "スキップ",
	},
	"out.suffix": {
		" - Conversion Job Manager", " - Job Manager conversion", " - Conversión Job Manager", " - Job Manager-Konvertierung",
		" - Conversione Job Manager", " - Conversão Job Manager", " - Job Manager-conversie", " - Job Manager 转换", " - Job Manager 変換",
	},
	"file.recap": {
		"Récapitulatif réglages.csv", "Settings summary.csv", "Resumen de ajustes.csv", "Übersicht Einstellungen.csv", "Riepilogo impostazioni.csv",
		"Resumo das configurações.csv", "Overzicht instellingen.csv", "设置汇总.csv", "設定一覧.csv",
	},
	"file.readme": {
		"LISEZ-MOI.txt", "README.txt", "LEAME.txt", "LIESMICH.txt", "LEGGIMI.txt", "LEIA-ME.txt", "LEESMIJ.txt", "说明.txt", "お読みください.txt",
	},
	"fmt.date": {
		"02/01/2006 15:04", "2006-01-02 15:04", "02/01/2006 15:04", "02.01.2006 15:04", "02/01/2006 15:04", "02/01/2006 15:04", "02-01-2006 15:04", "2006/01/02 15:04", "2006/01/02 15:04",
	},
	"log.copy": {
		"Copie de la sauvegarde…", "Copying the backup…", "Copiando la copia de seguridad…", "Sicherung wird kopiert…", "Copia del backup…",
		"Copiando o backup…", "Back-up kopiëren…", "正在复制备份…", "バックアップをコピーしています…",
	},
	"err.mkOut": {
		"création du dossier de sortie : %w", "creating the output folder: %w", "creación de la carpeta de salida: %w", "Anlegen des Ausgabeordners: %w",
		"creazione della cartella di uscita: %w", "criação da pasta de saída: %w", "uitvoermap aanmaken: %w", "创建输出文件夹：%w", "出力フォルダーの作成：%w",
	},
	"err.copy": {
		"copie de la sauvegarde : %w", "copying the backup: %w", "copia de la copia de seguridad: %w", "Kopieren der Sicherung: %w",
		"copia del backup: %w", "cópia do backup: %w", "back-up kopiëren: %w", "复制备份：%w", "バックアップのコピー：%w",
	},
	"log.recapFail": {
		"⚠ Récapitulatif non écrit : ", "⚠ Summary not written: ", "⚠ Resumen no escrito: ", "⚠ Übersicht nicht geschrieben: ",
		"⚠ Riepilogo non scritto: ", "⚠ Resumo não gravado: ", "⚠ Overzicht niet geschreven: ", "⚠ 未能写入汇总：", "⚠ 一覧を書き込めませんでした：",
	},

	// ---------------------------------------------------------------- SQLite
	"err.noSqlite": {
		"composant SQLite introuvable (winsqlite3.dll). Windows 10 ou 11 est nécessaire",
		"SQLite component not found (winsqlite3.dll). Windows 10 or 11 is required",
		"componente SQLite no encontrado (winsqlite3.dll). Se necesita Windows 10 u 11",
		"SQLite-Komponente nicht gefunden (winsqlite3.dll). Windows 10 oder 11 ist erforderlich",
		"componente SQLite non trovato (winsqlite3.dll). È necessario Windows 10 o 11",
		"componente SQLite não encontrado (winsqlite3.dll). É necessário o Windows 10 ou 11",
		"SQLite-component niet gevonden (winsqlite3.dll). Windows 10 of 11 is vereist",
		"未找到 SQLite 组件 (winsqlite3.dll)。需要 Windows 10 或 11",
		"SQLite コンポーネント (winsqlite3.dll) が見つかりません。Windows 10 または 11 が必要です",
	},
	"err.sqliteIncomplete": {
		"SQLite incomplet : %v", "incomplete SQLite: %v", "SQLite incompleto: %v", "SQLite unvollständig: %v", "SQLite incompleto: %v",
		"SQLite incompleto: %v", "SQLite onvolledig: %v", "SQLite 不完整：%v", "SQLite が不完全です：%v",
	},
	"err.database": {
		"base de données : %s", "database: %s", "base de datos: %s", "Datenbank: %s", "database: %s", "banco de dados: %s", "database: %s", "数据库：%s", "データベース：%s",
	},
}
