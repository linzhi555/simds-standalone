package main

import (
	"errors"
	"fmt"
)

type Component interface {
	Component() string
}
type Components map[string]Component
type ECS struct {
	Entities map[string]Components
	System   map[string]func(*ECS)
}

func NewEcs() *ECS {
	return &ECS{
		Entities: make(map[string]Components),
		System:   make(map[string]func(*ECS)),
	}
}

func (e *ECS) AddSystem(name string, f func(*ECS)) {
	e.System[name] = f

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

func (e *ECS) Update() {
	for _, f := range e.System {
		f(e)
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
