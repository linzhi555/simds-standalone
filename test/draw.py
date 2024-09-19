import csv
import matplotlib.pyplot as plt
import pandas as pd
import os

# ------------------------------------------------------------------
# -------------------------begin private----------------------------
# ------------------------------------------------------------------

FONT_SIZE = 20
LEGEND_SIZE = 15
LINE_WIDTH = 2.0


def _savefig(outputPath, png: str):
    plt.savefig(os.path.join(outputPath, png))


class markerGenerator():
    def __init__(self):
        self.count = 0
        self.markers = ['o', '^', 'x', 'v', 's', 'D']

    def next(self) -> str:
        result = self.markers[self.count]
        self.count += 1
        return result


def _rate_curves(filename: str) -> tuple:
    """速率类型的曲线"""
    t = []
    amount = []
    with open(filename, 'r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            t.append(int(row[0])/1000)
            amount.append(int(row[1]))
    return t, amount


def _cost_CDF_curves(filename: str) -> tuple:
    """CDF 类型的曲线"""
    costs = []
    percents = []
    with open(filename, 'r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            cost = pd.Timedelta(row[2]).total_seconds() * 1000
            costs.append(cost)
            percents.append(float(row[0]))
    return costs, percents


def _cost_time_curves(filename: str) -> tuple:
    """时间轴 类型的曲线"""
    t = []
    costs = []
    with open(filename, 'r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            t.append(int(row[0])/1000)
            costs.append(pd.Timedelta(row[2]).total_seconds() * 1000)
    return t, costs


def _cluster_status_curves(filename: str) -> tuple:
    """生成集群状态曲线"""
    t = []
    avg_cpu = []
    avg_ram = []
    var_cpu = []
    var_ram = []
    with open(filename, 'r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            t.append(int(row[0]) / 1000)

            avg_cpu.append(float(row[1]) * 100)
            avg_ram.append(float(row[2]) * 100)
            var_cpu.append(float(row[3]))
            var_ram.append(float(row[4]))
    return t, avg_cpu, avg_ram, var_cpu, var_ram


# ------------------------------------------------------------------
# -------------------------begin public-----------------------------
# ------------------------------------------------------------------

def draw_muilt_lantencyCurve(tests: list, outdir: str):
    """画多个实验的任务调度延迟曲线对比图"""
    marker = markerGenerator()
    plt.clf()

    for test in tests:
        t, latency = _cost_time_curves(
            os.path.join(test[0], "./_taskLatencyTimeCurve.log"))

        plt.plot(t, latency, lw=LINE_WIDTH, marker=marker.next(),
                 markevery=8, markersize=7, label=test[1])
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Worst Task Lantency (ms)", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.19, right=0.93, bottom=0.15, top=0.95)
    _savefig(outdir, './lantency_compare.png')

    plt.yscale("log", base=10)
    _savefig(outdir, './lantency_compare_log.png')


def draw_muilt_avg_resource(tests: list, outdir: str):
    """画多个实验的集群平均负载曲线对比图"""

    # for drawing cpu ulitzation
    marker = markerGenerator()
    plt.clf()
    for test in tests:

        t, avg_cpu, avg_ram, var_cpu, var_ram = _cluster_status_curves(
            os.path.join(test[0], "_clusterStatus.log"))

        plt.plot(t, avg_cpu, lw=LINE_WIDTH, label=t[1],
                 marker=marker.next(), markevery=8, markersize=7)

    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("CPU Utilization (%)", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.13, right=0.93, top=0.95)
    _savefig(outdir, './cpu_load_compare.png')

    # for drawing memory ulitzation
    marker = markerGenerator()
    plt.clf()
    for test in tests:

        t, avg_cpu, avg_ram, var_cpu, var_ram = _cluster_status_curves(
            os.path.join(test[0], "_clusterStatus.log"))

        plt.plot(t, avg_ram, lw=LINE_WIDTH, label=t[1],
                 marker=marker.next(), markevery=8, markersize=7)

    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Memory Utilization (%)", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.13, right=0.93, top=0.95)
    _savefig(outdir, './memory_load_compare.png')


def draw_muilt_var_resource(tests: list, outdir: str):
    """画多个实验的集群负载方差对比图"""

    marker = markerGenerator()
    plt.clf()
    for test in tests:
        t, avg_cpu, avg_ram, var_cpu, var_ram = _cluster_status_curves(
            os.path.join(test[0], "_clusterStatus.log"))
        plt.plot(t, var_cpu, lw=LINE_WIDTH, label=test[1],
                 marker=marker.next(), markevery=8, markersize=7)
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Cluster CPU Utilization Variance", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.20, right=0.93, top=0.95)

    _savefig(outdir, './cpu_variance_compare.png')

    marker = markerGenerator()
    plt.clf()
    for t in tests:
        t, avg_cpu, avg_ram, var_cpu, var_ram = _cluster_status_curves(
            os.path.join(test[0], "_clusterStatus.log"))
        plt.plot(t, var_ram, lw=LINE_WIDTH, label=test[1],
                 marker=marker.next(), markevery=8, markersize=7)
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Cluster Memory Utilization Variance", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.20, right=0.93, top=0.95)

    _savefig(outdir, './memory_variance_compare.png')


def draw_muilt_net_busy(tests: list, outdir: str):
    """画多个实验的网络繁忙程度对比图"""
    plt.clf()
    for t in tests:
        staus = _rate_curves(
            os.path.join(t[0], "_allNetRate.log"))
        plt.plot(staus[0], staus[1], lw=1.0, label=t[1])
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Net Request Rate \n (number/s)", fontsize=FONT_SIZE)
    plt.yscale("log", base=10)
    # plt.ticklabel_format(style='plain')
    # plt.yscale("log",base=10)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.legend(fontsize=LEGEND_SIZE, bbox_to_anchor=(
        1.1, 1), loc='upper right',)
    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.3, right=0.93, top=0.95)
    _savefig(outdir, './net_busy_compare_cluster.png')

    plt.clf()
    for t in tests:
        staus = _rate_curves(
            os.path.join(t[0], "_busiestHostNetRate.log"))
        plt.plot(staus[0], staus[1], lw=1.0, label=t[1])
    plt.ylabel("Net Request Rate \n (number/s)", fontsize=FONT_SIZE)
    plt.yscale("log", base=10)
    # plt.ticklabel_format(style='plain')
    # plt.yscale("log",base=10)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.legend(fontsize=LEGEND_SIZE, bbox_to_anchor=(
        1.1, 1), loc='upper right',)
    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.3, right=0.93, top=0.95)
    _savefig(outdir, './net_busy_compare_most_busy.png')


