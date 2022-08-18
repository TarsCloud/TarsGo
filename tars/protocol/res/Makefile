OS=$(shell uname -s)
all:
	tars2go -without-trace=true -add-servant=false -tarsPath github.com/TarsCloud/TarsGo/tars *.tars
ifeq ($(OS),Darwin)
	sed -i '' 's|"endpointf"|"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"|g' queryf/QueryF.tars.go
else
	sed -i 's|"endpointf"|"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"|g' queryf/QueryF.tars.go
endif
	#override file
