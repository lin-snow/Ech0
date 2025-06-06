package service

import model "github.com/lin-snow/ech0/internal/model/setting"

type SettingServiceInterface interface {
	GetSetting(setting *model.SystemSetting) error
	UpdateSetting(userid uint, newSetting *model.SystemSettingDto) error
}
