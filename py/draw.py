import csv
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import os

FONT_SIZE = 20
LEGEND_SIZE = 12
LINE_WIDTH = 2.0
FAIL_TASK_LATENCY = 5000
INFINITY = 100000


def net_commuication_rate_curves(filename: str) -> list:
    """输入日志输出网络请求速率曲线"""
    records = []
    with open(filename, 'r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            records.append(int(row[0]))
    intervals = [0] * (records[-1] // 10 + 1)
    t = [i / 100 for i in range(0, len(intervals))]
    for i in records:
        intervals[i // 10] += 1
    intervals = [100 * i for i in intervals]

    # 滤波使曲线光滑
    oldy = np.array(intervals)
    y = np.array(intervals)
    halfSampleNum = 4
    for i in range(halfSampleNum, len(intervals) - halfSampleNum):
        y[i] = oldy[i - halfSampleNum:i +
                    halfSampleNum].sum() / (2 * halfSampleNum)
    return [t, y]


def task_submit_rate_curves(filename: str) -> list:
    """输入日志输出任务提交请求速率曲线"""
    records = []
    with open(filename, 'r') as logfile:
        l = logfile.readline()
        while l:
            records.append(int(l.split(" ")[0]))
            l = logfile.readline()
    intervals = [0] * (records[-1] // 10 + 1)
    t = [i / 100 for i in range(0, len(intervals))]
    for i in records:
        intervals[i // 10] += 1
    intervals = [100 * i for i in intervals]
    return [t, intervals]


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
            # task latency larger or equal than FAIL_TASK_LATENCY ms is seen as failed
            if latency > FAIL_TASK_LATENCY - 1:
                max_latency.append(INFINITY)
            else:
                max_latency.append(latency)
            avg_cpu.append(float(row[2]) * 100)
            avg_ram.append(float(row[3]) * 100)
            var_cpu.append(float(row[4]))
            var_ram.append(float(row[5]))
    return [t, max_latency, avg_cpu, avg_ram, var_cpu, var_ram]


def draw_in_current_test_folder():
    """在当前实验结果文件下画集群状态图"""
    fig = plt.figure(figsize=(10, 15))
    ax1 = fig.add_subplot(311)
    status = cluster_status_curves("./cluster_status.log")
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

    res = net_commuication_rate_curves("./network_event.log")
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

    res = task_submit_rate_curves("./task_speed.log")
    ax5 = plt.twinx()
    ax5.plot(res[0], res[1], lw=LINE_WIDTH, color='b', label="task submission")
    ax5.set_ylabel(
        "Task Submission Rate, (amount/s)",
        fontsize=FONT_SIZE,
    )
    ax5.legend(fontsize=LEGEND_SIZE)

    plt.grid(True)
    plt.savefig('./cluster_status.png')


def draw_muilt_lantencyCurve(tests: list):
    """画多个实验的任务调度延迟曲线对比图"""
    plt.cla()

    for test in tests:
        status = cluster_status_curves(
            os.path.join(test[0], "cluster_status.log"))

        t = status[0]
        latency = status[1]
        plt.plot(t, latency, lw=LINE_WIDTH, label=test[1])

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

    plt.grid(True)
    plt.subplots_adjust(left=0.13, right=0.93)
    plt.savefig('./lantency_compare.png')


def draw_muilt_avg_resource(tests: list):
    """画多个实验的集群平均负载曲线对比图"""
    plt.cla()
    for t in tests:
        staus = cluster_status_curves(os.path.join(t[0], "cluster_status.log"))
        plt.plot(staus[0], staus[2], lw=LINE_WIDTH, label=t[1])
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Resource Utilization (%)", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.grid(True)
    plt.subplots_adjust(left=0.12, right=0.93)
    plt.savefig('./load_compare.png')


def draw_muilt_var_resource(tests: list):
    """画多个实验的集群负载方差对比图"""
    plt.cla()
    for t in tests:
        staus = cluster_status_curves(os.path.join(t[0], "cluster_status.log"))
        plt.plot(staus[0], staus[4], lw=LINE_WIDTH, label=t[1])
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Cluster Utilization Variance", fontsize=FONT_SIZE)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.grid(True)
    plt.subplots_adjust(left=0.12, right=0.93)
    plt.savefig('./variance_compare.png')


def draw_muilt_net_busy(tests: list):
    """画多个实验的网络繁忙程度对比图"""
    plt.cla()
    for t in tests:
        staus = net_commuication_rate_curves(
            os.path.join(t[0], "network_event.log"))
        plt.plot(staus[0], staus[1], lw=1.5, label=t[1])
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Net Request Rate \n (number/s)", fontsize=FONT_SIZE)
    plt.ticklabel_format(style='plain')
    #plt.yscale("log",base=10)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.grid(True)
    plt.subplots_adjust(left=0.22, right=0.93)
    plt.savefig('./net_busy_compare_cluster.png')

    plt.cla()
    for t in tests:
        staus = net_commuication_rate_curves(
            os.path.join(t[0], "network_most_busy.log"))
        plt.plot(staus[0], staus[1], lw=1.5, label=t[1])
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Net Request Rate \n (number/s)", fontsize=FONT_SIZE)
    plt.ticklabel_format(style='plain')
    #plt.yscale("log",base=10)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.grid(True)
    plt.subplots_adjust(left=0.22, right=0.93)
    plt.savefig('./net_busy_compare_most_busy.png')


def draw_task_submission_rate(tests: list):
    """画多个实验的任务提交速率图"""
    plt.cla()
    staus = task_submit_rate_curves(
        os.path.join(tests[0][0], "./task_speed.log"))
    plt.plot(staus[0], staus[1], lw=LINE_WIDTH)
    plt.ylabel("Task Submission Rate \n (number/s)", fontsize=FONT_SIZE)
    plt.subplots_adjust(left=0.2)
    plt.xlabel("Time (s)", fontsize=FONT_SIZE)

    plt.grid(True)
    plt.subplots_adjust(left=0.22, right=0.93)
    plt.savefig('./task_submission_rate.png')


def draw_task_latency_CDF(tests: list):
    """画多个实验的任务延迟的累积概率分布函数"""
    plt.cla()
    for t in tests:
        staus = task_latency_CDF_curves(os.path.join(t[0], "latencyCurve.log"))
        plt.plot(staus[0], staus[1], lw=LINE_WIDTH, label=t[1])
        #if max(staus[0]) >= FAIL_TASK_LATENCY-1:
    plt.legend(fontsize=LEGEND_SIZE)
    plt.ylabel("Cumulative Probability", fontsize=FONT_SIZE)
    plt.xlabel("Task Latency(ms)", fontsize=FONT_SIZE)

    plt.grid(True)
    plt.subplots_adjust(left=0.15, right=0.94, bottom=0.15)

    plt.savefig('./latency_CDF_compare_full.png')
    plt.ylim(0.95, 1.002)
    plt.savefig('./latency_CDF_compare.png')


if __name__ == "__main__":
    draw_in_current_test_folder()
