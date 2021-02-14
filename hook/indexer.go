package hook

//Indexer Index object path for a given event type
type Indexer interface {
	Index(eventType string, fieldPath string)
}