def draw_task_latency_CDF(tests: list, outdir: str):
    """画多个实验的任务延迟的累积概率分布函数"""
    marker = markerGenerator()
    plt.clf()
    for t in tests:
        costs, percents = _cost_CDF_curves(
            os.path.join(t[0], "_taskLatencyCDF.log"))
        plt.plot(costs, percents, lw=LINE_WIDTH, label=t[1],
                 marker=marker.next(), markevery=0.2, markersize=7)

    plt.legend(fontsize=LEGEND_SIZE, loc="lower left")

    plt.ylabel("Cumulative Probability", fontsize=FONT_SIZE)
    plt.xlabel("Task Latency(ms)", fontsize=FONT_SIZE)

    plt.xscale("log", base=10)
    plt.grid(True)
    plt.subplots_adjust(left=0.18, right=0.93, bottom=0.15, top=0.95)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.ylim(0.95, 1.002)
    _savefig(outdir, './latency_CDF_compare.png')

    plt.ylim(0.0, 1.1)
    _savefig(outdir, './latency_CDF_compare_full.png')


class testDataList:
    def __init__(self):
        self.tests = []

    def add_data(self, dataDir: str, testLabel: str):
        self.tests.append([dataDir, testLabel])


def all(datalist: testDataList, outfolder: str):
    draw_muilt_lantencyCurve(datalist.tests, outfolder)
    draw_muilt_avg_resource(datalist.tests, outfolder)
    draw_muilt_var_resource(datalist.tests, outfolder)
    draw_muilt_net_busy(datalist.tests, outfolder)
    draw_task_latency_CDF(datalist.tests, outfolder)


if __name__ == "__main__":
    tests = testDataList()
    os.system("mkdir -p pngs")
    tests.add_data(".", "test")
    all(tests, "pngs")
    pass
