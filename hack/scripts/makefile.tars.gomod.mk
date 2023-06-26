#-------------------------------------------------------------------------------
#fix cgo compile error
export LC_ALL   = en_US.UTF-8
export LANG     = en_US.UTF-8
#-------------------------------------------------------------------------------

GOPATH ?= $(shell go env GOPATH)
GOROOT ?= $(shell go env GOROOT)
GO      = ${GOROOT}/bin/go
GO_MINOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
GO_MINOR_VERSION16 = 16

#-------------------------------------------------------------------------------
libpath=${subst :, ,$(GOPATH)}
GOMODULENAME:= $(shell head -n1 go.mod | awk '{print $$2}')

ifeq (,$(findstring -outdir,$(J2GO_FLAG)))
    J2GO_FLAG   += -outdir=tars-protocol
endif

ifeq (,$(findstring -module,$(J2GO_FLAG)))
    J2GO_FLAG   += -module=${GOMODULENAME}
endif

PB2GO     	:= $(firstword $(subst :, , $(GOPATH)))/bin/protoc

#-------------------------------------------------------------------------------

TARS_SRC     := $(wildcard *.tars)
PRO_SRC     += $(wildcard *.proto)
GO_SRC      := $(wildcard *.go)

#----------------------------------------------------------------------------------

copyfile = if test -z "$(APP)" || test -z "$(TARGET)"; then \
               echo "['APP' or 'TARGET' option is empty.]"; exit 1; \
           	else \
		       	if test ! -d $(2); then \
              		echo "[No such dir:$(2), now we create it.]";\
    				mkdir -p $(2);\
				fi; \
         		echo "[Copy file $(1) -> $(2)]"; \
         		cp -v $(1) $(2); \
			fi;

ALL: $(TARGET)
#----------------------------------------------------------------------------------
$(TARGET): TARSBUILD PROBUILD $(GO_SRC)
	$(GO) mod tidy
	$(GO) build $(GO_BUILD_FLAG) -o $@

#----------------------------------------------------------------------------------
ifneq ($(strip $(TARS_SRC)),)
TARSBUILD: $(TARS_SRC) tars2go
	@echo "\033[33;1m$(TARS2GO)\033[0m \033[36;1m ${TARS_SRC} \033[0m..."
	$(TARS2GO) $(J2GO_FLAG) $(TARS_SRC)
else
TARSBUILD: $(TARS_SRC)
	@echo "no tars file"
endif

ifneq ($(PRO_SRC),)
PROBUILD: $(PRO_SRC) protoc-gen-go protoc-gen-go-tarsrpc
	@echo "\033[33;1mprotoc\033[0m \033[36;1m ${PRO_SRC} \033[0m..."
	@echo $(PB2GO) ${PB2GO_FLAG} $(addprefix --proto_path=, $(sort $(dir $(PRO_SRC)))) $(PRO_SRC)
	$(foreach file,$(PRO_SRC),$(eval echo $(PB2GO) ${PB2GO_FLAG} --proto_path=$(dir $(file)) $(file)))
	for file in $(sort $(PRO_SRC));\
	do \
		dirname=$$(dirname $$file);\
		$(PB2GO) ${PB2GO_FLAG} --go_out=$$dirname --go-tarsrpc_out=$$dirname --proto_path=$$dirname $$file;\
	done
else
PROBUILD: $(PRO_SRC)
	@echo "no proto file"
endif

#----------------------------------------------------------------------------------
tar: $(TARGET) $(CONFIG)
	@if [ -d $(TARGET)_tmp_dir ]; then \
		echo "dir has exist:$(TARGET)_tmp_dir, abort."; \
		exit 1; \
	else \
		mkdir -p $(TARGET)_tmp_dir/$(TARGET);\
		cp -rf $(TARGET) $(CONFIG) $(TARGET)_tmp_dir/$(TARGET)/; \
		cd $(TARGET)_tmp_dir; tar --exclude=".svn" --exclude="_svn" -czvf $(TARGET).tgz $(TARGET)/; cd ..; \
		if [ -f "$(TARGET).tgz" ]; then \
			mv -vf $(TARGET).tgz $(TARGET).`date +%Y%m%d%H%M%S`.tgz; \
		fi; \
		mv $(TARGET)_tmp_dir/$(TARGET).tgz ./; \
		rm -rf $(TARGET)_tmp_dir; \
		echo "tar cvfz $(TARGET).tgz ..."; \
	fi

