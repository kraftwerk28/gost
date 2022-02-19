package blocks

type UpdateChan chan int

func (u *UpdateChan) SendUpdate() {
	*u <- 0
}

func CombineUpdateChans(chans []UpdateChan) UpdateChan {
	ch := UpdateChan(make(chan int))
	for i := range chans {
		go func(c UpdateChan) {
			for {
				v := <-c
				ch <- v
			}
		}(chans[i])
	}
	return ch
}

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
