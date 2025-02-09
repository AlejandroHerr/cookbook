package intelligence

type Config struct {
	OpenAIKey string `env:"OPENAI_KEY,notEmpty,required" json:"-"`
}
