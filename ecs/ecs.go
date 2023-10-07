package ecs

import (
	"fmt"
	"simds-standalone/common"
)

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
		UpdateCount:0,
		Entities:   make(map[EntityName]componetIndex),
		Components: make(map[ComponentName]ComponentList),
		Systems:    make([]System, 0),
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
		common.AssertTypeIsNotPointer(c)
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
func (ecs *ECS) GetEntitiesHasComponenet(componentNeed ComponentName) []EntityName {
	var result map[EntityName]struct{} = make(map[EntityName]struct{})
	for _, node := range ecs.Components[componentNeed] {
		result[node.belong] = struct{}{}
	}

	keys := make([]EntityName, 0, len(result))
	for en := range result {
		keys = append(keys, en)
	}
	return keys
}

// Get the information of a commponet of entityNeed,rember the ret is a value not a pointer
func (ecs *ECS) GetComponet(entityNeed EntityName, componentNeed ComponentName) (ret Component) {

	index := ecs.Entities[entityNeed][componentNeed]
	return ecs.Components[componentNeed][index].componet

}

func (e *ECS) Update() {
	for _, system := range e.Systems {
		fmt.Println(system.Name)
		system.Function(e)
	}
	e.UpdateCount +=1;
}

//func (e *ECS) String() string {
//	var s = ""
//	for name, Components := range e.Entities {
//		s += "***" + string(name) + "***" + "\n"
//		for _, node := range Components {
//			s += string(node.componet.Component()) + "\n"
//			s += fmt.Sprint(node.componet) + "\n"
//		}
//	}
//	return s
//}
