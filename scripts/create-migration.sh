#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <migration_name>" >&2
  exit 1
fi

cd "$(dirname "$0")/.."
mkdir -p migrations

maximum=0
shopt -s nullglob
for file in migrations/*.sql; do
  base="${file##*/}"
  if [[ "$base" =~ ^([0-9]{6})_ ]]; then
    number=$((10#${BASH_REMATCH[1]}))
    if ((number > maximum)); then maximum=$number; fi
  fi
done

printf -v version '%06d' "$((maximum + 1))"
name="$1"

touch "migrations/${version}_${name}.up.sql"
touch "migrations/${version}_${name}.down.sql"
echo "Created migrations/${version}_${name}.{up,down}.sql"
