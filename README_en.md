# Distributed cluster simulator stand-alone version

simds-standalone(simulator of distribute cluster - standalone edition)
It can be used to simulate the task scheduling, task running, task information communication and other behaviors of a distributed cluster. Supports both centralized clustering and
Compared.

## Instructions

- Install Go >= 1.19, make
- ``` go get ``` install dependencies
- Modify the config.yaml content as needed
- Run a certain test
    - Centralized cluster testing ``` make centerTest ```
    - Run state sharing cluster test ```make shareTest ```
    - Distributed cluster test ```make dcssTest ```
- The log information of all components during the simulation process is in ./components.log
- Simulation logs of all tasks (including submission, start, end) are saved in ./tasks_event.log
- The analysis results are in ./target/{experiment completion time}/

## Introduction to simulator principles
Time model: The simulator uses the tick mechanism to update the cluster status. The simulation time increases by 0.1ms after each tick. Execute all components of the cluster every tick
Update function to complete a status update of the cluster. Each update updates all components of the cluster in parallel (no locks), improving simulation efficiency.

Node model: Each node in the cluster is an entity, and each entity has one or more components. The visual understanding is that there can be multiple APPs on a node, each
Components have input and output pipes that symbolize network interfaces. There are multiple types of components, namely Scheduler, Taskgen, ResourceManager, StateStorage,
These components store information that may be used by corresponding functions. There are two methods for each type of component in the cluster, Setup and Update. Setup is used for cluster initialization, and
Update is the update function called each tick mentioned earlier. You can change the Setup and Update methods registered to each component to achieve changes in cluster behavior.

Network model: The cluster has a virtual entity named networker1, which has the MockNetworke component. This component is network-connected to all other node components in the cluster.
That is, this component is the receiving end of all component output pipes and the sending end of all input pipes. When a component is sent to another node within an update function, this
After the information is stuffed into the pipeline, the next time the MockNetworke component executes the update function, the information will be retrieved, stuffed into the cache, and the sender and receiver of the information will be checked to determine this
The information is cached for how many ticks to simulate network delay, and then inserted into the corresponding receiver pipeline according to the receiver address. If the receiver and sender belong to the same entity,
The person sends directly without waiting (local communication).

