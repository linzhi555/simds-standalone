import csv
import matplotlib.pyplot as plt
from datetime import datetime
import pandas as pd
import os

FONT_SIZE=20

#def parseRFCnano(dt_str):
#    # 解析日期时间部分
#    dt = dt_str.split('+')[0]
#    x = datetime.strptime(dt[:-3],'%Y-%m-%dT%H:%M:%S.%f')
#    return x

def net_commuication_rate_curves(filename):
    #os.system("grep  'Info network.*sended' ./components.log > ./net.log")
    records = []
    with open(filename,'r') as logfile:
        l=logfile.readline()
        while l:
            records.append(int(l.split(" ")[0]))
            l=logfile.readline()
    intervals = [0]*(records[-1]//10+1)
    t=[i/100 for i in range (0,len(intervals))]
    for i in records:
        intervals[i//10] += 1
    intervals = [100 * i for i in intervals]
    return [t,intervals]

def task_submit_rate_curves(filename):
    #os.system("grep  'TaskGen : send task to' ./components.log > ./task_speed.log")
    records = []
    with open(filename,'r') as logfile:
        l=logfile.readline()
        while l:
            records.append(int(l.split(" ")[0]))
            l=logfile.readline()
    intervals = [0]*(records[-1]//10+1)
    t=[i/100 for i in range (0,len(intervals))]
    for i in records:
        intervals[i//10] += 1
    intervals = [100 * i for i in intervals]
    return [t,intervals]




def cluster_status_curves(filename):
    t = []
    avg_latency = []
    avg_cpu = []
    avg_ram = []
    var_cpu = []
    var_ram = []
    with open(filename,'r') as csvfile:
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
    return [t,avg_latency,avg_cpu,avg_ram,var_cpu,var_ram]



def draw_cluster_status():
    
    fig = plt.figure(figsize=(10,15))
    ax1 = fig.add_subplot(311)
    status = cluster_status_curves("./cluster_status.log")
    t=status[0]
    avg_latency= status[1]
    avg_cpu=status[2]
    avg_ram=status[3]
    var_cpu=status[4]
    var_ram=status[5]

    ax1.plot(t,avg_cpu,lw=1,color='r',label="cpu average")
    ax1.plot(t,avg_ram,lw=1,color='b',label="memory average")
    ax1.set_ylabel("resource usage percentage unit: %",fontsize=FONT_SIZE)
    ax1.set_xlabel("time unit: s",fontsize=FONT_SIZE)
    ax1.legend(loc="upper left")

    ax2 = plt.twinx()
    ax2.plot(t,avg_latency,lw=1,color='y',label="task latency")
    ax2.set_ylabel("task lantency unit: ms",fontsize=FONT_SIZE)
    ax2.legend(loc="upper right")

    if max(avg_latency)>10000:
        ax2.set_yscale("log",base=10)

    ax3 = fig.add_subplot(312)
    ax3.plot(t,var_cpu,lw=1,label="cpu variance")
    ax3.plot(t,var_ram,lw=1,label="ram variance")
    ax3.set_ylabel("resource variance",fontsize=FONT_SIZE)
    ax3.set_xlabel("time unit: s",fontsize=FONT_SIZE)
    ax3.legend()
    
    res=net_commuication_rate_curves("./net.log")
    ax4 = fig.add_subplot(313)
    ax4.plot(res[0],res[1],lw=1,color='y',label="all type of request")
    ax4.set_ylabel("request rate, (amount/s)",fontsize=FONT_SIZE,)
    ax4.set_xlabel("time (s)",fontsize=FONT_SIZE)
    ax4.legend(loc="upper left")

    res=task_submit_rate_curves("./task_speed.log")
    ax5 = plt.twinx()
    ax5.plot(res[0],res[1],lw=1,color='b',label="task submission")
    ax5.set_ylabel("task submission rate, (amount/s)",fontsize=FONT_SIZE,)
    ax5.legend()
    plt.savefig('./cluster_status.png')

def draw_muilt_lantencyCurve(tests):
    plt.cla()

    hasFailTask=False
    for t in tests:
        status = cluster_status_curves(os.path.join(t[0],"cluster_status.log"))
        #staus[0] = staus[0][::10]
        #staus[1] = staus[1][::10]
        plt.plot(status[0],status[1],lw=1,label=t[1])
        if max(status[1])>20000:
            hasFailTask=True
    if hasFailTask:
        plt.yscale("log",base=10)
        #y=list(range(0,2000,200))
        #yticks=y.copy()
        #ylabels=y.copy()
        #ylabels[-1] = r"$\infty$"
        #_,axes = plt.subplots()
        #axes.set_yticks(yticks)
        #axes.set_yticklabels(ylabels)


    plt.legend()
    plt.ylabel("worst task lantency (ms)",fontsize=FONT_SIZE)
    plt.xlabel("time (s)",fontsize=FONT_SIZE)

    plt.savefig('./lantency_compare.png')

def draw_muilt_avg_resource(tests):
    plt.cla()
    for t in tests:
        staus = cluster_status_curves(os.path.join(t[0],"cluster_status.log"))
        plt.plot(staus[0],staus[2],lw=1,label=t[1])
    plt.legend()
    plt.ylabel("resource utilization (%)",fontsize=FONT_SIZE)
    plt.xlabel("time (s)",fontsize=FONT_SIZE)

    plt.savefig('./load_compare.png')

def draw_muilt_var_resource(tests):
    plt.cla()
    for t in tests:
        staus = cluster_status_curves(os.path.join(t[0],"cluster_status.log"))
        plt.plot(staus[0],staus[4],lw=1,label=t[1])
    plt.legend()
    plt.ylabel("cluster utilization variance",fontsize=FONT_SIZE)
    plt.xlabel("time (s)",fontsize=FONT_SIZE)
    plt.savefig('./variance_compare.png')

def draw_muilt_net_busy(tests):
    plt.cla()
    for t in tests:
        staus = net_commuication_rate_curves(os.path.join(t[0],"net.log"))
        plt.plot(staus[0],staus[1],lw=1,label=t[1])
    plt.legend()
    plt.ylabel("net request rate \n (number/s)",fontsize=FONT_SIZE)
    plt.subplots_adjust(left=0.2) 
    plt.xlabel("time (s)",fontsize=FONT_SIZE)
    plt.savefig('./net_busy_compare.png')

def draw_task_submission_rate(tests):
    plt.cla()
    staus = task_submit_rate_curves(os.path.join(tests[0][0],"./task_speed.log"))
    plt.plot(staus[0],staus[1],lw=1)
    plt.ylabel("task submission rate \n (number/s)",fontsize=FONT_SIZE)
    plt.subplots_adjust(left=0.2) 
    plt.xlabel("time (s)",fontsize=FONT_SIZE)
    plt.savefig('./task_submission_rate.png')



if __name__ == "__main__":
    draw_cluster_status()