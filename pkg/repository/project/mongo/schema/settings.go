package schema

type Settings struct {
	ProjectID string `bson:"project_id" json:"project_id"`
	Name      string `bson:"name,omitempty" json:"name,omitempty"`
	Data      []byte `bson:"data,omitempty" json:"data,omitempty"`
}
