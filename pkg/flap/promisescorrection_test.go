package flap

import (
	"testing"
	"reflect"
	"bytes"
)

func TestPromisesCorrectionFromTo(t *testing.T) {

	var buff bytes.Buffer

	var pc promisesCorrection
	for i:=10.0; i <=30; i+=10 {
		pc.change(Kilometres(i),Kilometres(i*2))
	}

	err := pc.To(&buff)
	if err != nil {
		t.Error("To failed",err)
	}

	var pc2 promisesCorrection
	err = pc2.From(&buff)
	if err != nil {
		t.Error("From failed",err)
	}

	if !reflect.DeepEqual(pc,pc2) {
		t.Error("Deserialised doesnt equal serialized",pc ,pc2)
	}

}

