# Docs du domaine

Comment les skills d'ingénierie doivent utiliser la documentation du domaine en explorant le code.

## Avant d'explorer, lire

- **`../GLOSSARY.md`** : le glossaire métier commun à l'API et à l'app, dans le dossier parent BaladeEcoLogique (hors git).
- **`../docs/adr/`** : les décisions transverses qui touchent les deux repos (contrat d'API, authentification, rôles…).
- **`docs/adr/`** : les décisions propres à ce repo. Lire celles qui touchent la zone sur laquelle on va travailler.

Si l'un de ces fichiers n'existe pas, **continuer sans rien dire**. Ne pas signaler leur absence, ne pas proposer de les créer d'avance. Le skill `/domain-modeling` (atteint via `/grill-with-docs` et `/improve-codebase-architecture`) les crée au fil de l'eau, quand un terme ou une décision est réellement tranché.

Si ce repo est cloné seul (sans le dossier parent), les fichiers `../` sont absents : continuer de la même façon.

## Où écrire

- Un terme métier (Balade, Atelier, rôle…) va dans `../GLOSSARY.md`, partagé par les deux repos.
- Une décision qui impose quelque chose aux deux repos va dans `../docs/adr/`.
- Une décision purement interne à ce repo va dans `docs/adr/`.

## Structure

```
BaladeEcoLogique/              (hors git)
├── GLOSSARY.md
├── docs/adr/                  ← décisions transverses
├── api/                       (git : Feldrise/api-balades)
│   └── docs/adr/              ← décisions propres à l'API
└── balade/                    (git : Feldrise/application-balades)
    └── docs/adr/              ← décisions propres à l'app
```

## Utiliser le vocabulaire du glossaire

Quand ta production nomme un concept du domaine (titre d'issue, proposition de refactor, hypothèse, nom de test), utilise le terme tel que défini dans `GLOSSARY.md`. Ne dérive pas vers des synonymes que le glossaire écarte explicitement.

Si le concept dont tu as besoin n'est pas encore dans le glossaire, c'est un signal : soit tu inventes un vocabulaire que le projet n'utilise pas (à reconsidérer), soit il y a un vrai manque (à noter pour `/domain-modeling`).

## Signaler les conflits avec un ADR

Si ta production contredit un ADR existant, dis-le explicitement plutôt que de passer outre en silence :

> _Contredit l'ADR-0007 (…), mais mérite d'être rouvert parce que…_
