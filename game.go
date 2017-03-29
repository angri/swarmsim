package main

import (
	"image/color"

	"math"
	"math/rand"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

type Scene struct{
	backgroundColor *color.RGBA
	swarmSystem *SwarmSystem
	sim Simulator
	world *ecs.World
}

func (*Scene) Preload() {}

type ActorEntity struct {
	*Actor

	*ecs.BasicEntity
	*common.RenderComponent
	*common.SpaceComponent
}

type StaticEntity struct {
	*Static

	*ecs.BasicEntity
	*common.RenderComponent
	*common.SpaceComponent
}

func randomColor(contrastTo *color.RGBA) *color.RGBA {
	randByte := func() uint8 { return uint8(rand.Intn(256)) }
	colorVal := func(v uint8) float64 { return float64(v) / 255 }
	getLuminance := func (c *color.RGBA) float64 {
		return colorVal(c.R) * 0.3 + colorVal(c.G) * 0.59 + colorVal(c.B) * 0.11
	}
	ctl := getLuminance(contrastTo)
	for {
		c := color.RGBA{randByte(), randByte(), randByte(), 255}
		if math.Abs(getLuminance(&c) - ctl) > 0.3 {
			return &c
		}
	}
}

func (s *Scene) addActor(w *ecs.World, actor *Actor) {
	actorEntity := new(ActorEntity)
	actorEntity.Actor = actor
	basicEntity := ecs.NewBasic()
	actorEntity.BasicEntity = &basicEntity
	actorEntity.SpaceComponent = &common.SpaceComponent{
		Position: engo.Point{0, 0}, Width: 60, Height: 60,
	}
	actorEntity.RenderComponent = &common.RenderComponent{
		Drawable: common.ComplexTriangles{
			Points: []engo.Point{{-0.25, 0.7}, {0, 0}, {0.25, 0.7}, },
		},
		Color: randomColor(s.backgroundColor),
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(actorEntity.BasicEntity, actorEntity.RenderComponent,
				actorEntity.SpaceComponent)
		case *SwarmSystem:
			sys.Add(actorEntity)
		}
	}
}

func (s *Scene) addStatic(w *ecs.World, static *Static) {
	staticEntity := new(StaticEntity)
	staticEntity.Static = static
	basicEntity := ecs.NewBasic()
	staticEntity.BasicEntity = &basicEntity
	staticEntity.SpaceComponent = &common.SpaceComponent{
		Position: engo.Point{float32(static.PosX), float32(static.PosY)},
		Width: 10, Height: 10,
	}
	staticEntity.RenderComponent = &common.RenderComponent{
		Drawable: common.Circle{}, Color: &color.RGBA{180, 180, 180, 255},
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(staticEntity.BasicEntity, staticEntity.RenderComponent,
				staticEntity.SpaceComponent)
		}
	}
}

func (s *Scene) ActorSpawned(actor *Actor) {
	s.addActor(s.world, actor)
}

func (s *Scene) StaticSpawned(static *Static) {
	s.addStatic(s.world, static)
}

func (s *Scene) Setup(w *ecs.World) {
	s.world = w

	s.backgroundColor = &color.RGBA{35, 35, 35, 255}
	common.SetBackground(s.backgroundColor)

	w.AddSystem(&common.RenderSystem{})
	s.swarmSystem = new(SwarmSystem)
	s.swarmSystem.SetSimulator(s.sim)
	w.AddSystem(s.swarmSystem)

	s.sim.SetWorld(s)
}

func (*Scene) Type() string {
	return "GameWorld"
}

type SwarmSystem struct {
	entities []*ActorEntity
	sim Simulator
}

func (s *SwarmSystem) Add(shape *ActorEntity) {
	s.entities = append(s.entities, shape)
}

func (s *SwarmSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range s.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}

	if delete >= 0 {
		s.entities = append(s.entities[:delete], s.entities[delete+1:]...)
	}
}

func (s *SwarmSystem) SetSimulator(sim Simulator) {
	s.sim = sim
}

func (s *SwarmSystem) Update(dt float32) {
	s.sim.Tick(float64(dt))

	for _, e := range s.entities {
		e.SpaceComponent.Position.X = float32(e.Actor.PosX)
		e.SpaceComponent.Position.Y = float32(e.Actor.PosY)
		e.SpaceComponent.Rotation = float32(e.Actor.Heading)
	}
}

func (s *Scene) RunGame(sim Simulator) {
	opts := engo.RunOptions{
		Title:          "swarm",
		Width:          1000,
		Height:         440,
		StandardInputs: true,
	}

	s.sim = sim
	engo.Run(opts, s)
}
