package itchio

import (
	"bytes"
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

var gameTraitToIndex map[string]int
var gameTraits []string

func init() {
	typ := reflect.TypeOf(GameTraits2{})
	gameTraits = make([]string, typ.NumField())
	gameTraitToIndex = make(map[string]int)
	for i := 0; i < typ.NumField(); i++ {
		trait := typ.Field(i).Tag.Get("trait")
		gameTraitToIndex[trait] = i
		gameTraits[i] = trait
	}
}

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

//-----------------

func (gt GameTraits2) MarshalJSON2() ([]byte, error) {
	var traits []string
	val := reflect.ValueOf(gt)
	for i, trait := range gameTraits {
		if val.Field(i).Bool() {
			traits = append(traits, trait)
		}
	}
	return json.Marshal(traits)
}

func (gt *GameTraits2) UnmarshalJSON2(data []byte) error {
	var traits []string
	err := json.Unmarshal(data, &traits)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(gt).Elem()
	for _, trait := range gameTraits {
		val.Field(gameTraitToIndex[trait]).SetBool(true)
	}
	return nil
}

//-----------------

func (gt GameTraits2) MarshalJSON3() ([]byte, error) {
	var bb bytes.Buffer
	bb.WriteByte('[')

	first := true
	val := reflect.ValueOf(gt)
	for i, trait := range gameTraits {
		if val.Field(i).Bool() {
			if first {
				first = false
			} else {
				bb.WriteByte(',')
			}
			bb.WriteByte('"')
			bb.WriteString(trait)
			bb.WriteByte('"')
		}
	}
	bb.WriteByte(']')
	return bb.Bytes(), nil
}

func (gt *GameTraits2) UnmarshalJSON3(data []byte) error {
	i := 0
	val := reflect.ValueOf(gt).Elem()
	for i < len(data) {
		switch data[i] {
		case '"':
			j := i + 1
		scanString:
			for {
				switch data[j] {
				case '"':
					trait := string(data[i+1 : j])
					i = j + 1
					val.Field(gameTraitToIndex[trait]).SetBool(true)
					break scanString
				default:
					j++
				}
			}
		case ']':
			return nil
		default:
			i++
		}
	}
	return nil
}

//-----------------

func (gt GameTraits2) MarshalJSON4() ([]byte, error) {
	var bb bytes.Buffer
	bb.WriteByte('[')

	first := true
	if gt.PlatformAndroid {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"p_android"`)
	}
	if gt.PlatformWindows {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"p_windows"`)
	}
	if gt.PlatformLinux {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"p_linux"`)
	}
	if gt.PlatformOSX {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"p_osx"`)
	}
	if gt.HasDemo {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"has_demo"`)
	}
	if gt.CanBeBought {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"can_be_bought"`)
	}
	if gt.InPressSystem {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"in_press_system"`)
	}
	bb.WriteByte(']')
	return bb.Bytes(), nil
}

func (gt *GameTraits2) UnmarshalJSON4(data []byte) error {
	i := 0
	for i < len(data) {
		switch data[i] {
		case '"':
			j := i + 1
		scanString:
			for {
				switch data[j] {
				case '"':
					trait := data[i+1 : j]
					switch trait[0] {
					case 'p':
						switch trait[2] {
						case 'w':
							gt.PlatformWindows = true
						case 'l':
							gt.PlatformLinux = true
						case 'o':
							gt.PlatformOSX = true
						case 'a':
							gt.PlatformAndroid = true
						}
					case 'h':
						gt.HasDemo = true
					case 'c':
						gt.CanBeBought = true
					case 'i':
						gt.InPressSystem = true
					}
					i = j + 1
					break scanString
				default:
					j++
				}
			}
		case ']':
			return nil
		default:
			i++
		}
	}
	return nil
}
