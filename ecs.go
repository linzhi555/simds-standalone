package main

import (
	"fmt"
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

type ComponentList []*ComponentListNode

type SystemName string
type System struct {
	Name     SystemName
	Function func(*ECS)
}

type ECS struct {
	Entities   map[EntityName]ComponentList
	Components map[ComponentName]ComponentList
	Systems    []System
}

func NewEcs() *ECS {
	return &ECS{
		Entities:   make(map[EntityName]ComponentList),
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
	e.Entities[name] = ComponentList{}

	for _, c := range cs {
		newNode := ComponentListNode{c, name}
		if _, ok := e.Components[c.Component()]; !ok {
			e.Components[c.Component()] = ComponentList{&newNode}
		} else {
			e.Components[c.Component()] = append(e.Components[c.Component()], &newNode)
		}
		e.Entities[name] = append(e.Entities[name], &newNode)
	}

}

func (ecs *ECS) ApplyToAllComponent(name ComponentName, f func(ecs *ECS, entity EntityName, component Component)) {
	threadNum := 0
	finishChan := make(chan bool, len(ecs.Components[name]))
	for _, node := range ecs.Components[name] {
		threadNum += 1
		go func(c Component, e EntityName) {
			f(ecs, e, c)
			finishChan <- true
		}(node.componet, node.belong)
	}
	for i := 0; i < threadNum; i++ {
		<-finishChan
	}

}
func (ecs *ECS) GetEntitiesHasComponenet(componentNeed ComponentName) []EntityName {
	var result []EntityName
	for e, components := range ecs.Entities {
		for _, node := range components {
			if node.componet.Component() == componentNeed {
				result = append(result, e)
				break
			}
		}
	}
	return result
}

func (ecs *ECS) GetComponet(entityNeed EntityName, componentNeed ComponentName) (ret Component) {
	for e, components := range ecs.Entities {
		if e != entityNeed {
			continue
		}

		for _, node := range components {
			if node.componet.Component() == componentNeed {
				return node.componet
			}
		}
	}
	panic("the entity" + string(entityNeed) + "dones not have" + string(componentNeed))
}

func (e *ECS) Update() {
	for _, system := range e.Systems {
		fmt.Println(system.Name)
		system.Function(e)
	}
}

func (e *ECS) String() string {
	var s = ""
	for name, Components := range e.Entities {
		s += "***" + string(name) + "***" + "\n"
		for _, node := range Components {
			s += string(node.componet.Component()) + "\n"
			s += fmt.Sprint(node.componet) + "\n"
		}
	}
	return s
}
