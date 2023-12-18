import os
import shutil
INPUT_DIR= "./net_shape_test/target"
OUTPUT_DIR="./net_shape_test/epic"

CONVERTIONS=[
    ("node_num/all/lantency_compare.png",     "nodes_1k_lantency_compare.png"),
    ("node_num/all/lantency_compare.png",     "nodes_4k_lantency_compare.png"),
    ("node_num/all/lantency_compare.png",     "nodes_7k_lantency_compare.png"),
    ("node_num/all/lantency_compare.png",    "nodes_10k_lantency_compare.png"),
    ######

    ("all/neibor_random_p/lantency_compare.png",     "neibor_random_p_lantency_compare.png"),
    ("all/neibor_random_p/latency_CDF_compare.png",     "neibor_random_p_latency_CDF_compare.png"),
    #("node_num/all/lantency_compare.png",     "nodes_4k_lantency_compare.png"),
    #("node_num/all/lantency_compare.png",     "nodes_7k_lantency_compare.png"),
    #("node_num/all/lantency_compare.png",    "nodes_10k_lantency_compare.png"),
    #######



    #("./test_compose/1_node_num_all/nodes_1k/latency_CDF_compare.png",          "nodes_1k_latency_CDF_compare.png"),
    #("./test_compose/1_node_num_all/nodes_4k/latency_CDF_compare.png",          "nodes_4k_latency_CDF_compare.png"),
    #####
    #("./test_compose/1_node_num_all/nodes_1k/net_busy_compare_cluster.png",    "nodes_1k_net_busy_compare.png"),
    #("./test_compose/1_node_num_all/nodes_4k/net_busy_compare_cluster.png",    "nodes_4k_net_busy_compare.png"),
    #("./test_compose/1_node_num_all/nodes_1k/net_busy_compare_most_busy.png",  "nodes_1k_net_busy_compare_most_busy.png"),
    #("./test_compose/1_node_num_all/nodes_4k/net_busy_compare_most_busy.png",  "nodes_4k_net_busy_compare_most_busy.png"),
    ##### 
    #("./test_compose/1_node_num_all/nodes_1k/variance_compare.png",         "nodes_1k_variance_compare.png"),
    #("./test_compose/1_node_num_all/nodes_4k/variance_compare.png",         "nodes_4k_variance_compare.png"),
    #####
    #("./test_compose/1_node_num_all/dcss/latency_CDF_compare.png",          "dcss_latency_CDF_compare.png"),
    #("./test_compose/1_node_num_all/dcss/net_busy_compare_cluster.png",     "dcss_net_busy_compare.png"),
    #("./test_compose/1_node_num_all/dcss/load_compare.png",                 "dcss_load_compare.png" ),
    #("./test_compose/1_node_num_all/dcss/variance_compare.png",             "dcss_variance_compare.png" ),
    #####
    #("./test_compose/3_neibors_all/dcss/latency_CDF_compare.png",           "3_neighbors_all_latency_CDF_compare.png"   ),
    #("./test_compose/3_neibors_all/dcss/net_busy_compare_cluster.png",      "3_neibors_all_dcss_net_busy_compare.png" ),
    ##### 
    #("./test_compose/6_net_latency_all/dcss/latency_CDF_compare.png",       "6_net_latency_all_latency_CDF_compare.png" ),
    #("./test_compose/6_net_latency_all/dcss/net_busy_compare_cluster.png",  "6_net_latency_all_net_busy_compare.png" ),
    ##### 
    #("./test_compose/2_divide_policy_all/dcss/latency_CDF_compare.png",     "2_divide_policy_latency_CDF_compare.png"   ),
    #("./test_compose/2_divide_policy_all/dcss/net_busy_compare_cluster.png","2_divide_policy_net_busy_compare.png" ),
    ##### 
    #("./test_compose/4_utilization_all/dcss/latency_CDF_compare.png",       "4_utilization_all_latency_CDF_compare.png"  ),
    #("./test_compose/4_utilization_all/dcss/net_busy_compare_cluster.png",  "4_utilization_all_net_busy_compare.png" ),
]

if not os.path.exists(OUTPUT_DIR):
    os.mkdir(OUTPUT_DIR)

for convertion in CONVERTIONS:
    inputFile = os.path.join(INPUT_DIR,convertion[0])
    outputFile= os.path.join(OUTPUT_DIR,convertion[1])
    shutil.copyfile(inputFile,outputFile)



