

execute_process(COMMAND go env GOPATH OUTPUT_VARIABLE GOPATH)
string(REGEX REPLACE "\n$" "" GOPATH "${GOPATH}")
set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} "${GOPATH}/src/github.com/TarsCloud/TarsGo/cmake")

include(${GOPATH}/src/github.com/TarsCloud/TarsGo/cmake/golang.cmake)

set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)

if(WIN32)
	set(TARS2CPP "${GOPATH}/bin/tars2go.exe")
	set(TARS_PATH "${GOPATH}/bin/tars2go.exe")
else()
	set(TARS2CPP "${GOPATH}/bin/tars2go")
	set(TARS_PATH "${GOPATH}/bin/tars2go")
endif()

set(TARS_WEB_HOST "" CACHE STRING "set web host")
IF (TARS_WEB_HOST STREQUAL "")
	set(TARS_WEB_HOST "http://web.tars.com")
ENDIF ()

set(TARS_TOKEN "" CACHE STRING "set web token")

# set(TARS_RELEASE "${CMAKE_BINARY_DIR}/run-release.cmake")
set(TARS_UPLOAD "${CMAKE_BINARY_DIR}/run-upload.cmake")
set(TARS_TAR "${CMAKE_BINARY_DIR}/run-tar.cmake")

# FILE(WRITE ${TARS_RELEASE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo release all)\n")
FILE(WRITE ${TARS_UPLOAD} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo upload all)\n")
FILE(WRITE ${TARS_TAR} "")

function(gen_tars TARGET)

	file(GLOB_RECURSE TARS_INPUT *.tars)

	set(OUT_TARS_H_LIST)

	if (TARS_INPUT)

		foreach(TARS_FILE ${TARS_INPUT})
			get_filename_component(TARS_NAME ${TARS_FILE} NAME_WE)
			get_filename_component(TARS_PATH ${TARS_FILE} PATH)

			set(CUR_TARS_GEN ${TARS_PATH}/${TARS_NAME}.h)

			add_custom_command(
					OUTPUT ${CUR_TARS_GEN}
					WORKING_DIRECTORY ${TARS_PATH}
					COMMAND ${TARS2CPP} -outdir tars-protocol "${TARS_TOOL_FLAG}" "${TARS_FILE}"
					COMMENT "${TARS2CPP} -outdir tars-protocol ${TARS_TOOL_FLAG} ${TARS_FILE}"
					DEPENDS ${TARS2CPP} ${TARS_FILE}
			)

			list(APPEND OUT_TARS_H_LIST ${CUR_TARS_GEN})

		endforeach()

		add_custom_target(${TARGET} ALL DEPENDS ${OUT_TARS_H_LIST})

		set_directory_properties(PROPERTIES ADDITIONAL_MAKE_CLEAN_FILES "${OUT_TARS_H_LIST}")

	endif()

endfunction()

