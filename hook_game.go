package itchio

import (
	"reflect"
)

// GameHookFunc is used to transform API results from
// what they currently are to what we expect them to be.
func GameHookFunc(
	f reflect.Type,
	t reflect.Type,
	data any) (any, error) {

	if t != reflect.TypeFor[Game]() {
		return data, nil
	}

	if gameMap, ok := data.(map[string]any); ok {
		// API v2 briefly returned traits, which were 100%
		// amos's terrible idea - they're going away, but
		// in the meantime let's convert them to something saner.
		if traitsAny, ok := gameMap["traits"]; ok {
			platforms := make(map[string]any)
			if traits, ok := traitsAny.([]any); ok {
				for _, traitAny := range traits {
					if trait, ok := traitAny.(string); ok {
						switch trait {
						case "p_osx":
							platforms["osx"] = ArchitecturesAll
						case "p_windows":
							platforms["windows"] = ArchitecturesAll
						case "p_linux":
							platforms["linux"] = ArchitecturesAll
						case "can_be_bought":
							gameMap["canBeBought"] = true
						case "has_demo":
							gameMap["hasDemo"] = true
						case "in_press_system":
							gameMap["inPressSystem"] = true
						}
					}
				}
			}
			gameMap["platforms"] = platforms
			delete(gameMap, "traits")
			return gameMap, nil
		}
	}

	return data, nil
}
