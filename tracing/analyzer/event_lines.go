package analyzer

import (
	"fmt"
	"path"
	"simds-standalone/common"
	"sort"
	"time"
)

const INF = 999 * time.Hour

type EventLines interface {
	GetID(index int) string
	Len() int
	GetType(index int) string
	GetHappenTime(index int) time.Time
}

type CostRecord struct {
	Time_ms int
	Id      string
	Cost    time.Duration
}

type CostList []CostRecord

func (l CostList) SortByOccurTime() {
	sort.Slice(l, func(i, j int) bool {
		return l[i].Time_ms < l[j].Time_ms
	})
}

func (l CostList) SortByCost() {
	sort.Slice(l, func(i, j int) bool {
		return l[i].Cost < l[j].Cost
	})
}

func (l CostList) Output(outdir string, name string) {

	//time curve
	timePath := path.Join(outdir, name+"TimeCurve.log")
	common.RemoveIfExisted(timePath)

	err := common.AppendLineCsvFile(timePath, []string{"time_ms", "id", "cost_us"})
	if err != nil {
		panic(err)
	}

	l.SortByOccurTime()

	var records CostList
	var lastRecordMs int = l[0].Time_ms
	for _, item := range l {
		if (item.Time_ms - lastRecordMs) > 100 {
			maxlatency := time.Duration(0)
			var maxlatencyItem CostRecord

			for _, record := range records {
				if record.Cost > maxlatency {
					maxlatencyItem = record
				}
			}

			err := common.AppendLineCsvFile(timePath, []string{
				fmt.Sprint(maxlatencyItem.Time_ms),
				fmt.Sprint(maxlatencyItem.Id),
				fmt.Sprint(maxlatencyItem.Cost)},
			)

			if err != nil {
				panic(err)
			}
			lastRecordMs = item.Time_ms // begin next records
			records = CostList{}        //clear
		} else {
			records = append(records, item)
		}
	}

	// cdf log
	cdfPath := path.Join(outdir, name+"CDF.log")
	common.RemoveIfExisted(cdfPath)
	l.SortByCost()
	err = common.AppendLineCsvFile(cdfPath, []string{"percent", "id", name})
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(l); i += max(len(l)/10000, 1) {
		item := l[i]
		err := common.AppendLineCsvFile(cdfPath, []string{
			fmt.Sprintf("%.4f", float32(i+1)/float32(len(l))),
			fmt.Sprint(item.Id),
			fmt.Sprint(item.Cost)},
		)
		if err != nil {
			panic(err)
		}
	}

	// metrig log
	metricPath := path.Join(outdir, name+"Metric.log")
	common.RemoveIfExisted(metricPath)

	var sum time.Duration = 0
	for _, item := range l {
		sum += item.Cost
	}
	average := time.Duration(int64(sum) / int64(len(l)))
	high_90_per := l[len(l)*9/10].Cost
	high_99_per := l[len(l)*99/100].Cost

	err = common.AppendLineCsvFile(metricPath, []string{"average", "P90", "P99", "sampleNum", "sum"})
	if err != nil {
		panic(err)
	}
	err = common.AppendLineCsvFile(metricPath, []string{
		fmt.Sprint(average),
		fmt.Sprint(high_90_per),
		fmt.Sprint(high_99_per),
		fmt.Sprint(len(l)),
		fmt.Sprint(sum)},
	)

	if err != nil {
		panic(err)
	}
}

func (l CostList) RemoveFails() CostList {
	for i := 0; i < len(l); i++ {
		if l[i].Cost == INF {
			return l[0:i]
		}
	}
	return l
}

func AnalyzeStageDuration(events EventLines, event1, event2 string) CostList {
	events1map := make(map[string]time.Time)

	start := events.GetHappenTime(0)
	var res CostList
	for i := 0; i < events.Len(); i++ {
		switch events.GetType(i) {
		case event1:
			events1map[events.GetID(i)] = events.GetHappenTime(i)
		case event2:
			var temp CostRecord
			t2 := events.GetHappenTime(i)
			id := events.GetID(i)
			temp.Id = id

			if t1, ok := events1map[id]; ok {
				temp.Time_ms = int(t1.Sub(start).Milliseconds())
				temp.Cost = t2.Sub(t1)
				res = append(res, temp)
				delete(events1map, id)
			}
		default:
		}
	}

	for id, t1 := range events1map {
		var temp CostRecord
		temp.Time_ms = int(t1.Sub(start).Milliseconds())
		temp.Id = id
		temp.Cost = INF
		res = append(res, temp)
	}

	res.SortByCost()
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

	common.RemoveIfExisted(outfile)
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

func AnalyzeEventRate(events EventLines, evntype string, interval_ms int) RateList {
	if events.Len() == 0 {
		return RateList{}
	}

	start := events.GetHappenTime(0)
	end := events.GetHappenTime(events.Len() - 1)
	tointerval := func(t time.Time) int { return int(t.Sub(start) / (time.Duration(interval_ms) * time.Millisecond)) }
	result := make(RateList, tointerval(end)+1)

	for i := 0; i < len(result); i++ {
		result[i].Time_ms = i * interval_ms
	}

	for i := 0; i < events.Len(); i++ {
		if events.GetType(i) == evntype {
			result[tointerval(events.GetHappenTime(i))].Amount++
		}
	}

	return result
}
