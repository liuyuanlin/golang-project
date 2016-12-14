//
//  CRBuilding.h
//  Robot
//
//  Created by PU on 13-4-18.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#ifndef __Robot__CRBuilding__
#define __Robot__CRBuilding__

#include <iostream>
#include <vector>
#include <map>

#include "LogicAvatar.h"
#include "CRobot.h"
#include "LogicFloat2.h"

Use_NS_GameLogic
using namespace std;

struct RobotBuildingInfo
{
    string  templateName;
    unsigned char x;
    unsigned char y;
    unsigned char level;
    int     arg1;
    int     arg2;
};

class CRBuilding
{
public:
    static CRBuilding* Instance();
    void Init();
    
    //add by wyc 2013-12-26 read all village config file
    void ReadAllVillageXML();
    void ReadXML(unsigned int uIndex);
    
    void ReadRobotVillageXML();
    //add by wyc 2013-12-23 读取指定建筑配置
    void ReadRobotVillageXML(unsigned int nIndex);
    void ReadMaxBuildingCount();
    
    void SyncBuildings(LogicAvatar* logicAvater, const rpc::VillageInfo &info, bool battlefield);
    RobotBuildingInfo* GetBuildingInfo(rpc::BuildingId_IdType typeId, unsigned char curCount);
    LogicBuilding* CreateNewBuilding(CRobot* robot, RobotBuildingInfo* buildInfo);
    //add by wyc 2013-12-23
    LogicBuilding* CreateNewBuilding(LogicAvatar* pAvart, RobotBuildingInfo* buildInfo);
    
    ResType GetResourceCostType(rpc::BuildingId_IdType typeId);
    unsigned int GetResourceCost(rpc::BuildingId_IdType typeId, int level);
    
    bool UpgradeBuilding(CRobot* robot, LogicBuilding* logicBuilding);
    //add by wyc 2013-12-23
    bool UpgradeBuilding(LogicAvatar* pAvatar, LogicBuilding* logicBuilding);
    bool FinishNowBuilding(CRobot* robot, LogicBuilding* logicBuilding);
    //add by wyc 2013-12-23
    bool FinishNowBuilding(LogicAvatar* pAvatar, LogicBuilding* logicBuilding);
    
    bool RemoveBuilding(CRobot* robot, LogicBuilding* logicBuilding);
    
    //获取当前本数的配置数量
    unsigned int GetCountOnLevel(unsigned int uLevel);
    //获取当前本数的一个村庄配置
    unsigned int GetIndexOnLevel(unsigned int uLevel, unsigned int uIndex);
    
private:
    map<string, vector<RobotBuildingInfo> > m_BuildingInfoMap;
    
    map<int, vector<int> >      m_mapVillageInfo;
    unsigned int m_MapIndex;
};
#endif /* defined(__Robot__CRBuilding__) */
