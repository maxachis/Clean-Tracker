package service

import "github.com/maxachis/clean-tracker/wails/model"

func (s *Service) GetSettings() (model.Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Store.GetSettings()
}

func (s *Service) UpdateSettings(req model.UpdateSettingsRequest) (model.Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.LastSessionMinutes < 1 {
		return model.Settings{}, ErrInvalidLength
	}
	updated := model.Settings{LastSessionMinutes: req.LastSessionMinutes, MissSound: req.MissSound}
	if err := s.Store.UpdateSettings(updated); err != nil {
		return model.Settings{}, err
	}
	return updated, nil
}
