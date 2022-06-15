package providers

import (
	"reflect"
	"time"

	"github.com/up9inc/mizu/tap/api"
)

type GeneralStats struct {
	EntriesCount        int
	EntriesVolumeInGB   float64
	FirstEntryTimestamp int
	LastEntryTimestamp  int
}

type TimeFrameStatsValue map[string]ProtocolStats

type ProtocolStats struct {
	MethodsStats map[string]*SizeAndEntriesCount
	Color        string
}

type SizeAndEntriesCount struct {
	EntriesCount  int
	VolumeInBytes int
}

type AccumulativeStatsCounter struct {
	Name            string `json:"name"`
	EntriesCount    int    `json:"entriesCount"`
	VolumeSizeBytes int    `json:"volumeSizeBytes"`
}

type AccumulativeStatsProtocol struct {
	AccumulativeStatsCounter
	Color   string                      `json:"color"`
	Methods []*AccumulativeStatsCounter `json:"methods"`
}

type BucketStats map[time.Time]TimeFrameStatsValue

var (
	generalStats = GeneralStats{}
	bucketsStats = BucketStats{}
)

func ResetGeneralStats() {
	generalStats = GeneralStats{}
}

func GetGeneralStats() GeneralStats {
	return generalStats
}

func GetAccumulativeStats() []*AccumulativeStatsProtocol {
	result := make(map[string]*AccumulativeStatsProtocol, 0)
	methodsPerProtocolAggregated := make(map[string]map[string]*AccumulativeStatsCounter, 0)

	for _, countersOfTimeFrame := range bucketsStats {
		for protocolName, value := range countersOfTimeFrame {

			if _, found := result[protocolName]; !found {
				result[protocolName] = &AccumulativeStatsProtocol{
					AccumulativeStatsCounter: AccumulativeStatsCounter{
						Name:            protocolName,
						EntriesCount:    0,
						VolumeSizeBytes: 0,
					},
					Color:   value.Color,
				}
			}
			if _, found := methodsPerProtocolAggregated[protocolName]; !found {
				methodsPerProtocolAggregated[protocolName] = map[string]*AccumulativeStatsCounter{}
			}

			for method, countersValue := range value.MethodsStats {
				if _, found := methodsPerProtocolAggregated[protocolName][method]; !found {
					methodsPerProtocolAggregated[protocolName][method] = &AccumulativeStatsCounter{
						Name:            method,
						EntriesCount:    0,
						VolumeSizeBytes: 0,
					}
				}

				result[protocolName].AccumulativeStatsCounter.EntriesCount += countersValue.EntriesCount
				methodsPerProtocolAggregated[protocolName][method].EntriesCount += countersValue.EntriesCount
				result[protocolName].AccumulativeStatsCounter.VolumeSizeBytes += countersValue.VolumeInBytes
				methodsPerProtocolAggregated[protocolName][method].VolumeSizeBytes += countersValue.VolumeInBytes
			}
		}
	}

	finalResult := make([]*AccumulativeStatsProtocol, 0)
	for _, value := range result {
		methodsForProtocol := make([]*AccumulativeStatsCounter, 0)
		for _, methodValue := range methodsPerProtocolAggregated[value.Name] {
			methodsForProtocol = append(methodsForProtocol, methodValue)
		}
		value.Methods = methodsForProtocol
		finalResult = append(finalResult, value)
	}
	return finalResult
}

func EntryAdded(size int, summery *api.BaseEntry) {
	generalStats.EntriesCount++
	generalStats.EntriesVolumeInGB += float64(size) / (1 << 30)

	currentTimestamp := int(time.Now().Unix())

	if reflect.Value.IsZero(reflect.ValueOf(generalStats.FirstEntryTimestamp)) {
		generalStats.FirstEntryTimestamp = currentTimestamp
	}

	addToBucketStats(size, summery)

	generalStats.LastEntryTimestamp = currentTimestamp
}

func addToBucketStats(size int, summery *api.BaseEntry) {
	entryTimeBucketRounded := time.Unix(summery.Timestamp, 0).Round(time.Minute * 5)
	if _, found := bucketsStats[entryTimeBucketRounded]; !found {
		bucketsStats[entryTimeBucketRounded] = TimeFrameStatsValue{}
	}
	if _, found := bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation]; !found {
		bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation] = ProtocolStats{
			MethodsStats: map[string]*SizeAndEntriesCount{},
			Color:        summery.Protocol.BackgroundColor,
		}
	}
	if _, found := bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation].MethodsStats[summery.Method]; !found {
		bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation].MethodsStats[summery.Method] = &SizeAndEntriesCount{
			VolumeInBytes: 0,
			EntriesCount:  0,
		}
	}

	bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation].MethodsStats[summery.Method].EntriesCount += 1
	bucketsStats[entryTimeBucketRounded][summery.Protocol.Abbreviation].MethodsStats[summery.Method].VolumeInBytes += size
}
