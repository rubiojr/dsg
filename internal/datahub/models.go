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

// Dataset represents a DataHub dataset entity
type Dataset struct {
	SchemaMetadata         SchemaMetadataContainer         `json:"schemaMetadata"`
	Key                    DatasetKeyContainer             `json:"datasetKey"`
	GlobalTags             GlobalTagsContainer             `json:"globalTags"`
	GlossaryTerms          GlossaryTermsContainer          `json:"glossaryTerms"`
	URN                    string                          `json:"urn"`
	EditableSchemaMetadata EditableSchemaMetadataContainer `json:"editableSchemaMetadata,omitempty"`
}

type EditableSchemaMetadata struct {
	EditableSchemaFieldInfo []EditableSchemaFieldInfo `json:"editableSchemaFieldInfo"`
}

type EditableSchemaFieldInfo struct {
	FieldPath     string                      `json:"fieldPath"`
	GlossaryTerms FieldGlossaryTermsContainer `json:"glossaryTerms"`
}

type EditableSchemaMetadataContainer struct {
	Value EditableSchemaMetadata `json:"value"`
}

// SchemaMetadataContainer wraps SchemaMetadata with a value field
type SchemaMetadataContainer struct {
	Value SchemaMetadata `json:"value"`
}

// SchemaMetadata contains metadata about the dataset schema
type SchemaMetadata struct {
	SchemaName     string         `json:"schemaName"`
	Platform       string         `json:"platform"`
	Version        int            `json:"version"`
	Hash           string         `json:"hash"`
	PlatformSchema PlatformSchema `json:"platformSchema"`
	Fields         []SchemaField  `json:"fields"`
}

// PlatformSchema contains platform-specific schema information
type PlatformSchema struct {
	MySqlDDL MySqlDDL `json:"com.linkedin.schema.MySqlDDL"`
}

// MySqlDDL contains MySQL-specific DDL information
type MySqlDDL struct {
	TableSchema string `json:"tableSchema"`
}

// SchemaField represents a field in the schema
type SchemaField struct {
	FieldPath      string                       `json:"fieldPath"`
	Description    string                       `json:"description"`
	Type           FieldTypeContainer           `json:"type"`
	NativeDataType string                       `json:"nativeDataType"`
	Recursive      bool                         `json:"recursive"`
	GlossaryTerms  *FieldGlossaryTermsContainer `json:"glossaryTerms,omitempty"`
}

type FieldGlossaryTermsContainer struct {
	Terms      []TermAssociation `json:"terms"`
	AuditStamp AuditStamp        `json:"auditStamp"`
}

// FieldTypeContainer wraps a field type with a type field
type FieldTypeContainer struct {
	Type FieldType `json:"type"`
}

// FieldType represents the type of a field, which can be one of several types
type FieldType struct {
	StringType *struct{} `json:"com.linkedin.schema.StringType,omitempty"`
	NumberType *struct{} `json:"com.linkedin.schema.NumberType,omitempty"`
}

// DatasetKeyContainer wraps DatasetKey with a value field
type DatasetKeyContainer struct {
	Value DatasetKey `json:"value"`
}

// DatasetKey contains key information about the dataset
type DatasetKey struct {
	Platform string `json:"platform"`
	Name     string `json:"name"`
	Origin   string `json:"origin"`
}

// GlobalTagsContainer wraps GlobalTags with a value field
type GlobalTagsContainer struct {
	Value GlobalTags `json:"value"`
}

// GlobalTags contains tags associated with the dataset
type GlobalTags struct {
	Tags []TagAssociation `json:"tags"`
}

// TagAssociation represents a tag associated with an entity
type TagAssociation struct {
	Tag string `json:"tag"`
}

// GlossaryTermsContainer wraps GlossaryTerms with a value field
type GlossaryTermsContainer struct {
	Value GlossaryTerms `json:"value"`
}

// GlossaryTerms contains glossary terms associated with an entity
type GlossaryTerms struct {
	Terms      []TermAssociation `json:"terms"`
	AuditStamp AuditStamp        `json:"auditStamp"`
}

// TermAssociation represents a glossary term associated with an entity
type TermAssociation struct {
	URN string `json:"urn"`
}

// AuditStamp contains audit information about when changes were made
type AuditStamp struct {
	Time  int64  `json:"time"`
	Actor string `json:"actor"`
}
