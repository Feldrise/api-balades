#!/usr/bin/env bash

# Restore PostgreSQL backup in local dev (docker compose)
# Usage:
#   ./docker/scripts/restore-local.sh docker/backups/backup_20260430_144806.sql
#   ./docker/scripts/restore-local.sh docker/backups/backup_20260430_144806.sql.gz

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$REPO_ROOT/docker-compose.yml"
DB_SERVICE="db"
DB_CONTAINER_PATTERN="balade-db-dev"
DB_USER="balade"
DB_NAME="balade"

if [ $# -lt 1 ]; then
  echo "❌ Erreur: fichier de sauvegarde requis."
  echo "Usage: $0 <fichier_backup.sql|fichier_backup.sql.gz>"
  exit 1
fi

BACKUP_FILE="$1"
if [[ "$BACKUP_FILE" != /* ]]; then
  BACKUP_FILE="$REPO_ROOT/$BACKUP_FILE"
fi

if [ ! -f "$BACKUP_FILE" ]; then
  echo "❌ Erreur: fichier introuvable: $BACKUP_FILE"
  exit 1
fi

if ! docker compose -f "$COMPOSE_FILE" ps --status running --format '{{.Name}}' | grep -q "^${DB_CONTAINER_PATTERN}$"; then
  echo "❌ Erreur: le conteneur DB local n'est pas démarré."
  echo "Lance d'abord: docker compose -f docker-compose.yml up -d db"
  exit 1
fi

echo "⚠️  ATTENTION: cette opération écrase les données actuelles de la base locale '$DB_NAME'."
echo "📦 Fichier à restaurer: $BACKUP_FILE"
echo "Tape 'RESTORE' pour confirmer:"
read -r confirmation

if [ "$confirmation" != "RESTORE" ]; then
  echo "🚫 Restauration annulée."
  exit 0
fi

BACKUP_DIR="$REPO_ROOT/docker/backups"
mkdir -p "$BACKUP_DIR"
SAFETY_BACKUP="$BACKUP_DIR/pre_restore_$(date +%Y%m%d_%H%M%S).sql"

echo "💾 Sauvegarde préventive en cours: $SAFETY_BACKUP"
docker compose -f "$COMPOSE_FILE" exec -T "$DB_SERVICE" \
  pg_dump -U "$DB_USER" "$DB_NAME" > "$SAFETY_BACKUP"

echo "🧹 Nettoyage du schéma public..."
docker compose -f "$COMPOSE_FILE" exec -T "$DB_SERVICE" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<SQL
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO ${DB_USER};
GRANT ALL ON SCHEMA public TO public;
SQL

echo "🔄 Restauration en cours..."
if [[ "$BACKUP_FILE" == *.sql.gz ]]; then
  gunzip -c "$BACKUP_FILE" | docker compose -f "$COMPOSE_FILE" exec -T "$DB_SERVICE" \
    psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1
elif [[ "$BACKUP_FILE" == *.sql ]]; then
  docker compose -f "$COMPOSE_FILE" exec -T "$DB_SERVICE" \
    psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 < "$BACKUP_FILE"
else
  echo "❌ Erreur: format non supporté. Utilise .sql ou .sql.gz"
  exit 1
fi

echo "✅ Restauration terminée."
echo "🔍 Vérification rapide..."
docker compose -f "$COMPOSE_FILE" exec -T "$DB_SERVICE" \
  psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT NOW() AS restored_at;"

echo "🎉 Base locale restaurée avec succès."
