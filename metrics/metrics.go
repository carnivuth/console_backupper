package metrics

import "github.com/prometheus/client_golang/prometheus"

type ConsoleBackupperMetrics struct {
	NumBackups prometheus.Counter
}
