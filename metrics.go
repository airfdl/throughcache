package throughcache

var (
	metricsClient Metricor = new(emptyMetric)
	emptyMetricor          = new(emptyMetric)
)

type emptyMetric struct{}

func (*emptyMetric) EmitTimer(name string, value interface{}, tags map[string]string) error {
	return nil
}
func (*emptyMetric) EmitCounter(name string, value interface{}, tags map[string]string) error {
	return nil
}

func MetricTODO() Metricor {
	return emptyMetricor
}

type Metricor interface {
	EmitTimer(string, interface{}, map[string]string) error
	EmitCounter(string, interface{}, map[string]string) error
}

func SetMetrics(m Metricor) {
	metricsClient = m
}
