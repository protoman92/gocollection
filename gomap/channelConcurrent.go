package gomap

type containsRequest struct {
	key     Key
	foundCh chan<- bool
}

type deleteRequest struct {
	key   Key
	lenCh chan<- int
}

type getResult struct {
	element Value
	found   bool
}

type getRequest struct {
	key     Key
	valueCh chan<- *getResult
}

type setRequest struct {
	key   Key
	value Value
	lenCh chan<- int
}

type channelConcurrentMap struct {
	storage         Map
	accessMapCh     chan chan Map
	accessStorageCh chan chan map[Key]Value
	clearCh         chan chan interface{}
	containsCh      chan *containsRequest
	deleteCh        chan *deleteRequest
	lenCh           chan chan int
	getCh           chan *getRequest
	setCh           chan *setRequest
}

// This operation blocks until the underlying Map is received.
func (cm *channelConcurrentMap) UnderlyingMap() Map {
	accessCh := make(chan Map, 0)
	cm.accessMapCh <- accessCh
	return <-accessCh
}

// This operation blocks until the underlying storage is received.
func (cm *channelConcurrentMap) UnderlyingStorage() map[Key]Value {
	accessCh := make(chan map[Key]Value, 0)
	cm.accessStorageCh <- accessCh
	return <-accessCh
}

// This operation blocks until some result is received.
func (cm *channelConcurrentMap) Clear() {
	requestCh := make(chan interface{}, 0)
	cm.clearCh <- requestCh
	<-requestCh
}

// This operation blocks until a value is received.
func (cm *channelConcurrentMap) Contains(key Key) bool {
	foundCh := make(chan bool, 0)
	cm.containsCh <- &containsRequest{key: key, foundCh: foundCh}
	return <-foundCh
}

// This operation blocks until some value is received.
func (cm *channelConcurrentMap) Delete(key Key) int {
	lenCh := make(chan int, 0)
	cm.deleteCh <- &deleteRequest{key: key, lenCh: lenCh}
	return <-lenCh
}

// This operaton blocks until some value is received.
func (cm *channelConcurrentMap) Get(key Key) (Value, bool) {
	valueCh := make(chan *getResult, 0)
	cm.getCh <- &getRequest{key: key, valueCh: valueCh}
	result := <-valueCh
	return result.element, result.found
}

// This operaton blocks until some value is received.
func (cm *channelConcurrentMap) IsEmpty() bool {
	return cm.Length() == 0
}

// This operaton blocks until some value is received.
func (cm *channelConcurrentMap) Length() int {
	requestCh := make(chan int, 0)
	cm.lenCh <- requestCh
	return <-requestCh
}

// This operaton blocks until some value is received.
func (cm *channelConcurrentMap) Set(key Key, value Value) int {
	lenCh := make(chan int, 0)
	cm.setCh <- &setRequest{key: key, value: value, lenCh: lenCh}
	return <-lenCh
}

func (cm *channelConcurrentMap) loopMap() {
	for {
		select {
		case ar := <-cm.accessMapCh:
			ar <- cm.storage

		case ar := <-cm.accessStorageCh:
			ar <- cm.storage.UnderlyingStorage()

		case cr := <-cm.clearCh:
			cm.storage.Clear()
			cr <- true

		case cr := <-cm.containsCh:
			cr.foundCh <- cm.storage.Contains(cr.key)

		case dr := <-cm.deleteCh:
			dr.lenCh <- cm.storage.Delete(dr.key)

		case lr := <-cm.lenCh:
			lr <- cm.storage.Length()

		case gr := <-cm.getCh:
			element, found := cm.storage.Get(gr.key)
			gr.valueCh <- &getResult{element: element, found: found}

		case sr := <-cm.setCh:
			sr.lenCh <- cm.storage.Set(sr.key, sr.value)
		}
	}
}

// NewChannelConcurrentMap returns a new channel-based ConcurrentMap.
func NewChannelConcurrentMap(storage Map) ConcurrentMap {
	cm := &channelConcurrentMap{
		storage:         storage,
		accessMapCh:     make(chan chan Map, 0),
		accessStorageCh: make(chan chan map[Key]Value, 0),
		clearCh:         make(chan chan interface{}, 0),
		containsCh:      make(chan *containsRequest, 0),
		deleteCh:        make(chan *deleteRequest, 0),
		lenCh:           make(chan chan int, 0),
		getCh:           make(chan *getRequest, 0),
		setCh:           make(chan *setRequest, 0),
	}

	go cm.loopMap()
	return newConcurrentMap(cm)
}