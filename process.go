package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	cdcURL       = "https://www.cdc.gov/coronavirus/2019-ncov/json/county-map-data.json"
	githubURL    = "https://raw.githubusercontent.com/balsama/us_counties_data/master/data/counties.json"
	cdcStatesURL = "https://www.cdc.gov/covid-data-tracker/Content/CoronaViewJson_01/US_MAP_DATA.json"

	cdcLocalFile       = "./data/CDC.json"
	githubLocalFile    = "./data/Github.json"
	resultJSONFileName = "./data/risk.json"
	resultCSVFileName  = "./data/risk.csv"
)

type county struct {
	Name               string
	State              string
	StateCode          string
	Fips               int
	Population         int
	Area               int
	Density            int
	Cases              int
	Deaths             int
	PercentOfState     float64
	CasesByPopulation  int
	CasesByArea        int
	DeathsByPopulation int
	DeathsByArea       int
	RiskIndex          float64
}

type cdc struct {
	Data []countyCDC `json:"data"`
}

type countyCDC struct {
	Name    string `json:"county_name"`
	State   string `json:"state"`
	Fips    int    `json:"fips"`
	Cases   string `json:"cases"`
	Deaths  string `json:"deaths"`
	Percent string `json:"cases_percent"`
	Rate    string `json:"rate_per_100k"`
}

type countyGithub struct {
	Name       string `json:"name"`
	State      string `json:"state"`
	Fips       string `json:"fips"`
	Population int    `json:"population"`
	Area       int    `json:"area"`
	Density    int    `json:"density"`
}

func getCDCFile(counties *cdc, url, fileName string) {
	var body []byte

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		if body, err = ioutil.ReadFile(fileName); err != nil {
			fmt.Println(err)
		}
	} else {
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			fmt.Println(err)
		}
		body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	}

	if err = json.Unmarshal(body, &counties); err != nil {
		fmt.Println(err)
	}

	if err = ioutil.WriteFile(fileName, body, 0644); err != nil {
		fmt.Println(err)
	}
}

func getGithubFile(counties *map[string]countyGithub, url, fileName string) {
	var body []byte

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		if body, err = ioutil.ReadFile(fileName); err != nil {
			fmt.Println(err)
		}
	} else {
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			fmt.Println(err)
		}
		body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	}

	if err = json.Unmarshal(body, &counties); err != nil {
		fmt.Println(err)
	}

	if err = ioutil.WriteFile(fileName, body, 0644); err != nil {
		fmt.Println(err)
	}
}

func main() {
	var (
		CDCCounties             cdc
		GithubCounties          map[string]countyGithub
		result                  []county
		completeCounty          county
		tempFips, cases, deaths int
		area, percent           float64
		resultCsv               []string
	)

	getCDCFile(&CDCCounties, cdcURL, cdcLocalFile)
	getGithubFile(&GithubCounties, githubURL, githubLocalFile)

	// Mix results
	for _, g := range GithubCounties {
		for _, c := range CDCCounties.Data {
			tempFips, _ = strconv.Atoi(g.Fips)
			if tempFips == c.Fips {
				completeCounty.Name = strings.ReplaceAll(c.Name, " County", "")
				completeCounty.State = g.State
				completeCounty.StateCode = c.State
				completeCounty.Fips = c.Fips
				completeCounty.Population = g.Population
				area = float64(g.Area) * 2.59
				completeCounty.Area = int(area)
				completeCounty.Density = g.Population / completeCounty.Area
				if c.Cases != "<20" {
					cases, _ = strconv.Atoi(c.Cases)
				} else {
					cases = 10
				}
				completeCounty.Cases = cases

				if c.Deaths != "<20" {
					deaths, _ = strconv.Atoi(c.Deaths)
				} else {
					deaths = 10
				}
				completeCounty.Deaths = deaths
				if c.Percent != "Not Calculated" {
					percent, _ = strconv.ParseFloat(strings.ReplaceAll(c.Percent, " %", ""), 64)
				} else {
					percent = 0.0
				}
				completeCounty.PercentOfState = percent

				completeCounty.CasesByPopulation = completeCounty.Cases * 100000 / completeCounty.Population
				completeCounty.CasesByArea = completeCounty.Cases * 1000 / completeCounty.Area
				completeCounty.DeathsByPopulation = completeCounty.Deaths * 100000 / completeCounty.Population
				completeCounty.DeathsByArea = completeCounty.Deaths * 1000 / completeCounty.Area
				completeCounty.RiskIndex = math.Round(float64((completeCounty.DeathsByArea+1)*(completeCounty.CasesByPopulation+1))*100) / 10000
				result = append(result, completeCounty)
			}
		}
	}

	jsonResult, err := json.MarshalIndent(result, "", "\t")
	if err != nil {
		fmt.Println(err)
	}

	if err := ioutil.WriteFile(resultJSONFileName, jsonResult, 0644); err != nil {
		fmt.Println(err)
	}

	csvFile, err := os.Create(resultCSVFileName)
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)

	resultCsv = append(resultCsv, "Name")
	resultCsv = append(resultCsv, "State")
	resultCsv = append(resultCsv, "State Code")
	resultCsv = append(resultCsv, "Fips")
	resultCsv = append(resultCsv, "Population")
	resultCsv = append(resultCsv, "Area")
	resultCsv = append(resultCsv, "Density")
	resultCsv = append(resultCsv, "Cases")
	resultCsv = append(resultCsv, "Deaths")
	resultCsv = append(resultCsv, "Percent Of State")
	resultCsv = append(resultCsv, "Cases By Population")
	resultCsv = append(resultCsv, "Cases By Area")
	resultCsv = append(resultCsv, "Deaths By Population")
	resultCsv = append(resultCsv, "Deaths By Area")
	resultCsv = append(resultCsv, "Risk Index")

	_ = csvWriter.Write(resultCsv)

	for j := 0; j < len(result); j++ {
		values := reflect.ValueOf(result[j])
		resultCsv = nil
		for i := 0; i < values.NumField(); i++ {
			value := values.Field(i)
			switch value.Kind() {
			case reflect.String:
				resultCsv = append(resultCsv, value.String())
			case reflect.Int:
				resultCsv = append(resultCsv, fmt.Sprintf("%d", value.Int()))
			case reflect.Float64:
				resultCsv = append(resultCsv, fmt.Sprintf("%.2f", value.Float()))
			}
		}
		fmt.Println(resultCsv)
		_ = csvWriter.Write(resultCsv)
	}

	csvWriter.Flush()

}
