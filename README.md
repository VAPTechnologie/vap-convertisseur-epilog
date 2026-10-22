<p align="center"><img src="docs/logo.png" alt="VAP" width="320"></p>

[![Tests](https://github.com/VAPTechnologie/vap-convertisseur-epilog/actions/workflows/tests.yml/badge.svg)](https://github.com/VAPTechnologie/vap-convertisseur-epilog/actions/workflows/tests.yml)

# VAP Convertisseur Epilog

Vos anciens réglages Epilog (fichiers **.DAT** des anciens pilotes) passent dans **Epilog Job Manager** en un clic, déjà rangés par machine, sans rien ressaisir.

**[⬇ Télécharger la dernière version (V2.0)](../../releases/latest)**

Gratuit · Windows 10 et 11 · 9 langues · aucune installation · interface sombre VAP

![Le convertisseur](docs/capture.png)

## Machines prises en charge

- Mini 18, Mini 24 et Helix
- Zing 16 et Zing 24
- Fusion 32 et Fusion 40, CO2 et Fibre

## Comment l'utiliser

1. Dans Job Manager, faites une sauvegarde de la base de données (« Backup Database »).
2. Dans le convertisseur, choisissez le dossier des anciens .DAT, puis cette sauvegarde.
3. Cliquez sur « Convertir », puis restaurez la nouvelle base dans Job Manager (« Restore Database »).

La notice complète est dans le programme (bouton « Notice / procédure »). Vos fichiers d'origine ne sont jamais modifiés.

> [!WARNING]
> **Avant d'utiliser un réglage converti en production**
> - Gardez intacte la sauvegarde de Job Manager faite à l'étape 1 : c'est votre retour arrière.
> - Contrôlez chaque réglage sur la machine et le matériau concernés, avant toute production : puissance, vitesse, fréquence, source CO2 ou Fibre, nombre de passes.
> - L'assistance d'air est imposée sur Marche pour toutes les découpes, par sécurité.
> - Un réglage laser incorrect peut abîmer la pièce ou provoquer un départ de feu : ne laissez jamais la machine sans surveillance.

## Machines vérifiées

Chaque valeur convertie a été contrôlée dans Job Manager, réglage par réglage, avec les pilotes Epilog d'origine.

| Machine | Réglages contrôlés | État |
|---|---|---|
| Mini 18 / Mini 24 | 98 réglages + réglages de test du pilote | ✅ vérifiée |
| Helix | 80 réglages | ✅ vérifiée |
| Zing 16 / Zing 24 | 82 réglages + réglages de test du pilote | ✅ vérifiée |
| Fusion 32 / Fusion 40 (CO2 et Fibre) | 67 réglages + réglages de test du pilote | ✅ vérifiée |

En plus, 3 121 anciens réglages (66 dossiers, tous ces modèles) ont été convertis puis comparés valeur par valeur à un décodage indépendant des .DAT : tous sont conformes. Les deux modèles de chaque gamme utilisent le même pilote et le même format de .DAT.

Cette comparaison sur les vrais fichiers est relancée à la main avant chaque version : ces fichiers ne sont pas publiés, et elle ne détaille pas les paramètres des couleurs (seule l'assistance d'air y est contrôlée). À chaque modification, GitHub lance en plus des conversions complètes sur des fichiers de test inclus dans le dépôt : chaque pilote, gravure, découpe et combiné, CO2 et Fibre, gestion des couleurs, fichiers refusés, doublons, remplacement d'un dossier, originaux inchangés.

Vous avez un autre modèle, ou un réglage qui ne correspond pas ? Vos fichiers .DAT nous aideraient : écrivez à **vapfrenchtec@gmail.com**.

## Questions fréquentes

**Windows affiche « Windows a protégé votre ordinateur ».**
Le programme est gratuit et n'a pas de signature payante, donc Windows ne le connaît pas. Cliquez sur « Informations complémentaires », puis sur « Exécuter quand même ».

**Mon antivirus se méfie du fichier.**
Comparez l'empreinte SHA-256 du fichier téléchargé avec celle indiquée sur la page de la version. Dans PowerShell :
`Get-FileHash -Algorithm SHA256 "VAP.Convertisseur.Epilog.exe"`

**Comment revenir en arrière ?**
Dans Job Manager, restaurez la sauvegarde faite à l'étape 1 (« Restore Database »). Le convertisseur ne modifie jamais cette sauvegarde : il crée une nouvelle base à côté.

**Un réglage converti ne correspond pas à l'ancien.**
Signalez-le (voir plus bas) en joignant le fichier .DAT : nous corrigerons le convertisseur.

## Ce qui est refusé ou signalé

- Un .DAT dont la structure ne correspond à aucun pilote relevé est refusé : il n'est jamais converti avec des valeurs inventées.
- La source laser (CO2 ou Fibre) est imposée quand la machine n'en a qu'une. Un réglage Fibre n'est jamais envoyé vers une machine sans Fibre.
- Une puissance ou une vitesse hors limites, y compris dans la gestion des couleurs, fait refuser le réglage.
- Une valeur remplacée (résolution invalide, fréquence hors plage, position « Page-Centre » sans équivalent, passes lues dans le nom…) est indiquée dans le journal et dans la colonne « Remarque » du récapitulatif.
- Deux réglages de même nom mais différents sont tous les deux gardés : le second est renommé « nom (2) ».

## Ce qui est repris

Puissance, vitesse, résolution, fréquence, type de travail, trame, sens de gravure, autofocus, tri des vecteurs, mode Tampon et 3D, position du travail, gestion des couleurs, nombre de passes, et pour la Fusion : source Fibre, épaisseur et décalage, compensations et grille de découpe. L'assistance d'air est mise sur Marche pour toutes les découpes.

Le détail des versions est dans [NOTES_DE_VERSION.md](NOTES_DE_VERSION.md).

## Un problème ou une idée ?

Cliquez sur « Suggestion / bug » dans le convertisseur, ou ouvrez l'onglet **[Issues](../../issues/new/choose)**. Joignez si possible le fichier .DAT concerné.

Si la conversion a échoué, joignez aussi le fichier « rapport d'erreur.txt » créé à côté du dossier des .DAT : il permet de trouver la cause sans votre sauvegarde complète. Il contient les noms de vos fichiers .DAT, de vos machines et de vos dossiers Job Manager ; votre dossier personnel Windows y est masqué. Relisez-le avant de l'envoyer.

Pas de compte GitHub ? Écrivez à **vapfrenchtec@gmail.com**.
## Soutenir le projet

Le convertisseur est gratuit. Pour soutenir son développement : bouton « Faire un don » dans le programme, ou [paypal.me/VAPTechnologie](https://www.paypal.com/paypalme/VAPTechnologie).

## Pour les développeurs

Programme Windows écrit en Go, sans dépendance externe (interface Win32, base SQLite par le `winsqlite3.dll` du dossier système de Windows, jamais par une DLL placée à côté du programme). Sources dans [`sources/`](sources/) :

```
go-winres make --in res/winres.json --arch amd64 --out rsrc
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-H windowsgui -s -w" -o "VAP Convertisseur Epilog.exe" .
```

Licence MIT — © 2026 VAP Technologie (voir [LICENSE](LICENSE)).

Epilog, Job Manager, Mini, Helix, Zing et Fusion sont des marques d'Epilog Laser. Ce projet n'est pas affilié à Epilog Laser.
