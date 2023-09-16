package main

import (
	"errors"
	"fmt"
)

type Component interface {
	Component() string
}
type Components map[string]Component

type System struct {
	Name     string
	Function func(*ECS)
}

type ECS struct {
	Entities map[string]Components
	Systems  []System
}

func NewEcs() *ECS {
	return &ECS{
		Entities: make(map[string]Components),
		Systems:  make([]System, 0),
	}
}

func (e *ECS) AddSystem(name string, f func(*ECS)) {
	e.Systems = append(e.Systems, System{Name: name, Function: f})

}

func (e *ECS) AddEntities(name string, cs ...Component) error {
	if _, alreadyHas := e.Entities[name]; alreadyHas {
		return errors.New("the ecs has already this entity")
	}
	e.Entities[name] = make(Components)

	for _, c := range cs {
		e.Entities[name][c.Component()] = c
	}
	return nil
}

func (e *ECS) AddComponetsToEntity(entity string, cs ...Component) error {

	if _, has := e.Entities[entity]; !has {
		return errors.New("the entity can not be found")
	}

	for _, c := range cs {
		e.Entities[entity][c.Component()] = c
	}

	return nil
}

func (ecs *ECS) ComponentTick(name string, f func(ecs *ECS, entity string, component Component)) {
	for e, components := range ecs.Entities {
		for n, c := range components {
			if n == name {
				f(ecs, e, c)
			}
		}
	}

}

func (ecs *ECS) GetComponetOfEntity(entityNeed, componentNeed string) (ret Component, ok bool) {
	for e, components := range ecs.Entities {
		if e != entityNeed {
			continue
		}
		for n, c := range components {
			if n == componentNeed {
				return c, true
			}
		}
	}
	return nil, false
}

func (e *ECS) Update() {
	for _, system := range e.Systems {
		system.Function(e)
	}
}

func (e *ECS) String() string {
	var s = ""
	for name, Components := range e.Entities {
		s += name + "\n"
		for componentName, Component := range Components {
			s += componentName + "\n"
			s += fmt.Sprint(Component) + "\n"
		}
	}
	return s
}
