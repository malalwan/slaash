package models

// holds data sent from handlers to templates
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32
	Data      map[string]interface{} // we use interfaces when the type is not known
	CSRFToken string
	Flash     string
	Warning   string
	Error     string
}
