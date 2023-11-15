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



policy=[0,"always","random",4,7,10]
for cluster in ["dcss"]:
    tests = []
    for i in range(1,3):
        folder = "./target/test{id}_{c}/".format(c=cluster,id=i)
        tests.append([folder, "divide policy: {p}".format(p=policy[i])])
    draw.draw_task_submission_rate(tests)
    draw.draw_muilt_lantencyCurve(tests)
    draw.draw_muilt_avg_resource (tests)
    draw.draw_muilt_var_resource (tests)
    draw.draw_muilt_net_busy (tests)

    draw.draw_task_latency_CDF(tests)

    os.system("mkdir -p ./target/all/{c} && mv *.png ./target/all/{c}".format(c=cluster))


