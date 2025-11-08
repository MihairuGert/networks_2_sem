package ui

import (
	"errors"
	"snake-game/internal/domain"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

var (
	ErrNoPath            = errors.New("no path specified")
	ErrUnsupportedFormat = errors.New("unsupported config file format")
	ErrInvalidConfig     = errors.New("invalid configuration")
)

// ParseConfig parses a configuration file and returns a GameConfig
func ParseConfig(path string) (*domain.GameConfig, error) {
	if len(path) == 0 {
		return nil, ErrNoPath
	}

	k := koanf.New(".")

	var parser koanf.Parser
	switch {
	case len(path) > 5 && path[len(path)-5:] == ".json":
		parser = json.Parser()
	case len(path) > 5 && path[len(path)-5:] == ".yaml":
		parser = yaml.Parser()
	default:
		return nil, ErrUnsupportedFormat
	}

	if err := k.Load(file.Provider(path), parser); err != nil {
		return nil, err
	}

	var config domain.GameConfig
	if err := k.Unmarshal("", &config); err != nil {
		return nil, err
	}

	return &config, nil
}

//func validateConfig(config domain.GameConfig) error {
//	if config.Width < 5 || config.Width > 100 {
//		return errors.New("width must be between 5 and 100")
//	}
//
//	if config.Height < 5 || config.Height > 100 {
//		return errors.New("height must be between 5 and 100")
//	}
//
//	if config.FoodStatic < 0 || config.FoodStatic > 100 {
//		return errors.New("food_static must be between 0 and 100")
//	}
//
//	if config.StateDelayMs < 100 || config.StateDelayMs > 3000 {
//		return errors.New("state_delay_ms must be between 100 and 3000")
//	}
//
//	return nil
//}
