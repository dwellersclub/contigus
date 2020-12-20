package models

//Event event generated
type Event struct {
	Data          Data
	Source        Source
	EmitterConfig EmitterConfig
	Origin        Origin
}

//Data event data
type Data struct {
	Header   []byte
	Content  []byte
	Type     string
	EncKeyID string
}

//EmitterConfig config to dispatch event
type EmitterConfig struct {
	ID     string
	Config interface{}
}

//Origin event origin
type Origin struct {
	CorrolationID string
	ServerID      string
}

//Source event source
type Source struct {
	Date int64
	Type string
	ID   string
}
