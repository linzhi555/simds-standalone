package main

type Component interface {
	Component() string
}
type Components map[string]Component
type ECS struct {
	Entities map[string]Components
	System   map[string]func(Component)
}

func (e *ECS)AddSystem(name string,f func(Component)){

}

func (e *ECS) AddEntities() {


}
func (e *ECS) Ticks() {
	

}


