package state

import "strconv"
import . "luago/api"

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_type
// lua-5.3.4/src/lapi.c#lua_type()
func (self *luaState) Type(idx int) LuaType {
	if absIdx := self.absIndex(idx); absIdx > 0 {
		val := self.get(idx)
		return typeOf(val)
	} else {
		return LUA_TNONE
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_typename
// lua-5.3.4/src/lapi.c#lua_typename()
func (self *luaState) TypeName(tp LuaType) string {
	switch tp {
	case LUA_TNONE:
		return "no value"
	case LUA_TNIL:
		return "nil"
	case LUA_TBOOLEAN:
		return "boolean"
	case LUA_TNUMBER:
		return "number"
	case LUA_TSTRING:
		return "string"
	case LUA_TTABLE:
		return "table"
	case LUA_TFUNCTION:
		return "function"
	case LUA_TUSERDATA:
		return "userdata"
	case LUA_TTHREAD:
		return "thread"
	default:
		panic("unreachable!")
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnumber
// lua-5.3.4/src/lapi.c#lua_isnumber()
func (self *luaState) IsNumber(idx int) bool {
	_, ok := self.ToNumberX(idx)
	return ok
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isstring
// lua-5.3.4/src/lapi.c#lua_isstring()
func (self *luaState) IsString(idx int) bool {
	t := self.Type(idx)
	return t == LUA_TSTRING || t == LUA_TNUMBER
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_iscfunction
// lua-5.3.4/src/lapi.c#lua_iscfunction()
func (self *luaState) IsGoFunction(idx int) bool {
	val := self.get(idx)
	t := fullTypeOf(val)
	return t == LUA_TLGF || t == LUA_TGCL
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isinteger
// lua-5.3.4/src/lapi.c#lua_isinteger()
func (self *luaState) IsInteger(idx int) bool {
	val := self.get(idx)
	return fullTypeOf(val) == LUA_TNUMINT
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isuserdata
// http://www.lua.org/manual/5.3/manual.html#lua_islightuserdata
// lua-5.3.4/src/lapi.c#lua_isuserdata()
func (self *luaState) IsUserData(idx int) bool {
	return self.Type(idx) == LUA_TUSERDATA
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnone
// lua-5.3.4/src/lua.h#lua_isnone()
func (self *luaState) IsNone(idx int) bool {
	return self.Type(idx) == LUA_TNONE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnoneornil
// lua-5.3.4/src/lua.h#lua_isnoneornil()
func (self *luaState) IsNoneOrNil(idx int) bool {
	return self.Type(idx) <= 0
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnil
// lua-5.3.4/src/lua.h#lua_isnil()
func (self *luaState) IsNil(idx int) bool {
	return self.Type(idx) == LUA_TNIL
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isboolean
// lua-5.3.4/src/lua.h#lua_isboolean()
func (self *luaState) IsBoolean(idx int) bool {
	return self.Type(idx) == LUA_TBOOLEAN
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_istable
// lua-5.3.4/src/lua.h#lua_istable()
func (self *luaState) IsTable(idx int) bool {
	return self.Type(idx) == LUA_TTABLE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isfunction
// lua-5.3.4/src/lua.h#lua_isfunction()
func (self *luaState) IsFunction(idx int) bool {
	return self.Type(idx) == LUA_TFUNCTION
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isthread
// lua-5.3.4/src/lua.h#lua_isthread()
func (self *luaState) IsThread(idx int) bool {
	return self.Type(idx) == LUA_TTHREAD
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tonumberx
func (self *luaState) ToNumberX(idx int) (float64, bool) {
	val := self.get(idx)
	return castToNumber(val)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tonumber
func (self *luaState) ToNumber(idx int) float64 {
	n, _ := self.ToNumberX(idx)
	return n
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tointegerx
func (self *luaState) ToIntegerX(idx int) (int64, bool) {
	val := self.get(idx)
	return castToInteger(val)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tointeger
func (self *luaState) ToInteger(idx int) int64 {
	i, _ := self.ToIntegerX(idx)
	return i
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_toboolean
func (self *luaState) ToBoolean(idx int) bool {
	val := self.get(idx)
	return castToBoolean(val)
}

// [-0, +0, m]
// http://www.lua.org/manual/5.3/manual.html#lua_tostring
// http://www.lua.org/manual/5.3/manual.html#lua_tolstring
func (self *luaState) ToString(idx int) (string, bool) {
	val := self.get(idx)

	s := ""
	switch x := val.(type) {
	case int64:
		s = strconv.FormatInt(x, 10) // todo
	case float64:
		s = strconv.FormatFloat(x, 'f', -1, 64) // todo
	case string:
		return x, true
	default:
		return "", false
	}

	// val is a number
	self.CheckStack(1)
	self.PushString(s)
	self.Replace(idx)
	return s, true
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tocfunction
func (self *luaState) ToGoFunction(idx int) GoFunction {
	val := self.get(idx)
	switch x := val.(type) {
	case GoFunction:
		return x
	default:
		return nil
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tothread
func (self *luaState) ToThread(idx int) LuaState {
	val := self.get(idx)
	if val != nil {
		if ls, ok := val.(*luaState); ok {
			return ls
		}
	}
	return nil
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_touserdata
func (self *luaState) ToUserData(idx int) UserData {
	val := self.get(idx)
	if val != nil {
		if ud, ok := val.(*userData); ok {
			return ud.data
		}
	}
	return nil
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_topointer
func (self *luaState) ToPointer(idx int) interface{} {
	return self.get(idx)

}