#生成带tars文件的可执行程序
macro(gen_server APP TARGET)

	set(UPLOAD_FILES ${ARGN})

	include_directories(${PROJECT_SOURCE_DIR})

	# FILE(GLOB_RECURSE SRC_FILES  "*.go")

	add_go_executable(${TARGET})
	file(GLOB_RECURSE TARS_INPUT *.tars)

	if(TARS_INPUT)
		gen_tars(tars-${TARGET})
		add_dependencies(${TARGET} tars-${TARGET})
	endif()

	#make tar #########################################################################
	#must create tmp directory, avoid linux cmake conflict!
	SET(RUN_TAR_COMMAND_FILE "${CMAKE_BINARY_DIR}/run-tar-${TARGET}.cmake")
	FILE(WRITE ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo mkdir -p ${CMAKE_BINARY_DIR}/tmp/${TARGET})\n")
	FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo rm -rf ${CMAKE_BINARY_DIR}/tmp/${TARGET})\n")
	FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E make_directory ${CMAKE_BINARY_DIR}/tmp/${TARGET})\n")
	FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo copy bin/${TARGET} ${CMAKE_BINARY_DIR}/tmp/${TARGET}/)\n")

	IF(WIN32)
		FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E copy bin/${TARGET}.exe ${CMAKE_BINARY_DIR}/tmp/${TARGET}/)\n")
	ELSE()
		FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E copy bin/${TARGET} ${CMAKE_BINARY_DIR}/tmp/${TARGET}/)\n")
	ENDIF()	

	foreach(UPLOAD ${UPLOAD_FILES})
		FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo copy ${UPLOAD} ${CMAKE_BINARY_DIR}/tmp/${TARGET}/)\n")
		IF(WIN32)
			FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E copy ${UPLOAD} ${CMAKE_BINARY_DIR}/tmp/${TARGET}/)\n")
		ELSE()
			FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E copy ${UPLOAD} ${CMAKE_BINARY_DIR}/tmp/${TARGET}/)\n")
		ENDIF()	
	endforeach(UPLOAD ${UPLOAD_FILES})

	FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E chdir ${CMAKE_BINARY_DIR}/tmp/ tar czfv ${TARGET}.tgz ${TARGET}/)\n")
	FILE(APPEND ${RUN_TAR_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E copy_if_different ${CMAKE_BINARY_DIR}/tmp/${TARGET}.tgz ${CMAKE_BINARY_DIR}/${TARGET}.tgz)\n")

	add_custom_command(OUTPUT ${CMAKE_BINARY_DIR}/.timestamp
			WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
			COMMAND ${CMAKE_COMMAND} -P ${RUN_TAR_COMMAND_FILE}
			DEPENDS ${TARGET}
			COMMENT "tar czfv ${TARGET}.tgz")

	add_custom_target(${TARGET}-tar DEPENDS ${CMAKE_BINARY_DIR}/.timestamp)

	FILE(APPEND ${TARS_TAR} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -P ${RUN_TAR_COMMAND_FILE})\n")

	#make upload #########################################################################
	SET(RUN_UPLOAD_COMMAND_FILE "${PROJECT_BINARY_DIR}/run-upload-${TARGET}.cmake")
	IF(WIN32)
		FILE(WRITE ${RUN_UPLOAD_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo ${TARS_WEB_HOST}/api/upload_and_publish -Fsuse=@${TARGET}.tgz -Fapplication=${APP} -Fmodule_name=${TARGET} -Fcomment=developer-auto-upload)\n")
		FILE(APPEND ${RUN_UPLOAD_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${TARS_PATH}/thirdparty/bin/curl.exe ${TARS_WEB_HOST}/api/upload_and_publish?ticket=${TARS_TOKEN} -Fsuse=@${TARGET}.tgz -Fapplication=${APP} -Fmodule_name=${TARGET} -Fcomment=developer-auto-upload)\n")
		FILE(APPEND ${RUN_UPLOAD_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND -E echo \n---------------------------------------------------------------------------)\n")
	ELSE()
		FILE(WRITE ${RUN_UPLOAD_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo ${TARS_WEB_HOST}/api/upload_and_publish -Fsuse=@${TARGET}.tgz -Fapplication=${APP} -Fmodule_name=${TARGET} -Fcomment=developer-auto-upload)\n")
		FILE(APPEND ${RUN_UPLOAD_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND curl ${TARS_WEB_HOST}/api/upload_and_publish?ticket=${TARS_TOKEN} -Fsuse=@${TARGET}.tgz -Fapplication=${APP} -Fmodule_name=${TARGET} -Fcomment=developer-auto-upload)\n")
		FILE(APPEND ${RUN_UPLOAD_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo \n---------------------------------------------------------------------------)\n")

	ENDIF()

	#执行命令
	add_custom_target(${TARGET}-upload
			WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
			DEPENDS ${TARGET}-tar
			COMMAND cmake -P ${RUN_UPLOAD_COMMAND_FILE}
			COMMENT "upload ${APP}.${TARGET}.tgz and publish...")

	FILE(APPEND ${TARS_UPLOAD} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -P ${RUN_UPLOAD_COMMAND_FILE})\n")

	# #make release #########################################################################
	# SET(RUN_RELEASE_COMMAND_FILE "${PROJECT_BINARY_DIR}/run-release-${TARGET}.cmake")

	# if (TARS_INPUT)
	# 	foreach(TARS_FILE ${TARS_INPUT})
	# 		get_filename_component(TARS_NAME ${TARS_FILE} NAME_WE)
	# 		get_filename_component(TARS_PATH ${TARS_FILE} PATH)

	# 		set(CUR_TARS_GEN ${TARS_PATH}/${TARS_NAME}.h)

	# 		if(WIN32)
	# 			FILE(WRITE ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E make_directory c:\\tarsproto\\protocol\\${APP}\\${TARGET})\n")
	# 			FILE(APPEND ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo cp -rf ${CUR_TARS_GEN} c:\\tarsproto\\protocol\\${APP}\\${TARGET})\n")
	# 			FILE(APPEND ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E copy ${CUR_TARS_GEN} c:\\tarsproto\\protocol\\${APP}\\${TARGET})\n")
	# 		elseif(APPLE)
	# 			FILE(WRITE ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E make_directory $ENV{HOME}/tarsproto/protocol/${APP}/${TARGET})\n")
	# 			FILE(APPEND ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo cp -rf ${CUR_TARS_GEN} $ENV{HOME}/tarsproto/protocol/${APP}/${TARGET})\n")
	# 			FILE(APPEND ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E copy ${CUR_TARS_GEN} $ENV{HOME}/tarsproto/protocol/${APP}/${TARGET})\n")
	# 		elseif(UNIX)
	# 			FILE(WRITE ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E make_directory /home/tarsproto/${APP}/${TARGET})\n")
	# 			FILE(APPEND ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E echo cp -rf ${CUR_TARS_GEN} /home/tarsproto/${APP}/${TARGET})\n")
	# 			FILE(APPEND ${RUN_RELEASE_COMMAND_FILE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -E copy ${CUR_TARS_GEN} /home/tarsproto/${APP}/${TARGET})\n")
	# 		endif()
	# 	endforeach(TARS_FILE ${TARS_INPUT})

	# 	add_custom_target(${TARGET}-release
	# 		WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
	# 		DEPENDS ${TARGET}
	# 		COMMAND cmake -P ${RUN_RELEASE_COMMAND_FILE}
	# 		COMMENT "call ${RUN_RELEASE_COMMAND_FILE}")
        
	# 	FILE(APPEND ${TARS_RELEASE} "EXECUTE_PROCESS(COMMAND ${CMAKE_COMMAND} -P ${RUN_RELEASE_COMMAND_FILE})\n")
	# endif ()
endmacro()

add_custom_target(upload
		WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
		COMMAND cmake -P ${TARS_UPLOAD})

# add_custom_target(release
# 		WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
# 		COMMAND cmake -P ${TARS_RELEASE})

add_custom_target(tar
		WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
		COMMAND cmake -P ${TARS_TAR})

message("-------------------------------------------------------------------------------------")
message("CMAKE_SOURCE_DIR:          ${CMAKE_SOURCE_DIR}")
message("CMAKE_BINARY_DIR:          ${CMAKE_BINARY_DIR}")
message("PROJECT_SOURCE_DIR:        ${PROJECT_SOURCE_DIR}")
message("CMAKE_BUILD_TYPE:          ${CMAKE_BUILD_TYPE}")
message("PLATFORM:                  ${PLATFORM}")
message("TARS2CPP:                  ${TARS2CPP}")
message("TARS_WEB_HOST:             ${TARS_WEB_HOST}")
message("TARS_TOKEN:                ${TARS_TOKEN}")
message("-------------------------------------------------------------------------------------")

