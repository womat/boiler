package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"

	"github.com/womat/debug"
	"github.com/womat/tools"
)

type yamlData struct {
	Timestamp              time.Time `yaml:"Timestamp"`
	Runtime                float64   `yaml:"Runtime"`
	SolarPumpRuntime       float64   `yaml:"SolarPumpRuntime"`
	HeatPumpRuntime        float64   `yaml:"HeatPumpRuntime"`
	CirculationPumpRuntime float64   `yaml:"CirculationPumpRuntime"`
	HeatingRodRuntime      float64   `yaml:"HeatingRodRuntime"`
}

func (d *CurrentMeasurements) load(fileName string) (err error) {
	// if file doesn't exists, create an empty file
	if !tools.FileExists(fileName) {
		s := yamlData{}

		// marshal the byte slice which contains the yaml file's content into SaveMeters struct
		var data []byte
		data, err = yaml.Marshal(&s)
		if err != nil {
			return
		}

		if err = ioutil.WriteFile(fileName, data, 0600); err != nil {
			return
		}
	}

	// read the yaml file as a byte array.
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	// unmarshal the byte slice which contains the yaml file's content into SaveMeters struct
	s := yamlData{}
	if err = yaml.Unmarshal(data, &s); err != nil {
		return
	}

	func() {
		d.Lock()
		defer d.Unlock()
		d.Timestamp = s.Timestamp
		d.Runtime = s.Runtime
		d.SolarPumpRuntime = s.SolarPumpRuntime
		d.HeatPumpRuntime = s.HeatPumpRuntime
		d.CirculationPumpRuntime = s.CirculationPumpRuntime
		d.HeatingRodRuntime = s.HeatingRodRuntime
	}()

	return
}

func (d *CurrentMeasurements) save(fileName string) error {
	debug.DebugLog.Println("saveMeasurements measurements to file")

	s := yamlData{
		Timestamp:              d.Timestamp,
		Runtime:                d.Runtime,
		SolarPumpRuntime:       d.SolarPumpRuntime,
		HeatPumpRuntime:        d.HeatPumpRuntime,
		CirculationPumpRuntime: d.CirculationPumpRuntime,
		HeatingRodRuntime:      d.HeatingRodRuntime,
	}

	// marshal the byte slice which contains the yaml file's content into SaveMeters struct
	data, err := yaml.Marshal(&s)
	if err != nil {
		debug.ErrorLog.Printf("backupMeasurements marshal: %v\n", err)
		return err
	}

	if err := ioutil.WriteFile(fileName, data, 0600); err != nil {
		debug.ErrorLog.Printf("backupMeasurements write file: %v\n", err)
		return err
	}

	return nil
}
