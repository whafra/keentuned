package utils

import (
	"encoding/json"
	"fmt"
)

// Parse2Map parse to map from body by key
func Parse2Map(key string, body interface{}) map[string]interface{} {
	var originMap map[string]interface{}
	switch t := body.(type) {
	case []byte:
		if err := json.Unmarshal(t, &originMap); err != nil {
			return nil
		}
	case map[string]interface{}:
		originMap = t
	default:
	}

	info, ok := originMap[key]
	if !ok {
		return nil
	}

	retMap, ok := info.(map[string]interface{})
	if !ok {
		return nil
	}

	return retMap
}

// ParseValue parse value from body by key 
func ParseValue(key string, body []byte) (interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	value, ok := result[key]
	if !ok {
		fmt.Printf("response assert %v is not ok\n", key)
		return value, fmt.Errorf("response assert %v is not ok", key)
	}

	return value, nil
}

//  Map2Struct  convert map to dest struct 
func Map2Struct(m map[string]interface{}, dest interface{}) error {
	byte, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal [%+v] err:[%v]", m, err)
	}

	if err = json.Unmarshal(byte, dest); err != nil {
		return fmt.Errorf("unmarshal [%+v] err:[%v]", string(byte), err)
	}
	return nil
}

// Interface2Map convert source interface info to map
func Interface2Map(source interface{}) (map[string]interface{}, error) {
	var bodyBytes []byte
	var err error
	bodyBytes, ok := source.([]byte)
	if !ok {
		bodyBytes, err = json.Marshal(source)
		if err != nil {
			return nil, fmt.Errorf("marshal [%+v] err:[%v]", source, err)
		}
	}

	var retMap = make(map[string]interface{})
	if err = json.Unmarshal(bodyBytes, &retMap); err != nil {
		return nil, fmt.Errorf("unmarshal [%+v] err:[%v]", string(bodyBytes), err)
	}
	return retMap, nil
}

//  Struct2Map  convert source struct to map 
func Struct2Map(source interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("struct to map error")
	}

	var resultMap map[string]interface{}
	if err = json.Unmarshal(bytes, &resultMap); err != nil {
		return nil, fmt.Errorf("unmarshal to map error")
	}

	return resultMap, nil
}

// Parse2Struct parse to dest struct from body by key
func Parse2Struct(key string, body interface{}, dest interface{}) error {
	return Map2Struct(Parse2Map(key, body), dest)
}
