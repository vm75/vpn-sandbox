package utils

import (
	"encoding/json"
	"reflect"
)

func ObjectToMap(obj interface{}, out *map[string]interface{}) error {
	in, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(in, out)
	if err != nil {
		return err
	}
	return nil
}

func MapToObject(obj map[string]interface{}, out interface{}) error {
	in, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(in, out)
	if err != nil {
		return err
	}
	return err
}

func AreEqual(obj1, obj2 interface{}) bool {
	return reflect.DeepEqual(obj1, obj2)
}

func HasChanged(obj interface{}, config map[string]interface{}) bool {
	var currentConfig map[string]interface{}
	ObjectToMap(obj, &currentConfig)
	return !AreEqual(currentConfig, config)
}
