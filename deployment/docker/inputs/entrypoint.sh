#!/bin/sh

set -e

# Start AESM service in background
LD_LIBRARY_PATH=/opt/intel/sgx-aesm-service/aesm /opt/intel/sgx-aesm-service/aesm/aesm_service --no-daemon --no-syslog &

AESM_PID=$!

# Forward all arguments to gramine-sgx
gramine-sgx "$APP"
STATUS=$?

# Cleanup
kill "$AESM_PID"
wait "$AESM_PID" 2>/dev/null

exit $STATUS