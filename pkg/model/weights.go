package model

import (
	"errors"
	"sort"
	"math/rand"
	//"fmt"
)

var EWEIGHTNOTFOUND = errors.New("Weight not found")
var ENOWEIGHTSDEFINED = errors.New("No Weights defined")

type weight uint64

type Weights struct
{
	Scale [] weight
}

// add adds a new entry to the end of a scale of accummulating Weights
func (self *Weights) add(w weight) {
	if self.Scale == nil {
		self.Scale = make([]weight,0,500)
	}
	l := len(self.Scale)
	if l == 0 {
		self.Scale = append(self.Scale,w)
	} else {
		self.Scale= append(self.Scale,self.Scale[l-1]+w)
	}
}

// addMultiple adds multiple weights with the same value. Does nothing if
// number of multiples requested is less than 1
func (self *Weights) addMultiple(w weight,multiples int) {
	for  x:=multiples; x > 0;  x-- {
		self.add(w)
	}
}

// find finds the entry within the scale that includes the given value and returns 
// its index
func (self  *Weights) find(w weight) (int,error) {
	i := sort.Search(len(self.Scale), func(i int) bool { return self.Scale[i] >= w})
	if i < len(self.Scale)  {
		return i,nil
	} else {
		return -1,EWEIGHTNOTFOUND
	}
}

// choose chooses an entry randomly, appropriately biasing selection according to weight
// size
func (self *Weights) choose() (int,error) {
	tw,err := self.topWeight()
	if err != nil {
		return -1,ENOWEIGHTSDEFINED
	}
	if tw == 0 {
		return -1,EWEIGHTNOTFOUND
	}
	return self.find(weight(rand.Intn(int(tw))))
}

// topWeight returns the highest weight value
func (self *Weights) topWeight() (weight,error) {
	if  len(self.Scale) == 0 {
		return 0, ENOWEIGHTSDEFINED
	} else { 
		return self.Scale[len(self.Scale)-1],nil
	}
}

// reset removes all of the existing weights whilst keeping memory
// of slice allocated
func (self *Weights) reset() {
	if  len(self.Scale)>0 {
		self.Scale = self.Scale[:0]
	}
}

