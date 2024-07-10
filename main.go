package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

type ProverSpeedResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    float64 `json:"data,string"`
}

func getProverSpeed(address string, startTime int64, endTime int64) (float64, error) {
	url := "http://localhost:8088/api/v1/provers/prover_speed_address"
	requestBody := fmt.Sprintf(`{"address":"%s","start_time":%d,"end_time":%d}`, address, startTime, endTime)

	resp, err := http.Post(url, "application/json", strings.NewReader(requestBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response ProverSpeedResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, err
	}

	if response.Code != 0 {
		return 0, fmt.Errorf("API error: %s", response.Message)
	}

	return response.Data, nil
}

func main() {
	address := flag.String("address", "", "The address to query")
	startDateStr := flag.String("start", "", "Start date (format: 2006-01-02 15:04:05)")
	endDateStr := flag.String("end", "", "End date (format: 2006-01-02 15:04:05)")
	flag.Parse()

	if *address == "" || *startDateStr == "" || *endDateStr == "" {
		fmt.Println("Address, start date, and end date must be provided")
		flag.Usage()
		return
	}

	location, _ := time.LoadLocation("Asia/Shanghai")
	startDate, err := time.ParseInLocation("2006-01-02 15:04:05", *startDateStr, location)
	if err != nil {
		fmt.Printf("Invalid start date: %v\n", err)
		return
	}

	endDate, err := time.ParseInLocation("2006-01-02 15:04:05", *endDateStr, location)
	if err != nil {
		fmt.Printf("Invalid end date: %v\n", err)
		return
	}

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf("Failed to add sheet: %v\n", err)
		return
	}

	headers := []string{"start_time", "start_time_BJ", "end_time", "end_time_BJ", "power"}
	row := sheet.AddRow()
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
	}

	for current := startDate; current.Before(endDate); current = current.Add(time.Hour) {
		startTime := current.Unix()
		endTime := current.Add(time.Hour).Unix()

		power, err := getProverSpeed(*address, startTime, endTime)
		if err != nil {
			fmt.Printf("Error fetching prover speed: %v\n", err)
			continue
		}

		startTimeStr := strconv.FormatInt(startTime, 10)
		endTimeStr := strconv.FormatInt(endTime, 10)
		startTimeBJ := current.Format("2006-01-02 15:04:05")
		endTimeBJ := current.Add(time.Hour).Format("2006-01-02 15:04:05")

		row := sheet.AddRow()
		row.AddCell().Value = startTimeStr
		row.AddCell().Value = startTimeBJ
		row.AddCell().Value = endTimeStr
		row.AddCell().Value = endTimeBJ
		row.AddCell().Value = fmt.Sprintf("%f", power)
	}

	// Save the spreadsheet
	if err := file.Save("ProverPowerCalculator.xlsx"); err != nil {
		fmt.Println(err)
	}
}
