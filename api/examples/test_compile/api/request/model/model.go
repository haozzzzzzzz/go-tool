package model

type anonymousModel struct {
	ExportedFieldOfAnonymousModel string `json:"exported_field_of_anonymous_model" form:"exported_field_of_anonymous_model"`
}
type Model2 struct {
	Model2Field1         string         `json:"model_2_field_1" form:"model_2_field_1"`
	ExportAnonymousModel anonymousModel `json:"export_anonymous_model"`
}
