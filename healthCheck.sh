#!/bin/bash

# Perform the first GET request
if ! curl -f http://localhost:8080/api/v1/health/router > /dev/null 2>&1; then
    echo "Health check failed for /health/router"
    exit 1
fi

# Perform the second GET request
if ! curl -f http://localhost:8080/api/v1/health/redis > /dev/null 2>&1; then
    echo "Health check failed for /health/redis"
    exit 1
fi

echo "Health Check Passed!"
exit 0
