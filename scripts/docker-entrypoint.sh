#!/bin/sh
set -e

if [ "${RUN_MIGRATIONS_ON_START:-true}" = "true" ]; then
	if [ -z "${DB_HOST}" ] || [ -z "${DB_NAME}" ] || [ -z "${DB_USER}" ]; then
		echo "WARN: skipping database migrations — set DB_HOST, DB_NAME, and DB_USER" >&2
	elif [ -z "${DB_PASSWORD}" ]; then
		echo "WARN: skipping database migrations — DB_PASSWORD is not set" >&2
	else
		sslmode="${DB_SSLMODE:-prefer}"
		echo "running database migrations (host=${DB_HOST}, sslmode=${sslmode})..."
		migrate -path=/migrations \
			-database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT:-5432}/${DB_NAME}?sslmode=${sslmode}" \
			up
		echo "database migrations complete"
	fi
fi

exec /usr/local/bin/brenox-engine "$@"
