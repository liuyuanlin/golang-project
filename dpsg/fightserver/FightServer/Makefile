ROOT_DIR=../../../..

include $(ROOT_DIR)/Makefile.include

FILES=$(wildcard *.cpp)
MODULES=$(FILES:%.cpp=%.o)


INCLUDE= -I$(ROOT_DIR)/engine/src/Common
INCLUDE2= -I$(ROOT_DIR)/engine/src/Network/net
INCLUDE3= -I$(ROOT_DIR)/engine/src
INCLUDE4= -I$(ROOT_DIR)/rpc/proto
INCLUDE5= -I$(ROOT_DIR)/engine/src/Network/snappy-1.0.5
INCLUDE6= -I$(ROOT_DIR)/engine/src/Network/protobuf-2.4.1/src
INCLUDE7= -I$(ROOT_DIR)/client/dpsg/Classes/GameLogic

APP_DEPEND_LIB=-lsnappy -lprotobuf -levent -lCommon -lnet -lxml2 -lGameLogic -lrpc
TARGET=FightServer

$(TARGET):$(MODULES)
	@$(CC) $(APP) $(TARGET) $(MODULES) -L$(ROOT_DIR)/$(OUTPUT_DIR) $(DEPEND_LIB_COMMON) $(APP_DEPEND_LIB)
	@echo "linking ....." $(TARGET)
	@cp $(TARGET) $(ROOT_DIR)/$(OUTPUT_DIR)
	@echo "Copy " $(TARGET) " to " $(ROOT_DIR)/$(OUTPUT_DIR)

$(MODULES):%.o:%.cpp
	@$(CC) $(MODULE) $(INCLUDE) $(INCLUDE2) $(INCLUDE3) $(INCLUDE4) $(INCLUDE5) $(INCLUDE6) $(INCLUDE7) $< -o $@
	@echo Compiling ..... $<

clean :
	rm $(TARGET) *.o
