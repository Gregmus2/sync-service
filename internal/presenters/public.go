package presenters

type Public struct {
}

func NewAPI() TemplateServer {
	return &Public{}
}
