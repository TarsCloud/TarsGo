all:
	tars2go -add-servant=false *.tars
	#override file
	sed -i 's|"endpointf"|"tars/protocol/res/endpointf"|g' queryf/QueryF_IF.go
