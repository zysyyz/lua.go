package codegen

import "luago/binchunk"

type upvalInfo struct {
	locVarSlot int
	upvalIndex int
	index      int
}

type locVarInfo struct {
	prev    *locVarInfo
	name    string
	level   int
	slot    int
	startPc int
	endPc   int
}

type scope struct {
	parent    *scope
	usedRegs  int
	maxRegs   int
	level     int // blockLevel
	locVars   []*locVarInfo
	locNames  map[string]*locVarInfo
	upvalues  map[string]upvalInfo
	constants map[interface{}]int
	breaks    [][]int
}

func newScope(parent *scope) *scope {
	return &scope{
		parent:    parent,
		locVars:   make([]*locVarInfo, 0, 8),
		locNames:  map[string]*locVarInfo{},
		upvalues:  map[string]upvalInfo{},
		constants: map[interface{}]int{},
		breaks:    make([][]int, 1),
	}
}

/* constants */

func (self *scope) indexOfConstant(k interface{}) int {
	if idx, found := self.constants[k]; found {
		return idx
	}

	idx := len(self.constants)
	self.constants[k] = idx
	return idx
}

/* registers */

func (self *scope) allocReg() int {
	self.usedRegs++
	if self.usedRegs > self.maxRegs {
		self.maxRegs = self.usedRegs
	}
	return self.usedRegs - 1
}

func (self *scope) freeReg() {
	if self.usedRegs > 0 {
		self.usedRegs--
	} else {
		panic("usedRegs == 0 !")
	}
}

func (self *scope) allocRegs(n int) int {
	if n > 0 {
		slot := self.allocReg()
		for i := 1; i < n; i++ {
			self.allocReg()
		}
		return slot
	} else {
		panic("n <= 0 !")
	}
}

func (self *scope) freeRegs(n int) {
	if n >= 0 {
		for i := 0; i < n; i++ {
			self.freeReg()
		}
	} else {
		panic("n < 0 !")
	}
}

/* lexical scope */

func (self *scope) incrLevel(breakable bool) {
	self.level++
	if breakable {
		self.breaks = append(self.breaks, []int{})
	} else {
		self.breaks = append(self.breaks, nil)
	}
}

func (self *scope) decrLevel(endPc int) []int {
	self.level--
	for _, locVar := range self.locNames {
		if locVar.level > self.level { // out of scope
			locVar.endPc = endPc
			self.removeLocVar(locVar)
		}
	}

	breaks := self.breaks[len(self.breaks)-1]
	self.breaks = self.breaks[:len(self.breaks)-1]
	return breaks
}

func (self *scope) removeLocVar(locVar *locVarInfo) {
	self.freeReg()
	if locVar.prev == nil {
		delete(self.locNames, locVar.name)
	} else if locVar.prev.level == locVar.level {
		self.removeLocVar(locVar.prev)
	} else {
		self.locNames[locVar.name] = locVar.prev
	}
}

func (self *scope) addLocVar(name string, startPc int) int {
	newVar := &locVarInfo{
		prev:    self.locNames[name],
		name:    name,
		level:   self.level,
		slot:    self.allocReg(),
		startPc: startPc,
		endPc:   0,
	}

	self.locVars = append(self.locVars, newVar)
	self.locNames[name] = newVar

	return newVar.slot
}

func (self *scope) slotOfLocVar(name string) int {
	if locVar, found := self.locNames[name]; found {
		return locVar.slot
	} else {
		return -1
	}
}

func (self *scope) addBreakJmp(pc int) {
	for i := self.level; i >= 0; i-- {
		if self.breaks[i] != nil { // breakable
			self.breaks[i] = append(self.breaks[i], pc)
			return
		}
	}

	panic("<break> at line ? not inside a loop!")
}

/* upvalues */

func (self *scope) setupEnv() {
	self.upvalues["_ENV"] = upvalInfo{
		locVarSlot: 0,
		upvalIndex: -1,
		index:      0,
	}
}

func (self *scope) indexOfUpval(name string) int {
	if upval, ok := self.upvalues[name]; ok {
		return upval.index
	}
	if self.parent != nil {
		if locVar, found := self.parent.locNames[name]; found {
			upval := upvalInfo{
				upvalIndex: -1,
				locVarSlot: locVar.slot,
				index:      len(self.upvalues),
			}
			self.upvalues[name] = upval
			return upval.index
		}
		if idx := self.parent.indexOfUpval(name); idx >= 0 {
			upval := upvalInfo{
				locVarSlot: -1,
				upvalIndex: idx,
				index:      len(self.upvalues),
			}
			self.upvalues[name] = upval
			return upval.index
		}
	}
	return -1
}

/* summarize */

func (self scope) getConstants() []interface{} {
	consts := make([]interface{}, len(self.constants))
	for k, idx := range self.constants {
		consts[idx] = k
	}
	return consts
}

func (self *scope) getLocVars() []binchunk.LocVar {
	locVars := make([]binchunk.LocVar, len(self.locVars))
	for i, locVar := range self.locVars {
		locVars[i] = binchunk.LocVar{
			VarName: locVar.name,
			StartPc: uint32(locVar.startPc),
			EndPc:   uint32(locVar.endPc),
		}
	}
	return locVars
}

func (self scope) getUpvalues() []binchunk.Upvalue {
	upvals := make([]binchunk.Upvalue, len(self.upvalues))
	for _, uv := range self.upvalues {
		if uv.locVarSlot >= 0 { // instack
			upvals[uv.index] = binchunk.Upvalue{1, byte(uv.locVarSlot)}
		} else {
			upvals[uv.index] = binchunk.Upvalue{0, byte(uv.upvalIndex)}
		}
	}
	return upvals
}

func (self scope) getUpvalueNames() []string {
	names := make([]string, len(self.upvalues))
	for name, uv := range self.upvalues {
		names[uv.index] = name
	}
	return names
}
