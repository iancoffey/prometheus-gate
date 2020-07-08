# prometheus-gate

`prometheus-gate` is a small utility image that is intended to block Pipeline eecutions based on the results Prometheus RangeQuery.

 This tool uses a Range Query, a target Value and a duration of time, and enforces that the metrics match the criteria before exiting 0. If the criteria are not met by the timeout, it will exit non-zero.

So, you can gate your deployment from staging to QA until the staging environment has had 3 replicas available consistently for 10 minutes. Or, you could gate your ML models Canary deployment until the new models R2 score proves to be > N for 30 minutes.

Anything that can emit metrics to Prometheus can be used to form your query, there is no limit to the type of metrics that can be checked with this.

## Configuration

**PROMETHEUS_ENDPOINT**: URI of your Prometheus API server

**RANGE_QUERY**: The actual Range Query

**RANGE_TIME**: The timeframe we should ensure is at the target threshold

**TARGET_VALUE**: The actual Value we are asserting - eg, 3 replicas, 94(ms), 10(GB)

**TARGET_STRATEGY**: Valid values are `min`, `max`, `equal`

**TIMEOUT**: How long to let the process continue checking until it gives up

**TICK_TIME**: This process will hit the Prometheus API at this internal