TARS_WEB_HOST   ?= http://localhost:3000
TARS_WEB_TOKEN  ?= ""
UPLOAD_USER     ?= $(shell whoami)
UPLOAD_OS       ?= linux
upload: export GOOS=${UPLOAD_OS}
upload: tar
	@echo "$(TARGET).tgz --- $(APP).$(TARGET).tgz  OS: ${GOOS}"
	curl ${TARS_WEB_HOST}/api/upload_and_publish?ticket=${TARS_WEB_TOKEN} -Fsuse=@${TARGET}.tgz -Fapplication=${APP} -Fmodule_name=${TARGET} -Fcomment=uploaded-by-${UPLOAD_USER}
	@echo "\n---------------------------------------------------------------------------\n"


HELP += $(HELP_TAR)

ifneq ($(TARS_SRC),)

SERVER_NAME := $(TARGET)

endif
#----------------------------------------------------------------------------------

clean:
	rm -vf $(DEPEND_TARS_OBJ) $(INVOKE_DEPEND_TARS_OBJ) $(LOCAL_OBJ) $(TARGET) $(TARGETS) $(DEP_FILE) ${CLEANFILE} .*.d.tmp gmon.out
	rm -vf *$(TARGET)*.tgz

cleanall:
	rm -vf $(DEPEND_TARS_H) $(DEPEND_TARS_CPP) $(DEPEND_TARS_OBJ) $(LOCAL_OBJ) $(HCE_H) $(HCE_CPP) $(TARGET) $(TARGETS) $(DEP_FILE) ${CLEANFILE} *.o .*.d.tmp .*.d gmon.out
	rm -vf *$(TARGET)*.tgz

HELP += $(HELP_CLEAN)
HELP += $(HELP_CLEANALL)

HELP_CLEAN    = "\n\033[1;33mclean\033[0m:\t\t[remove $(LOCAL_OBJ) $(TARGET)]"
HELP_CLEANALL = "\n\033[1;33mcleanall\033[0m:\t[clean & rm .*.d]"
HELP_TAR      = "\n\033[1;33mtar\033[0m:\t\t[will do 'tar $(TARGET).tgz $(RELEASE_FILE)']"

help:
	@echo $(HELP)"\n"

#-------------------------------------------------------------------------------
tars2go:
ifeq (, $(shell which tars2go))
	@{ \
	set -e ;\
	export GO111MODULE=on; \
	TARS2GO_TMP_DIR=$$(mktemp -d);\
	cd $$TARS2GO_DIR;\
	go mod init tmp;\
	if [ $(GO_MAJOR_VERSION) -gt $(GO_MAJOR_VERSION16) ]; then  \
	go install github.com/TarsCloud/TarsGo/tars/tools/tars2go@latest;\
	else \
	go get github.com/TarsCloud/TarsGo/tars/tools/tars2go@latest;\
	fi;\
	rm -rf $$TARS2GO_TMP_DIR ;\
	}
TARS2GO=$(shell go env GOPATH)/bin/tars2go
else
TARS2GO=$(shell which tars2go)
endif

protoc-gen-go:
ifeq (, $(shell which protoc-gen-go))
	@{ \
	set -e ;\
	export GO111MODULE=on; \
	PROTOC_GEN_GO_TMP_DIR=$$(mktemp -d);\
	cd $$PROTOC_GEN_GO_TMP_DIR;\
	go mod init tmp;\
	if [ $(GO_MAJOR_VERSION) -gt $(GO_MAJOR_VERSION16) ]; then  \
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0;\
	else \
	go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0;\
	fi;\
	rm -rf $$PROTOC_GEN_GO_TMP_DIR ;\
	}
PROTOC_GEN_GO=$(shell go env GOPATH)/bin/protoc-gen-go
else
PROTOC_GEN_GO=$(shell which protoc-gen-go)
endif

protoc-gen-go-tarsrpc:
ifeq (, $(shell which protoc-gen-go-tarsrpc))
	@{ \
	set -e ;\
	export GO111MODULE=on; \
	PROTOC_GEN_GO_TARSRPC_TMP_DIR=$$(mktemp -d);\
	cd $$PROTOC_GEN_GO_TARSRPC_TMP_DIR;\
	go mod init tmp;\
	if [ $(GO_MAJOR_VERSION) -gt $(GO_MAJOR_VERSION16) ]; then  \
	go install github.com/TarsCloud/TarsGo/tars/tools/protoc-gen-go-tarsrpc@latest;\
	else \
	go get github.com/TarsCloud/TarsGo/tars/tools/protoc-gen-go-tarsrpc@latest; \
	fi;\
	rm -rf $$PROTOC_GEN_GO_TARSRPC_TMP_DIR ;\
	}
PROTOC_GEN_GO_TARSRPC=$(shell go env GOPATH)/bin/protoc-gen-go-tarsrpc
else
PROTOC_GEN_GO_TARSRPC=$(shell which protoc-gen-go-tarsrpc)
endif
