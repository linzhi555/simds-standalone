package core

// 一个简单的ECS实现
// ECS (entity componet system)是一种广泛用于游戏和模拟的架构

// EntityName 每个实体有唯一的Name
type EntityName string

// ComponentName ECS每个组件有对应的名称，根据名称进行更新等操作
type ComponentName string

// Component 泛型组件
type Component interface {
	Component() ComponentName
}

// ComponentListNode 是 ComponentList 的一个节点，每个节点有具体componet,以及该componet 归属的实体
type ComponentListNode struct {
	componet Component
	belong   EntityName
}

// ComponentList 一种列表，列表中存储相同类型的ComponetNode
type ComponentList []ComponentListNode

// SystemName 系统名
type SystemName string

// System 系统在ECS负责ECS更新的行为
type System struct {
	Name     SystemName
	Function func(*ECS)
}

type componetIndex map[ComponentName]int

// ECS 由 实体，组件，系统组成，UpdateCount 记录了这个ECS更新了多少
type ECS struct {
	UpdateCount uint64
	Entities    map[EntityName]componetIndex
	Components  map[ComponentName]ComponentList
	Systems     []System
}

// NewEcs 创建一个空的ECS
func NewEcs() *ECS {
	return &ECS{
		UpdateCount: 0,
		Entities:    make(map[EntityName]componetIndex),
		Components:  make(map[ComponentName]ComponentList),
		Systems:     make([]System, 0),
	}
}

// AddSystem 为ECS加入System
func (ecs *ECS) AddSystem(name SystemName, f func(*ECS)) {
	ecs.Systems = append(ecs.Systems, System{Name: name, Function: f})
}

// AddEntities 为ECS加入实体
func (ecs *ECS) AddEntities(name EntityName, cs ...Component) {
	if _, alreadyHas := ecs.Entities[name]; alreadyHas {
		panic("the entity is already existed")
	}
	ecs.Entities[name] = make(componetIndex)

	for _, c := range cs {
		ecs.Components[c.Component()] = append(ecs.Components[c.Component()], ComponentListNode{c, name})
		ecs.Entities[name][c.Component()] = len(ecs.Components[c.Component()]) - 1
	}

}

// ApplyToAllComponent 传入函数指针对ecs 所有 某种类型的组件进行并行更新
func (ecs *ECS) ApplyToAllComponent(name ComponentName, f func(ecs *ECS, e EntityName, componet Component) Component) {
	const RenderThreadNum = 100
	finishChan := make(chan bool, RenderThreadNum)
	for i := 0; i < RenderThreadNum; i++ {
		go func(id int) {
			for j := range ecs.Components[name] {
				if j%RenderThreadNum == id {
					entity := ecs.Components[name][j].belong
					ecs.Components[name][j].componet = f(ecs, entity, ecs.Components[name][j].componet)
				}
			}
			finishChan <- true
		}(i)
	}

	for i := 0; i < RenderThreadNum; i++ {
		<-finishChan
	}

}

// Update 调用ECS 所有system 完成一次更新
func (ecs *ECS) Update() {
	for _, system := range ecs.Systems {
		system.Function(ecs)
	}
	ecs.UpdateCount++
}

// Update 调用ECS 所有system 完成n次更新
func (ecs *ECS) UpdateNtimes(n uint64) {
	for i := uint64(0); i < n; i++ {
		ecs.Update()
	}
}
