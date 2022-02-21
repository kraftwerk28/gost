package core

type UpdateChan chan int

type I3barBlockletCtor func() I3barBlocklet

var Builtin = map[string]I3barBlockletCtor{}

func RegisterBlocklet(name string, ctor I3barBlockletCtor) {
	Builtin[name] = ctor
}

// TODO: send some more information, not just integer
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

type I3barBlockletConfigurable interface {
	I3barBlocklet
	GetConfig() interface{}
}

type I3barBlockletListener interface {
	I3barBlocklet
	OnEvent(event *I3barClickEvent)
}

type I3barBlockletRunnable interface {
	I3barBlocklet
	Run()
}

type I3barBlockletSelfUpdater interface {
	I3barBlocklet
	UpdateChan() UpdateChan
}

type PluginLoadFunc func() I3barBlocklet

// A helper wrapper around a blocklet
type BlockletMgr struct {
	b I3barBlocklet
}

func NewFromPluginName(name string) *BlockletMgr {
	// TODO: get builtin blocklet by name
	return &BlockletMgr{}
}
