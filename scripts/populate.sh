#!/bin/bash

set -e

# Configuration
INFLUX_HOST="http://localhost:8086"
INFLUX_ORG="activity"
INFLUX_BUCKET="activity"
INFLUX_TOKEN="secret"
LOG_FILE="$(dirname "$0")/../logs/activity.log"

if ! command -v jq &> /dev/null; then
    echo "jq is required but not installed. Install with: brew install jq"
    exit 1
fi

# Process each line of the log file
while IFS= read -r line; do
    # Parse JSON and extract data with proper format for InfluxDB line protocol
    echo "$line" | jq -r '
        .apps[] as $app |
        "app_activity,app_name=" +
        ($app.name | gsub(" "; "\\ ")) +
        " is_active=" +
        (if $app.is_active then "true" else "false" end) +
        " " +
        (.created_at | tostring) + "000000000"
    ' | while read -r point; do
        # Send data to InfluxDB
        curl -s -X POST "${INFLUX_HOST}/api/v2/write?org=${INFLUX_ORG}&bucket=${INFLUX_BUCKET}&precision=ns" \
            -H "Authorization: Token ${INFLUX_TOKEN}" \
            -H "Content-Type: text/plain; charset=utf-8" \
            --data-raw "$point"
    done
done < "$LOG_FILE"

echo
echo "Import completed"
