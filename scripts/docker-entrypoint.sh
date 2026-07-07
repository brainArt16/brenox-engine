#!/bin/sh
set -e

# golang-migrate does not support pgx's "prefer" sslmode.
migrate_sslmode() {
	mode="${DB_SSLMODE:-disable}"
	case "$mode" in
	prefer) echo disable ;;
	*) echo "$mode" ;;
	esac
}

if [ "${RUN_MIGRATIONS_ON_START:-true}" = "true" ]; then
	if [ -z "${DB_HOST}" ] || [ -z "${DB_NAME}" ] || [ -z "${DB_USER}" ]; then
		echo "WARN: skipping database migrations — set DB_HOST, DB_NAME, and DB_USER" >&2
	elif [ -z "${DB_PASSWORD}" ]; then
		echo "WARN: skipping database migrations — DB_PASSWORD is not set" >&2
	else
		sslmode="$(migrate_sslmode)"
		echo "running database migrations (host=${DB_HOST}, sslmode=${sslmode})..."
		migrate -path=/migrations \
			-database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT:-5432}/${DB_NAME}?sslmode=${sslmode}" \
			up
		echo "database migrations complete"
	fi
fi

exec /usr/local/bin/brenox-engine "$@"
