package common

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
)

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

func CsvToList(path string) (table [][]string, tabletop []string) {
	fs, err := os.Open(path)
	if err != nil {
		log.Println("can not open ", path)
		panic(err)
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	table = make([][]string, 0)
	for i := 0; ; i++ {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			panic("fail to read" + err.Error())
		}
		if err == io.EOF {
			break
		}
		if i == 0 {
			tabletop = row
		} else {
			table = append(table, row)
		}
	}

	return table, tabletop
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

	if err != nil {
		return false
	}
	return true
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

	if err != nil {
		return false
	}
	return true
}
