package grouping

import (
	"github.com/ottemo/commerce/env"
	"github.com/ottemo/commerce/utils"
)

// setupConfig setups package configuration values for a system
func setupConfig() error {
	config := env.GetConfig()
	if config == nil {
		err := env.ErrorNew(ConstErrorModule, env.ConstErrorLevelStartStop, "b2c1c442-36b9-4994-b5d1-7c948a7552bd", "can't obtain config")
		return env.ErrorDispatch(err)
	}

	// validateNewRules validate structure of new rules
	validateNewRules := func(newRulesValues interface{}) (interface{}, error) {

		var rules []interface{}
		// taking rules as array
		if newRulesValues != "" && newRulesValues != nil {

			var err error
			switch value := newRulesValues.(type) {
			case string:
				rules, err = utils.DecodeJSONToArray(value)
				if err != nil {
					return nil, env.ErrorDispatch(err)
				}
			case []interface{}:
				rules = value
			default:
				err := env.ErrorNew(ConstErrorModule, ConstErrorLevel, "32417b78-6881-471e-922c-d7d7d4ddef1f", "can't convert to array")
				return nil, env.ErrorDispatch(err)
			}

			// checking rules array
			for _, rule := range rules {
				ruleItem := utils.InterfaceToMap(rule)

				if !utils.KeysInMapAndNotBlank(ruleItem, "group", "into") {
					err := env.ErrorNew(ConstErrorModule, ConstErrorLevel, "7912df05-8ea7-451e-83bd-78e9e201378e", "keys 'group' and 'into' should be not null")
					return nil, env.ErrorDispatch(err)
				}

				// checking product specification arrays
				for _, groupingValue := range []interface{}{ruleItem["group"], ruleItem["into"]} {
					groupingElement := utils.InterfaceToArray(groupingValue)

					for _, productValue := range groupingElement {
						productElement := utils.InterfaceToMap(productValue)

						if !utils.KeysInMapAndNotBlank(productElement, "pid", "qty") {
							err := env.ErrorNew(ConstErrorModule, ConstErrorLevel, "6b9deedd-39d1-46b0-9157-9b8d96bda858", "keys 'qty' and 'pid' should be not null")
							return nil, env.ErrorDispatch(err)
						}
					}
				}
			}

			currentRules = rules
		}

		return newRulesValues, nil
	}

	// grouping rules config setup
	//----------------------------
	err := config.RegisterItem(env.StructConfigItem{
		Path:    ConstGroupingConfigPath,
		Value:   ``,
		Type:    env.ConstConfigTypeJSON,
		Editor:  "multiline_text",
		Options: "",
		Label:   "Rules for grouping items",
		Description: `Rules must be in JSON format:
[
	{
		"group": [{ "pid": "id1", "qty": 1 }, ...],
		"into":  [{ "pid": "id2", "qty": 1, "options": {"color": "red"}, ...]
	}, ...
]`,
		Image: "",
	}, env.FuncConfigValueValidator(validateNewRules))

	if err != nil {
		return env.ErrorDispatch(err)
	}
	return nil
}
