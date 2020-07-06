# prometheus-gate

`prometheus-gate` is a small utility image that is intended to block Pipeline eecutions based on the results Prometheus RangeQuery.

A threshold and timeframe are set, and if the conditions are met for the entire time-period, the gate exits 0.
