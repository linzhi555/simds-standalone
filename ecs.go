package main

type EntityName string
type ComponentName string
type Component interface {
	Component() ComponentName
}

type ComponentListNode struct {
	componet Component
	belong   EntityName
}

type ComponentList []ComponentListNode

type SystemName string
type System struct {
	Name     SystemName
	Function func(*ECS)
}

type componetIndex map[ComponentName]int

type ECS struct {
	UpdateCount uint64
	Entities    map[EntityName]componetIndex
	Components  map[ComponentName]ComponentList
	Systems     []System
}

func NewEcs() *ECS {
	return &ECS{
		UpdateCount: 0,
		Entities:    make(map[EntityName]componetIndex),
		Components:  make(map[ComponentName]ComponentList),
		Systems:     make([]System, 0),
	}
}

func (e *ECS) AddSystem(name SystemName, f func(*ECS)) {
	e.Systems = append(e.Systems, System{Name: name, Function: f})
}

func (e *ECS) AddEntities(name EntityName, cs ...Component) {
	if _, alreadyHas := e.Entities[name]; alreadyHas {
		panic("the entity is already existed")
	}
	e.Entities[name] = make(componetIndex)

	for _, c := range cs {
		e.Components[c.Component()] = append(e.Components[c.Component()], ComponentListNode{c, name})
		e.Entities[name][c.Component()] = len(e.Components[c.Component()]) - 1
	}

}

func (ecs *ECS) ApplyToAllComponent(name ComponentName, f func(ecs *ECS, e EntityName, componet Component) Component) {
	const RenderThreadNum = 100
	finishChan := make(chan bool, RenderThreadNum)
	for i := 0; i < RenderThreadNum; i++ {

		go func(id int) {
			for j, _ := range ecs.Components[name] {
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

func (e *ECS) Update() {
	for _, system := range e.Systems {
		system.Function(e)
	}
	e.UpdateCount += 1
}
