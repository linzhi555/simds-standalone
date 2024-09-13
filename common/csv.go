package common

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

// CountLines 函数用于返回指定文件的行数
func CountLines(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lineCount, nil
}

func RemoveIfExisted(file string) {
	if IsFileExist(file) {
		err := os.Remove(file)
		if err != nil {
			panic(err)
		}
	}
}

func IsFileExist(filename string) bool {
	var exist = true
	if res, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	} else if res.IsDir() {
		exist = false
	}
	return exist
}

func AppendLineCsvFile(path string, line []string) error {

	File, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("can't open file")
	}
	defer File.Close()

	WriterCsv := csv.NewWriter(File)
	var addContent [][]string
	addContent = append(addContent, line)

	// write one line , append mode
	err = WriterCsv.WriteAll(addContent)
	if err != nil {
		return err
	}
	WriterCsv.Flush() // flush the file
	return nil
}

// maybe return patial csv read error
func CsvToList(path string) (table [][]string, tabletop []string, err error) {
	fs, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer fs.Close()
	return BytesCsvToList(fs)
}

func IterateCsv(input io.Reader, f1 func(tabletop []string), f2 func(row []string)) error {
	r := csv.NewReader(input)

	errorLines := []int{}
	allnum := 0
	var tabletop []string

	for i := 0; ; i++ {
		row, err := r.Read()
		allnum++
		if err != nil && err != io.EOF {
			errorLines = append(errorLines, i)
			continue
		}
		if err == io.EOF {
			break
		}
		if i == 0 {
			tabletop = row
			if f1 != nil {
				f1(tabletop)
			}
		} else {
			if f2 != nil {
				f2(row)
			}
		}
	}

	if len(errorLines) > 0 {
		if tabletop == nil {
			return errors.New("read csv error")
		}

		errinfo := fmt.Sprintf("partial error: error / all = %d / %d   \n", len(errorLines), allnum)
		for i := 0; i < 100 && i < len(errorLines); i++ {
			errinfo += fmt.Sprintf("%d th error line: %d\n", i, errorLines[i])
		}
		return errors.New(errinfo)
	} else {
		return nil
	}
}

func BytesCsvToList(input io.Reader) (table [][]string, tabletop []string, err error) {
	table = make([][]string, 0)

	err = IterateCsv(input,
		func(t []string) {
			tabletop = t
		},
		func(row []string) {
			table = append(table, row)
		},
	)

	return table, tabletop, err
}

func ListToCsv(table [][]string, tabletop []string, outpath string) {
	//将表头加入到slice前

	temp := make([][]string, 0)
	temp = append(temp, tabletop)
	temp = append(temp, table...)

	f, err := os.Create(outpath)

	if err != nil {
		log.Fatal(err)
	}

	w := csv.NewWriter(f)

	err = w.WriteAll(temp)
	if err != nil {
		panic(err)
	}
	w.Flush()

}

func Float64_to_str(f float64) string {
	return strconv.FormatFloat(f, 'f', 5, 64)
}

func Str_to_float64(s string) float64 {

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Println("string to float panic")
		panic(err)

	}
	return f
}

func IsLegalFloat64(s string) bool {

	_, err := strconv.ParseFloat(s, 64)

	return err == nil
}

func Int64_to_str(i int64) string {

	return strconv.FormatInt(i, 10)
}

func Str_to_int64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)

	if err != nil {

		log.Println("string to int panic")
		panic(err)
	}
	return i
}

func IsLegalInt64(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)

	return err == nil
}
