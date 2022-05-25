package param

import (
	m "keentune/daemon/modules"
	"os"
	"fmt"
        "keentune/daemon/common/log"
	"encoding/csv"
	"io"
	"bufio"
	"strings"
)

// Stop run param stop service
func (s *Service) Stop(request string, reply *string) error {
	filePath := "/var/keentune/tuning_jobs.csv"
	fs, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		log.Errorf("", "Can not open the file, err: %v\n",err)
		return fmt.Errorf("Can not open the file.")
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	var csvData []string
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}

		for index, value := range row {
			switch index {
			case 3:
				csvData = append(csvData, value)
			}
		}
	}

	var status string
	for _, v := range csvData {
		status += v
	}

	if strings.Contains(status, "running") {
		StrReplace("running", "abort", filePath)
		m.StopSig <- os.Interrupt
	} else {
		log.Errorf("", "No running job can stop.")
		return fmt.Errorf("No running job can stop.")
	}

	return nil
}

//StrReplace Modify csv file data
func StrReplace(src string, dest string, filePath string) {
	out, _ := os.OpenFile(filePath, os.O_RDWR, 0666)
	defer out.Close()
	in, _ := os.Open(filePath)
	defer in.Close()

	br := bufio.NewReader(in)
	for {
		line, _, err := br.ReadLine()
		if err ==io.EOF {
			break
		}
		newLine := strings.Replace(string(line), src, dest, -1)
		_, err = out.WriteString(newLine)
		if err != nil {
			log.Errorf("", "replace err.")
			os.Exit(1)
		}
	}
}
