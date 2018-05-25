package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

type Stations map[string]Station

type Station struct {
	StationName string `json:"stationName"`
	Direction1 direction `json:"direction1"`
	Direction2 direction `json:"direction2"`
	WalkingDistanceMinutes int
}

type Direction struct {
	Name string `json:"name"`
	Times departures `json:"times"`
}

type Departures []departure

type Departure struct {
	Route       string `json:"route"`
	LastStation string `json:"lastStation"`
	Minutes     int    `json:"minutes"`
}

func (s *Stations) Update(stationCode string, station Station) {
	s.Stations[stationCode] = Station
}

var (
	stationCodes = [7]string{"1/142", "R/R27", "2/230", "4/420", "J/M23"}
	updates      = make(chan update)
	display      = make(chan bool)
	handler      = make(chan departures)
)

func getDepartures(stationCode string) error {
	for {
		getTime := Station{}

		if resp, err := http.Get("https://mtasubwaytime.info/getTime/" + stationCode); err == nil {
			defer resp.Body.Close()
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				if err := json.Unmarshal(body, &getTime); err == nil {
					departures := departures{}

					for _, departure := range append(getTime.Direction1.Times, getTime.Direction2.Times...) {
						if getTime.StationName != departure.LastStation {
							departure.StationName = getTime.StationName
							departures = append(departures, departure)
						}
					}

					updates <- update{stationCode, departures}
				}
			}
		}

		time.Sleep(15 * time.Second)
	}
}

func main() {
	go func() {
		for _, stationCode := range stationCodes {
			go getDepartures(stationCode)
			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		stations := make(Stations)
		for {
			select {
			case <-display:
				handler <- index.Departures
			case update := <-updates:
				index.Update(update)
			}
		}
	}()

	t := template.Must(template.New("/").Parse(`{{range .}}
    <tr>
			<td><img src='http://subwaytime.mta.info/img/{{.Route}}_sm.png'></td>
			<td>{{.StationName}}</td>
			<td>{{.LastStation}}</td>
			<td align='right'>{{.Minutes}}</td>
		</tr>{{end}}`))

	http.HandleFunc("/departures", func(w http.ResponseWriter, r *http.Request) {
		display <- true
		t.Execute(w, <-handler)
	})
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.ListenAndServe(":8080", nil)
}
