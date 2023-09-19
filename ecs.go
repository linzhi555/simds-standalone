package main

import (
	"errors"
	"fmt"
)

type EntityName string
type ComponentName string

type Component interface {
	Component() ComponentName
}
type Components map[ComponentName]Component

type SystemName string
type System struct {
	Name     SystemName
	Function func(*ECS)
}

type ECS struct {
	Entities map[EntityName]Components
	Systems  []System
}

func NewEcs() *ECS {
	return &ECS{
		Entities: make(map[EntityName]Components),
		Systems:  make([]System, 0),
	}
}

func (e *ECS) AddSystem(name SystemName, f func(*ECS)) {
	e.Systems = append(e.Systems, System{Name: name, Function: f})
}

func (e *ECS) AddEntities(name EntityName, cs ...Component) error {
	if _, alreadyHas := e.Entities[name]; alreadyHas {
		return errors.New("the ecs has already this entity")
	}
	e.Entities[name] = make(Components)

	for _, c := range cs {
		e.Entities[name][c.Component()] = c
	}
	return nil
}

func (e *ECS) AddComponetsToEntity(entity EntityName, cs ...Component) error {

	if _, has := e.Entities[entity]; !has {
		return errors.New("the entity can not be found")
	}

	for _, c := range cs {
		e.Entities[entity][c.Component()] = c
	}

	return nil
}

func (ecs *ECS) ApplyToAllComponent(name ComponentName, f func(ecs *ECS, entity EntityName, component Component)) {
	for e, components := range ecs.Entities {
		for n, c := range components {
			if n == name {
				f(ecs, e, c)
			}
		}
	}

}
func (ecs *ECS) GetEntitiesHasComponenet(componentNeed ComponentName) []EntityName {
	var result []EntityName
	for e, components := range ecs.Entities {
		for n := range components {
			if n == componentNeed {
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
		for n, c := range components {
			if n == componentNeed {
				return c
			}
		}
	}
	panic("the entity" + string(entityNeed) + "dones not have" + string(componentNeed))
	return nil
}

func (e *ECS) Update() {
	for _, system := range e.Systems {
		system.Function(e)
	}
}

func (e *ECS) String() string {
	var s = ""
	for name, Components := range e.Entities {
		s += string(name) + "\n"
		for componentName, Component := range Components {
			s += string(componentName) + "\n"
			s += fmt.Sprint(Component) + "\n"
		}
	}
	return s
}
