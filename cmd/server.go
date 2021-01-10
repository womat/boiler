package main

import (
	"time"

	"github.com/womat/debug"

	"boiler/pkg/boiler"
)

func (r *solarPumpRuntime) serveSave(f string, p time.Duration) {
	for range time.Tick(p) {
		_ = r.current.save(f)
	}
}

func (r *solarPumpRuntime) serveCalc(p time.Duration) {
	runtime := func(state boiler.State, lastStateDate *time.Time, lastState *boiler.State) (runTime float64) {
		if state != boiler.Off {
			if *lastState == boiler.Off {
				*lastStateDate = time.Now()
			}
			runTime = time.Since(*lastStateDate).Hours()
			*lastStateDate = time.Now()
		}
		*lastState = state
		return
	}

	ticker := time.NewTicker(p)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		debug.DebugLog.Println("get data")

		v, err := r.handler.Read()
		if err != nil {
			debug.ErrorLog.Printf("get solar data: %v", err)
			continue
		}

		debug.DebugLog.Println("calc runtime")

		func() {
			r.current.Lock()
			defer r.current.Unlock()

			r.current.Timestamp = v.Timestamp
			r.current.State = v.State

			r.current.WaterTempOut = v.WaterTempOut
			r.current.WaterTempHeatPumpRegister = v.WaterTempHeatPumpRegister
			r.current.WaterTempHeatingRodRegister = v.WaterTempHeatingRodRegister

			r.current.SolarPumpState = v.SolarPumpState
			r.current.SolarCollectorTemp = v.SolarCollectorTemp
			r.current.SolarFlow = v.SolarFlow
			r.current.SolarReturn = v.SolarReturn

			r.current.HeatPumpState = v.HeatPumpState
			r.current.HeatPumpFlow = v.SolarFlow
			r.current.HeatPumpReturn = v.SolarReturn

			r.current.HeatingRodState = v.HeatingRodState
			r.current.HeatingRodPower = v.HeatingRodPower
			r.current.HeatingRodEnergy = v.HeatingRodEnergy

			r.current.CirculationPumpState = v.CirculationPumpState
			r.current.CirculationFlow = v.CirculationFlow
			r.current.CirculationReturn = v.CirculationReturn

			r.current.Runtime += runtime(r.current.State, &r.last.stateDate, &r.last.state)
			r.current.SolarPumpRuntime += runtime(r.current.SolarPumpState, &r.last.solarPumpStateDate, &r.last.solarPumpState)
			r.current.HeatPumpRuntime += runtime(r.current.HeatPumpState, &r.last.heatPumpStateDate, &r.last.heatPumpState)
			r.current.HeatingRodRuntime += runtime(r.current.HeatingRodState, &r.last.heatingRodStateDate, &r.last.heatingRodState)
			r.current.CirculationPumpRuntime += runtime(r.current.CirculationPumpState, &r.last.circulationPumpStateDate, &r.last.circulationPumpState)
		}()
	}
}
