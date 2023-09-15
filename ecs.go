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
	System   map[string]func(Component)
}

func NewEcs() *ECS {
	return &ECS{
		Entities:make(map[string]Components),
		System: make(map[string]func(Component)),
	}
}


func (e *ECS)AddSystem(name string,f func(Component)){
	e.System[name]=f

}

func (e *ECS) AddEntities(name string,cs ... Component) error {
	if _,alreadyHas := e.Entities[name];alreadyHas{
		return errors.New("the ecs has already this entity")
	}
	e.Entities[name]=make(Components)

	for _,c := range cs {
		e.Entities[name][c.Component()]=c
	}
	return nil
}

func (e *ECS)AddComponetsToEntity(entity string,c Component)error{

	if _,has := e.Entities[entity]; !has {
		return errors.New("the entity can not be found")
	}
	e.Entities[entity][c.Component()]=c
	return nil
}


func (e *ECS) Ticks() {
	for _,Components :=range e.Entities{
		for componentName,Component := range Components{
			e.System[componentName](Component)
		}
	}
}

func (e *ECS)String()string{
	var s = ""
	for name,Components :=range e.Entities{
		s += name+ "\n"
		for componentName,Component := range Components{
			s += componentName + "\n"
			s += fmt.Sprint(Component) + "\n"
		}
	}
	return s
}
