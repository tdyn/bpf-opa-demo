package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	"github.com/iovisor/gobpf/elf"
	"github.com/iovisor/gobpf/pkg/cpuonline"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
)

type perfEventHeader struct {
	Ts      uint64
	Tid     uint64
	Len     uint32
	Type    uint16
	Nparams uint32
}

var (
	logger = klogr.New()
)

func main() {
	klog.InitFlags(nil)

	cpuRange, err := cpuonline.Get()
	if err != nil {
		panic(err)
	}

	// 0 indexed, so adding 1
	noCPU := int(cpuRange[len(cpuRange)-1]) + 1

	module, err := load_bpf_file(noCPU, "bpf/probe.o")
	if err != nil {
		panic(err)
	}

	err = populateSettingsMap(module)
	if err != nil {
		logger.Error(err, "failed to populate settings map")
		panic(err)
	}

	err = populateSyscallTableMap(module)
	if err != nil {
		logger.Error(err, "error populating syscall table map")
		panic(err)
	}

	err = populateFillerTableMap(module)
	if err != nil {
		logger.Error(err, "error populating syscall table map")
		panic(err)
	}

	err = populateEventTableMap(module)
	if err != nil {
		logger.Error(err, "error populating event table map")
		panic(err)
	}

	perfMap, err := readFromPerfMap(module)
	if err != nil {
		logger.Error(err, "error reading from perf map")
		panic(err)
	}
	defer perfMap.PollStop()

	sig := make(chan os.Signal)
	<-sig
}

func load_bpf_file(noCPU int, filepath string) (*elf.Module, error) {
	m := elf.NewModule(filepath)
	err := m.Load(map[string]elf.SectionParams{
		"maps/perf_map": elf.SectionParams{
			MapMaxEntries: noCPU,
		},
		"maps/frame_scratch_map": elf.SectionParams{
			MapMaxEntries: noCPU,
		},
		"maps/tmp_scratch_map": elf.SectionParams{
			MapMaxEntries: noCPU,
		},
		"maps/local_state_map": elf.SectionParams{
			MapMaxEntries: noCPU,
		},
	})
	if err != nil {
		return nil, err
	}

	progMap := m.Map("tail_map")
	if progMap == nil {
		return nil, errors.New("tail map is nil")
	}

	progs := m.RawTracePointPrograms()

	for _, prog := range progs {
		log := logger.WithValues("name", prog.Name, "fd", prog.Fd())

		if strings.Contains(prog.Name, "filler/") {
			eventName := prog.Name[len("raw_tracepoint/filler/"):]
			key, found := lookupFillerID(eventName)
			if !found {
				log.Info("filler id not found")
				return nil, errors.New("filler id not found")
			}

			value := prog.Fd()

			err = m.UpdateElement(progMap, unsafe.Pointer(&key), unsafe.Pointer(&value), 0)
			if err != nil {
				log.Error(err, "failed to update prog map", "key", key, "value", value)
				return nil, err
			}

			log.Info("successfully loaded program to tail_map")

		} else {
			err = m.AttachRawTracepoint(prog.Name)
			if err != nil {
				log.Error(err, "failed to attach program", "progname", prog.Name)
				return nil, err
			}
		}
	}

	return m, nil
}

func populateSettingsMap(m *elf.Module) error {
	log := logger.WithName("[popultae-settings-map]")

	type bpfSettings struct {
		capture_enabled bool
	}

	settingsMap := m.Map("settings_map")
	key := 0
	settings := bpfSettings{}
	err := m.LookupElement(settingsMap, unsafe.Pointer(&key), unsafe.Pointer(&settings))
	if err != nil {
		log.Error(err, "failed to lookup settings map", "key", key, "settings", settings)
		return err
	}

	settings.capture_enabled = true
	err = m.UpdateElement(settingsMap, unsafe.Pointer(&key), unsafe.Pointer(&settings), 0)
	if err != nil {
		log.Error(err, "failed to update settings map", "key", key, "settings", settings)
		return err
	}

	return nil
}

func readFromPerfMap(module *elf.Module) (*elf.PerfMap, error) {
	receiveChan := make(chan []byte)
	lostChan := make(chan uint64)

	perfMap, err := elf.InitPerfMap(module, "perf_map", receiveChan, lostChan)
	if err != nil {
		return nil, err
	}

	evtDataCh := make(chan []byte, 100000)
	go processEventData(evtDataCh)

	go func() {
		for {
			select {
			case data := <-receiveChan:
				evtDataCh <- data

			case data := <-lostChan:
				logger.V(6).Info("lost events", "count", data)
			}
		}
	}()

	perfMap.PollStart()

	return perfMap, nil
}

func processEventData(evtDataCh chan []byte) {
	logger := klogr.New().WithName("[parse-event-data]")

	for {
		data := <-evtDataCh

		out := &perfEventHeader{}
		err := binary.Read(bytes.NewReader(data), binary.LittleEndian, out)
		if err != nil {
			logger.V(6).Info("error reading event header data", "error", err)
			continue
		}

		// continue reading from perf map until we have all the necessary data
		for len(data) < int(out.Len) {
			data = append(data, <-evtDataCh...)
		}

		data = data[binary.Size(out):]
		paramLens := make([]uint16, out.Nparams)

		err = binary.Read(bytes.NewReader(data), binary.LittleEndian, paramLens)
		if err != nil {
			logger.Error(err, "error reading param length data")
			continue
		}
		spew.Dump(paramLens)

		data = data[binary.Size(paramLens):]

		params := make([][]byte, out.Nparams)

		for i := 0; i < int(out.Nparams); i++ {
			params[i] = make([]byte, paramLens[i])
			params[i] = data[:paramLens[i]]

			switch i {
			case 0:
				fmt.Printf("ret = %v\n", binary.LittleEndian.Uint32(params[i]))
			case 1:
				fmt.Printf("dirfd = %v\n", binary.LittleEndian.Uint32(params[i]))
			case 2:
				fmt.Printf("name = %v\n", string(params[i]))
			case 3:
				fmt.Printf("flags = %v\n", binary.LittleEndian.Uint32(params[i]))
			case 4:
				fmt.Printf("mode = %v\n", binary.LittleEndian.Uint32(params[i]))
			case 5:
				fmt.Printf("device = %v\n", binary.LittleEndian.Uint32(params[i]))
			}
			data = data[paramLens[i]:]
		}
	}
}