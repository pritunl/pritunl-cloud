package demo

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/metric"
)

var (
	chartLock  = sync.Mutex{}
	chartStore = map[bson.ObjectID]uint64{}
)

type chartSeries struct {
	key string
	sum bool
	val func(minute int64) float64
}

func chartSeed(instId bson.ObjectID) (seed uint64) {
	chartLock.Lock()
	defer chartLock.Unlock()

	seed, ok := chartStore[instId]
	if !ok {
		hash := fnv.New64a()
		hash.Write(instId[:])
		seed = hash.Sum64()
		chartStore[instId] = seed
	}

	return
}

func chartHash(seed uint64, key string, minute int64) uint64 {
	var b [16]byte

	binary.BigEndian.PutUint64(b[0:8], seed)
	binary.BigEndian.PutUint64(b[8:16], uint64(minute))

	hash := fnv.New64a()
	hash.Write(b[:])
	hash.Write([]byte(key))

	return hash.Sum64()
}

func chartValue(seed uint64, key string, minute int64) float64 {
	base := chartHash(seed, key, 0)
	phase1 := float64(base%997) / 997 * 2 * math.Pi
	phase2 := float64((base>>16)%997) / 997 * 2 * math.Pi
	phase3 := float64((base>>32)%997) / 997 * 2 * math.Pi

	noise := float64(chartHash(seed, key, minute)%1000)/1000*0.08 - 0.04

	m := float64(minute)
	val := 0.5 +
		0.22*math.Sin(2*math.Pi*m/1440+phase1) +
		0.14*math.Sin(2*math.Pi*m/360+phase2) +
		0.08*math.Sin(2*math.Pi*m/45+phase3) +
		noise

	return math.Min(math.Max(val, 0), 1)
}

func chartGauge(seed uint64, key string, min, max float64) func(
	minute int64) float64 {

	return func(minute int64) float64 {
		return min + (max-min)*chartValue(seed, key, minute)
	}
}

func chartWindow(val func(minute int64) float64, window int64) func(
	minute int64) float64 {

	return func(minute int64) float64 {
		total := 0.0
		for i := int64(0); i < window; i++ {
			total += val(minute - i)
		}
		return total / float64(window)
	}
}

func buildChart(start, end time.Time, interval time.Duration,
	series []chartSeries) metric.ChartData {

	start = start.Add(time.Duration(start.UnixMilli()%
		interval.Milliseconds()) * -time.Millisecond)
	end = end.Add(time.Duration(end.UnixMilli()%
		interval.Milliseconds()) * -time.Millisecond)

	chart := metric.NewChart(start, end, interval)

	intvMs := interval.Milliseconds()
	minuteMs := time.Minute.Milliseconds()
	startMs := start.UnixMilli()
	endMs := end.UnixMilli()

	for bucket := startMs; bucket <= endMs; bucket += intvMs {
		for _, ser := range series {
			total := 0.0
			count := 0

			for ms := bucket; ms < bucket+intvMs &&
				ms <= endMs; ms += minuteMs {

				val := ser.val(ms / minuteMs)
				if ser.sum {
					val = math.Round(val)
				}
				total += val
				count++
			}

			if count == 0 {
				continue
			}

			if ser.sum {
				chart.Add(ser.key, bucket, uint64(total))
			} else {
				chart.Add(ser.key, bucket,
					math.Round(total/float64(count)*100)/100)
			}
		}
	}

	return chart.Export()
}

func chartSystemSeries(seed uint64) []chartSeries {
	return []chartSeries{
		{"cpu_usage", false, chartGauge(seed, "cpu", 5, 90)},
		{"mem_usage", false, chartGauge(seed, "mem", 25, 75)},
		{"swap_usage", false, chartGauge(seed, "swap", 0, 15)},
		{"huge_usage", false, func(minute int64) float64 {
			return 0
		}},
	}
}

func chartLoadSeries(seed uint64) []chartSeries {
	load1 := chartGauge(seed, "load", 10, 80)
	return []chartSeries{
		{"load1", false, load1},
		{"load5", false, chartWindow(load1, 5)},
		{"load15", false, chartWindow(load1, 15)},
	}
}

func chartDiskSeries(seed uint64) []chartSeries {
	return []chartSeries{
		{"/", false, chartGauge(seed, "disk", 40, 65)},
	}
}

func chartIoSeries(seed uint64, dev string,
	readMax, writeMax float64) []chartSeries {

	return []chartSeries{
		{dev + "-br", true, chartGauge(seed, "diskio-br", 0, readMax)},
		{dev + "-bw", true, chartGauge(seed, "diskio-bw", 0, writeMax)},
		{dev + "-tr", true, chartGauge(seed, "diskio-tr", 0, 6000)},
		{dev + "-tw", true, chartGauge(seed, "diskio-tw", 0, 9000)},
		{dev + "-ti", true, chartGauge(seed, "diskio-ti", 0, 12000)},
	}
}

func chartNetSeries(seed uint64, iface string,
	sentMax, recvMax float64) []chartSeries {

	return []chartSeries{
		{iface + "-bs", true, chartGauge(seed, iface+"-bs", 0, sentMax)},
		{iface + "-br", true, chartGauge(seed, iface+"-br", 0, recvMax)},
	}
}

func GetChartData(instId bson.ObjectID, typ string, start, end time.Time,
	interval time.Duration) (data metric.ChartData, err error) {

	seed := chartSeed(instId)

	var series []chartSeries

	switch typ {
	case "system":
		series = chartSystemSeries(seed)
	case "load":
		series = chartLoadSeries(seed)
	case "disk":
		series = chartDiskSeries(seed)
	case "diskio":
		series = chartIoSeries(seed, "vda", 40e6, 25e6)
	case "network":
		series = append(
			chartNetSeries(seed, "int0", 30e6, 80e6),
			chartNetSeries(seed, "ext0", 30e6, 80e6)...,
		)
	default:
		err = &errortypes.UnknownError{
			errors.New("demo: Unknown chart resource type"),
		}
		return
	}

	data = buildChart(start, end, interval, series)

	return
}

func GetNodeChartData(ndeId bson.ObjectID, typ string, start, end time.Time,
	interval time.Duration) (data metric.ChartData, err error) {

	seed := chartSeed(ndeId)

	var series []chartSeries

	switch typ {
	case "system":
		series = chartSystemSeries(seed)
	case "load":
		series = chartLoadSeries(seed)
	case "disk":
		series = chartDiskSeries(seed)
	case "diskio":
		series = chartIoSeries(seed, "nvme0n1", 400e6, 250e6)
	case "network":
		series = chartNetSeries(seed, "pritunlbr0", 300e6, 800e6)
	default:
		err = &errortypes.UnknownError{
			errors.New("demo: Unknown chart resource type"),
		}
		return
	}

	data = buildChart(start, end, interval, series)

	return
}
