#!/usr/bin/env sh
set -e

host="$1"
shift
port="$1"
shift
cmd="$@"

if [ -z "$host" ] || [ -z "$port" ]; then
  echo "usage: $0 host port [cmd...]"
  exit 1
fi

echo "Waiting for $host:$port..."
until nc -z "$host" "$port"; do
  sleep 1
done

echo "$host:$port is available"
exec $cmd
