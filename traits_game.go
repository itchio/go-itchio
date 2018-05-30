package itchio

import (
	"encoding/json"
	"log"
	"reflect"
)

type GameTrait string

const (
	GameTraitPlatformWindows GameTrait = "p_windows"
	GameTraitPlatformLinux   GameTrait = "p_linux"
	GameTraitPlatformOSX     GameTrait = "p_osx"
	GameTraitPlatformAndroid GameTrait = "p_android"
	GameTraitCanBeBought     GameTrait = "can_be_bought"
	GameTraitHasDemo         GameTrait = "has_demo"
	GameTraitInPressSystem   GameTrait = "in_press_system"
)

type GameTraits map[GameTrait]bool

var _ json.Marshaler = (GameTraits)(nil)
var _ json.Unmarshaler = (*GameTraits)(nil)

func (gt GameTraits) MarshalJSON() ([]byte, error) {
	var traits []GameTrait
	for k, v := range gt {
		if v {
			traits = append(traits, k)
		}
	}
	return json.Marshal(traits)
}

func (gtp *GameTraits) UnmarshalJSON(data []byte) error {
	gt := make(GameTraits)
	var traits []GameTrait
	err := json.Unmarshal(data, &traits)
	if err != nil {
		return err
	}

	for _, k := range traits {
		gt[k] = true
	}
	*gtp = gt
	return nil
}

func GameTraitHookFunc(
	f reflect.Type,
	t reflect.Type,
	data interface{}) (interface{}, error) {

	log.Printf("Hook called with f %v, t %v, data %v", f, t, data)

	if f.Kind() != reflect.Slice {
		return data, nil
	}

	if t != reflect.TypeOf(GameTraits{}) {
		return data, nil
	}

	gt := make(GameTraits)
	var traits = data.([]interface{})
	for _, k := range traits {
		if trait, ok := k.(GameTrait); ok {
			gt[trait] = true
		}
	}

	log.Printf("Converted! before = %v, after = %v", traits, gt)

	return gt, nil
}
