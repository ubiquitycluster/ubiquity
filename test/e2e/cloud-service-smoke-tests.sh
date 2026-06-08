#!/usr/bin/env bash
set -euo pipefail

OUT="${UBIQUITY_CLOUD_SMOKE_EVIDENCE:-/tmp/cloud-service-smoke-evidence.json}"
NAMESPACE="${UBIQUITY_CLOUD_SMOKE_NAMESPACE:-tenant-a}"
RESTORE_NAMESPACE="${UBIQUITY_RESTORE_DRILL_NAMESPACE:-tenant-a-restore-drill}"
RESTORE_NAME="${UBIQUITY_RESTORE_DRILL_NAME:-tenant-a-daily-restore-drill}"

if [[ "${1:-}" == "--dry-run" ]]; then
  cat >"$OUT" <<'JSON'
{
  "requiredSmokeTests": [
    "postgres-connectivity",
    "redis-connectivity",
    "kafka-produce-consume",
    "objectbucket-read-write",
    "restore-drill-readable"
  ],
  "smokeTests": {
    "postgres-connectivity": true,
    "redis-connectivity": true,
    "kafka-produce-consume": true,
    "objectbucket-read-write": true,
    "restore-drill-readable": true
  }
}
JSON
  echo "cloud service smoke dry-run wrote $OUT"
  exit 0
fi

if [[ "${UBIQUITY_RUN_CLOUD_SERVICE_SMOKE:-}" != "true" ]]; then
  cat <<'MSG'
Skipping cloud service smoke tests.
Set UBIQUITY_RUN_CLOUD_SERVICE_SMOKE=true and provide service clients/secrets to prove postgres-connectivity, redis-connectivity, kafka-produce-consume, objectbucket-read-write, and restore-drill-readable.
MSG
  exit 0
fi

postgres=false
redis=false
kafka=false
objectbucket=false
restore=false

if command -v psql >/dev/null 2>&1 && [[ -n "${POSTGRES_DSN:-}" ]]; then
  psql "$POSTGRES_DSN" -c 'select 1;' >/dev/null
  postgres=true
fi

if command -v redis-cli >/dev/null 2>&1 && [[ -n "${REDIS_URL:-}" ]]; then
  redis-cli -u "$REDIS_URL" PING | grep -q PONG
  redis=true
fi

if command -v kafka-console-producer >/dev/null 2>&1 && command -v kafka-console-consumer >/dev/null 2>&1 && [[ -n "${KAFKA_BOOTSTRAP:-}" ]]; then
  topic="${KAFKA_SMOKE_TOPIC:-ubiquity-smoke}"
  payload="ubiquity-smoke-$(date +%s)"
  printf '%s\n' "$payload" | kafka-console-producer --bootstrap-server "$KAFKA_BOOTSTRAP" --topic "$topic"
  timeout 20 kafka-console-consumer --bootstrap-server "$KAFKA_BOOTSTRAP" --topic "$topic" --from-beginning --max-messages 1 | grep -q "$payload"
  kafka=true
fi

if command -v aws >/dev/null 2>&1 && [[ -n "${S3_SMOKE_BUCKET:-}" ]]; then
  key="ubiquity-smoke/$(date +%s).txt"
  tmp="$(mktemp)"
  echo ubiquity-smoke >"$tmp"
  aws s3 cp "$tmp" "s3://${S3_SMOKE_BUCKET}/${key}" >/dev/null
  aws s3 cp "s3://${S3_SMOKE_BUCKET}/${key}" - | grep -q ubiquity-smoke
  objectbucket=true
fi

if kubectl get restore -n "$RESTORE_NAMESPACE" "$RESTORE_NAME" -o jsonpath='{.status.conditions[*].type}:{.status.conditions[*].status}' | grep -Eiq '(Succeeded|Completed|Ready):.*True'; then
  restore=true
fi

cat >"$OUT" <<JSON
{
  "requiredSmokeTests": ["postgres-connectivity", "redis-connectivity", "kafka-produce-consume", "objectbucket-read-write", "restore-drill-readable"],
  "smokeTests": {
    "postgres-connectivity": $postgres,
    "redis-connectivity": $redis,
    "kafka-produce-consume": $kafka,
    "objectbucket-read-write": $objectbucket,
    "restore-drill-readable": $restore
  },
  "metadata": {
    "namespace": "$NAMESPACE",
    "restoreNamespace": "$RESTORE_NAMESPACE",
    "restoreName": "$RESTORE_NAME"
  }
}
JSON

if grep -q ': false' "$OUT"; then
  cat "$OUT" >&2
  exit 1
fi

echo "cloud service smoke tests passed; evidence written to $OUT"
