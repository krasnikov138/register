package app

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	DateLayout = "2006-01-02"
)

type DateList []time.Time

func ReadDates(fname string) DateList {
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dates := make([]time.Time, 0, 128)

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \n")

		if len(line) == 0 {
			continue
		}

		if t, err := time.Parse(DateLayout, line); err != nil {
			log.Printf("Can not parse date line '%s' - skipped, err: %s", line, err)
		} else {
			dates = append(dates, t)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error during reading holiday file %s: %s\n", fname, err)
	}

	sort.Slice(
		dates,
		func(i int, j int) bool {
			return dates[i].Compare(dates[j]) < 0
		},
	)

	return dates
}

// generate working dates between startDate and endDate (endDate is not included)
// based on holidays array
func GenerateWorkingDays(startDate time.Time, endDate time.Time, holidays DateList) DateList {
	result := make(DateList, 0, 128)
	for date := startDate; date.Compare(endDate) < 0; date = date.AddDate(0, 0, 1) {
		if date.Weekday() == time.Sunday || date.Weekday() == time.Saturday {
			continue
		}
		_, isHoliday := sort.Find(
			len(holidays),
			func(i int) int { return holidays[i].Compare(date) },
		)
		if !isHoliday {
			result = append(result, date)
		}
	}

	return result
}

func PrintDateList(dates DateList) {
	for _, val := range dates {
		fmt.Println(val.Format(DateLayout))
	}
}

func ParseDateSlice(dates []string, layout string) ([]time.Time, error) {
	result := make([]time.Time, 0, len(dates))
	for _, date := range dates {
		parsedDate, err := time.Parse(layout, date)
		if err != nil {
			return nil, err
		}
		result = append(result, parsedDate)
	}

	return result, nil
}
