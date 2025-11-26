package packtrack

// MetricsHooks provides optional callbacks for observability.
type MetricsHooks struct {
	OnIngestSuccess func(count int)
	OnIngestFailure func(count int)
	OnQueueDepth    func(depth int)
}
