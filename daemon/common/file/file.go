/*
Copyright © 2021 KeenTune

Package file for daemon, this package contains the file. The package implements the processing of files, including the transformation of file types, path acquisition, and directory existence judgment.
*/
package file

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const recommendReg = "^recommend:+$"

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

// DecoratePath cut the end of the path separator "/"
func DecoratePath(path string) string {
	if len(path) == 0 {
		return path
	}

	if string(path[len(path)-1]) == "/" {
		return strings.TrimSuffix(path, "/")
	}

	return path
}

// ReadFile2Map ...
func ReadFile2Map(path string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read [%v] file:%v", path, err)
	}

	var retMap map[string]interface{}
	err = json.Unmarshal(bytes, &retMap)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal err: %v", err)
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

	fullPath := path + "/" + fileName

	resultBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal info %v to bytes err:[%v]", info, err)
	}

	err = ioutil.WriteFile(fullPath, resultBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write to file [%v] err:[%v]", fileName, err)
	}

	return nil
}

//  WalkFilePath walk file path for file or path list by onlyDir
// return file list  while onlyDir is false, else return path list.
func WalkFilePath(folder, match string, onlyDir bool, separators ...string) ([]string, error) {
	var result []string
	var separator string
	if len(separators) != 0 {
		separator = separators[0]
	}

	filepath.Walk(folder, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		pathSections := strings.Split(path, "/")
		if len(pathSections) == 0 {
			return fmt.Errorf("get path [%v] Sections is null", path)
		}

		var fileName string

		if fi.IsDir() && onlyDir {
			fileName = pathSections[len(pathSections)-1]
			if fileName != "" && !strings.Contains(fileName, strings.Trim(separator, "/")) {
				result = append(result, fileName)
			}
		}

		if !fi.IsDir() && strings.Contains(path, match) && !onlyDir {
			fileName = pathSections[len(pathSections)-1]
			result = append(result, fileName)
		}

		return nil
	})

	return result, nil
}

// ConvertConfFileToJson convert conf file to json
func ConvertConfFileToJson(fileName string) (string, map[string]map[string]interface{}, error) {
	paramBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", nil, fmt.Errorf("read file err: %v", err)
	}

	if len(paramBytes) == 0 {
		return "", nil, fmt.Errorf("read file is empty")
	}

	var resultMap = make(map[string]map[string]interface{})
	var domainMap = make(map[string][]map[string]interface{})

	commonDomain := ""
	recommends := ""
	var tmpRecommendMap = make(map[string][]string)
	for _, line := range strings.Split(string(paramBytes), "\n") {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		if strings.Contains(line, "[") {
			commonDomain = strings.Trim(strings.Trim(strings.TrimSpace(line), "]"), "[")
			continue
		}

		recommend, param, err := readLine(line)
		if err != nil {
			fmt.Printf("read line [%v] err:%v\n", line, err)
			continue
		}

		if len(recommend) != 0 {
			tmpRecommendMap[commonDomain] = append(tmpRecommendMap[commonDomain], recommend)
			continue
		}

		domainMap[commonDomain] = append(domainMap[commonDomain], param)
	}

	for key, value := range tmpRecommendMap {
		recommends += fmt.Sprintf("[%v]\n%v\n", key, strings.Join(value, ""))
	}

	if len(domainMap) == 0 {
		if recommends != "" {
			return recommends, nil, nil
		}

		return recommends, nil, fmt.Errorf("domain '%v' content is empty", commonDomain)
	}

	for domain, paramSlice := range domainMap {
		if len(paramSlice) == 0 {
			return recommends, nil, fmt.Errorf("domain '%v' content is empty", commonDomain)
		}

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

	return recommends, resultMap, nil
}

func readLine(line string) (string, map[string]interface{}, error) {
	paramSlice := strings.Split(line, ":")
	partLen := len(paramSlice)
	switch {
	case partLen <= 1:
		return "", nil, fmt.Errorf("param %v length %v is invalid, required: 2", paramSlice, len(paramSlice))
	case partLen == 2:
		return getParam(paramSlice)
	default:
		newSlice := []string{paramSlice[0]}
		newSlice = append(newSlice, strings.Join(paramSlice[1:], ":"))
		return getParam(newSlice)
	}
}

func getParam(paramSlice []string) (string, map[string]interface{}, error) {
	var param map[string]interface{}
	var recommend string
	paramName := strings.TrimSpace(paramSlice[0])
	valueStr := strings.ReplaceAll(strings.TrimSpace(paramSlice[1]), "\"", "")
	matched, _ := regexp.MatchString(recommendReg, strings.ToLower(valueStr))
	if matched {
		recommend = fmt.Sprintf("\t%v: %v\n", paramName, strings.TrimPrefix(valueStr, "recommend:"))
		return recommend, nil, nil
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		param = map[string]interface{}{
			"value": valueStr,
			"dtype": "string",
			"name":  paramName,
		}
		return recommend, param, nil
	}

	param = map[string]interface{}{
		"value": value,
		"dtype": "int",
		"name":  paramName,
	}
	return recommend, param, nil
}

func Save2CSV(path, fileName string, data map[string][]float32) error {
	if !IsPathExist(path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("save to %v cvs file, make dir err:[%v]\n", fileName, err)
			return fmt.Errorf("make dir err:[%v]", err)
		}
	}

	if len(data) == 0 {
		fmt.Printf("save to %v cvs file, data length is 0\n", fileName)
		return nil
	}

	fullName := fmt.Sprintf("%s/%s", path, fileName)

	newFile, err := os.Create(fullName)
	if err != nil {
		fmt.Printf("creat %v cvs file failed\n", fullName)
		return fmt.Errorf("create [%v] failed", fullName)
	}

	defer newFile.Close()

	w := csv.NewWriter(newFile)
	var headers []string
	for name, _ := range data {
		headers = append(headers, name)
	}

	dataPair := make([][]string, len(data[headers[0]]))

	sort.Strings(headers)
	contents := [][]string{
		headers,
	}

	for index := 0; index < len(headers); index++ {
		for row, value := range data[headers[index]] {
			dataPair[row] = append(dataPair[row], fmt.Sprintf("%v", value))
		}
	}

	contents = append(contents, dataPair...)
	w.WriteAll(contents)
	w.Flush()
	return nil
}

func GetWalkPath(folder, match string) (string, error) {
	var result string
	filepath.Walk(folder, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.IsDir() && strings.Contains(path, match) {
			result = path
		}

		return nil
	})

	return result, nil
}

func GetPlainName(fileName string) string {
	if !strings.Contains(fileName, "/") || fileName == "" {
		return fileName
	}

	parts := strings.Split(fileName, "/")
	return parts[len(parts)-1]
}

