package main

// Notice (fenêtre d'aide) et LISEZ-MOI, dans les 9 langues.
// Ordre : fr, en, es, de, it, pt, nl, zh, ja.
// Les noms des menus Job Manager sont rappelés en anglais entre parenthèses
// (Backup Database / Restore Database) pour être reconnus quelle que soit la
// langue de Job Manager.

func noticeText() string { return notices[cur] }

// readmeText : arguments = fichier .epidb, dossier, dossier, machine, date.
func readmeText() string { return readmes[cur] }

var notices = [nLang]string{
	// ------------------------------------------------------------------ FR
	`AVANT DE COMMENCER — SAUVEGARDE OBLIGATOIRE
La conversion part d'une sauvegarde de la base Job Manager qui contient déjà la machine de destination.

MACHINES CONCERNÉES
  Anciens réglages .DAT lus (anciens pilotes Epilog, avant Job Manager) :
    • Mini 18 / Mini 24 et Helix (CO2)
    • Zing 16 / Zing 24 (CO2)
    • Fusion 32 / Fusion 40 (CO2)
    • Fusion 32 Fibre / Fusion 40 Fibre (choisir « Fibre » dans Source laser)
  Machine de destination : la même machine, déclarée dans Job Manager
  (Mini 18, Mini 24, Helix, Zing 16, Zing 24, Fusion 32, Fusion 40).
  Non concernées : Fusion Pro, Fusion Edge, Fusion Maker (dernière génération, sans .DAT).
  Plus de 3 100 anciens réglages contrôlés valeur par valeur (Mini 18, Mini 24, Helix, Zing 16, Zing 24,
  Fusion 32, Fusion 40). Vérifiez toujours chaque réglage sur la machine avant production.

ÉTAPE 1 — Préparer Job Manager (sur le PC relié à la machine)
  1. Ouvrir Epilog Job Manager.
  2. Vérifier que la machine de destination est présente dans Job Manager.
     Si elle n'y est pas : l'ajouter, avec le nom voulu (ex. « MINI24 40 watts »).
  3. Faire une sauvegarde de la base de données :
     Paramètres du programme (roue crantée) > Sauvegarder la base de données (Backup Database).
     Enregistrer le fichier .epidb (bureau ou clé USB).
  4. Garder cette sauvegarde de côté : c'est votre retour arrière en cas de problème.

ÉTAPE 2 — Préparer les anciens réglages
  5. Copier le dossier des anciens fichiers .DAT de la machine (un dossier par machine).
     Emplacement habituel : Documents\Epilog\Engraving_Settings

ÉTAPE 3 — Convertir avec ce logiciel
  6. Section 1 : « Parcourir… » et choisir le dossier des .DAT.
     Cocher « Inclure les sous-dossiers » si les réglages sont rangés dans des sous-dossiers.
  7. Section 2 : « Parcourir… » et choisir la sauvegarde .epidb faite à l'étape 1.
  8. Section 3 : vérifier la machine de destination et le nom du dossier
     (par défaut, le nom de la machine).
  9. Section 4 : régler les options (assistance d'air, tri des vecteurs, position du dossier,
     source laser : choisir « Fibre » pour une machine fibre).
 10. Cliquer sur « Convertir ».
     Un dossier « … - Conversion Job Manager » est créé à côté du dossier des .DAT.
     Il contient la nouvelle base (.epidb), le récapitulatif (Excel) et un LISEZ-MOI.

ÉTAPE 4 — Importer dans Job Manager
 11. Job Manager > Paramètres du programme > Restaurer la base de données (Restore Database).
 12. Choisir le fichier .epidb du dossier « … - Conversion Job Manager ».
     Job Manager redémarre.
 13. Dans un travail, cliquer sur l'icône « Importer les paramètres du travail » (en bas à droite).
     Le dossier de la machine apparaît en tête de liste, avec les sous-dossiers Métal / Bois / Plastique.

À SAVOIR
  • Les fichiers d'origine (.DAT et sauvegarde) ne sont jamais modifiés.
  • La restauration remplace toute la base Job Manager : utilisez une sauvegarde récente (étape 1)
    pour ne perdre aucun travail ni réglage.
  • Tampon : talus et épaisseur prennent les valeurs par défaut de Job Manager
    (voir la colonne « Remarque » du récapitulatif).
  • Tester quelques matériaux sur une chute avant de lancer la production.
  • En cas de souci : restaurer la sauvegarde d'origine faite à l'étape 1.`,

	// ------------------------------------------------------------------ EN
	`BEFORE YOU START — A BACKUP IS REQUIRED
The conversion starts from a backup of the Job Manager database that already contains the target machine.

SUPPORTED MACHINES
  Old .DAT settings read (legacy Epilog drivers, before Job Manager):
    • Mini 18 / Mini 24 and Helix (CO2)
    • Zing 16 / Zing 24 (CO2)
    • Fusion 32 / Fusion 40 (CO2)
    • Fusion 32 Fiber / Fusion 40 Fiber (choose "Fiber" in Laser source)
  Target machine: the same machine, set up in Job Manager
  (Mini 18, Mini 24, Helix, Zing 16, Zing 24, Fusion 32, Fusion 40).
  Not supported: Fusion Pro, Fusion Edge, Fusion Maker (latest generation, no .DAT files).
  Over 3,100 old settings checked value by value (Mini 18, Mini 24, Helix, Zing 16, Zing 24,
  Fusion 32, Fusion 40). Always check each setting on the machine before production.

STEP 1 — Prepare Job Manager (on the PC connected to the machine)
  1. Open Epilog Job Manager.
  2. Check that the target machine is listed in Job Manager.
     If it is not: add it, with the name you want (e.g. "MINI24 40 watts").
  3. Back up the database:
     Program settings (gear icon) > Backup Database.
     Save the .epidb file (desktop or USB stick).
  4. Keep this backup aside: it is your way back if anything goes wrong.

STEP 2 — Prepare the old settings
  5. Copy the machine's folder of old .DAT files (one folder per machine).
     Usual location: Documents\Epilog\Engraving_Settings

STEP 3 — Convert with this software
  6. Section 1: "Browse…" and choose the .DAT folder.
     Tick "Include subfolders" if the settings are stored in subfolders.
  7. Section 2: "Browse…" and choose the .epidb backup made in step 1.
  8. Section 3: check the target machine and the folder name
     (by default, the machine name).
  9. Section 4: set the options (air assist, vector sorting, folder position,
     laser source: choose "Fiber" for a fiber machine).
 10. Click "Convert".
     A "… - Job Manager conversion" folder is created next to the .DAT folder.
     It contains the new database (.epidb), the summary (Excel) and a README.

STEP 4 — Import into Job Manager
 11. Job Manager > Program settings > Restore Database.
 12. Choose the .epidb file from the "… - Job Manager conversion" folder.
     Job Manager restarts.
 13. In a job, click the job settings icon (bottom right).
     The machine folder appears at the top of the list, with the Metal / Wood / Plastic subfolders.

GOOD TO KNOW
  • The original files (.DAT and backup) are never modified.
  • Restoring replaces the whole Job Manager database: use a recent backup (step 1)
    so you do not lose any job or setting.
  • Stamp: shoulder and thickness use the Job Manager default values
    (see the "Note" column of the summary).
  • Test a few materials on scrap before starting production.
  • If anything goes wrong: restore the original backup made in step 1.`,

	// ------------------------------------------------------------------ ES
	`ANTES DE EMPEZAR — COPIA DE SEGURIDAD OBLIGATORIA
La conversión parte de una copia de seguridad de la base de datos de Job Manager que ya contiene la máquina de destino.

MÁQUINAS COMPATIBLES
  Ajustes .DAT antiguos leídos (antiguos controladores Epilog, anteriores a Job Manager):
    • Mini 18 / Mini 24 y Helix (CO2)
    • Zing 16 / Zing 24 (CO2)
    • Fusion 32 / Fusion 40 (CO2)
    • Fusion 32 Fibra / Fusion 40 Fibra (elija «Fibra» en Fuente láser)
  Máquina de destino: la misma máquina, configurada en Job Manager
  (Mini 18, Mini 24, Helix, Zing 16, Zing 24, Fusion 32, Fusion 40).
  No compatibles: Fusion Pro, Fusion Edge, Fusion Maker (última generación, sin .DAT).
  Más de 3100 ajustes antiguos comprobados valor por valor (Mini 18, Mini 24, Helix, Zing 16, Zing 24,
  Fusion 32, Fusion 40). Compruebe siempre cada ajuste en la máquina antes de producir.

PASO 1 — Preparar Job Manager (en el PC conectado a la máquina)
  1. Abra Epilog Job Manager.
  2. Compruebe que la máquina de destino aparece en Job Manager.
     Si no aparece: añádala con el nombre deseado (p. ej. «MINI24 40 vatios»).
  3. Haga una copia de seguridad de la base de datos:
     Configuración del programa (rueda dentada) > Copia de seguridad de la base de datos (Backup Database).
     Guarde el archivo .epidb (escritorio o memoria USB).
  4. Guarde esta copia aparte: es su vuelta atrás en caso de problema.

PASO 2 — Preparar los ajustes antiguos
  5. Copie la carpeta de los antiguos archivos .DAT de la máquina (una carpeta por máquina).
     Ubicación habitual: Documentos\Epilog\Engraving_Settings

PASO 3 — Convertir con este programa
  6. Sección 1: «Examinar…» y elija la carpeta de los .DAT.
     Marque «Incluir subcarpetas» si los ajustes están en subcarpetas.
  7. Sección 2: «Examinar…» y elija la copia .epidb hecha en el paso 1.
  8. Sección 3: compruebe la máquina de destino y el nombre de la carpeta
     (por defecto, el nombre de la máquina).
  9. Sección 4: ajuste las opciones (asistencia de aire, orden de vectores, posición de la carpeta,
     fuente láser: elija «Fibra» para una máquina de fibra).
 10. Haga clic en «Convertir».
     Se crea una carpeta «… - Conversión Job Manager» junto a la carpeta de los .DAT.
     Contiene la nueva base (.epidb), el resumen (Excel) y un LEAME.

PASO 4 — Importar en Job Manager
 11. Job Manager > Configuración del programa > Restaurar base de datos (Restore Database).
 12. Elija el archivo .epidb de la carpeta «… - Conversión Job Manager».
     Job Manager se reinicia.
 13. En un trabajo, haga clic en el icono de ajustes del trabajo (abajo a la derecha).
     La carpeta de la máquina aparece al principio de la lista, con las subcarpetas Metal / Madera / Plástico.

A TENER EN CUENTA
  • Los archivos originales (.DAT y copia de seguridad) nunca se modifican.
  • La restauración reemplaza toda la base de Job Manager: use una copia reciente (paso 1)
    para no perder ningún trabajo ni ajuste.
  • Sello: talud y grosor toman los valores por defecto de Job Manager
    (vea la columna «Observación» del resumen).
  • Pruebe algunos materiales en un retal antes de empezar la producción.
  • En caso de problema: restaure la copia de seguridad original del paso 1.`,

	// ------------------------------------------------------------------ DE
	`VOR DEM START — SICHERUNG ERFORDERLICH
Die Konvertierung geht von einer Sicherung der Job Manager-Datenbank aus, die die Zielmaschine bereits enthält.

UNTERSTÜTZTE MASCHINEN
  Gelesene alte .DAT-Einstellungen (alte Epilog-Treiber, vor Job Manager):
    • Mini 18 / Mini 24 und Helix (CO2)
    • Zing 16 / Zing 24 (CO2)
    • Fusion 32 / Fusion 40 (CO2)
    • Fusion 32 Faser / Fusion 40 Faser („Faser“ unter Laserquelle wählen)
  Zielmaschine: dieselbe Maschine, in Job Manager angelegt
  (Mini 18, Mini 24, Helix, Zing 16, Zing 24, Fusion 32, Fusion 40).
  Nicht unterstützt: Fusion Pro, Fusion Edge, Fusion Maker (neueste Generation, ohne .DAT).
  Über 3.100 alte Einstellungen Wert für Wert geprüft (Mini 18, Mini 24, Helix, Zing 16, Zing 24,
  Fusion 32, Fusion 40). Jede Einstellung vor der Produktion an der Maschine prüfen.

SCHRITT 1 — Job Manager vorbereiten (am PC, der mit der Maschine verbunden ist)
  1. Epilog Job Manager öffnen.
  2. Prüfen, ob die Zielmaschine in Job Manager vorhanden ist.
     Falls nicht: sie mit dem gewünschten Namen hinzufügen (z. B. „MINI24 40 Watt“).
  3. Die Datenbank sichern:
     Programmeinstellungen (Zahnrad) > Datenbank sichern (Backup Database).
     Die .epidb-Datei speichern (Desktop oder USB-Stick).
  4. Diese Sicherung aufbewahren: Sie ist Ihr Rückweg, falls etwas schiefgeht.

SCHRITT 2 — Alte Einstellungen vorbereiten
  5. Den Ordner mit den alten .DAT-Dateien der Maschine kopieren (ein Ordner pro Maschine).
     Üblicher Speicherort: Dokumente\Epilog\Engraving_Settings

SCHRITT 3 — Mit diesem Programm konvertieren
  6. Abschnitt 1: „Durchsuchen…“ und den .DAT-Ordner wählen.
     „Unterordner einbeziehen“ ankreuzen, wenn die Einstellungen in Unterordnern liegen.
  7. Abschnitt 2: „Durchsuchen…“ und die in Schritt 1 erstellte .epidb-Sicherung wählen.
  8. Abschnitt 3: Zielmaschine und Ordnernamen prüfen
     (standardmäßig der Maschinenname).
  9. Abschnitt 4: Optionen einstellen (Air Assist, Vektorsortierung, Ordnerposition,
     Laserquelle: „Faser“ für eine Fasermaschine wählen).
 10. Auf „Konvertieren“ klicken.
     Neben dem .DAT-Ordner wird ein Ordner „… - Job Manager-Konvertierung“ angelegt.
     Er enthält die neue Datenbank (.epidb), die Übersicht (Excel) und eine LIESMICH-Datei.

SCHRITT 4 — In Job Manager importieren
 11. Job Manager > Programmeinstellungen > Datenbank wiederherstellen (Restore Database).
 12. Die .epidb-Datei aus dem Ordner „… - Job Manager-Konvertierung“ wählen.
     Job Manager startet neu.
 13. In einem Auftrag auf das Symbol für die Auftragseinstellungen klicken (unten rechts).
     Der Maschinenordner erscheint oben in der Liste, mit den Unterordnern Metall / Holz / Kunststoff.

GUT ZU WISSEN
  • Die Originaldateien (.DAT und Sicherung) werden nie verändert.
  • Die Wiederherstellung ersetzt die gesamte Job Manager-Datenbank: Verwenden Sie eine aktuelle
    Sicherung (Schritt 1), damit keine Aufträge oder Einstellungen verloren gehen.
  • Stempel: Schulter und Stärke übernehmen die Standardwerte von Job Manager
    (siehe Spalte „Hinweis“ der Übersicht).
  • Vor der Produktion einige Materialien an einem Reststück testen.
  • Bei Problemen: die Originalsicherung aus Schritt 1 wiederherstellen.`,

	// ------------------------------------------------------------------ IT
	`PRIMA DI INIZIARE — BACKUP OBBLIGATORIO
La conversione parte da un backup del database di Job Manager che contiene già la macchina di destinazione.

MACCHINE INTERESSATE
  Vecchie impostazioni .DAT lette (vecchi driver Epilog, prima di Job Manager):
    • Mini 18 / Mini 24 e Helix (CO2)
    • Zing 16 / Zing 24 (CO2)
    • Fusion 32 / Fusion 40 (CO2)
    • Fusion 32 Fibra / Fusion 40 Fibra (scegliere «Fibra» in Sorgente laser)
  Macchina di destinazione: la stessa macchina, configurata in Job Manager
  (Mini 18, Mini 24, Helix, Zing 16, Zing 24, Fusion 32, Fusion 40).
  Non interessate: Fusion Pro, Fusion Edge, Fusion Maker (ultima generazione, senza .DAT).
  Oltre 3.100 vecchie impostazioni controllate valore per valore (Mini 18, Mini 24, Helix, Zing 16, Zing 24,
  Fusion 32, Fusion 40). Controllare sempre ogni impostazione sulla macchina prima della produzione.

PASSO 1 — Preparare Job Manager (sul PC collegato alla macchina)
  1. Aprire Epilog Job Manager.
  2. Verificare che la macchina di destinazione sia presente in Job Manager.
     Se non c'è: aggiungerla con il nome desiderato (es. «MINI24 40 watt»).
  3. Eseguire un backup del database:
     Impostazioni del programma (ingranaggio) > Backup del database (Backup Database).
     Salvare il file .epidb (desktop o chiavetta USB).
  4. Conservare questo backup: è il vostro ritorno indietro in caso di problemi.

PASSO 2 — Preparare le vecchie impostazioni
  5. Copiare la cartella dei vecchi file .DAT della macchina (una cartella per macchina).
     Posizione abituale: Documenti\Epilog\Engraving_Settings

PASSO 3 — Convertire con questo programma
  6. Sezione 1: «Sfoglia…» e scegliere la cartella dei .DAT.
     Spuntare «Includi sottocartelle» se le impostazioni sono in sottocartelle.
  7. Sezione 2: «Sfoglia…» e scegliere il backup .epidb fatto al passo 1.
  8. Sezione 3: verificare la macchina di destinazione e il nome della cartella
     (per impostazione predefinita, il nome della macchina).
  9. Sezione 4: regolare le opzioni (assistenza aria, ordinamento vettori, posizione della cartella,
     sorgente laser: scegliere «Fibra» per una macchina a fibra).
 10. Fare clic su «Converti».
     Accanto alla cartella dei .DAT viene creata una cartella «… - Conversione Job Manager».
     Contiene il nuovo database (.epidb), il riepilogo (Excel) e un LEGGIMI.

PASSO 4 — Importare in Job Manager
 11. Job Manager > Impostazioni del programma > Ripristina database (Restore Database).
 12. Scegliere il file .epidb della cartella «… - Conversione Job Manager».
     Job Manager si riavvia.
 13. In un lavoro, fare clic sull'icona delle impostazioni del lavoro (in basso a destra).
     La cartella della macchina compare in cima all'elenco, con le sottocartelle Metallo / Legno / Plastica.

DA SAPERE
  • I file originali (.DAT e backup) non vengono mai modificati.
  • Il ripristino sostituisce l'intero database di Job Manager: usare un backup recente (passo 1)
    per non perdere lavori né impostazioni.
  • Timbro: spalla e spessore prendono i valori predefiniti di Job Manager
    (vedere la colonna «Nota» del riepilogo).
  • Provare alcuni materiali su uno scarto prima di avviare la produzione.
  • In caso di problemi: ripristinare il backup originale del passo 1.`,

	// ------------------------------------------------------------------ PT
	`ANTES DE COMEÇAR — BACKUP OBRIGATÓRIO
A conversão parte de um backup do banco de dados do Job Manager que já contém a máquina de destino.

MÁQUINAS COMPATÍVEIS
  Configurações .DAT antigas lidas (antigos drivers Epilog, antes do Job Manager):
    • Mini 18 / Mini 24 e Helix (CO2)
    • Zing 16 / Zing 24 (CO2)
    • Fusion 32 / Fusion 40 (CO2)
    • Fusion 32 Fibra / Fusion 40 Fibra (escolha "Fibra" em Fonte laser)
  Máquina de destino: a mesma máquina, configurada no Job Manager
  (Mini 18, Mini 24, Helix, Zing 16, Zing 24, Fusion 32, Fusion 40).
  Não compatíveis: Fusion Pro, Fusion Edge, Fusion Maker (última geração, sem .DAT).
  Mais de 3.100 configurações antigas verificadas valor a valor (Mini 18, Mini 24, Helix, Zing 16, Zing 24,
  Fusion 32, Fusion 40). Verifique sempre cada configuração na máquina antes de produzir.

ETAPA 1 — Preparar o Job Manager (no PC ligado à máquina)
  1. Abra o Epilog Job Manager.
  2. Verifique se a máquina de destino aparece no Job Manager.
     Se não aparecer: adicione-a com o nome desejado (ex. "MINI24 40 watts").
  3. Faça um backup do banco de dados:
     Configurações do programa (engrenagem) > Fazer backup do banco de dados (Backup Database).
     Salve o arquivo .epidb (área de trabalho ou pen drive).
  4. Guarde esse backup: é o seu caminho de volta em caso de problema.

ETAPA 2 — Preparar as configurações antigas
  5. Copie a pasta dos antigos arquivos .DAT da máquina (uma pasta por máquina).
     Local habitual: Documentos\Epilog\Engraving_Settings

ETAPA 3 — Converter com este programa
  6. Seção 1: "Procurar…" e escolha a pasta dos .DAT.
     Marque "Incluir subpastas" se as configurações estiverem em subpastas.
  7. Seção 2: "Procurar…" e escolha o backup .epidb feito na etapa 1.
  8. Seção 3: verifique a máquina de destino e o nome da pasta
     (por padrão, o nome da máquina).
  9. Seção 4: ajuste as opções (assistência de ar, ordem dos vetores, posição da pasta,
     fonte laser: escolha "Fibra" para uma máquina de fibra).
 10. Clique em "Converter".
     Uma pasta "… - Conversão Job Manager" é criada ao lado da pasta dos .DAT.
     Ela contém o novo banco de dados (.epidb), o resumo (Excel) e um LEIA-ME.

ETAPA 4 — Importar no Job Manager
 11. Job Manager > Configurações do programa > Restaurar banco de dados (Restore Database).
 12. Escolha o arquivo .epidb da pasta "… - Conversão Job Manager".
     O Job Manager reinicia.
 13. Num trabalho, clique no ícone das configurações do trabalho (canto inferior direito).
     A pasta da máquina aparece no topo da lista, com as subpastas Metal / Madeira / Plástico.

IMPORTANTE
  • Os arquivos originais (.DAT e backup) nunca são modificados.
  • A restauração substitui todo o banco de dados do Job Manager: use um backup recente (etapa 1)
    para não perder nenhum trabalho ou configuração.
  • Carimbo: talude e espessura usam os valores padrão do Job Manager
    (veja a coluna "Observação" do resumo).
  • Teste alguns materiais numa sobra antes de iniciar a produção.
  • Em caso de problema: restaure o backup original feito na etapa 1.`,

	// ------------------------------------------------------------------ NL
	`VOORDAT U BEGINT — BACK-UP VERPLICHT
De conversie gaat uit van een back-up van de Job Manager-database waarin de doelmachine al staat.

ONDERSTEUNDE MACHINES
  Gelezen oude .DAT-instellingen (oude Epilog-drivers, van vóór Job Manager):
    • Mini 18 / Mini 24 en Helix (CO2)
    • Zing 16 / Zing 24 (CO2)
    • Fusion 32 / Fusion 40 (CO2)
    • Fusion 32 Fiber / Fusion 40 Fiber (kies 'Fiber' bij Laserbron)
  Doelmachine: dezelfde machine, ingesteld in Job Manager
  (Mini 18, Mini 24, Helix, Zing 16, Zing 24, Fusion 32, Fusion 40).
  Niet ondersteund: Fusion Pro, Fusion Edge, Fusion Maker (nieuwste generatie, zonder .DAT).
  Meer dan 3.100 oude instellingen waarde voor waarde gecontroleerd (Mini 18, Mini 24, Helix, Zing 16, Zing 24,
  Fusion 32, Fusion 40). Controleer elke instelling altijd op de machine vóór productie.

STAP 1 — Job Manager voorbereiden (op de pc die met de machine verbonden is)
  1. Open Epilog Job Manager.
  2. Controleer of de doelmachine in Job Manager staat.
     Zo niet: voeg haar toe met de gewenste naam (bijv. 'MINI24 40 watt').
  3. Maak een back-up van de database:
     Programma-instellingen (tandwiel) > Database back-uppen (Backup Database).
     Sla het .epidb-bestand op (bureaublad of USB-stick).
  4. Bewaar deze back-up: daarmee kunt u terug als er iets misgaat.

STAP 2 — De oude instellingen voorbereiden
  5. Kopieer de map met de oude .DAT-bestanden van de machine (één map per machine).
     Gebruikelijke locatie: Documenten\Epilog\Engraving_Settings

STAP 3 — Converteren met dit programma
  6. Sectie 1: 'Bladeren…' en kies de .DAT-map.
     Vink 'Submappen meenemen' aan als de instellingen in submappen staan.
  7. Sectie 2: 'Bladeren…' en kies de .epidb-back-up uit stap 1.
  8. Sectie 3: controleer de doelmachine en de mapnaam
     (standaard de naam van de machine).
  9. Sectie 4: stel de opties in (luchtondersteuning, vectorsortering, positie van de map,
     laserbron: kies 'Fiber' voor een fibermachine).
 10. Klik op 'Converteren'.
     Naast de .DAT-map wordt een map '… - Job Manager-conversie' aangemaakt.
     Daarin staan de nieuwe database (.epidb), het overzicht (Excel) en een LEESMIJ.

STAP 4 — Importeren in Job Manager
 11. Job Manager > Programma-instellingen > Database herstellen (Restore Database).
 12. Kies het .epidb-bestand uit de map '… - Job Manager-conversie'.
     Job Manager start opnieuw.
 13. Klik in een taak op het pictogram voor taakinstellingen (rechtsonder).
     De machinemap staat bovenaan de lijst, met de submappen Metaal / Hout / Kunststof.

GOED OM TE WETEN
  • De originele bestanden (.DAT en back-up) worden nooit gewijzigd.
  • Herstellen vervangt de hele Job Manager-database: gebruik een recente back-up (stap 1)
    zodat u geen taken of instellingen verliest.
  • Stempel: schouder en dikte krijgen de standaardwaarden van Job Manager
    (zie de kolom 'Opmerking' in het overzicht).
  • Test enkele materialen op een restant voordat u met de productie begint.
  • Bij problemen: herstel de originele back-up uit stap 1.`,

	// ------------------------------------------------------------------ ZH
	`开始之前 — 必须先备份
转换基于 Job Manager 数据库的备份，该备份中必须已包含目标机器。

适用机器
  可读取的旧 .DAT 设置（Job Manager 之前的旧版 Epilog 驱动程序）：
    • Mini 18 / Mini 24 和 Helix（CO2）
    • Zing 16 / Zing 24（CO2）
    • Fusion 32 / Fusion 40（CO2）
    • Fusion 32 光纤 / Fusion 40 光纤（在“激光源”中选择“光纤”）
  目标机器：同一台机器，已在 Job Manager 中添加
  （Mini 18、Mini 24、Helix、Zing 16、Zing 24、Fusion 32、Fusion 40）。
  不适用：Fusion Pro、Fusion Edge、Fusion Maker（最新一代，没有 .DAT 文件）。
  已逐项核对 3100 多项旧设置（Mini 18、Mini 24、Helix、Zing 16、Zing 24、Fusion 32、Fusion 40）。
  正式生产前，请务必在机器上检查每项设置。

第 1 步 — 准备 Job Manager（在连接机器的电脑上）
  1. 打开 Epilog Job Manager。
  2. 确认目标机器已在 Job Manager 中。
     如果没有：用所需名称添加该机器（例如“MINI24 40 watts”）。
  3. 备份数据库：
     程序设置（齿轮图标）> 备份数据库 (Backup Database)。
     保存 .epidb 文件（桌面或 U 盘）。
  4. 妥善保留此备份：出现问题时可用它恢复。

第 2 步 — 准备旧设置
  5. 复制该机器的旧 .DAT 文件夹（每台机器一个文件夹）。
     常见位置：文档\Epilog\Engraving_Settings

第 3 步 — 使用本软件转换
  6. 第 1 部分：点击“浏览…”，选择 .DAT 文件夹。
     如果设置存放在子文件夹中，请勾选“包含子文件夹”。
  7. 第 2 部分：点击“浏览…”，选择第 1 步制作的 .epidb 备份。
  8. 第 3 部分：检查目标机器和文件夹名称
     （默认为机器名称）。
  9. 第 4 部分：设置选项（空气辅助、矢量排序、文件夹位置，
     激光源：光纤机器请选择“光纤”）。
 10. 点击“转换”。
     程序会在 .DAT 文件夹旁边创建“… - Job Manager 转换”文件夹，
     其中包含新数据库 (.epidb)、汇总表 (Excel) 和说明文件。

第 4 步 — 导入 Job Manager
 11. Job Manager > 程序设置 > 恢复数据库 (Restore Database)。
 12. 选择“… - Job Manager 转换”文件夹中的 .epidb 文件。
     Job Manager 将重新启动。
 13. 在作业中点击作业设置图标（右下角）。
     机器文件夹显示在列表顶部，包含 金属 / 木材 / 塑料 子文件夹。

注意事项
  • 原始文件（.DAT 和备份）始终不会被修改。
  • 恢复会替换整个 Job Manager 数据库：请使用最近的备份（第 1 步），
    以免丢失任何作业或设置。
  • 印章：坡度和厚度使用 Job Manager 的默认值
    （见汇总表的“备注”列）。
  • 正式生产前，请先在边角料上测试几种材料。
  • 如有问题：恢复第 1 步制作的原始备份。`,

	// ------------------------------------------------------------------ JA
	`始める前に — バックアップが必要です
変換は、変換先のマシンが登録済みの Job Manager データベースのバックアップをもとに行います。

対象マシン
  読み込める旧 .DAT 設定（Job Manager 以前の旧 Epilog ドライバー）：
    • Mini 18 / Mini 24 および Helix（CO2）
    • Zing 16 / Zing 24（CO2）
    • Fusion 32 / Fusion 40（CO2）
    • Fusion 32 ファイバー / Fusion 40 ファイバー（「レーザー光源」で「ファイバー」を選択）
  変換先のマシン：同じマシンを Job Manager に登録しておきます
  （Mini 18、Mini 24、Helix、Zing 16、Zing 24、Fusion 32、Fusion 40）。
  対象外：Fusion Pro、Fusion Edge、Fusion Maker（最新世代、.DAT なし）。
  3,100 件以上の旧設定を値ごとに確認済みです（Mini 18、Mini 24、Helix、Zing 16、Zing 24、Fusion 32、Fusion 40）。
  本番加工の前に、必ず各設定を機械で確認してください。

ステップ 1 — Job Manager の準備（マシンに接続された PC で）
  1. Epilog Job Manager を開きます。
  2. 変換先のマシンが Job Manager にあることを確認します。
     ない場合は、希望の名前で追加します（例：「MINI24 40 watts」）。
  3. データベースをバックアップします：
     プログラム設定（歯車アイコン）> データベースのバックアップ (Backup Database)。
     .epidb ファイルを保存します（デスクトップまたは USB メモリー）。
  4. このバックアップは保管しておいてください。問題が起きたときに元に戻せます。

ステップ 2 — 旧設定の準備
  5. マシンの旧 .DAT ファイルのフォルダーをコピーします（マシンごとに 1 フォルダー）。
     通常の場所：ドキュメント\Epilog\Engraving_Settings

ステップ 3 — このソフトで変換
  6. セクション 1：「参照…」で .DAT フォルダーを選択します。
     設定がサブフォルダーにある場合は「サブフォルダーを含める」にチェックします。
  7. セクション 2：「参照…」でステップ 1 のバックアップ (.epidb) を選択します。
  8. セクション 3：変換先のマシンとフォルダー名を確認します
     （既定ではマシン名）。
  9. セクション 4：オプションを設定します（エアアシスト、ベクターの並べ替え、フォルダーの位置、
     レーザー光源：ファイバー機の場合は「ファイバー」を選択）。
 10. 「変換」をクリックします。
     .DAT フォルダーの隣に「… - Job Manager 変換」フォルダーが作成されます。
     新しいデータベース (.epidb)、設定一覧 (Excel)、説明ファイルが入っています。

ステップ 4 — Job Manager へのインポート
 11. Job Manager > プログラム設定 > データベースの復元 (Restore Database)。
 12. 「… - Job Manager 変換」フォルダーの .epidb ファイルを選択します。
     Job Manager が再起動します。
 13. ジョブ内で、ジョブ設定のアイコン（右下）をクリックします。
     マシンのフォルダーが一覧の先頭に表示され、金属 / 木材 / プラスチック のサブフォルダーがあります。

ご注意
  • 元のファイル（.DAT とバックアップ）は変更されません。
  • 復元すると Job Manager のデータベース全体が置き換わります。ジョブや設定を失わないよう、
    最新のバックアップ（ステップ 1）を使用してください。
  • スタンプ：肩と厚さは Job Manager の既定値になります
    （設定一覧の「備考」列を参照）。
  • 本番の前に、端材でいくつかの材料をテストしてください。
  • 問題が起きた場合：ステップ 1 の元のバックアップを復元してください。`,
}

