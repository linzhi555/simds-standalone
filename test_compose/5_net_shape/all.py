import sys
sys.path.append("../../py")
import draw
import os

def longName(s):
    if s == "center":
        return "Centralized"
    if s == "share":
        return "Shared state"
    if s == "dcss":
        return "Beehive"




neiborsP=[0,0.3,0.6,1.0]


for cluster in ["dcss"]:
    tests = []
    for i in range(1,len(neiborsP)+1):
        folder = "./target/test{id}_{c}/".format(c=cluster,id=i)
        p = neiborsP[i-1]
        label = ""
        if p < 0.001:
            label = "Lattice Network"
        else:
            label = "WS Network p={}".format(p)
        tests.append([folder, label])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    draw.draw_task_latency_CDF(tests)
    os.system("mkdir -p ./target/all/{c} && mv *.png ./target/all/{c}".format(c=cluster))


