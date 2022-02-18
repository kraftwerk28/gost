package blocks

type UpdateChan <-chan int

type I3barBlocklet interface {
	Render() []I3barBlock
}

type I3barBlockletListener interface {
	I3barBlocklet
	OnEvent(event *I3barClickEvent)
}

type I3barBlockletRun interface {
	I3barBlocklet
	Run()
}

type I3barBlockletAutoUpdater interface {
	I3barBlocklet
	UpdateChan() UpdateChan
}
