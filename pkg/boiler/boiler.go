package boiler

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/womat/debug"
)

const httpRequestTimeout = 10 * time.Second

const (
	On           State = "on"
	Off          State = "off"
	HeatingUp    State = "heating up with "
	HeatRecovery State = "heat recovery"

	HeatingSolar    = "Solar"
	HeatingHeatPump = "Heatpump"
	HeatingRod      = "Rod"

	ThresholdHeatingRod = 250
)

type State string

type Values struct {
	Timestamp time.Time
	Runtime   float64
	State     State

	WaterTempOut                float64
	WaterTempHeatPumpRegister   float64
	WaterTempHeatingRodRegister float64

	SolarPumpState         State
	SolarPumpRuntime       float64
	SolarCollectorTemp     float64
	SolarFlow, SolarReturn float64

	HeatPumpState                State
	HeatPumpRuntime              float64
	HeatPumpFlow, HeatPumpReturn float64

	HeatingRodState     State
	HeatingRodRuntime   float64
	HeatingRodWaterTemp float64
	HeatingRodPower     float64
	HeatingRodEnergy    float64

	CirculationPumpState               State
	CirculationPumpRuntime             float64
	CirculationFlow, CirculationReturn float64
}

type Measurements struct {
	uvs232URL string
	meterURL  string
}

type uvs232URLBody struct {
	Timestamp time.Time `json:"Timestamp"`
	Runtime   float64   `json:"Runtime"`
	Measurand struct {
		Temperature1, Temperature2, Temperature3, Temperature4 float64
		Out1, Out2                                             bool
		RotationSpeed                                          float64
	} `json:"Data"`
}

type meterURLBody struct {
	Timestamp time.Time `json:"Time"`
	Runtime   float64   `json:"Runtime"`
	Measurand struct {
		E float64 `json:"e"`
		P float64 `json:"p"`
	} `json:"Measurand"`
}

func New() *Measurements {
	return &Measurements{}
}

func (m *Measurements) SetUVS232URL(url string) {
	m.uvs232URL = url
}

func (m *Measurements) SetMeterURL(url string) {
	m.meterURL = url
}

func (m *Measurements) Read() (v Values, err error) {
	var wg sync.WaitGroup
	start := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := m.readUVS232(&v); e != nil {
			err = e
		}

		debug.TraceLog.Printf("runtime to request UVS232 data: %vs", time.Since(start).Seconds())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := m.readMeter(&v); e != nil {
			err = e
		}

		debug.TraceLog.Printf("runtime to request meter data: %vs", time.Since(start).Seconds())
	}()

	wg.Wait()
	if err != nil {
		return
	}

	debug.DebugLog.Printf("runtime to request data: %vs", time.Since(start).Seconds())

	v.HeatPumpState = Off
	v.CirculationPumpState = Off

	if v.State != HeatRecovery {
		var s []string
		if v.SolarPumpState == On {
			s = append(s, HeatingSolar)
		}
		if v.HeatPumpState == On {
			s = append(s, HeatingHeatPump)
		}
		if v.HeatingRodState == On {
			s = append(s, HeatingRod)
		}
		if len(s) > 0 {
			v.State = HeatingUp + State(strings.Join(s, ","))
		} else {
			v.State = Off
		}
	}

	return
}

func (m *Measurements) readMeter(v *Values) (err error) {
	var r meterURLBody

	if err = read(m.meterURL, &r); err != nil {
		return
	}

	v.HeatingRodPower = r.Measurand.P
	v.HeatingRodEnergy = r.Measurand.E

	if r.Measurand.P > ThresholdHeatingRod {
		v.HeatingRodState = On
	} else {
		v.HeatingRodState = Off
	}

	return
}

func (m *Measurements) readUVS232(v *Values) (err error) {
	var r uvs232URLBody

	if err = read(m.uvs232URL, &r); err != nil {
		return
	}

	v.Timestamp = r.Timestamp
	v.SolarCollectorTemp = r.Measurand.Temperature1
	v.WaterTempHeatPumpRegister = r.Measurand.Temperature2

	if r.Measurand.Out1 {
		v.SolarPumpState = On
		if r.Measurand.Out2 {
			v.State = HeatRecovery
		}
	} else {
		v.SolarPumpState = Off
	}

	return
}

func read(url string, data interface{}) (err error) {
	done := make(chan bool, 1)
	go func() {
		// ensures that data is sent to the channel when the function is terminated
		defer func() {
			select {
			case done <- true:
			default:
			}
			close(done)
		}()

		debug.TraceLog.Printf("performing http get: %v\n", url)

		var res *http.Response
		if res, err = http.Get(url); err != nil {
			return
		}

		bodyBytes, _ := ioutil.ReadAll(res.Body)
		_ = res.Body.Close()

		if err = json.Unmarshal(bodyBytes, data); err != nil {
			return
		}
	}()

	// wait for API Data
	select {
	case <-done:
	case <-time.After(httpRequestTimeout):
		err = errors.New("timeout during receive data")
	}

	if err != nil {
		debug.ErrorLog.Println(err)
		return
	}
	return
}
