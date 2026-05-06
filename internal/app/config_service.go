package app

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/heybox/readerx/internal/domain"
	"github.com/heybox/readerx/internal/storage"
)

type ConfigService struct {
	store storage.Store
}

var allowedSettings = map[string]func(string) error{
	"reader.width": validatePositiveInt,
	"reader.theme": func(value string) error {
		switch value {
		case "default", "dark", "light":
			return nil
		default:
			return fmt.Errorf("reader.theme must be one of: default, dark, light")
		}
	},
	"search.limit": validatePositiveInt,
}

func NewConfigService(store storage.Store) *ConfigService {
	return &ConfigService{store: store}
}

func (s *ConfigService) Set(key, value string) error {
	validate, ok := allowedSettings[key]
	if !ok {
		return fmt.Errorf("unknown config key %q", key)
	}
	if err := validate(value); err != nil {
		return err
	}
	return s.store.SetSetting(key, value)
}

func (s *ConfigService) Get(key string) (string, error) {
	value, err := s.store.GetSetting(key)
	if errors.Is(err, storage.ErrNotFound) {
		return defaultSettingValue(key)
	}
	return value, err
}

func (s *ConfigService) List() ([]domain.Setting, error) {
	stored, err := s.store.ListSettings()
	if err != nil {
		return nil, err
	}
	byKey := map[string]domain.Setting{}
	for _, setting := range stored {
		byKey[setting.Key] = setting
	}
	keys := []string{"reader.width", "reader.theme", "search.limit"}
	settings := make([]domain.Setting, 0, len(keys))
	for _, key := range keys {
		if setting, ok := byKey[key]; ok {
			settings = append(settings, setting)
			continue
		}
		value, _ := defaultSettingValue(key)
		settings = append(settings, domain.Setting{Key: key, Value: value})
	}
	return settings, nil
}

func defaultSettingValue(key string) (string, error) {
	switch key {
	case "reader.width":
		return "80", nil
	case "reader.theme":
		return "default", nil
	case "search.limit":
		return "50", nil
	default:
		return "", fmt.Errorf("unknown config key %q", key)
	}
}

func validatePositiveInt(value string) error {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fmt.Errorf("value must be a positive integer")
	}
	return nil
}
