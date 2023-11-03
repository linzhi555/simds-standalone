import csv
import matplotlib.pyplot as plt
from datetime import datetime
import pandas as pd
import os

#def parseRFCnano(dt_str):
#    # 解析日期时间部分
#    dt = dt_str.split('+')[0]
#    x = datetime.strptime(dt[:-3],'%Y-%m-%dT%H:%M:%S.%f')
#    return x

def net_analyse():
    os.system("grep  'Info network.*sended' ./components.log > ./net.log")
    records = []
    with open('./net.log','r') as logfile:
        l=logfile.readline()
        while l:
            records.append(int(l.split(" ")[0]))
            l=logfile.readline()
    os.system("rm ./net.log")

    intervals = [0]*(records[-1]//10+1)
    t=[i/100 for i in range (0,len(intervals))]
    for i in records:
        intervals[i//10] += 1
    return [t,intervals]

def task_speed_analyse():
    os.system("grep  'TaskGen : send task to' ./components.log > ./task_speed.log")
    records = []
    with open('./task_speed.log','r') as logfile:
        l=logfile.readline()
        while l:
            records.append(int(l.split(" ")[0]))
            l=logfile.readline()
    os.system("rm ./task_speed.log")

    intervals = [0]*(records[-1]//10+1)
    t=[i/100 for i in range (0,len(intervals))]
    for i in records:
        intervals[i//10] += 1
    return [t,intervals]



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
            t.append(int(row[0])/1000)  
            latency = pd.Timedelta(row[1]).total_seconds()*1000
            avg_latency.append(latency)
            avg_cpu.append(float(row[2])*100)
            avg_ram.append(float(row[3])*100)
            var_cpu.append(float(row[4]))
            var_ram.append(float(row[5]))
    
    FONT_SIZE=10
    fig = plt.figure(figsize=(10,15))
    ax1 = fig.add_subplot(311)
    ax1.plot(t,avg_cpu,lw=1,color='r',label="cpu average")
    ax1.plot(t,avg_ram,lw=1,color='b',label="memory average")
    ax1.set_ylabel("resource usage percentage unit: %",fontsize=FONT_SIZE)
    ax1.set_xlabel("time unit: s",fontsize=FONT_SIZE)
    ax1.legend(loc="upper left")

    ax2 = plt.twinx()
    ax2.plot(t,avg_latency,lw=1,color='y',label="task latency")
    ax2.set_ylabel("task lantency unit: ms",fontsize=FONT_SIZE)
    ax2.legend(loc="upper right")
    


    ax3 = fig.add_subplot(312)
    ax3.plot(t,var_cpu,lw=1,label="cpu variance")
    ax3.plot(t,var_ram,lw=1,label="ram variance")
    ax3.set_ylabel("resource variance",fontsize=FONT_SIZE)
    ax3.set_xlabel("time unit: s",fontsize=FONT_SIZE)
    ax3.legend()
    
    res=net_analyse()
    ax4 = fig.add_subplot(313)
    ax4.plot(res[0],res[1],lw=1,label="all type of request")
    ax4.set_ylabel("request rate, unit: amount of recent 10ms",fontsize=FONT_SIZE,)
    ax4.set_xlabel("time unit: s",fontsize=FONT_SIZE)

    res=task_speed_analyse()
    ax4.plot(res[0],res[1],lw=1,label="task submission")
    ax4.legend()


    plt.savefig('./cluster_status.png')


if __name__ == "__main__":
    draw_cluster_status()
