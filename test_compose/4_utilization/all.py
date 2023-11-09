import sys
sys.path.append("../../py")
import draw
import os

def longName(s):
    if s == "center":
        return "centralized"
    if s == "share":
        return "share state"
    if s == "dcss":
        return "distributed"


task_rate=[0,6.0,6.2,6.4,6.6,6.8]
for cluster in ["center","share","dcss"]:
    tests = []
    for i in range(1,6):
        folder = "./target/test{id}_{c}/".format(c=cluster,id=i)
        tests.append([folder, "task rate factor {num}".format(num=task_rate[i])])
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    os.system("mkdir -p ./target/all/{c} && mv *.png ./target/all/{c}".format(c=cluster))


