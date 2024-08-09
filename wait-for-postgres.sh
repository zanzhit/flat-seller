#!/bin/sh

# Ожидаем пока Postgres не будет доступен
until pg_isready -h "$1" -p "$2" > /dev/null 2> /dev/null; do
  echo "Waiting for postgres at $1:$2..."
  sleep 1
done

# Выполняем оставшуюся команду
shift 2
exec "$@"
