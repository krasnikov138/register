package app

import (
	"math/rand"
	"sort"
	"time"
)

const timeLayout = "15:04:05"

func Choose[T any](items []T) T {
	return items[rand.Intn((len(items)))]
}

type Record struct {
	Started  *time.Time
	Finished *time.Time
	Duration *time.Duration
	Comment  string
}

// you can define your own generator using provided interface
type Generator interface {
	Gen(date time.Time) Record
}

type FixedWorkdayGenerator struct {
	Vacations       []time.Time
	StartedChoices  []time.Time
	WorkdayDuration time.Duration
}

func (gen *FixedWorkdayGenerator) Gen(date time.Time) Record {
	_, isVac := sort.Find(
		len(gen.Vacations),
		func(j int) int { return date.Compare(gen.Vacations[j]) },
	)

	if isVac {
		return Record{
			Started:  nil,
			Finished: nil,
			Duration: nil,
			Comment:  "vacationing",
		}
	}

	started := Choose(gen.StartedChoices)
	finished := started.Add(gen.WorkdayDuration)
	return Record{
		Started:  &started,
		Finished: &finished,
		Duration: &gen.WorkdayDuration,
		Comment:  "",
	}
}

func renderDate(d time.Time) float64 {
	return d.Sub(time.Date(1899, 12, 30, 0, 0, 0, 0, d.Location())).Hours() / 24.0
}

func durationFromHMS(h int, m int, s int) time.Duration {
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
}

func renderTime(t *time.Time) interface{} {
	if t == nil {
		return ""
	}
	return durationFromHMS(t.Clock()).Hours() / 24.0
}

func renderDuration(d *time.Duration) interface{} {
	if d == nil {
		return ""
	}
	return d.Hours() / 24.0
}

type ColumnNames struct {
	Date, Started, Finished, Duration, Comment, Month string
}

func GenerateTable(
	dates []time.Time,
	gen Generator,
	columns []string,
	mapping ColumnNames,
) *Table[interface{}] {
	result := NewTableCols[interface{}](len(dates), columns)

	datecol := result.GetColumn(mapping.Date)
	comments := result.GetColumn(mapping.Comment)
	started := result.GetColumn(mapping.Started)
	finished := result.GetColumn(mapping.Finished)
	durations := result.GetColumn(mapping.Duration)
	months := result.GetColumn(mapping.Month)

	for i, date := range dates {
		record := gen.Gen(date)

		datecol[i] = renderDate(date)
		started[i] = renderTime(record.Started)
		finished[i] = renderTime(record.Finished)
		durations[i] = renderDuration(record.Duration)
		comments[i] = record.Comment
		months[i] = date.Month().String()
	}

	return result
}
