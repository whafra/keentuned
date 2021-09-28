package file 
import (
	"io/ioutil"
	"encoding/json"
	"fmt"	
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// IsPathExist ...
func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

// DecoratePath ...
func DecoratePath(path string) string {
	var retPath string
	if len(path) > 0 {
		if string(path[len(path)-1]) == "/" {
			return path
		}

		retPath = path + "/"
		return retPath
	}

	return retPath
}

// ReadFile2Map ...
func ReadFile2Map(path string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read [%v] file err:%v\n", path, err)
	}

	var retMap map[string]interface{}
	err = json.Unmarshal(bytes, &retMap)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal err:%v\n", err)
	}

	return retMap, nil
}

// Dump2File ...
func Dump2File(path, fileName string, info interface{}) error {
	if !IsPathExist(path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("make dir err:[%v]", err)
		}
	}

	fullPath:= path + "/" + fileName	
	
	resultBytes, err := json.Marshal(info)
	if err != nil {
		return  fmt.Errorf("marshal info to bytes err:[%v] ", fileName, err)
	}

	err = ioutil.WriteFile(fullPath, resultBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write to file [%v] err:[%v] ", fileName, err)
	}

	return nil
}

//  WalkFilePath walk file path for file or path list by onlyDir
// return file list  while onlyDir is false, else return path list.
func WalkFilePath(folder , match string, onlyDir bool) ([]string, error) {
    var result []string

    filepath.Walk(folder, func(path string, fi os.FileInfo, err error) error {
        if err != nil {
            return err
        }

		pathSections :=strings.Split(path, "/")
		if len(pathSections) == 0 {
			return fmt.Errorf("get path [%v] Sections is null", path)
		}

		if fi.IsDir() && onlyDir {			
			fileName := pathSections[len(pathSections)-1]
			if fileName != ""{
				result = append(result, fileName)
			}			
		}

        if !fi.IsDir() && strings.Contains(path, match) && !onlyDir {			
			fileName := pathSections[len(pathSections)-1]
            result = append(result, fileName)
        }

        return nil
    })

    return result, nil
}

// ConvertConfFileToJson convert conf file to json
func ConvertConfFileToJson(fileName string) (map[string]interface{}, error){
	paramBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read file :%v err:%v\n", fileName, err)
	}

	var resultMap = make(map[string]interface{})
	var domainMap = make(map[string][]map[string]interface{})

	commonDomain := ""
	for _, line := range strings.Split(string(paramBytes), "\n") {
		if len(line) == 0 {
			continue
		}

		if strings.Contains(line, "[") {
			commonDomain = strings.Trim(strings.Trim(line, "]"), "[")
			continue
		}

		param, err := readLine(line)
		if err != nil {
			fmt.Printf("read line [%v] err:%v\n", line, err)
			continue
		}

		domainMap[commonDomain] = append(domainMap[commonDomain], param)
	}

	for domain, paramSlice := range domainMap {
		var paramMap = make(map[string]interface{})
		for _, paramInfo := range paramSlice {
			name, ok := paramInfo["name"].(string)
			if !ok {
				fmt.Printf("parse name from [%v] failed\n", paramInfo)
				continue
			}
			delete(paramInfo, "name")
			paramMap[name] = paramInfo
		}

		resultMap[domain] = paramMap
	}

	return resultMap, nil
}

func readLine(line string) (map[string]interface{}, error) {
	var param map[string]interface{}
	paramSlice := strings.Split(line, ":")
	if len(paramSlice) != 2 {
		return nil, fmt.Errorf("param slice %v length is less than 2", paramSlice)
	}

	paramName := strings.Trim(paramSlice[0]," ")
	valueStr := strings.ReplaceAll(strings.Trim(paramSlice[1]," "), "\"", "")
	
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		param = map[string]interface{}{
			"value": valueStr,
			"dtype": "string",
			"name":  paramName,
		}
		return param, nil
	}

	param = map[string]interface{}{
		"value": value,
		"dtype": "int",
		"name":  paramName,
	}

	return param, nil
}