import csv
import matplotlib.pyplot as plt
from datetime import datetime
import pandas as pd

def parseRFCnano(dt_str):
    # 解析日期时间部分
    dt = dt_str.split('+')[0]
    x = datetime.strptime(dt[:-3],'%Y-%m-%dT%H:%M:%S.%f')
    return x


def draw_cluster_status():
    t = []
    avg_latency = []
    avg_cpu = []
    avg_ram = []
    var_cpu = []
    var_ram = []
    with open('./cluster_status.log','r') as csvfile:
        plots = csv.reader(csvfile, delimiter=',')
        next(plots)
        for row in plots:
            t.append(parseRFCnano(row[0]))  
            latency = pd.Timedelta(row[1]).total_seconds()*1000
            avg_latency.append(latency)
            avg_cpu.append(float(row[2])*100)
            avg_ram.append(float(row[3])*100)
            var_cpu.append(float(row[4])*100)
            var_ram.append(float(row[5])*100)
    fig = plt.figure()
    ax1 = fig.add_subplot(111)
    ax1.plot(t,avg_cpu,lw=1,color='r')
    ax1.plot(t,avg_ram,lw=1,color='b')
    ax1.set_ylabel("resource usage percentage unit: %")

    ax2 = plt.twinx()
    ax2.plot(t,avg_latency,lw=1,color='y',antialiased=True)
    ax2.set_ylabel("task lantency unit: ms")
    
    plt.title('Data from CSV')
    plt.xlabel('Time')

    plt.savefig('./cluster_status.png')


if __name__ == "__main__":
    draw_cluster_status()
