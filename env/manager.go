package env

import (
	"errors"
)

var registeredConfig IConfig
var registeredIniConfig IINIConfig
var callbacksOnConfigStart = []func() error{}
var callbacksOnConfigIniStart = []func() error{}

// RegisterOnConfigStart allows the registration of callbacks to be executed upon application start.
func RegisterOnConfigStart(callback func() error) {
	callbacksOnConfigStart = append(callbacksOnConfigStart, callback)
}

// OnConfigStart executes the registered callbacks upon application start.
func OnConfigStart() error {
	for _, callback := range callbacksOnConfigStart {
		if err := callback(); err != nil {
			return err
		}
	}
	return nil
}

// RegisterOnConfigIniStart will register the ini file upon application start.
func RegisterOnConfigIniStart(callback func() error) {
	callbacksOnConfigIniStart = append(callbacksOnConfigIniStart, callback)
}

// OnConfigIniStart executes the registered callbacks to be run when the INI file has been initialized.
func OnConfigIniStart() error {
	for _, callback := range callbacksOnConfigIniStart {
		if err := callback(); err != nil {
			return err
		}
	}
	return nil
}

// RegisterIniConfig registers the new INI configuration file.
func RegisterIniConfig(IniConfig IINIConfig) error {
	if registeredIniConfig == nil {
		registeredIniConfig = IniConfig
	} else {
		return errors.New("There is other ini config already registered")
	}
	return nil
}

// RegisterConfig registers the configuration upon application start.
func RegisterConfig(Config IConfig) error {
	if registeredConfig == nil {
		registeredConfig = Config
	} else {
		return errors.New("There is other config already registered")
	}
	return nil
}

// GetConfig returns the registered configuration.
func GetConfig() IConfig { return registeredConfig }

// GetIniConfig returns the registered INI configuration that has been initialized.
func GetIniConfig() IINIConfig { return registeredIniConfig }

// ConfigEmptyValueValidator initializes an empty value validator.
func ConfigEmptyValueValidator(val interface{}) (interface{}, bool) { return val, true }
