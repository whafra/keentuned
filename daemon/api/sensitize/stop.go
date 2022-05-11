package sensitize

import (
//	m "keentune/daemon/modules"
	"os"
	"fmt"
        "keentune/daemon/common/log"
	"encoding/csv"
	"io"
	"bufio"
	"strings"
)

// Stop run sensitize stop service
func (s *Service) Stop(request string, reply *string) error {
	fs, err := os.OpenFile("/var/keentune/sensitize_jobs.csv", os.O_RDWR|os.O_CREATE, 0666)
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
		StrReplace("running", "abort")
		m.StopSig <- os.Interrupt
	} else {
		log.Errorf("", "No running job can stop.")
		return fmt.Errorf("No running job can stop.")
	}

	return nil
}

func StrReplace(src string, dest string) {
	out, _ := os.OpenFile("/var/keentune/sensitize_jobs.csv", os.O_RDWR|os.O_CREATE, 0766)
	defer out.Close()
	in, _ := os.Open("/var/keentune/sensitize_jobs.csv")
	defer in.Close()

	br := bufio.NewReader(in)
	for {
		line, _, err := br.ReadLine()
		if err ==io.EOF {
			break
		}
		newLine := strings.Replace(string(line), src, dest, -1)
		_, err = out.WriteString(newLine + "\n")
		if err != nil {
			log.Errorf("", "replace err.")
			os.Exit(1)
		}
	}
}
