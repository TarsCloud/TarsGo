APP       := _APP_
TARGET    := _SERVER_
MFLAGS    :=
DFLAGS    :=
CONFIG    := client
STRIP_FLAG:= N
J2GO_FLAG:= 

libpath=${subst :, ,$(GOPATH)}
$(foreach path,$(libpath),$(eval -include $(path)/src/github.com/TarsCloud/TarsGo/tars/makefile.tars.gomod))
