all:
	tars2go -add-servant=false -tarsPath github.com/TarsCloud/TarsGo/tars *.tars
	#override file
	sed -i 's|"endpointf"|"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"|g' queryf/QueryF_IF.go
