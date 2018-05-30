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

type GameTraits2 struct {
	PlatformWindows bool `trait:"p_windows"`
	PlatformLinux   bool `trait:"p_linux"`
	PlatformOSX     bool `trait:"p_osx"`
	PlatformAndroid bool `trait:"p_android"`
	CanBeBought     bool `trait:"can_be_bought"`
	HasDemo         bool `trait:"has_demo"`
	InPressSystem   bool `trait:"in_press_system"`
}

var _ json.Marshaler = GameTraits2{}
var _ json.Unmarshaler = (*GameTraits2)(nil)

func (gt GameTraits2) MarshalJSON() ([]byte, error) {
	var traits []string
	val := reflect.ValueOf(gt)
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).Bool() {
			traits = append(traits, typ.Field(i).Tag.Get("trait"))
		}
	}
	return json.Marshal(traits)
}

func (gt *GameTraits2) UnmarshalJSON(data []byte) error {
	var traits []string
	err := json.Unmarshal(data, &traits)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(gt).Elem()
	typ := val.Type()
	for _, t := range traits {
		for i := 0; i < typ.NumField(); i++ {
			tf := typ.Field(i)
			if tf.Tag.Get("trait") == t {
				val.Field(i).SetBool(true)
			}
		}
	}
	return nil
}
