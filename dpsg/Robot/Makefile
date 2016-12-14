ROOT_DIR=../../..

include $(ROOT_DIR)/Makefile.include

FILES=$(wildcard *.cpp)
MODULES=$(FILES:%.cpp=%.o)


SUBDIR=ConnectionMgr Robot

SUBSRC=$(shell find $(SUBDIR) -name '*.cpp')
SUBOBJ=$(SUBSRC:%.cpp=%.o)

INCLUDE1= -I$(ROOT_DIR)/engine/src/Common
INCLUDE2= -I$(ROOT_DIR)/engine/src/Network/net
INCLUDE3= -I$(ROOT_DIR)/engine/src
INCLUDE4= -I$(ROOT_DIR)/rpc/proto
INCLUDE5= -I$(ROOT_DIR)/engine/src/Network/snappy-1.0.5
INCLUDE6= -I$(ROOT_DIR)/engine/src/Network/protobuf-2.4.1/src
INCLUDE7= -I$(ROOT_DIR)/client/dpsg/Classes/GameLogic
INCLUDE8=-I$(ROOT_DIR)/server/src/Robot/ConnectionMgr
INCLUDE9=-I$(ROOT_DIR)/server/src/Robot/Robot

APP_DEPEND_LIB=-lsnappy -lprotobuf -levent -lCommon -lnet -lxml2 -lGameLogic -lrpc

TARGET=robot1

$(TARGET):$(MODULES) $(SUBOBJ)
	@$(CC) $(APP) $(TARGET) $(MODULES) $(SUBOBJ) -L$(ROOT_DIR)/$(OUTPUT_DIR) $(DEPEND_LIB_COMMON) $(APP_DEPEND_LIB)
	@echo Linking ..... $(TARGET)
	@cp $(TARGET) $(ROOT_DIR)/$(OUTPUT_DIR)
	@echo copy $(TARGET) to $(ROOT_DIR)/$(OUTPUT_DIR)

$(MODULES):%.o:%.cpp
	@$(CC) $(MODULE) $(INCLUDE1) $(INCLUDE2) $(INCLUDE3) $(INCLUDE4) $(INCLUDE5) $(INCLUDE6) $(INCLUDE7) $(INCLUDE8) $(INCLUDE9) $< -o $@;
	@echo Compiling ..... $<

$(SUBOBJ):%.o:%.cpp
	@$(CC) $(MODULE) $(INCLUDE1) $(INCLUDE2) $(INCLUDE3) $(INCLUDE4) $(INCLUDE5) $(INCLUDE6) $(INCLUDE7) $(INCLUDE8) $(INCLUDE9) $< -o $@;
	@echo Compiling ..... $<

clean:
	rm $(TARGET) *.o $(SUBOBJ)
