package datahub

type GlossaryTerm struct {
	URN  string           `json:"urn"`
	Info GlossaryTermInfo `json:"glossaryTermInfo"`
}

type GlossaryTermInfo struct {
	Value GlossaryTermValue `json:"value"`
}

type GlossaryTermValue struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Source     string `json:"termSource"`
}
