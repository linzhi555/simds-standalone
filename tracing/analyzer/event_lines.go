package analyzer

import (
	"fmt"
	"path"
	"simds-standalone/common"
	"sort"
	"time"
)

type EventLines interface {
	GetID(index int) string
	Len() int
	GetType(index int) string
	GetHappenTime(index int) time.Time
}

type StageCostList []struct {
	Id   string
	Cost time.Duration
}

func (l StageCostList) Len() int      { return len(l) }
func (l StageCostList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l StageCostList) Less(i, j int) bool {
	return l[i].Cost.Nanoseconds() < l[j].Cost.Nanoseconds()
}

func (l StageCostList) Output(outdir string, name string) {
	outPutLogPath := path.Join(outdir, name+"Curve.log")
	outPutMetricPath := path.Join(outdir, name+"Metric.log")

	err := common.AppendLineCsvFile(outPutLogPath, []string{"id", name})
	if err != nil {
		panic(err)
	}

	for _, line := range l {
		err := common.AppendLineCsvFile(outPutLogPath, []string{fmt.Sprint(line.Id), fmt.Sprint(line.Cost)})
		if err != nil {
			panic(err)
		}
	}

	var sum time.Duration = 0
	for _, line := range l {
		sum += line.Cost
	}
	average := time.Duration(int64(sum) / int64(l.Len()))
	high_90_per := l[l.Len()*9/10].Cost
	high_99_per := l[l.Len()*99/100].Cost

	err = common.AppendLineCsvFile(outPutMetricPath, []string{"average", "90%high", "99%high"})
	if err != nil {
		panic(err)
	}
	err = common.AppendLineCsvFile(outPutMetricPath, []string{fmt.Sprint(average), fmt.Sprint(high_90_per), fmt.Sprint(high_99_per)})

	if err != nil {
		panic(err)
	}
}

func AnalyzeStageDuration(events EventLines, event1, event2 string) StageCostList {
	events1map := make(map[string]time.Time)
	events2map := make(map[string]time.Time)

	for i := 0; i < events.Len(); i++ {
		switch events.GetType(i) {
		case event1:
			events1map[events.GetID(i)] = events.GetHappenTime(i)
		case event2:
			events2map[events.GetID(i)] = events.GetHappenTime(i)
		default:
		}
	}

	var res StageCostList
	for id := range events1map {
		var temp struct {
			Id   string
			Cost time.Duration
		}
		temp.Id = id
		if _, ok := events2map[id]; ok {
			temp.Cost = events2map[id].Sub(events1map[id])
		} else {
			temp.Cost = FAIL
		}
		res = append(res, temp)

	}

	sort.Sort(res)
	return res
}

type RateList []struct {
	Time_ms int
	Amount  int
}

func (l RateList) Highest() int {
	maxamout := -1
	for _, line := range l {
		if line.Amount > maxamout {
			maxamout = line.Amount
		}
	}
	return maxamout

}

func (l RateList) Output(outdir string, name string) {
	outfile := path.Join(outdir, name+"Rate.log")

	err := common.AppendLineCsvFile(outfile, []string{"time_ms", "amout"})
	if err != nil {
		panic(err)
	}

	for _, line := range l {
		err := common.AppendLineCsvFile(outfile, []string{fmt.Sprint(line.Time_ms), fmt.Sprint(line.Amount)})
		if err != nil {
			panic(err)
		}
	}

}

func AnalyzeEventRate(events EventLines, evntype string, interval int) RateList {
	start := events.GetHappenTime(0)
	end := events.GetHappenTime(events.Len() - 1)
	tointerval := func(t time.Time) int { return int(t.Sub(start) / (time.Duration(interval) * time.Millisecond)) }
	result := make(RateList, tointerval(end)+1)

	for i := 0; i < len(result); i++ {
		result[i].Time_ms = i * interval
	}

	for i := 0; i < events.Len(); i++ {
		if events.GetType(i) == evntype {
			result[tointerval(events.GetHappenTime(i))].Amount++
		}
	}

	return result
}