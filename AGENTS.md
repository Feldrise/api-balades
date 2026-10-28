# AGENTS.md

Ce repo contient l'API Go du projet BaladeEcoLogique. Le dossier parent (hors git) contient le vocabulaire métier commun et les décisions qui touchent les deux repos.

## Agent skills

### Issue tracker

Issues GitHub du repo `Feldrise/api-balades` (CLI `gh`). Voir `docs/agents/issue-tracker.md`.

### Triage labels

Les 5 labels par défaut (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). Voir `docs/agents/triage-labels.md`.

### Domain docs

Multi-context : glossaire commun dans `../GLOSSARY.md` et ADR transverses dans `../docs/adr/` (hors git), ADR propres à ce repo dans `docs/adr/`. Voir `docs/agents/domain.md`.

## Commits

Les messages de commit suivent [gitmoji](https://gitmoji.dev) : l'emoji Unicode (pas le code `:sparkles:`) en tête, puis un message en anglais à l'impératif. Exemples : `✨ Add online payment`, `🐛 Fix issue with local date`, `💄 Fix admin ramble details UI`, `👷 Add Docker configuration`.
