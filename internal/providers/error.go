package providers

type Refreshable interface {
    TriggerRefresh()
}

type ErrorReporter interface {
    LastError() error
}
