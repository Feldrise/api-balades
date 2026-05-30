#!/bin/sh

# Database restore script for the Balade Ecologique API.
# Usage: ./restore.sh <backup_name>

set -eu

if [ "$#" -eq 0 ]; then
  echo "Error: backup name required"
  echo "Usage: ./restore.sh <backup_name>"
  echo ""
  echo "Available backups:"
  ls -la /backups/balade_backup_*.dump 2>/dev/null | awk '{print " - "$9}' | sed 's|.*/||' | sed 's|\.dump||' || echo "No backup found"
  exit 1
fi

BACKUP_NAME="$1"
BACKUP_DIR="/backups"
DUMP_FILE="$BACKUP_DIR/${BACKUP_NAME}.dump"
SQL_FILE="$BACKUP_DIR/${BACKUP_NAME}.sql.gz"

echo "Starting database restore..."
echo "Date: $(date)"
echo "Backup: ${BACKUP_NAME}"

if [ ! -f "$DUMP_FILE" ] && [ ! -f "$SQL_FILE" ]; then
  echo "Error: backup file not found"
  echo "Looked for: $DUMP_FILE or $SQL_FILE"
  exit 1
fi

echo "WARNING: this will overwrite the current database."
printf "Do you want to continue? Type 'yes': "
read -r confirmation

if [ "$confirmation" != "yes" ]; then
  echo "Restore cancelled"
  exit 0
fi

echo "Creating a safety backup before restoring..."
/scripts/backup.sh "pre_restore_$(date +%Y%m%d_%H%M%S)"

if [ -f "$DUMP_FILE" ]; then
  echo "Restoring from dump file..."
  pg_restore \
    --host="$PGHOST" \
    --username="$PGUSER" \
    --dbname="$PGDATABASE" \
    --clean \
    --if-exists \
    --no-owner \
    --no-privileges \
    --verbose \
    "$DUMP_FILE"
elif [ -f "$SQL_FILE" ]; then
  echo "Restoring from SQL file..."
  gunzip -c "$SQL_FILE" | psql \
    --host="$PGHOST" \
    --username="$PGUSER" \
    --dbname="$PGDATABASE"
fi

echo "Verifying database connection..."
if psql --host="$PGHOST" --username="$PGUSER" --dbname="$PGDATABASE" -c "SELECT 1;" >/dev/null 2>&1; then
  echo "Database reachable."
else
  echo "Database connection problem."
  exit 1
fi

echo "Restore completed successfully."
