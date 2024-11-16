package formats

import (
	"fmt"

	"github.com/jphastings/recipes/internal/llm"
	"github.com/spf13/viper"
)

func LoadOptions() (ParseOptions, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.recipes")
	viper.SetEnvPrefix("RECIPE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return ParseOptions{}, fmt.Errorf("unable to read config: %w", err)
	}

	o := ParseOptions{}

	if url := viper.GetString("LLM_URL"); url != "" {
		c, err := llm.NewLMStudioConnection(url, viper.GetString("LLM_MODEL"))
		if err == nil {
			o.LLM = c
		}
	}

	return o, nil
}
