#!/bin/sh
set -e

if [ "${RUN_MIGRATIONS_ON_START:-true}" = "true" ]; then
	if [ -n "${DB_HOST}" ] && [ -n "${DB_NAME}" ] && [ -n "${DB_USER}" ]; then
		sslmode="${DB_SSLMODE:-prefer}"
		echo "running database migrations (sslmode=${sslmode})..."
		migrate -path=/migrations \
			-database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT:-5432}/${DB_NAME}?sslmode=${sslmode}" \
			up
	fi
fi

exec /usr/local/bin/brenox-engine "$@"
