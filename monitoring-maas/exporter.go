package maas

// Connector represents a data source connection.
type Connector interface {
	Execute(query interface{}) (interface{}, error)
}

// Exporter is a minimal stub used for testing.
type Exporter interface {
	Start()
	Serve()
}
