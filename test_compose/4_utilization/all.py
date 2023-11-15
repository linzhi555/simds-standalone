import sys
sys.path.append("../../py")
import draw
import os

def longName(s):
    if s == "center":
        return "Centralized"
    if s == "share":
        return "Share state"
    if s == "dcss":
        return "Beehive"






load=["","80%","83%","85%","88%","90%"]
for cluster in ["center","share","dcss"]:
    tests = []
    for i in range(1,6):
        folder = "./target/test{id}_{c}/".format(c=cluster,id=i)
        tests.append([folder, "max utilization {num}".format(num=load[i])])
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    draw.draw_task_latency_CDF(tests)
    os.system("mkdir -p ./target/all/{c} && mv *.png ./target/all/{c}".format(c=cluster))


