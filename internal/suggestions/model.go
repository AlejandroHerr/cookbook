package suggestions

type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func NewSimpleOption(option string) Option {
	return Option{
		Label: option,
		Value: option,
	}
}
