package app

import (
	"math/rand"
	"sort"
	"time"
)

const timeLayout = "15:04:05"

func Generate[T any](length int, choices []T) []T {
	result := make([]T, length)
	for i := range result {
		result[i] = choices[rand.Intn(len(choices))]
	}
	return result
}

func Transform[In any, Out any](values []In, transformer func(In) Out) []Out {
	result := make([]Out, len(values))

	for i, el := range values {
		result[i] = transformer(el)
	}

	return result
}

func date2Number(d time.Time) float64 {
	return d.Sub(time.Date(1899, 12, 30, 0, 0, 0, 0, d.Location())).Hours() / 24.0
}

func durationFromHMS(h int, m int, s int) time.Duration {
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
}

func time2Number(t time.Time) float64 {
	return durationFromHMS(t.Clock()).Hours() / 24.0
}

type ColumnNames struct {
	Date, Started, Finished, Duration, Comment, Month string
}

func GenerateTable(
	dates []time.Time,
	vacations []time.Time,
	startedChoices []time.Time,
	workdayDuration time.Duration,
	columns []string,
	mapping ColumnNames,
) *Table[interface{}] {
	startedHours := Generate(len(dates), startedChoices)

	result := NewTableCols[interface{}](len(dates), columns)

	datecol := result.GetColumn(mapping.Date)
	comments := result.GetColumn(mapping.Comment)
	started := result.GetColumn(mapping.Started)
	finished := result.GetColumn(mapping.Finished)
	durations := result.GetColumn(mapping.Duration)
	months := result.GetColumn(mapping.Month)

	for i, date := range dates {
		datecol[i] = date2Number(date)
		months[i] = date.Month().String()

		_, isVac := sort.Find(len(vacations), func(j int) int { return date.Compare(vacations[j]) })
		if isVac {
			comments[i] = "vacationing"
			started[i] = ""
			finished[i] = ""
			durations[i] = ""
		} else {
			comments[i] = ""
			started[i] = time2Number(startedHours[i])
			finished[i] = time2Number(startedHours[i].Add(workdayDuration))
			durations[i] = workdayDuration.Hours() / 24.0
		}
	}

	return result
}
