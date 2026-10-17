# Notes de version

Le VAP Convertisseur Epilog reprend les anciens réglages des pilotes Epilog (fichiers .DAT) et les range automatiquement dans Epilog Job Manager. Vos années de réglages matières passent dans Job Manager en un clic, sans rien ressaisir.

## Prochaine version (en préparation)

Un réglage incomplet ou incompatible ne peut plus passer pour une conversion fidèle.

- **Source CO2 / Fibre contrôlée** : elle est imposée quand la machine n'en a qu'une. Un réglage Fibre n'est jamais envoyé vers une machine sans Fibre : il est ignoré avec un message clair.
- **Fichiers inconnus refusés** : un .DAT dont la structure ne correspond à aucun pilote relevé n'est plus converti avec des valeurs de remplacement.
- **Remarques visibles** : chaque valeur remplacée ou non reprise (résolution, fréquence, position « Page-Centre », passes lues dans le nom…) est indiquée dans le journal et le récapitulatif. Une puissance ou une vitesse de couleur hors limites fait refuser le réglage.
- **Doublons** : deux réglages de même nom mais différents sont gardés tous les deux, le second renommé « nom (2) ».
- **Dossiers Job Manager** : le dossier remplacé est celui de la machine choisie, et les sous-dossiers Métal, Bois, Plastique sont retrouvés par leur nom.
- **Assistance d'air** : réellement imposée sur Marche pour toutes les découpes.
- **Bilan final** : réglages convertis sans remarque, avec remarque, et fichiers ignorés. Le récapitulatif indique la source laser et le détail des couleurs.
- **Rapport d'erreur** : votre dossier personnel Windows y est masqué, et il s'ouvre pour être relu avant l'envoi.
- **Sécurité** : seule la bibliothèque SQLite de Windows est chargée.
- **Tests** : à chaque modification, des conversions complètes sur des fichiers de test inclus dans le dépôt.

## Version 1.7 — octobre 2026

Un programme plus léger, et vérifié automatiquement à chaque modification.

- **Deux fois plus léger** : le programme passe de 7,9 Mo à 4,2 Mo. La recherche de nouvelle version utilise maintenant le module Internet intégré à Windows. Rien ne change à l'utilisation.
- **Tests automatiques** : à chaque modification du convertisseur, GitHub le compile et lance les tests sur un Windows, avant toute publication.

## Version 1.6 — octobre 2026

Des contrôles avant la conversion, et plus de 3 100 réglages vérifiés.

- **Sauvegarde contrôlée dès son choix** : si la sauvegarde Job Manager n'a pas la structure attendue, un message clair le dit tout de suite et conseille d'en refaire une, au lieu d'une erreur technique en pleine conversion.
- **Mauvais dossier repéré** : si vous choisissez par erreur un dossier « … - Conversion Job Manager », le convertisseur propose le dossier des .DAT d'origine.
- **Nouvelle version signalée** : au démarrage, un lien discret apparaît quand une version plus récente est disponible sur GitHub.
- **Plus de 3 100 réglages contrôlés** : chaque valeur convertie a été comparée à un décodage indépendant des .DAT, sur toutes les machines prises en charge.
- **Rappel de prudence** : le message de fin de conversion rappelle de vérifier chaque réglage avant toute production.
- **Noms des modèles** : Mini 18 / Mini 24, Zing 16 / Zing 24, Fusion 32 / Fusion 40.

## Version 1.5 — octobre 2026

En cas de problème, un rapport d'erreur permet de trouver la cause à distance.

- **Rapport d'erreur** : si une conversion échoue, un fichier « rapport d'erreur.txt » est créé à côté du dossier des .DAT. Il décrit le PC, la sauvegarde Job Manager et les .DAT lus, sans aucune donnée personnelle.
- **Envoi simplifié** : le convertisseur propose de l'envoyer. Le fichier est sélectionné dans l'explorateur et l'adresse e-mail est copiée, il suffit de la coller dans Gmail, Outlook ou toute autre messagerie.
- **Suggestion / bug** : le bouton fonctionne aussi sans logiciel de messagerie installé.

## Version 1.4 — octobre 2026

La Zing et la Fusion rejoignent la Mini et l'Helix.

- **Zing et Fusion** : les réglages de ces deux machines sont maintenant convertis.
- **Fusion** : l'autofocus avec l'épaisseur et le décalage, la source Fibre et sa fréquence, les compensations de vitesse et de puissance, et la grille de découpe sont repris.
- **Position du travail** : le point de centrage choisi dans le pilote est conservé.
- **Taille de page** : celle de la machine, ou à défaut celle du réglage d'origine.
- **Drapeaux** : un drapeau devant chaque langue du menu.

## Version 1.3 — octobre 2026

Plus de réglages repris tels quels, sans retouche.

- **Passes** : un réglage nommé « 2 passes » devient bien un réglage à 2 passes.
- **Gestion des couleurs** : chaque couleur réglée dans le pilote devient un procédé Job Manager, avec sa vitesse, sa puissance et sa fréquence.

## Version 1.2 — octobre 2026

Les réglages convertis sont beaucoup plus fidèles à ceux du pilote d'origine.

- **Correction importante** : les réglages « gravure puis découpe » arrivaient en mode Tampon dans Job Manager. Ils sont maintenant en mode Standard.
- **Trame** : Floyd Steinberg, Jarvis, Stucki et Brighten sont reprises.
- **Sens de gravure** (haut vers bas ou bas vers haut) et **autofocus** sont repris.
- **Tampon et 3D** : le type de gravure est conservé.
- **Tri des vecteurs** : nouveau choix « Comme dans le .DAT ».

## Version 1.1 — octobre 2026

Le convertisseur parle 9 langues.

- **Langues** : français, anglais, espagnol, allemand, italien, portugais, néerlandais, chinois et japonais.
- Au premier lancement, la langue de Windows est choisie automatiquement.
- La notice et le récapitulatif des réglages sont traduits, avec les séparateurs habituels de chaque pays pour Excel.

## Version 1.0 — octobre 2026

Première version : les anciens réglages arrivent dans Job Manager, déjà rangés.

- **Conversion** des fichiers .DAT des anciens pilotes Epilog : puissance, vitesse, résolution, fréquence et type de travail.
- **Rangement** dans un dossier « Paramètres du travail » au nom de la machine, avec les sous-dossiers Métal, Bois et Plastique, placé en tête de liste.
- **Assistance d'air** sur toutes les découpes et tri des vecteurs de l'intérieur vers l'extérieur.
- **Épaisseur** lue dans le nom du réglage (par exemple « MDF 3 mm »).
- Un **récapitulatif** des réglages convertis et une notice pas à pas sont fournis.
- Vos fichiers d'origine ne sont jamais modifiés.
