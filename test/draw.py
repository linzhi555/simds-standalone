import csv
import matplotlib.pyplot as plt
import pandas as pd
import os

FONT_SIZE = 20
LEGEND_SIZE = 15
LINE_WIDTH = 2.0
FAIL_TASK_LATENCY = 5000
INFINITY = 100000


def savefig(outputPath, png: str):
    plt.savefig(os.path.join(outputPath, png))


class markerGenerator():
    def __init__(self):
        self.count = 0
        self.markers = ['o', '^', 'x', 'v', 's', 'D']

    def next(self):
        result = self.markers[self.count]
        self.count += 1
        return result


def rate_curves(filename: str) -> list:
    """任务提交请求速率曲线"""
    t = []
    amount = []
    with open(filename, 'r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            t.append(int(row[0])/1000)
            amount.append(int(row[1]))
    return [t, amount]


# def task_submit_rate_hist(filename: str) -> list:
#     """输入日志输出任务提交请求柱状图"""
#     records = []
#     with open(filename, 'r') as logfile:
#         line = logfile.readline()
#         while line:
#             records.append(int(re.split(",|\s", line)[0]))
#             line = logfile.readline()
#     taskstream = [i/1000 for i in records]
#     return taskstream


def task_latency_CDF_curves(filename: str) -> list:
    """任务延迟的累积分布函数 CDF"""
    records = []
    with open(filename, 'r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            latency = pd.Timedelta(row[1]).total_seconds() * 1000
            if latency > FAIL_TASK_LATENCY - 1:
                latency = INFINITY
            records.append(latency)
    ticks = [(i + 1) / len(records) for i in range(0, len(records))]
    return [records, ticks]


def cluster_status_curves(filename: str) -> list:
    """生成集群状态曲线"""
    t = []
    max_latency = []
    avg_cpu = []
    avg_ram = []
    var_cpu = []
    var_ram = []
    with open(filename, 'r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            t.append(int(row[0]) / 1000)
            latency = pd.Timedelta(row[1]).total_seconds() * 1000

            # task latency >= than FAIL_TASK_LATENCY ms is seen as failed
            if latency > FAIL_TASK_LATENCY - 1:
                max_latency.append(INFINITY)
            else:
                max_latency.append(latency)
            avg_cpu.append(float(row[2]) * 100)
            avg_ram.append(float(row[3]) * 100)
            var_cpu.append(float(row[4]))
            var_ram.append(float(row[5]))
    return [t, max_latency, avg_cpu, avg_ram, var_cpu, var_ram]


def draw_muilt_lantencyCurve(tests: list, outdir: str):
    """画多个实验的任务调度延迟曲线对比图"""
    marker = markerGenerator()
    plt.clf()

    for test in tests:
        status = cluster_status_curves(
            os.path.join(test[0], "_clusterStatus.log"))

        t = status[0]
        latency = status[1]
        plt.plot(t, latency, lw=LINE_WIDTH, marker=marker.next(),
                 markevery=8, markersize=7, label=test[1])

        if max(latency) > FAIL_TASK_LATENCY - 1:
            plt.yscale("log", base=10)
        failTaskT = []
        for i in range(len(latency)):
            if latency[i] > FAIL_TASK_LATENCY - 1:
                failTaskT.append(t[i])
        if len(failTaskT) > 0:
            plt.plot(failTaskT, [max(latency) for _ in range(len(failTaskT))],
                     'ro')

    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Worst Task Lantency (ms)", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.19, right=0.93, bottom=0.15, top=0.95)
    savefig(outdir, './lantency_compare.png')

    plt.yscale("log", base=10)
    savefig(outdir, './lantency_compare_log.png')


def draw_muilt_avg_resource(tests: list, outdir: str):
    """画多个实验的集群平均负载曲线对比图"""
    marker = markerGenerator()
    plt.clf()
    for t in tests:
        staus = cluster_status_curves(os.path.join(t[0], "_clusterStatus.log"))
        plt.plot(staus[0], staus[2], lw=LINE_WIDTH, label=t[1],
                 marker=marker.next(), markevery=8, markersize=7)
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Resource Utilization (%)", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.13, right=0.93, top=0.95)
    savefig(outdir, './load_compare.png')


def draw_muilt_var_resource(tests: list, outdir: str):
    """画多个实验的集群负载方差对比图"""
    marker = markerGenerator()
    plt.clf()
    for t in tests:
        staus = cluster_status_curves(os.path.join(t[0], "_clusterStatus.log"))
        plt.plot(staus[0], staus[4], lw=LINE_WIDTH, label=t[1],
                 marker=marker.next(), markevery=8, markersize=7)
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Cluster CPU Utilization Variance", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.20, right=0.93, top=0.95)

    savefig(outdir, './cpu_variance_compare.png')

    marker = markerGenerator()
    plt.clf()
    for t in tests:
        staus = cluster_status_curves(os.path.join(t[0], "_clusterStatus.log"))
        plt.plot(staus[0], staus[5], lw=LINE_WIDTH, label=t[1],
                 marker=marker.next(), markevery=8, markersize=7)
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Cluster Memory Utilization Variance", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    plt.subplots_adjust(left=0.20, right=0.93, top=0.95)

    savefig(outdir, './memory_variance_compare.png')


def draw_muilt_net_busy(tests: list, outdir: str):
    """画多个实验的网络繁忙程度对比图"""
    plt.clf()
    for t in tests:
        staus = rate_curves(
            os.path.join(t[0], "allNetRate.log"))
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
    savefig(outdir, './net_busy_compare_cluster.png')

    plt.clf()
    for t in tests:
        staus = rate_curves(
            os.path.join(t[0], "busiestHostNetRate.log"))
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
    savefig(outdir, './net_busy_compare_most_busy.png')


# def draw_task_submission_rate(tests: list, outdir: str):
#     """画多个实验的任务提交速率图"""
#     i = 0
#     for t in tests:
#         plt.clf()
#
#         marker = markerGenerator()
#         taskhist = task_submit_rate_hist(
#             os.path.join(t[0], "./task_speed.log"))
#         curve = [0 for _ in range(0, len(taskhist)+1)]
#         for i in taskhist:
#             curve[int(i)] += 1

#         staus = cluster_status_curves(os.path.join(t[0],
#                                           "cluster_status.log"))

#         ax1 = plt.axes()
#         ax1.hist(taskhist, bins=np.arange(37), histtype='bar',
#                  rwidth=0.7, color="c", label="tasks")
#         ax1.set_ylabel(
#             "Tasks Rate(amount/s)",
#             fontsize=FONT_SIZE,
#         )
#         # ax1.set_ylim(0,max(curve) * 1.2)
#         ax1.set_ylim(0, 11000)
#         ax1.set_xlabel("Time (s)", fontsize=FONT_SIZE)
#         ax2 = plt.twinx()
#         ax2.plot(staus[0], staus[2],
#                  lw=LINE_WIDTH, color="b",
#                  label="CPU", marker=marker.next(),
#                   markevery=8, markersize=7)

#         ax2.plot(staus[0], staus[3], lw=LINE_WIDTH, color="r",
#                  label="Memory", marker=marker.next(),
#                  markevery=8, markersize=7)

#         ax2.set_ylabel("Resources Utilization (%)", fontsize=FONT_SIZE)
#         ax2.set_ylim(0, 100)
#         ax2.legend(fontsize=LEGEND_SIZE, loc='upper right')
#         ax1.legend(fontsize=LEGEND_SIZE, loc='upper left')
#         plt.xticks(fontsize=FONT_SIZE*0.8)
#
#         plt.grid(True)
#         plt.subplots_adjust(left=0.15, right=0.85, top=0.95, bottom=0.15)
#         savefig(
#             outdir,
#             './task_submission_rate_{}.png'.format(t[1].replace(" ", "_")))
#         i += 1


def draw_task_latency_CDF(tests: list, outdir: str):
    """画多个实验的任务延迟的累积概率分布函数"""
    marker = markerGenerator()
    plt.clf()
    for t in tests:
        staus = task_latency_CDF_curves(os.path.join(t[0], "latencyCurve.log"))
        plt.plot(staus[0], staus[1], lw=LINE_WIDTH, label=t[1],
                 marker=marker.next(), markevery=0.2, markersize=7)
        # if max(staus[0]) >= FAIL_TASK_LATENCY-1:
    plt.legend(fontsize=LEGEND_SIZE, loc="lower left")

    plt.ylabel("Cumulative Probability", fontsize=FONT_SIZE)
    plt.xlabel("Task Latency(ms)", fontsize=FONT_SIZE)

    plt.xscale("log", base=10)
    plt.grid(True)
    plt.subplots_adjust(left=0.18, right=0.93, bottom=0.15, top=0.95)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.ylim(0.95, 1.002)
    savefig(outdir, './latency_CDF_compare.png')

    plt.ylim(0.0, 1.1)
    savefig(outdir, './latency_CDF_compare_full.png')


def indepent_draw():
    """
    在一个实验跑完后直接在target画得实验过程的图,用于大概描述该论实验的情况，
    不与其他实验进行结果对比分析
    """
    fig = plt.figure(figsize=(10, 15))
    ax1 = fig.add_subplot(311)
    status = cluster_status_curves("./_clusterStatus.log")
    t = status[0]
    avg_latency = status[1]
    avg_cpu = status[2]
    avg_ram = status[3]
    var_cpu = status[4]
    var_ram = status[5]

    ax1.plot(t, avg_cpu, lw=LINE_WIDTH, color='r', label="cpu average")
    ax1.plot(t, avg_ram, lw=LINE_WIDTH, color='b', label="memory average")
    ax1.set_ylabel("Resource Usage Percentage (%)", fontsize=FONT_SIZE)
    ax1.set_xlabel("Time (s)", fontsize=FONT_SIZE)
    ax1.legend(fontsize=LEGEND_SIZE, loc="upper left")

    ax2 = plt.twinx()
    ax2.plot(t, avg_latency, lw=LINE_WIDTH, color='y', label="task latency")
    ax2.set_ylabel("Task Lantency Unit: ms", fontsize=FONT_SIZE)
    ax2.legend(fontsize=LEGEND_SIZE, loc="upper right")

    if max(avg_latency) > FAIL_TASK_LATENCY - 1:
        ax2.set_yscale("log", base=10)

    ax3 = fig.add_subplot(312)
    ax3.plot(t, var_cpu, lw=LINE_WIDTH, label="cpu variance")
    ax3.plot(t, var_ram, lw=LINE_WIDTH, label="ram variance")
    ax3.set_ylabel("Resource Variance", fontsize=FONT_SIZE)
    ax3.set_xlabel("Time unit: s", fontsize=FONT_SIZE)
    ax3.legend(fontsize=LEGEND_SIZE)

    res = rate_curves("./_allNetRate.log")
    ax4 = fig.add_subplot(313)
    ax4.plot(res[0],
             res[1],
             lw=LINE_WIDTH,
             color='y',
             label="all type of request")
    ax4.set_ylabel(
        "Request Rate, (amount/s)",
        fontsize=FONT_SIZE,
    )
    ax4.set_xlabel("Time (s)", fontsize=FONT_SIZE)
    ax4.legend(fontsize=LEGEND_SIZE, loc="upper left")

    res = rate_curves("./_taskSubmitRate.log")
    ax5 = plt.twinx()
    ax5.plot(res[0], res[1], lw=LINE_WIDTH, color='b', label="task submission")
    ax5.set_ylabel(
        "Task Submission Rate, (amount/s)",
        fontsize=FONT_SIZE,
    )
    ax5.legend(fontsize=LEGEND_SIZE)

    plt.yticks(fontsize=FONT_SIZE*0.8)
    plt.xticks(fontsize=FONT_SIZE*0.8)
    plt.grid(True)
    savefig(".", './_brief.png')


if __name__ == "__main__":
    indepent_draw()
