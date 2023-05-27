#!/bin/bash


HEALTH_ENDPOINTS=(
  "http://localhost:3501/v1.0/healthz" # Timeline
  "http://localhost:3502/v1.0/healthz" # User
  "http://localhost:3504/v1.0/healthz" # Login
  "http://localhost:3505/v1.0/healthz" # Krakend
  "http://localhost:3506/v1.0/healthz" # Status
)

MAX_RETRIES=10
RETRY_INTERVAL=5

check_health() {
  local endpoint=$1
  local response
  response=$(curl -s -o /dev/null -w "%{http_code}" "$endpoint")
  if [ $response -eq 204 ]; then
    echo "Endpoint $endpoint is healthy"
    return 0
  else
    echo "Endpoint $endpoint is not healthy, received status code $response"
    return 1
  fi
}

check_all_health() {
  retries=0
  while [ $retries -lt $MAX_RETRIES ]; do
    all_healthy=true

    for endpoint in "${HEALTH_ENDPOINTS[@]}"; do
      if ! check_health $endpoint; then
        all_healthy=false
      fi
    done

    if $all_healthy; then
      echo "All health endpoints are healthy"
      break
    else
      echo "Retrying in $RETRY_INTERVAL seconds..."
      sleep $RETRY_INTERVAL
      retries=$((retries + 1))
    fi
  done

  if [ $retries -eq $MAX_RETRIES ]; then
    echo "Timed out waiting for all health endpoints to be healthy"
    kill $DAPR_PID
    exit 1
  fi
}

# Start all Services
dapr run -f test-e2e/dapr_multirun.yaml &
DAPR_PID=$!

# Wait for them to be ready
check_all_health

# Run E2E
dapr run --app-id e2e-service --app-port 9999 --dapr-http-port 9998 --resources-path ./status/config/test-components -- go test test-e2e/e2e_test.go
EXIT_CODE=$?

# Clean up and exit
kill $DAPR_PID
exit $EXIT_CODE
