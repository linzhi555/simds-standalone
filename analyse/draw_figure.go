package main

import (
	"image/color"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// MyTicks 自定义Ticks
type MyTicks struct{ precise int }

// Ticks returns Ticks in the specified range.
func (m MyTicks) Ticks(min, max float64) []plot.Tick {
	if max <= min {
		panic("illegal range")
	}
	var ticks []plot.Tick

	// 1. from the range we get 20 float number, they have same gap between neigbours,but they are float with long tails numbers
	// 2. get the float number with short tail number  near the number ,for example ,from 21.0244  to 21.02 when the precise is 2 using the 'FormatFloat'
	// 3. change the string to the tick we need
	//delta, _ := strconv.ParseFloat(strconv.FormatFloat((max-min)/20, 'f', m.precise, 64), 64)
	delta := (max - min) / 20

	var temptick float64
	for temptick = 0; temptick <= max; temptick += delta {
		ticks = append(ticks, plot.Tick{Value: temptick, Label: strconv.FormatFloat(temptick, 'f', m.precise, 64)})
	}
	ticks = append(ticks, plot.Tick{Value: temptick, Label: strconv.FormatFloat(temptick, 'f', m.precise, 64)})

	return ticks
}

func outputLatencyResultFigure(fileName string, l TaskStageCostList) {

	p := plot.New()

	p.Title.Text = "Task Scheduling Latency Diagram"
	p.X.Label.Text = "X: the percent of the tasks (%)"
	p.Y.Label.Text = "Y: the corresponding time (ms)"

	p.Y.Tick.Marker = MyTicks{precise: 2}
	p.X.Tick.Marker = MyTicks{precise: 1}

	xys := calPercentCurves(l)
	newCurve, err := plotter.NewScatter(xys)
	newCurve.Color = color.RGBA{R: 255, B: 0, A: 255}

	if err != nil {
		panic(err)
	}

	p.Add(newCurve)

	// Save the plot to a PNG file.
	if err := p.Save(10*vg.Inch, 10*vg.Inch, fileName); err != nil {
		panic(err)
	}
}

func calPercentCurves(latencyList TaskStageCostList) plotter.XYs {
	allNum := len(latencyList)

	var pts plotter.XYs

	for i := 0; i < allNum; i++ {
		var temp plotter.XY
		temp.Y = float64(latencyList[i].Cost.Microseconds()) / 1000
		temp.X = float64(i) / float64(allNum) * float64(100)
		pts = append(pts, temp)
	}
	return pts
}

func OutputAverageCpuRamCurve(fileName string, statusLine []ClusterStatus) {
	p := plot.New()

	p.Title.Text = "cluster Metric curve"
	p.X.Label.Text = "X: the time"
	p.Y.Label.Text = "Y: cpu percent"

	linecpu, err := plotter.NewScatter(calAverageCpuCurve(statusLine))
	lineram, err := plotter.NewScatter(calAverageRamCurve(statusLine))
	if err != nil {
		panic(err)
	}

	linecpu.Color = color.RGBA{R: 255, B: 0, A: 255}
	lineram.Color = color.RGBA{G: 255, B: 0, A: 255}

	p.Add(linecpu, lineram)

	p.Legend.Add("average cpu used percentage", linecpu)
	p.Legend.Add("average ram used percentage", lineram)

	p.Y.Tick.Marker = MyTicks{precise: 2}
	p.X.Tick.Marker = MyTicks{precise: 1}
	// Save the plot to a PNG file.
	if err := p.Save(10*vg.Inch, 10*vg.Inch, fileName); err != nil {
		panic(err)
	}

}

func calAverageCpuCurve(statusLine []ClusterStatus) plotter.XYs {
	allNum := len(statusLine)
	pts := make(plotter.XYs, allNum)
	for i := 0; i < allNum; i++ {
		pts[i].X = statusLine[i].Time.Sub(statusLine[0].Time).Seconds()
		pts[i].Y = float64(statusLine[i].Metric.CpuUsedPerAverage)
	}
	return pts
}

func calAverageRamCurve(statusLine []ClusterStatus) plotter.XYs {
	allNum := len(statusLine)
	pts := make(plotter.XYs, allNum)
	for i := 0; i < allNum; i++ {
		pts[i].X = statusLine[i].Time.Sub(statusLine[0].Time).Seconds()
		pts[i].Y = float64(statusLine[i].Metric.RamUsedPerAverage)
	}
	return pts
}

func OutputVarianceCpuRamCurve(fileName string, statusLine []ClusterStatus) {
	p := plot.New()

	p.Title.Text = "cluster Metric curve"
	p.X.Label.Text = "X: the time"
	p.Y.Label.Text = "Y: variance"

	err := plotutil.AddLinePoints(p,
		"variance cpu used percentage", calVarianceCpuCurve(statusLine),
		"variance ram used percentage", calVarianceRamCurve(statusLine),
	)

	if err != nil {
		panic(err)
	}

	p.Y.Tick.Marker = MyTicks{precise: 4}
	p.X.Tick.Marker = MyTicks{precise: 1}
	// Save the plot to a PNG file.
	if err := p.Save(10*vg.Inch, 10*vg.Inch, fileName); err != nil {
		panic(err)
	}

}

func calVarianceCpuCurve(statusLine []ClusterStatus) plotter.XYs {
	allNum := len(statusLine)
	pts := make(plotter.XYs, allNum)
	for i := 0; i < allNum; i++ {
		pts[i].X = statusLine[i].Time.Sub(statusLine[0].Time).Seconds()
		pts[i].Y = float64(statusLine[i].Metric.CpuUsedPerVariance)
	}
	return pts
}

func calVarianceRamCurve(statusLine []ClusterStatus) plotter.XYs {
	allNum := len(statusLine)
	pts := make(plotter.XYs, allNum)
	for i := 0; i < allNum; i++ {
		pts[i].X = statusLine[i].Time.Sub(statusLine[0].Time).Seconds()
		pts[i].Y = float64(statusLine[i].Metric.RamUsedPerVariance)
	}
	return pts
}
