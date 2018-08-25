
package holder

import(
    "sync"
)

type Holder struct{
    locks         map[string] *sync.Mutex
    creationQueue chan func()
    mutexCreator  func()
}

func New() *Holder {
    
    h := new( Holder )
    
    h.locks         = make( map[string] *sync.Mutex )
    h.creationQueue = make( chan func() )
    h.mutexCreator  = func() {
        for queue := range h.creationQueue {
            queue()
        }
    }
    
    go h.mutexCreator()
    
    return h
    
}

func ( h *Holder ) HoldAt( key string ) {
    
    var(
        pMutex *sync.Mutex
        holder  sync.WaitGroup
    )
    
    holder.Add( 1 )
    h.creationQueue <- func() {
        if _, ok := h.locks[key]; !ok {
            var newMutex sync.Mutex
            h.locks[key] = &newMutex
        }
        holder.Done()
    }
    holder.Wait()
    
    pMutex = h.locks[key]
    pMutex.Lock()
    
}

func ( h *Holder ) UnholdAt( key string ) {
    
    var(
        m *sync.Mutex
    )
    
    m = h.locks[key]
    m.Unlock()
    
}
