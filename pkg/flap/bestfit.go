package flap

import (
	//"errors"
	//"math"
	//"sort"
	//"time"
)

type bestFit struct {
	yvalues  []Kilometres
	m	float64
	c	float64
}

func (self *bestFit) addYValue(y Kilometres) error {
	return ENOTIMPLEMENTED
}

func (self *bestFit) calculateLine() error {
	return ENOTIMPLEMENTED
}

func (self *bestFit) estimateDate(y Kilometres) (EpochTime,error) {
	return EpochTime(0), ENOTIMPLEMENTED
}

