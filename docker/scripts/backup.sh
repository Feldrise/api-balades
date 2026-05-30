#!/bin/sh

# Database backup script for the Balade Ecologique API.
# Usage: ./backup.sh [optional_name]

set -eu

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME=${1:-"balade_backup_${DATE}"}

echo "Starting database backup..."
echo "Date: $(date)"
echo "Backup name: ${BACKUP_NAME}"

mkdir -p "$BACKUP_DIR"

echo "Creating binary backup..."
pg_dump \
  --host="$PGHOST" \
  --username="$PGUSER" \
  --dbname="$PGDATABASE" \
  --format=custom \
  --no-owner \
  --no-privileges \
  --verbose \
  --file="$BACKUP_DIR/${BACKUP_NAME}.dump"

echo "Creating plain SQL backup..."
pg_dump \
  --host="$PGHOST" \
  --username="$PGUSER" \
  --dbname="$PGDATABASE" \
  --no-owner \
  --no-privileges \
  --clean \
  --if-exists \
  --verbose \
  --file="$BACKUP_DIR/${BACKUP_NAME}.sql"

echo "Compressing SQL backup..."
gzip "$BACKUP_DIR/${BACKUP_NAME}.sql"

if [ -f "$BACKUP_DIR/${BACKUP_NAME}.dump" ] && [ -f "$BACKUP_DIR/${BACKUP_NAME}.sql.gz" ]; then
  echo "Backup created successfully:"
  ls -lh "$BACKUP_DIR/${BACKUP_NAME}".* | awk '{print " - "$9" ("$5")"}'
else
  echo "Backup creation failed"
  exit 1
fi

echo "Cleaning up old backups, keeping the 7 most recent..."
cd "$BACKUP_DIR"
ls -t balade_backup_*.dump 2>/dev/null | tail -n +8 | xargs -r rm -f
ls -t balade_backup_*.sql.gz 2>/dev/null | tail -n +8 | xargs -r rm -f

echo "Backup completed successfully."
