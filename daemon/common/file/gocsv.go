package file

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/pkg/errors"
)

var (
	Empty         = errors.New("record is empty")
	NotExist      = errors.New("record not found")
	HeaderIsEmpty = errors.New("header is empty")
	AlreadyExist  = errors.New("job or record is already existence")
)

func LoadCsv(fileName string) (dataframe.DataFrame, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	defer f.Close()
	df := dataframe.ReadCSV(f, dataframe.DetectTypes(false))
	if df.Err != nil {
		return dataframe.DataFrame{}, df.Err
	}

	return df, nil
}

// CreatCSV create csv file with header
func CreatCSV(fileName string, header []string) error {
	if len(header) == 0 {
		return HeaderIsEmpty
	}

	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()

	return addRecord(fileName, header)

}

func addRecord(fileName string, header []string) error {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer f.Close()
	w := csv.NewWriter(f)
	err = w.Write(header)
	if err != nil {
		return err
	}

	w.Flush()
	return nil
}

func Append(fileName string, record []string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	defer f.Close()
	w := csv.NewWriter(f)
	err = w.Write(record)
	if err != nil {
		return err
	}

	w.Flush()
	return nil
}

// Insert Null value must be assigned "-"
func Insert(fileName string, record []string) error {
	df, err := LoadCsv(fileName)
	if err == nil {
		_, primaryName, err := GetPrimaryName(df)
		if err != nil {
			return err
		}

		oldJobs := df.Col(primaryName).Records()
		if IsInSlice(record[0], oldJobs) {
			return fmt.Errorf("'%v' %v", record[0], AlreadyExist)
		}
	}

	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()

	return Append(fileName, record)
}

// GetPrimaryName get name of COL.1 as the primary name
func GetPrimaryName(df dataframe.DataFrame) ([][]string, string, error) {
	records := df.Records()
	if len(records) < 2 || len(records[0]) == 0 {
		return nil, "", NotExist
	}

	primaryName := df.Records()[0][0]
	return records, primaryName, nil
}

func UpdateRow(fileName, jobName string, info map[int]interface{}) error {
	df, err := LoadCsv(fileName)
	if err != nil {
		if err.Error() == "load records: empty DataFrame" {
			return Empty
		}

		return err
	}

	records, primaryName, err := GetPrimaryName(df)
	if err != nil {
		return err
	}

	oldJobs := df.Col(primaryName).Records()
	if !IsInSlice(jobName, oldJobs) {
		return NotExist
	}

	rowIdx := 0
	var flush []string
	for index, row := range records {
		if row[0] == jobName {
			rowIdx = index
			flush = row
			break
		}
	}

	if len(flush) == 0 || rowIdx < 1 {
		return NotExist
	}

	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()

	for index, value := range info {
		flush[index] = fmt.Sprintf("%v", value)
	}

	header := records[0]
	newDF := df.Set(
		series.Ints([]int{rowIdx - 1}),
		dataframe.LoadRecords([][]string{header, flush}, dataframe.DetectTypes(false)),
	)

	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer f.Close()

	return newDF.WriteCSV(f)
}

func DeleteRow(fileName string, primaryKeys []string) error {
	df, err := LoadCsv(fileName)
	if err != nil {
		if err.Error() == "load records: empty DataFrame" {
			return Empty
		}
		return err
	}

	records, primaryName, err := GetPrimaryName(df)
	if err != nil {
		return err
	}

	var failedNames string
	for _, key := range primaryKeys {
		oldJobs := df.Col(primaryName).Records()
		if !IsInSlice(key, oldJobs) {
			failedNames += fmt.Sprintf("'%v' ", key)
		}
	}

	if len(failedNames) != 0 {
		return fmt.Errorf("record %v not found", failedNames)
	}

	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()

	var contents [][]string
	for index, row := range records {
		if index == 0 || !IsInSlice(row[0], primaryKeys) {
			contents = append(contents, row)
		}
	}

	f, err := os.OpenFile(fileName, os.O_WRONLY|syscall.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer f.Close()
	w := csv.NewWriter(f)
	err = w.WriteAll(contents)
	if err != nil {
		return err
	}
	w.Flush()
	return nil
}

func IsInSlice(obj string, pond []string) bool {
	for _, item := range pond {
		if obj == item {
			return true
		}
	}

	return false
}

func HasRecord(fileName, header, match string) bool {
	df, err := LoadCsv(fileName)
	if err != nil {
		return false
	}

	pond := df.Col(header).Records()

	return IsInSlice(match, pond)
}

func IsJobRunning(fileName, jobName string) bool {
	df, err := LoadCsv(fileName)
	if err != nil {
		return false
	}

	_, primaryName, err := GetPrimaryName(df)
	if err != nil {
		return false
	}

	status := getRecord(df, primaryName, jobName, "status")

	return status == "running"

}

func getRecord(df dataframe.DataFrame, header1, value1, header2 string) (value2 string) {
	record1 := df.Col(header1).Records()
	record2 := df.Col(header2).Records()
	if len(record1) != len(record2) {
		return value2
	}

	var index int
	var matchFlag bool
	for i, s := range record1 {
		if s == value1 {
			index = i
			matchFlag = true
			break
		}
	}
	if matchFlag {
		return record2[index]
	}

	return value2
}

func GetRecord(fileName, header1, value1, header2 string) string {
	df, err := LoadCsv(fileName)
	if err != nil {
		return ""
	}

	return getRecord(df, header1, value1, header2)
}

func GetAllRecords(fileName string) ([][]string, error) {
	df, err := LoadCsv(fileName)
	if err != nil {
		return nil, err
	}

	return df.Records(), nil
}

func GetOneRecord(fileName, matched, primary string) ([]string, error) {
	df, err := LoadCsv(fileName)
	if err != nil {
		return nil, err
	}

	records := df.Records()
	if len(records) < 2 {
		return nil, NotExist
	}

	rows := df.Col(primary).Records()
	for r, record := range rows {
		if record == matched {
			return records[r+1], nil
		}
	}

	return nil, NotExist
}

