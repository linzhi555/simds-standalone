package main

var centralizedSystem []System

func addCentralizedSystem(n SystemName, f func(*ECS)) {
	centralizedSystem = append(centralizedSystem, System{n, f})
}
func RegisteCentralizedsystemToEcs(e *ECS) {
	for _, s := range commonSystem {
		e.AddSystem(s.Name, s.Function)
	}
	for _, s := range centralizedSystem {
		e.AddSystem(s.Name, s.Function)
	}
}

const SSchedulerUpdate = "SchedulerUpdateSystem"

func init() { addCentralizedSystem(SSchedulerUpdate, SchedulerUpdateSystem) }
func SchedulerUpdateSystem(ecs *ECS) {
	ecs.ApplyToAllComponent("Scheduler", SchedulerTicks)
}

const schedulerDelay = 10 * MiliSecond

func SchedulerTicks(ecs *ECS, entity EntityName, c Component) {
	scheduler := c.(*Scheduler)
	timeNow := GetEntityTime(ecs, entity)

	for !scheduler.Net.In.Empty() {
		newMessage, err := scheduler.Net.In.Dequeue()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "TaskDispense" {
			task := newMessage.Body.(TaskInfo)
			task.InQueneTime = timeNow
			task.Status = "WaitSchedule"
			scheduler.Tasks[task.Id] = &task
			LogInfo(ecs, entity, scheduler.Net.Addr, "received TaskDispense", task)
		}

		if newMessage.Content == "TaskFinish" {
			taskInfo := newMessage.Body.(TaskInfo)
			scheduler.Workers[newMessage.From].SubAllocated(taskInfo.CpuRequest, taskInfo.MemoryRequest)
			LogInfo(ecs, entity, scheduler.Net.Addr, "received TaskFinish", newMessage.From, taskInfo)
		}

	}

	for taskid, task := range scheduler.Tasks {
		switch task.Status {
		case "WaitSchedule":
			dstWorker, ok := schdulingAlgorithm(scheduler, task)
			if ok {
				task.Worker = dstWorker
				task.Status = "Scheduling"
				scheduler.Workers[task.Worker].AddAllocated(task.CpuRequest, task.MemoryRequest)
			} else {
			}
			break

		case "Scheduling":
			if timeNow-task.InQueneTime > schedulerDelay {
				task.Status = "Scheduled"
			}
			break
		case "Scheduled":
			newMessage := Message{
				From:    scheduler.Net.Addr,
				To:      task.Worker,
				Content: "TaskAllocate",
				Body:    *task,
			}
			scheduler.Net.Out.InQueue(newMessage)
			task.Status = "Allocated"
			LogInfo(ecs, entity, scheduler.Net.Addr, "sendtask to", task.Worker, task)
			break
		case "Allocated":
			delete(scheduler.Tasks, taskid)
			break
		default:
			panic("wrong task status")
		}
	}

}

// schedule the task,if it  can not find a worker for the task,return "",false
// else return "addr of some worker",true
func schdulingAlgorithm(scheduler *Scheduler, task *TaskInfo) (dstAddr string, ok bool) {
	dstAddr = ""

	keys := make([]string, 0, len(scheduler.Workers))
	for k := range scheduler.Workers {
		keys = append(keys, k)
	}
	shuffleStringSlice(keys)
	for _, workerAddr := range keys {
		nodeinfo := scheduler.Workers[workerAddr]
		if nodeinfo.CanAllocate(task.CpuRequest, task.MemoryRequest) {
			dstAddr = workerAddr
		}
	}

	if dstAddr == "" {
		return dstAddr, false
	}
	return dstAddr, true
}
