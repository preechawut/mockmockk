package http

type CreateProjectRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type RenameProjectRequest struct {
	Name string `json:"name"`
}