var readmes = [nLang]string{
	// FR
	`CONVERTISSEUR DE RÉGLAGES EPILOG - VAP
=====================================

Contenu de ce dossier
---------------------
- %s
    Sauvegarde Job Manager contenant vos réglages convertis
    (dossier « %s » dans Paramètres du travail).
- %s
    Liste des réglages convertis (s'ouvre dans Excel).

Installation dans Job Manager
-----------------------------
1. Gardez votre sauvegarde d'origine de côté (retour arrière possible).
2. Job Manager > Paramètres du programme > Restaurer la base de données.
3. Choisissez le fichier .epidb de ce dossier. Job Manager redémarre.
4. Dans un travail : icône « Importer les paramètres du travail »
   > dossier « %s ».

À vérifier après import
-----------------------
- Tampon : talus et épaisseur aux valeurs par défaut de Job Manager (voir colonne Remarque).
- Testez quelques matériaux sur une chute avant production.

Machine : %s
Converti le %s
`,
	// EN
	`EPILOG SETTINGS CONVERTER - VAP
===============================

Contents of this folder
-----------------------
- %s
    Job Manager backup containing your converted settings
    ("%s" folder in Job Settings).
- %s
    List of the converted settings (opens in Excel).

Installing in Job Manager
-------------------------
1. Keep your original backup aside (so you can go back).
2. Job Manager > Program settings > Restore Database.
3. Choose the .epidb file from this folder. Job Manager restarts.
4. In a job: job settings icon (bottom right)
   > "%s" folder.

Check after import
------------------
- Stamp: shoulder and thickness at the Job Manager default values (see the Note column).
- Test a few materials on scrap before production.

Machine: %s
Converted on %s
`,
	// ES
	`CONVERTIDOR DE AJUSTES EPILOG - VAP
===================================

Contenido de esta carpeta
-------------------------
- %s
    Copia de seguridad de Job Manager con sus ajustes convertidos
    (carpeta «%s» en Ajustes del trabajo).
- %s
    Lista de los ajustes convertidos (se abre en Excel).

Instalación en Job Manager
--------------------------
1. Guarde aparte su copia de seguridad original (para poder volver atrás).
2. Job Manager > Configuración del programa > Restaurar base de datos (Restore Database).
3. Elija el archivo .epidb de esta carpeta. Job Manager se reinicia.
4. En un trabajo: icono de ajustes del trabajo (abajo a la derecha)
   > carpeta «%s».

Comprobar tras la importación
-----------------------------
- Sello: talud y grosor con los valores por defecto de Job Manager (ver columna Observación).
- Pruebe algunos materiales en un retal antes de la producción.

Máquina: %s
Convertido el %s
`,
	// DE
	`EPILOG-EINSTELLUNGSKONVERTER - VAP
==================================

Inhalt dieses Ordners
---------------------
- %s
    Job Manager-Sicherung mit Ihren konvertierten Einstellungen
    (Ordner „%s“ in den Auftragseinstellungen).
- %s
    Liste der konvertierten Einstellungen (öffnet sich in Excel).

Installation in Job Manager
---------------------------
1. Bewahren Sie Ihre Originalsicherung auf (Rückweg möglich).
2. Job Manager > Programmeinstellungen > Datenbank wiederherstellen (Restore Database).
3. Die .epidb-Datei aus diesem Ordner wählen. Job Manager startet neu.
4. In einem Auftrag: Symbol für Auftragseinstellungen (unten rechts)
   > Ordner „%s“.

Nach dem Import prüfen
----------------------
- Stempel: Schulter und Stärke mit den Standardwerten von Job Manager (siehe Spalte Hinweis).
- Vor der Produktion einige Materialien an einem Reststück testen.

Maschine: %s
Konvertiert am %s
`,
	// IT
	`CONVERTITORE DI IMPOSTAZIONI EPILOG - VAP
=========================================

Contenuto di questa cartella
----------------------------
- %s
    Backup di Job Manager con le impostazioni convertite
    (cartella «%s» in Impostazioni del lavoro).
- %s
    Elenco delle impostazioni convertite (si apre in Excel).

Installazione in Job Manager
----------------------------
1. Conservate il backup originale (per poter tornare indietro).
2. Job Manager > Impostazioni del programma > Ripristina database (Restore Database).
3. Scegliete il file .epidb di questa cartella. Job Manager si riavvia.
4. In un lavoro: icona delle impostazioni del lavoro (in basso a destra)
   > cartella «%s».

Da verificare dopo l'importazione
---------------------------------
- Timbro: spalla e spessore ai valori predefiniti di Job Manager (vedere colonna Nota).
- Provate alcuni materiali su uno scarto prima della produzione.

Macchina: %s
Convertito il %s
`,
	// PT
	`CONVERSOR DE CONFIGURAÇÕES EPILOG - VAP
=======================================

Conteúdo desta pasta
--------------------
- %s
    Backup do Job Manager com as suas configurações convertidas
    (pasta "%s" em Configurações do trabalho).
- %s
    Lista das configurações convertidas (abre no Excel).

Instalação no Job Manager
-------------------------
1. Guarde o seu backup original (para poder voltar atrás).
2. Job Manager > Configurações do programa > Restaurar banco de dados (Restore Database).
3. Escolha o arquivo .epidb desta pasta. O Job Manager reinicia.
4. Num trabalho: ícone das configurações do trabalho (canto inferior direito)
   > pasta "%s".

Verificar após a importação
---------------------------
- Carimbo: talude e espessura nos valores padrão do Job Manager (ver coluna Observação).
- Teste alguns materiais numa sobra antes da produção.

Máquina: %s
Convertido em %s
`,
	// NL
	`EPILOG-INSTELLINGENCONVERTER - VAP
==================================

Inhoud van deze map
-------------------
- %s
    Job Manager-back-up met uw geconverteerde instellingen
    (map '%s' in Taakinstellingen).
- %s
    Lijst van de geconverteerde instellingen (opent in Excel).

Installeren in Job Manager
--------------------------
1. Bewaar uw originele back-up (zodat u terug kunt).
2. Job Manager > Programma-instellingen > Database herstellen (Restore Database).
3. Kies het .epidb-bestand uit deze map. Job Manager start opnieuw.
4. In een taak: pictogram voor taakinstellingen (rechtsonder)
   > map '%s'.

Controleren na import
---------------------
- Stempel: schouder en dikte op de standaardwaarden van Job Manager (zie kolom Opmerking).
- Test enkele materialen op een restant vóór de productie.

Machine: %s
Geconverteerd op %s
`,
	// ZH
	`EPILOG 设置转换器 - VAP
=======================

本文件夹内容
------------
- %s
    包含已转换设置的 Job Manager 备份
    （作业设置中的“%s”文件夹）。
- %s
    已转换设置的列表（可用 Excel 打开）。

在 Job Manager 中安装
---------------------
1. 妥善保留原始备份（以便恢复）。
2. Job Manager > 程序设置 > 恢复数据库 (Restore Database)。
3. 选择本文件夹中的 .epidb 文件。Job Manager 将重新启动。
4. 在作业中：作业设置图标（右下角）
   > “%s”文件夹。

导入后请检查
------------
- 印章：坡度和厚度为 Job Manager 默认值（见“备注”列）。
- 正式生产前，请先在边角料上测试几种材料。

机器：%s
转换时间：%s
`,
	// JA
	`EPILOG 設定変換ツール - VAP
===========================

このフォルダーの内容
--------------------
- %s
    変換した設定を含む Job Manager バックアップ
    （ジョブ設定の「%s」フォルダー）。
- %s
    変換した設定の一覧（Excel で開けます）。

Job Manager へのインストール
----------------------------
1. 元のバックアップは保管しておいてください（元に戻すため）。
2. Job Manager > プログラム設定 > データベースの復元 (Restore Database)。
3. このフォルダーの .epidb ファイルを選択します。Job Manager が再起動します。
4. ジョブ内で：ジョブ設定のアイコン（右下）
   > 「%s」フォルダー。

インポート後の確認
------------------
- スタンプ：肩と厚さは Job Manager の既定値（「備考」列を参照）。
- 本番の前に、端材でいくつかの材料をテストしてください。

マシン：%s
変換日時：%s
`,
}
