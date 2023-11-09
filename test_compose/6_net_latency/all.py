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


nodes_num=[0,1,2,4,8,16]
for cluster in ["center","share","dcss"]:
    tests = []
    for i in range(1,6):
        folder = "./target/test{id}_{c}/".format(c=cluster,id=i)
        tests.append([folder, "net latency {num}ms".format(num=nodes_num[i])])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)
    os.system("mkdir -p ./target/all/{c} && mv *.png ./target/all/{c}".format(c=cluster))


