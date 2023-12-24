package cmd

import (
	"log"
	"time"

	"github.com/krasnikov138/register/app"
	"github.com/spf13/viper"
)

func parseCmdDate(name string, defaultValue time.Time) time.Time {
	var err error

	date := viper.GetString(name)

	result := defaultValue
	if date != "" {
		result, err = time.Parse(app.DateLayout, date)
		if err != nil {
			log.Fatalf("%s date provided in wrong format: '%s'", name, date)
		}
	}
	return result
}

func checkColumns[T any](table *app.Table[T]) {
	requiredColumns := viper.GetStringSlice("columns")

	notFound := make([]string, 0)
	for _, col := range requiredColumns {
		if table.GetColumnIdx(col) == -1 {
			notFound = append(notFound, col)
		}
	}

	if len(notFound) != 0 {
		log.Fatalf("Required columns %v are not found in google sheet", notFound)
	}
}

func initConfig(configFile string) {
	viper.SetConfigName(configFile)
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Fatalln(err)
	}
}
