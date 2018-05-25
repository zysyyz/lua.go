#!/bin/sh
set -ex

go install luago/standalone/lua
./bin/lua ./test/lua-5.3.4-tests/attrib.lua     | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/bitwise.lua    | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/calls.lua      | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/closure.lua    | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/constructs.lua | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/events.lua     | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/goto.lua       | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/locals.lua     | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/nextvar.lua    | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/math.lua       | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/sort.lua       | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/strings.lua    | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/utf8.lua       | grep -q ok
./bin/lua ./test/lua-5.3.4-tests/vararg.lua     | grep -q OK
./bin/lua ./test/lua-5.3.4-tests/verybig.lua    | grep -q OK
