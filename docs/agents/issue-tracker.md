# Issue tracker : GitHub

Les issues et les specs de ce repo sont des issues GitHub sur `Feldrise/api-balades`. Toutes les opérations passent par la CLI `gh`.

## Conventions

- **Créer une issue** : `gh issue create --title "..." --body "..."`. Utiliser un heredoc pour les corps sur plusieurs lignes.
- **Lire une issue** : `gh issue view <numéro> --json number,title,body,labels,comments`.
- **Lister les issues** : `gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'`, avec les filtres `--label` et `--state` adaptés.
- **Rattacher une sous-issue à une issue parente** : `gh issue create --parent <parent> ...`, ou `gh issue edit <parent> --add-sub-issue <enfant>` après coup (`gh` 2.94+). Avec un `gh` plus ancien : `gh api --method POST repos/<owner>/<repo>/issues/<parent>/sub_issues -F sub_issue_id=<id-db-enfant>` (id de base de données, comme dans **Blocage** plus bas). Sans sous-issues, mettre `Part of #<parent>` en tête du corps de l'enfant.
- **Commenter** : `gh issue comment <numéro> --body "..."`
- **Ajouter / retirer des labels** : `gh issue edit <numéro> --add-label "..."` / `--remove-label "..."`
- **Fermer** : `gh issue close <numéro> --comment "..."`

`gh` déduit le repo depuis `git remote -v` quand il est lancé dans ce dossier.

## Fonctionnalités qui touchent l'API et l'app

Le projet est découpé en deux repos : `Feldrise/api-balades` (API Go) et `Feldrise/application-balades` (app Flutter). Quand un travail touche les deux, créer une issue dans chaque repo (`gh issue create --repo Feldrise/<repo> ...`) et les relier par une référence complète (`Feldrise/api-balades#12`) en tête de chaque corps.

## Pull requests comme surface de triage

**PRs as a request surface: no.** _(Passer à `yes` si ce repo traite les PR externes comme des demandes de fonctionnalité ; `/triage` lit ce flag.)_

Avec `yes`, les PR suivent les mêmes labels et états que les issues, via les équivalents `gh pr` :

- **Lire une PR** : `gh pr view <numéro> --comments`, et `gh pr diff <numéro>` pour le diff.
- **Lister les PR externes à trier** : `gh api --paginate 'repos/{owner}/{repo}/pulls?state=open' --jq '.[] | select(.author_association | IN("OWNER","MEMBER","COLLABORATOR") | not) | {number, title, author: .user.login, author_association, labels: [.labels[].name]}'`.
- **Commenter / labelliser / fermer** : `gh pr comment`, `gh pr edit --add-label`/`--remove-label`, `gh pr close`.

GitHub partage la même numérotation entre issues et PR, donc un `#42` seul peut désigner l'une ou l'autre : essayer `gh pr view 42`, puis `gh issue view 42`.

## Quand un skill dit « publier dans l'issue tracker »

Créer une issue GitHub.

## Quand un skill dit « récupérer le ticket concerné »

Le lire comme dans **Lire une issue** ci-dessus.

## Opérations Wayfinding

Utilisées par `/wayfinder`. La **carte** est une issue unique dont les **enfants** sont les tickets.

- **Carte** : une issue unique avec le label `wayfinder:map`, dont le corps contient Notes / Decisions-so-far / Fog. `gh issue create --label wayfinder:map`.
- **Ticket enfant** : une issue rattachée à la carte comme sous-issue GitHub (voir **Rattacher une sous-issue**). Si les sous-issues ne sont pas disponibles, ajouter l'enfant à une task list dans le corps de la carte et mettre `Part of #<carte>` en tête du corps de l'enfant. Labels : `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`). Une fois pris, le ticket est assigné au dev qui le pilote.
- **Blocage** : les **dépendances natives** des issues GitHub, la représentation canonique et visible dans l'UI. Ajouter un lien avec `gh api --method POST repos/<owner>/<repo>/issues/<enfant>/dependencies/blocked_by -F issue_id=<id-db-bloquant>`, où `<id-db-bloquant>` est l'**id de base de données** numérique du bloquant (`gh api repos/<owner>/<repo>/issues/<n> --jq .id`, _pas_ le `#numéro` ni le `node_id`). GitHub expose `issue_dependencies_summary.blocked_by` (bloquants ouverts uniquement, c'est le vrai verrou). Si les dépendances ne sont pas disponibles, mettre une ligne `Blocked by: #<n>, #<n>` en tête du corps de l'enfant. Un ticket est débloqué quand tous ses bloquants sont fermés.
- **Requête de frontière** : lister les enfants ouverts de la carte (`gh issue list --state open`, restreint aux sous-issues / à la task list de la carte), écarter ceux qui ont un bloquant ouvert (`issue_dependencies_summary.blocked_by > 0`, ou une issue ouverte dans la ligne `Blocked by`) ou un assigné ; le premier dans l'ordre de la carte l'emporte.
- **Prise** : `gh issue edit <n> --add-assignee @me`, première écriture de la session.
- **Résolution** : `gh issue comment <n> --body "<réponse>"`, puis `gh issue close <n>`, puis ajouter un pointeur de contexte (résumé + lien) dans Decisions-so-far de la carte.
