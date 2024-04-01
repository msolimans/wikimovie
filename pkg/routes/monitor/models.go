package monitor

type ServiceState struct {
	Healthy bool
	Error   string
}
type StatusReport struct {
	ElasticSearch *ServiceState
}
