# Notes de version

Le VAP Convertisseur Epilog reprend les anciens réglages des pilotes Epilog (fichiers .DAT) et les range automatiquement dans Epilog Job Manager. Vos années de réglages matières passent dans Job Manager en un clic, sans rien ressaisir.

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
