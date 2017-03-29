package main

import (
	"math"
	"math/rand"
)


type Simulator interface {
	SetWorld(World)
	Tick(float64)
}

type World interface {
	ActorSpawned(*Actor)
	StaticSpawned(*Static)
}

type Emitter interface {
	GetPos() (x, y float64)
}

type Static struct {
	PosX float64
	PosY float64
}

func (a *Static) GetPos() (x, y float64) {
	return a.PosX, a.PosY
}

type Actor struct {
	PosX float64
	PosY float64
	Heading float64
	TargetAngle float64

	attractor Emitter
	repulsors []Emitter
}

func (a *Actor) GetPos() (float64, float64) {
	return a.PosX, a.PosY
}

func (a *Actor) ProbeRepulsingField(x, y float64) (fx, fy float64) {
	dx := a.PosX - x
	dy := a.PosY - y
	const minDistanceNoticeable = 150
	const fieldScalingFactor = 0.2
	distance := math.Hypot(dx, dy)
	if distance > minDistanceNoticeable {
		return 0, 0
	}
	scalarNorm := (minDistanceNoticeable / distance - 1) * fieldScalingFactor
	return dx * scalarNorm, dy * scalarNorm
}

func (a *Actor) ProbeAttractingField(x, y float64) (fx, fy float64) {
	dx := x - a.PosX
	dy := y - a.PosY
	distance := math.Hypot(dx, dy)
	const attractingPower = 10
	scalarNorm := (1 / distance) * attractingPower
	return dx * scalarNorm, dy * scalarNorm
}

func angleBetweenAngles(a1, a2 float64) float64 {
	res := a2 - a1 + 180
	for res > 360 {
		res -= 360
	}
	return res - 180
}

func (a *Actor) PlanAhead() {
	ax, ay := a.attractor.GetPos()
	fx, fy := a.ProbeAttractingField(ax, ay)
	for _, r := range a.repulsors {
		rx, ry := r.GetPos()
		rfx, rfy := a.ProbeRepulsingField(rx, ry)
		fx += rfx
		fy += rfy
	}
	direction := math.Atan2(fy, fx) / math.Pi * 180 + 90
	for direction < 0 {
		direction += 360
	}
	a.TargetAngle = direction
}

func (a *Actor) AddRepulsor(r Emitter) {
	a.repulsors = append(a.repulsors, r)
}

func (a *Actor) SetAttractor(e Emitter) {
	a.attractor = e
}

type Sim struct {
	world World
	actors []*Actor
}

func (s *Sim) SetWorld(w World) {
	s.world = w

	cx, cy := 500.0, 220.0
	attr := &Static{500, 300}

	statics := make([]*Static, 100)
	for i := range statics {
		angle := (math.Pi * 2.0 / float64(len(statics))) * float64(i)
		sx := cx + math.Hypot(cx, cy) * math.Sin(angle)
		sy := cy + math.Hypot(cx, cy) * math.Cos(angle)
		st := &Static{sx, sy}
		statics[i] = st
		w.StaticSpawned(st)
	}

	s.actors = make([]*Actor, 6)
	for i := range s.actors {
		a := new(Actor)
		a.PosX = float64(rand.Intn(int(cx * 2 + 1)))
		a.PosY = float64(rand.Intn(int(cy * 2 + 1)))
		a.Heading = float64(rand.Intn(360))
		s.actors[i] = a
		for _, st := range statics {
			a.AddRepulsor(st)
		}
	}

	for i, a := range s.actors {
		if i == 0 {
			a.SetAttractor(s.actors[len(s.actors) - 1])
		} else {
			a.SetAttractor(s.actors[i - 1])
		}
		a.SetAttractor(attr)
		for j, aa := range s.actors {
			if i != j {
				a.AddRepulsor(aa)
			}
		}
	}

	for _, a := range s.actors {
		w.ActorSpawned(a)
	}

}

func (s *Sim) Tick(timePassed float64) {
	const speedScale = 2.0
	speed := 90.0 * speedScale
	moved := speed * timePassed
	for _, a := range s.actors {
		a.PlanAhead()

		aba := angleBetweenAngles(a.Heading, a.TargetAngle)
		const maxDegreesPerSecond = 120 * speedScale
		maxDegrees := maxDegreesPerSecond * timePassed
		if aba > maxDegrees {
			aba = maxDegrees
		}
		if aba < - maxDegrees {
			aba = - maxDegrees
		}
		a.Heading += aba

		dx := math.Sin(a.Heading / 180 * math.Pi)
		dy := - math.Cos(a.Heading / 180 * math.Pi)
		a.PosX += dx * moved
		a.PosY += dy * moved
	}
}
