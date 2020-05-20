package extensions

import (
	"log"
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

const (
	PRIVATE_BYTES   = "\\Process(??APP_WIN32_PROC??)\\Private Bytes"
	AVAILABLE_BYTES = "\\Memory\\Available Bytes"
	PROCESSOR_TIME  = "\\Processor(_Total)\\% Processor Time"
	PROCESS_TIME    = "\\Process(??APP_WIN32_PROC??)\\% Processor Time"
)

var quitChan chan struct{}

func trackProcessorCpu() {
	times, err := cpu.Percent(0, false)
	var percent float64
	if err != nil {
		log.Printf("Failed to get CPU usage %s\n", err)
	} else {
		percent = times[0]
		if GetCurrentClient() != nil {
			GetCurrentClient().TrackMetric(PROCESSOR_TIME, percent)
		} else {
			log.Printf("Percentage of Processor Time %.2f\n", percent)
		}
	}
}

func trackProcessCpu(curProcess *process.Process) {
	percent, err := curProcess.CPUPercent()
	if err == nil {
		if GetCurrentClient() != nil {
			GetCurrentClient().TrackMetric(PROCESS_TIME, percent)
		} else {
			log.Printf("Percentage of Process(%d) Time %.2f\n", curProcess.Pid, percent)
		}
	}
}

func trackProcessMemory(curProcess *process.Process) {
	memory, err := curProcess.MemoryInfo()
	if err == nil {
		if GetCurrentClient() != nil {
			GetCurrentClient().TrackMetric(PRIVATE_BYTES, float64(memory.RSS))
		} else {
			log.Printf("Process(%d) RSS %d\n", curProcess.Pid, memory.RSS)
		}
	}
}

func trackServerMemory() {
	memory, err := mem.VirtualMemory()
	if err == nil {
		if GetCurrentClient() != nil {
			GetCurrentClient().TrackMetric(AVAILABLE_BYTES, float64(memory.Available))
		} else {
			log.Printf("Available memory %d\n", memory.Available)
		}
	}
}

func StartCollectPerformance(seconds time.Duration) {
	quitChan = make(chan struct{})
	ticker := time.NewTicker(seconds * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				trackProcessorCpu()
				trackServerMemory()
				pid := os.Getpid()
				curProcess, err := process.NewProcess(int32(pid))
				if err == nil {
					trackProcessCpu(curProcess)
					trackProcessMemory(curProcess)
				}

			case <-quitChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func StopCollectPerformance() {
	if quitChan != nil {
		close(quitChan)
	}
}
