package mr

//
// a word-count application "plugin" for MapReduce.
//
// go build -buildmode=plugin wc.go
//

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/luc/tdfs"
)

//
// The map function is called once for each file of input. The first
// argument is the name of the input file, and the second is the
// file's complete contents. You should ignore the input file name,
// and look only at the contents argument. The return value is a slice
// of key/value pairs.
//
func Map(filename string, contents string) []KeyValue {
	return Task3Map(filename, contents)
}

//
// The reduce function is called once for each key generated by the
// map tasks, with a list of all the values created for that key by
// any map task.
//
func Reduce(key string, values []string) string {
	return Task3Reduce(key, values)
}

func Partition(key string, totalPartition int) int {
	return int(tdfs.GetHashInt([]byte(key)) % uint32(totalPartition))
}

// function to detect word separators.
func ff(r rune) bool { return !unicode.IsPrint(r) }

func wordCountMap(filename string, contents string) []KeyValue {

	// split contents into an array of words.
	words := strings.FieldsFunc(contents, ff)

	kva := []KeyValue{}
	for _, w := range words {
		kv := KeyValue{w, "1"}
		kva = append(kva, kv)
	}
	return kva
}

func wordCountReduce(key string, values []string) string {
	// return the number of occurrences of this word.
	return strconv.Itoa(len(values))
}

func Task1Map(filename string, contents string) []KeyValue {
	reader := bytes.NewBuffer([]byte(contents))

	kva := []KeyValue{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			MyPanic("XX task1Map fail at read file ", err)
		}

		fields := strings.FieldsFunc(line, ff)
		if len(fields) == 14 {
			kva = append(kva, KeyValue{fields[1], fields[0]})
		}
	}
	return kva
}

func Task1Reduce(key string, values []string) string {
	callNum := len(values)

	// values are an array of duplicate date values
	dateNum := 0
	keys := make(map[string]bool) // use a map to mock a Set
	for _, date := range values {
		if _, ok := keys[date]; !ok {
			keys[date] = true
			dateNum += 1
		}
	}

	return fmt.Sprintf("%.3f", float32(callNum)/float32(dateNum))
}

func Task2Map(filename string, contents string) []KeyValue {
	reader := bytes.NewBuffer([]byte(contents))
	kva := []KeyValue{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			MyPanic("XX task2Map fail at read file ", err)
		}

		fields := strings.FieldsFunc(line, ff)
		if len(fields) == 14 {
			kva = append(kva, KeyValue{fields[12], fields[4]})
		}
	}
	return kva

}

func Task2Reduce(key string, values []string) string {
	total := float32(len(values))

	optrArr := [4]float32{}

	for _, optr := range values {
		switch optr {
		case "1":
			optrArr[1] += 1
		case "2":
			optrArr[2] += 1
		case "3":
			optrArr[3] += 1
		default:
			optrArr[0] += 1
		}
	}

	return fmt.Sprintf("%.3f %.3f %.3f %.3f", optrArr[0]/total, optrArr[1]/total, optrArr[2]/total, optrArr[3]/total)
}

func Task3Map(filename string, contents string) []KeyValue {
	reader := bytes.NewBuffer([]byte(contents))
	kva := []KeyValue{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			MyPanic("XX task3Map fail at read file ", err)
		}

		fields := strings.FieldsFunc(line, ff)
		if len(fields) == 14 {
			kva = append(kva, KeyValue{fields[1], fields[9] + "-" + fields[11]})
		}
	}
	return kva
}

func Task3Reduce(key string, values []string) string {
	var spanTime [8]float32

	for _, val := range values {
		parts := strings.Split(val, "-")
		startTime, timeSpan := parts[0], parts[1]
		// handle error time format
		if startTime == "00:00:00" {
			continue
		}
		startHour, _, _ := parseTime(startTime)
		timeSpanNum, _ := strconv.ParseFloat(timeSpan, 32)
		startIndex := startHour / 3
		spanTime[startIndex] += float32(timeSpanNum)
	}

	var total float32
	for _, time := range spanTime {
		total += time
	}

	return fmt.Sprintf("%.3f %.3f %.3f %.3f %.3f %.3f %.3f %.3f", spanTime[0]/total, spanTime[1]/total, spanTime[2]/total, spanTime[3]/total, spanTime[4]/total, spanTime[5]/total, spanTime[6]/total, spanTime[7]/total)
}

func parseTime(timeString string) (hourNum int, minuteNum int, secondNum int) {
	parts := strings.Split(timeString, ":")
	hour, minute, second := parts[0], parts[1], parts[2]
	hourNum, _ = strconv.Atoi(hour)
	minuteNum, _ = strconv.Atoi(minute)
	secondNum, _ = strconv.Atoi(second)
	return hourNum, minuteNum, secondNum
}
