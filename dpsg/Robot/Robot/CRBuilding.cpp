//
//  CRBuilding.cpp
//  Robot
//
//  Created by PU on 13-4-18.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#include "CRBuilding.h"

#include <fstream>
#include <stdlib.h>
#include "XMLParser.h"
#include "RUserData.h"
#include "ConfigManager.h"
#include "CConnectionMgr.h"
#include "LogicFinishNowCommand.h"
#include "LogicCharacter.h"

Use_NS_GameLogic;

#define CREATE_SINGLEBUILDING(typeId,className,funcName) \
{ \
    const ::rpc::Position& pos = village.funcName().p(); \
    unsigned int level = village.funcName().level(); \
    unsigned int hp = village.funcName().hp(); \
    unsigned int food = village.funcName().storage_food(); \
    unsigned int gold = village.funcName().storage_gold(); \
    LogicBuildingData* pData = iNew LogicBuildingData(); \
    pData->_bEnemy = true; \
    pData->_hp = hp; \
    pData->_idx = 0; \
    pData->_foodCount = food; \
    pData->_goldCount = gold; \
    pData->_categoryName = BuildingIdEnumToString(typeId); \
    LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
    pObj->SetGridPosition(pos.x(),pos.y()); \
    pVillage->AddBuilding(typeId,pObj); \
} \

#define CREATE_MULTIBUILDING_RESOURCE_PRODUCE(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className >& buildingList = village.funcName(); \
    for (size_t i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        unsigned int hp = _building.hp(); \
        string cateName = BuildingIdEnumToString(typeId); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_bEnemy = true; \
        pData->_hp = hp; \
        pData->_idx = i; \
        pData->_goldCount = _building.res_count(); \
        pData->_foodCount = _building.res_count(); \
        pData->_categoryName = cateName; \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_MULTIBUILDING_GOLDSTORED(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village.funcName(); \
    for (size_t i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        unsigned int hp = _building.hp(); \
        unsigned int gold = _building.storage_gold(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_bEnemy = true; \
        pData->_hp = hp; \
        pData->_idx = i; \
        pData->_goldCount = gold; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_MULTIBUILDING_FOODSTORED(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village.funcName(); \
    for (size_t i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        unsigned int hp = _building.hp(); \
        unsigned int food = _building.storage_food(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_bEnemy = true; \
        pData->_hp = hp; \
        pData->_idx = i; \
        pData->_foodCount = food; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_MULTIBUILDING_NOLEVEL(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village.funcName(); \
    for (size_t i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int hp = _building.hp(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_bEnemy = true; \
        pData->_hp = hp; \
        pData->_idx = i; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,1,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_MULTIBUILDING(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village.funcName(); \
    for (size_t i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        unsigned int hp = _building.hp(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_bEnemy = true; \
        pData->_hp = hp; \
        pData->_idx = i; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_DECOBUILDING(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village.funcName(); \
    for (size_t i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_bEnemy = true; \
        pData->_hp = 0; \
        pData->_idx = i; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,1,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_BARRIERBUILDING(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village.funcName(); \
    for (size_t i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_bEnemy = true; \
        pData->_hp = 0; \
        pData->_idx = i; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,1,pData); \
        if(_building.has_gem_num()) \
            pObj->SetBarrierGem(_building.gem_num()); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

////////////////////////////////////////////////////////
#define STORE_SINGLE_BUILDINGINFO(typeId,funcName) \
{ \
    LogicBuilding* pBuilding = LogicVillage::Instance()->GetBuilding(typeId,0); \
    LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
    newvillage->mutable_##funcName()->set_hp(pData->_hp); \
} \

#define STORE_TOWNHALLINFO(typeId,funcName) \
{ \
    LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,0); \
    LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
    newvillage->mutable_##funcName()->set_hp(pData->_hp); \
    newvillage->mutable_##funcName()->set_storage_food(pData->_foodCount); \
    newvillage->mutable_##funcName()->set_storage_gold(pData->_goldCount); \
} \

#define STORE_MULTI_BUILDINGINFO(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = newvillage->mutable_##funcName(); \
    for (size_t i=0;i<buildingList->size();++i) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        newvillage->mutable_##funcName(i)->set_hp(pData->_hp); \
    } \
} \

#define STORE_MULTI_GOLDSTORAGE_BUILDINGINFO(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = newvillage->mutable_##funcName(); \
    for (size_t i=0;i<buildingList->size();++i) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        newvillage->mutable_##funcName(i)->set_hp(pData->_hp); \
        newvillage->mutable_##funcName(i)->set_gold_storage(pData->_hp); \
    } \
} \

//#define UPGRADESINGLEBUILDING(typeId,className,funcName) \
//{ \
//    unsigned int upgrageTime = m_VillageInfo.funcName().upgrade_time(); \
//    unsigned int level = village.funcName().level(); \
//    if(upgrageTime>0) \
//    { \
//        Building* pBuilding = GetBuilding(typeId, 0); \
//        pBuilding->GetLogicObject()->StartConstruction(-1); \
//        if(pBuilding->GetLogicObject()->GetLevel() == 1) \
//            pBuilding->GetLogicObject()->SetAsNewConstructed(true); \
//        int idx=0; \
//        stringstream code; \
//        rpc::BuildingId_IdType tId = typeId; \
//        code.write((char*)&tId, sizeof(rpc::BuildingId_IdType)); \
//        code.write((char*)&idx, sizeof(int)); \
//        code.write((char*)&level, sizeof(int)); \
//        unsigned int localTime = TimeSystem::Instance()->ServerTimeToLocalTime(upgrageTime); \
//        code.write((char*)&localTime,sizeof(int)); \
//        int zero = 0; \
//        code.write((char*)&zero,sizeof(int)); \
//        code.write((char*)&zero,sizeof(int)); \
//        ClientAvatar::Instance()->GetCommandManager()->PushCommand(Command_Build,code); \
//    } \
//} \

//#define UPGRADEMULTIBUILDING(typeId,className,funcName) \
//{ \
//const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = m_VillageInfo.funcName(); \
//for (size_t i=0;i<buildingList.size();++i) \
//{ \
//const ::rpc::className& _building = buildingList.Get(i); \
//unsigned int level = _building.level(); \
//unsigned int upgrageTime = _building.upgrade_time(); \
//if(upgrageTime>0) \
//{ \
//Building* pBuilding = GetBuilding(typeId, i); \
//pBuilding->GetLogicObject()->StartConstruction(-1); \
//if(pBuilding->GetLogicObject()->GetLevel() == 1) \
//pBuilding->GetLogicObject()->SetAsNewConstructed(true); \
//stringstream code; \
//rpc::BuildingId_IdType tId = typeId; \
//code.write((char*)&tId, sizeof(rpc::BuildingId_IdType)); \
//code.write((char*)&i, sizeof(int)); \
//code.write((char*)&level, sizeof(int)); \
//unsigned int localTime = TimeSystem::Instance()->ServerTimeToLocalTime(upgrageTime); \
//code.write((char*)&localTime,sizeof(int)); \
//int zero = 0; \
//code.write((char*)&zero,sizeof(int)); \
//code.write((char*)&zero,sizeof(int)); \
//ClientAvatar::Instance()->GetCommandManager()->PushCommand(Command_Build,code); \
//} \
//} \
//} \

CRBuilding* CRBuilding::Instance()
{
    static CRBuilding crBuilding;
    return &crBuilding;
}

void CRBuilding::Init()
{
    //ReadRobotVillageXML();
    ReadMaxBuildingCount();
    //随机一张配置的村庄
    m_MapIndex = random() % RUserData::Instance()->m_VillageCount + 1;
    
    //Read all data
    ReadAllVillageXML();
}

//读取所有村庄配置，并进行分类处理
void CRBuilding::ReadAllVillageXML()
{
    int nIndex = 0;
    int nAllCount = RUserData::Instance()->m_VillageCount;
    
    while (nIndex <= nAllCount) {
        ReadXML(nIndex);
        ++nIndex;
    }
}

//读取指定文件的townhall等级进行分类
void CRBuilding::ReadXML(unsigned int uIndex)
{
    char rv[64] = {};
    sprintf(rv, "/RobotVillage%d.xml", uIndex);
    ifstream file((RUserData::Instance()->m_Path + string(rv)).c_str());
    if(!file)
        return;
    
    file.seekg(0, ios::end);
    long long buffLen = file.tellg();
    file.seekg(0, ios::beg);
    char buffer[buffLen];
    file.read(buffer, (int)buffLen);
    
    XMLParser* parser = new XMLParser();
    if(parser->Parse(buffer, (int)buffLen))
    {
        if(parser->SetToFirstChild("building")) do
        {
            if ("TownHall" == parser->GetString("name"))
            {
                int nLevel = parser->GetInt("lvl");
                if (m_mapVillageInfo.find(nLevel) == m_mapVillageInfo.end())
                {
                    vector<int> v;
                    m_mapVillageInfo[nLevel] = v;
                }
                m_mapVillageInfo[nLevel].push_back(uIndex);
                break;
            }
        }
        while(parser->SetToNextChild("building"));
    }
    delete parser;

}

//
unsigned int CRBuilding::GetCountOnLevel(unsigned int uLevel)
{
    if (m_mapVillageInfo.find(uLevel) == m_mapVillageInfo.end()) {
        return 0;
    }
    return m_mapVillageInfo[uLevel].size();
}

//
unsigned int CRBuilding::GetIndexOnLevel(unsigned int uLevel, unsigned int uIndex)
{
    map<int, vector<int> >::iterator itr = m_mapVillageInfo.find(uLevel);
    
    if (itr == m_mapVillageInfo.end()) {
        return 0;
    }
    
    return itr->second.at(uIndex);
}

void CRBuilding::ReadRobotVillageXML()
{
    if(m_MapIndex > RUserData::Instance()->m_VillageCount)
        m_MapIndex = 1;
    char rv[64] = {};
    sprintf(rv, "/RobotVillage%d.xml", m_MapIndex++);
    //ifstream file((RUserData::Instance()->m_Path + "/RobotVillage.xml").c_str());
    ifstream file((RUserData::Instance()->m_Path + string(rv)).c_str());
    if(!file)
        return;
    
    m_BuildingInfoMap.clear();
    file.seekg(0, ios::end);
    long long buffLen = file.tellg();
    file.seekg(0, ios::beg);
    char buffer[buffLen];
    file.read(buffer, (int)buffLen);
   
    XMLParser* parser = new XMLParser();
    if(parser->Parse(buffer, (int)buffLen))
    {
        if(parser->SetToFirstChild("building")) do
        {
            RobotBuildingInfo buildInfo;
            buildInfo.templateName = parser->GetString("name");
            buildInfo.x = parser->GetInt("x");
            buildInfo.y = parser->GetInt("y");
            buildInfo.level = parser->GetInt("lvl");
            buildInfo.arg1 = CharacterClassNameToIdEnum(parser->GetString("arg1"));
            buildInfo.arg2 = parser->GetInt("arg2");
            
            if(m_BuildingInfoMap.find(buildInfo.templateName) == m_BuildingInfoMap.end())
            {
                vector<RobotBuildingInfo> v;
                m_BuildingInfoMap[buildInfo.templateName] = v;
            }
            m_BuildingInfoMap[buildInfo.templateName].push_back(buildInfo);
        }
        while(parser->SetToNextChild("building"));
    }
    delete parser;
}

void CRBuilding::ReadRobotVillageXML( unsigned int nIndex)
{
    if(nIndex > RUserData::Instance()->m_VillageCount)
        nIndex = 1;
    char rv[64] = {};
    sprintf(rv, "/RobotVillage%d.xml", nIndex++);
    //ifstream file((RUserData::Instance()->m_Path + "/RobotVillage.xml").c_str());
    ifstream file((RUserData::Instance()->m_Path + string(rv)).c_str());
    if(!file)
        return;
    
    m_BuildingInfoMap.clear();
    file.seekg(0, ios::end);
    long long buffLen = file.tellg();
    file.seekg(0, ios::beg);
    char buffer[buffLen];
    file.read(buffer, (int)buffLen);
    
    XMLParser* parser = new XMLParser();
    if(parser->Parse(buffer, (int)buffLen))
    {
        if(parser->SetToFirstChild("building")) do
        {
            RobotBuildingInfo buildInfo;
            buildInfo.templateName = parser->GetString("name");
            buildInfo.x = parser->GetInt("x");
            buildInfo.y = parser->GetInt("y");
            buildInfo.level = parser->GetInt("lvl");
            buildInfo.arg1 = CharacterClassNameToIdEnum(parser->GetString("arg1"));
            buildInfo.arg2 = parser->GetInt("arg2");
            
            if(m_BuildingInfoMap.find(buildInfo.templateName) == m_BuildingInfoMap.end())
            {
                vector<RobotBuildingInfo> v;
                m_BuildingInfoMap[buildInfo.templateName] = v;
            }
            m_BuildingInfoMap[buildInfo.templateName].push_back(buildInfo);
        }
        while(parser->SetToNextChild("building"));
    }
    delete parser;
}

void CRBuilding::ReadMaxBuildingCount()
{
    
}

void CRBuilding::SyncBuildings(LogicAvatar* logicAvater, const rpc::VillageInfo &info, bool battlefield)
{
    LogicVillage* pVillage = logicAvater->GetVillage();
    LogicGameObjectManager* pObjMgr = logicAvater->GetGameObjectManager();
    pVillage->InitVillage(44, 44, battlefield);
    const rpc::VillageInfo& village = info;
    
    CREATE_MULTIBUILDING_NOLEVEL(rpc::BuildingId_IdType_Worker, Worker, worker)//17 17
    CREATE_MULTIBUILDING_NOLEVEL(rpc::BuildingId_IdType_GeneralHouse, GeneralHouse, generalhouse)
    CREATE_SINGLEBUILDING(rpc::BuildingId_IdType_Center, Center, center)//20 20
    
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_AllianceCastle, AllianceCastle, alliancecastle)//10 30
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Laboratory, Laboratory, laboratory)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Wall, Wall, wall)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Barrack, Barrack, barrack)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_TroopHousing, TroopHousing, troophosing)//19 25
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_ArcherTower, ArcherTower, archertower)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_AirDefense, AirDefense, airdefense)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Cannon, Cannon, cannon)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Mortar, Mortar, mortar)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_WizardTower, WizardTower, wizardtower)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_TeslaTower, TeslaTower, teslatower)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_XBow, XBow, xbow)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_SpellForge, SpellForge, spellforge)
    
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Bomb, Bomb, bomb)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_GiantBomb, GiantBomb, giantbomb)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Eject, Eject, eject)
    
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco1, Deco1, deco1)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco2, Deco2, deco2)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco3, Deco3, deco3)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco4, Deco4, deco4)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco5, Deco5, deco5)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco6, Deco6, deco6)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco7, Deco7, deco7)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco8, Deco8, deco8)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco9, Deco9, deco9)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco10, Deco10, deco10)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco11, Deco11, deco11)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco12, Deco12, deco12)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco13, Deco13, deco13)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco14, Deco14, deco14)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco15, Deco15, deco15)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco16, Deco16, deco16)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco17, Deco17, deco17)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco18, Deco18, deco18)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco19, Deco19, deco19)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco20, Deco20, deco20)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco21, Deco21, deco21)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Deco22, Deco22, deco22)
    
    CREATE_BARRIERBUILDING(rpc::BuildingId_IdType_Barrier1, Barrier1, barrier1)
    CREATE_BARRIERBUILDING(rpc::BuildingId_IdType_Barrier2, Barrier2, barrier2)
    CREATE_BARRIERBUILDING(rpc::BuildingId_IdType_Barrier3, Barrier3, barrier3)
    CREATE_BARRIERBUILDING(rpc::BuildingId_IdType_Barrier4, Barrier4, barrier4)
    CREATE_BARRIERBUILDING(rpc::BuildingId_IdType_Barrier5, Barrier5, barrier5)
    CREATE_BARRIERBUILDING(rpc::BuildingId_IdType_Barrier6, Barrier6, barrier6)
    
    CREATE_MULTIBUILDING_RESOURCE_PRODUCE(rpc::BuildingId_IdType_GoldMine, GoldMine, goldmine)//24 20
    CREATE_MULTIBUILDING_RESOURCE_PRODUCE(rpc::BuildingId_IdType_Farm, Farm, farm)
    CREATE_MULTIBUILDING_GOLDSTORED(rpc::BuildingId_IdType_GoldStorage, GoldStorage, goldstorage)
    CREATE_MULTIBUILDING_FOODSTORED(rpc::BuildingId_IdType_FoodStorage, FoodStorage, foodstorage)
    
//    int iCount = village.troophosing_size();
//    for (int i = 0; i != iCount; ++i)
//    {
//        const rpc::Position& ppp = village.troophosing(i).p();
//        int j = 0;
//    }
//        
//
//    const rpc::Position& ppp = village.alliancecastle(0).p();
//    unsigned int food,maxFood,gold,maxGold;
//    pVillage->GetFoodStorage(food, maxFood);
//    pVillage->GetGoldStorage(gold, maxGold);
//    unsigned int yb = pVillage->GetYuanBaoStorage();
//    int i = 0;
    //upgrade buildings
    //    UPGRADESINGLEBUILDING(rpc::BuildingId_IdType_Center, Center, center)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_Laboratory, Laboratory, laboratory)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_Wall, Wall, wall)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_GoldMine, GoldMine, goldmine)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_Farm, Farm, farm)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_Barrack, Barrack, barrack)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_FoodStorage, FoodStorage, foodstorage)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_GoldStorage, GoldStorage, goldstorage)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_TroopHousing, TroopHousing, troophosing)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_ArcherTower, ArcherTower, archertower)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_AirDefense, AirDefense, airdefense)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_Cannon, Cannon, cannon)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_Mortar, Mortar, mortar)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_TeslaTower, TeslaTower, teslatower)
    //    UPGRADEMULTIBUILDING(rpc::BuildingId_IdType_XBow, XBow, xbow)
//    if(!battlefield)
//    {
        //        //restore unit production queue
        //        const google::protobuf::RepeatedPtrField<rpc::Barrack>& barrackList = m_VillageInfo.barrack();
        //        for (size_t i=0;i<barrackList.size();++i)
        //        {
        //            const rpc::Barrack& barrack = barrackList.Get(i);
        //            const google::protobuf::RepeatedPtrField<rpc::CharacterQueue> multiCharQueue = barrack.queue();
        //            int idx = i;
        //            for (size_t j=0; j<multiCharQueue.size(); ++j)
        //            {
        //                const rpc::Character& unitSlot = multiCharQueue.Get(j).character();
        //                for (size_t k=0; k<unitSlot.count(); ++k)
        //                {
        //                    int enumId = (int)unitSlot.type();
        //                    int level = 1;
        //                    if (m_pLab)
        //                    {
        //                        level = m_pLab->GetLogicObject()->GetComponent<LogicUnitUpgradeComponent>()->GetUnitLevel(enumId);
        //                    }
        //                    unsigned int startTime_ser = multiCharQueue.Get(j).start_time();
        //                    unsigned int srartTime_loc = TimeSystem::Instance()->ServerTimeToLocalTime(startTime_ser);
        //                    stringstream commandData;
        //                    commandData.write((char*)&idx, sizeof(int));
        //                    commandData.write((char*)&enumId, sizeof(int));
        //                    commandData.write((char*)&level, sizeof(int));
        //                    commandData.write((char*)&srartTime_loc, sizeof(int));
        //                    ClientAvatar::Instance()->GetCommandManager()->PushCommand(Command_TrainingUnit, commandData);
        //                }
        //
        //            }
        //        }
        //
        //        //restore units producted
        //        const google::protobuf::RepeatedPtrField<rpc::TroopHousing>& campList = m_VillageInfo.troophosing();
        //        for (size_t i=0;i<campList.size();++i)
        //        {
        //            const rpc::TroopHousing& camp = campList.Get(i);
        //            const google::protobuf::RepeatedPtrField<rpc::Character>& charList = camp.character();
        //            int dd = 0;
        //            for (size_t j=0;j<charList.size();++j)
        //            {
        //                const rpc::Character& unitSlot = charList.Get(j);
        //                dd+=unitSlot.count();
        //                rpc::CharacterType charType = unitSlot.type();
        //                for (size_t k=0; k<unitSlot.count(); ++k)
        //                {
        //                    int level = 1;
        //                    if (m_pLab)
        //                    {
        //                        level = m_pLab->GetLogicObject()->GetComponent<LogicUnitUpgradeComponent>()->GetUnitLevel(charType);
        //                    }
        //                    int space = atoi(ConfigManager::Instance()->GetAttribute(CHARACTERCONFIG, CharacterIdEnumToString(charType), 1, "HousingSpace").c_str());
        //                    if(GetBuilding(rpc::BuildingId_IdType_TroopHousing, i)->GetLogicObject()->GetComponent<LogicUnitStorageComponent>()->AddUnits(space))
        //                        OnStoreUnit(0,i,charType,level);
        //                }
        //            }
        //        }
        //        //restore laboratory research
        //        if(m_VillageInfo.laboratory().size())
        //        {
        //            const google::protobuf::RepeatedPtrField<rpc::LaboratoryInfo>& lvlInfoList = m_VillageInfo.laboratory(0).info();
        //            for(size_t i=0;i<lvlInfoList.size();++i)
        //            {
        //                rpc::LaboratoryInfo info = lvlInfoList.Get(i);
        //                int enumId = (int)info.type();
        //                unsigned int startTime_ser = info.upgrade_time();
        //                unsigned int srartTime_loc = TimeSystem::Instance()->ServerTimeToLocalTime(startTime_ser);
        //                if(startTime_ser)
        //                {
        //                    stringstream commandData;
        //                    commandData.write((char*)&enumId, sizeof(int));
        //                    commandData.write((char*)&srartTime_loc, sizeof(int));
        //                    ClientAvatar::Instance()->GetCommandManager()->PushCommand(Command_UnitUpgrade, commandData);
        //                    break;
        //                }
        //            }
        //        }
        //        //update resources
        //        CalculateAllResource();
//    }
}

ResType CRBuilding::GetResourceCostType(rpc::BuildingId_IdType typeId)
{
    string cateName = BuildingIdEnumToString(typeId);
    string resCostType = ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG, cateName,
                                                                 1, "BuildResource",false);
    ResType rt;
    if (resCostType == "Gold")
        rt = RT_Gold;
    else if(resCostType == "Food")
        rt = RT_Food;
    else if(resCostType == "Diamonds")
        rt = RT_YuanBao;
    else
        rt = RT_Invalid;
    return rt;
}
unsigned int CRBuilding::GetResourceCost(rpc::BuildingId_IdType typeId, int level)
{
    string cateName = BuildingIdEnumToString(typeId);
    int cost = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG, cateName,level,
                                                            "BuildCost",false).c_str());
    return cost;
}

RobotBuildingInfo* CRBuilding::GetBuildingInfo(rpc::BuildingId_IdType typeId, unsigned char curCount)
{
    map<string, vector<RobotBuildingInfo> >::iterator it = m_BuildingInfoMap.find(BuildingIdEnumToString(typeId));
    if(it == m_BuildingInfoMap.end() || it->second.size() <= curCount)
        return NULL;
    
    return &it->second.at(curCount);
}

//机器人创建建筑物接口，创建失败处理应该再这里
LogicBuilding* CRBuilding::CreateNewBuilding(CRobot* robot, RobotBuildingInfo* buildInfo)
{
    rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(buildInfo->templateName);
    int index = robot->m_LogicAvatar->GetVillage()->GetBuildingCount(typeId);
    int level = 1;
    stringstream code;
    code.write((char*)&typeId, sizeof(rpc::BuildingId_IdType));
    code.write((char*)&index, sizeof(int));
    code.write((char*)&level, sizeof(int));
    TimeValue tv;
    LogicTimer::GetTime(&tv);
    int startTime = tv._Seconds;
    code.write((char*)&startTime,sizeof(int));
    int rt = GetResourceCostType(typeId);
    int cost = GetResourceCost(typeId, level);
    code.write((char*)&rt, sizeof(int));
    code.write((char*)&cost, sizeof(int));
    if(robot->m_LogicAvatar->GetCommandManager()->PushCommand(Command_Build,code))
    {
        LogicBuildingData* pData = iNew LogicBuildingData();
        pData->_categoryName = buildInfo->templateName;
        pData->_idx = index;
        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG, pData->_categoryName, level, "Hitpoints").c_str());
        LogicBuilding* pObj = (LogicBuilding*)robot->m_LogicAvatar->GetGameObjectManager()->CreateObject(LOGICBUILDING,level,pData);
        //TODO:添加对pObj的处理
        pObj->SetAsNewConstructed(true);
        pObj->SetGridPosition(buildInfo->x,buildInfo->y);
        robot->m_LogicAvatar->GetVillage()->AddBuilding(typeId,pObj);
        
        robot->m_LogicAvatar->GetCommandManager()->Tick();
        
        //send build rpc
        rpc::CreateTo ct;
        ct.mutable_id()->set_type(typeId);
        ct.mutable_id()->set_index(0);
        ct.mutable_p()->set_x(buildInfo->x);
        ct.mutable_p()->set_y(buildInfo->y);
        CConnectionMgr::GetSingleton().Call(robot->m_connectorID, "CNServer.Create", &ct);
        return pObj;
    }
    return NULL;
}

//add by wyc 2013-12-23
LogicBuilding* CRBuilding::CreateNewBuilding(LogicAvatar* pAvart, RobotBuildingInfo* buildInfo)
{
    rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(buildInfo->templateName);
    int index = pAvart->GetVillage()->GetBuildingCount(typeId);
    int level = 1;
    stringstream code;
    code.write((char*)&typeId, sizeof(rpc::BuildingId_IdType));
    code.write((char*)&index, sizeof(int));
    code.write((char*)&level, sizeof(int));
    TimeValue tv;
    LogicTimer::GetTime(&tv);
    int startTime = tv._Seconds;
    code.write((char*)&startTime,sizeof(int));
    int rt = GetResourceCostType(typeId);
    int cost = GetResourceCost(typeId, level);
    code.write((char*)&rt, sizeof(int));
    code.write((char*)&cost, sizeof(int));
    if(pAvart->GetCommandManager()->PushCommand(Command_Build,code))
    {
        LogicBuildingData* pData = iNew LogicBuildingData();
        pData->_categoryName = buildInfo->templateName;
        pData->_idx = index;
        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG, pData->_categoryName, level, "Hitpoints").c_str());
        LogicBuilding* pObj = (LogicBuilding*)pAvart->GetGameObjectManager()->CreateObject(LOGICBUILDING,level,pData);
        //TODO:添加对pObj的处理
        if (NULL == pObj)
            return NULL;
        
        pObj->SetAsNewConstructed(true);
        pObj->SetGridPosition(buildInfo->x,buildInfo->y);
        pAvart->GetVillage()->AddBuilding(typeId,pObj);
        
        pAvart->GetCommandManager()->Tick();
        return pObj;
    }
    return NULL;
}

bool CRBuilding::UpgradeBuilding(CRobot* robot, LogicBuilding* logicBuilding)
{
    string cateName = logicBuilding->GetLogicData()->_categoryName;
    rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(cateName);
    int nexLevel = logicBuilding->GetLevel()+1;
    int index = logicBuilding->GetLogicData()->_idx;
    stringstream code;
    code.write((char*)&typeId, sizeof(rpc::BuildingId_IdType));
    code.write((char*)&index, sizeof(int));
    code.write((char*)&nexLevel, sizeof(int));
    TimeValue tv;
    LogicTimer::GetTime(&tv);
    int startTime = tv._Seconds;
    code.write((char*)&startTime,sizeof(int));
    int rt = GetResourceCostType(typeId);
    int cost = GetResourceCost(typeId,nexLevel);
    if(cost <= 0)
        return false;
    
    code.write((char*)&rt, sizeof(int));
    code.write((char*)&cost, sizeof(int));
    if(robot->m_LogicAvatar->GetCommandManager()->PushCommand(Command_Build,code))
    {
        robot->m_LogicAvatar->GetCommandManager()->Tick();
        
        //send upgrade rpc
        rpc::BuildingId msg;
        msg.set_type(typeId);
        msg.set_index(index);
        CConnectionMgr::GetSingleton().Call(robot->m_connectorID, "CNServer.Upgrade", &msg);
        return true;
    }
    return false;
}

//add by wyc 2013-12-23
bool CRBuilding::UpgradeBuilding(LogicAvatar* pAvatar, LogicBuilding* logicBuilding)
{
    string cateName = logicBuilding->GetLogicData()->_categoryName;
    rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(cateName);
    int nexLevel = logicBuilding->GetLevel()+1;
    int index = logicBuilding->GetLogicData()->_idx;
    stringstream code;
    code.write((char*)&typeId, sizeof(rpc::BuildingId_IdType));
    code.write((char*)&index, sizeof(int));
    code.write((char*)&nexLevel, sizeof(int));
    TimeValue tv;
    LogicTimer::GetTime(&tv);
    int startTime = tv._Seconds;
    code.write((char*)&startTime,sizeof(int));
    int rt = GetResourceCostType(typeId);
    int cost = GetResourceCost(typeId,nexLevel);
    if(cost <= 0)
        return false;
    
    code.write((char*)&rt, sizeof(int));
    code.write((char*)&cost, sizeof(int));
    if(pAvatar->GetCommandManager()->PushCommand(Command_Build,code))
    {
        pAvatar->GetCommandManager()->Tick();
        return true;
    }
    return false;
}

bool CRBuilding::FinishNowBuilding(CRobot* robot, LogicBuilding* logicBuilding)
{
    string cateName = logicBuilding->GetLogicData()->_categoryName;
    rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(cateName);
    unsigned int buildingIdx = logicBuilding->GetLogicData()->_idx;
    int finishType = LogicFinishNowCommand::FT_Build;
    stringstream commandData;
    commandData.write((char*)&typeId,sizeof(int));
    commandData.write((char*)&buildingIdx,sizeof(int));
    commandData.write((char*)&finishType, sizeof(int));
    if(robot->m_LogicAvatar->GetCommandManager()->PushCommand(Command_FinishNow, commandData))
    {
        robot->m_LogicAvatar->GetCommandManager()->Tick();
        //send rpc
        rpc::BuildingId msg;
        msg.set_type(typeId);
        msg.set_index(buildingIdx);
        CConnectionMgr::GetSingleton().Call(robot->m_connectorID, "CNServer.Buildings_FinishNow", &msg);
        return true;
    }
    return false;
}

//add by wyc 2013-12-23
bool CRBuilding::FinishNowBuilding(LogicAvatar* pAvatar, LogicBuilding* logicBuilding)
{
    string cateName = logicBuilding->GetLogicData()->_categoryName;
    rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(cateName);
    unsigned int buildingIdx = logicBuilding->GetLogicData()->_idx;
    int finishType = LogicFinishNowCommand::FT_Build;
    stringstream commandData;
    commandData.write((char*)&typeId,sizeof(int));
    commandData.write((char*)&buildingIdx,sizeof(int));
    commandData.write((char*)&finishType, sizeof(int));
    if(pAvatar->GetCommandManager()->PushCommand(Command_FinishNow, commandData))
    {
        pAvatar->GetCommandManager()->Tick();
        return true;
    }
    return false;
}

bool CRBuilding::RemoveBuilding(CRobot* robot, LogicBuilding* logicBuilding)
{
    rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(logicBuilding->GetLogicData()->_categoryName);
    int index = logicBuilding->GetLogicData()->_idx;
    int level = 1;
    int rt = GetResourceCostType(typeId);
    int cost = GetResourceCost(typeId, level);
    TimeValue tv;
    LogicTimer::GetTime(&tv);
    int startTime = tv._Seconds;
    stringstream code;
    code.write((char*)&typeId, sizeof(rpc::BuildingId_IdType));
    code.write((char*)&index, sizeof(int));
    code.write((char*)&level, sizeof(int));
    code.write((char*)&startTime,sizeof(int));
    code.write((char*)&rt, sizeof(int));
    code.write((char*)&cost, sizeof(int));
    
    if(robot->m_LogicAvatar->GetCommandManager()->PushCommand(Command_Build,code))
    {
        robot->m_LogicAvatar->GetCommandManager()->Tick();
        
        rpc::BuildingId msg;
        msg.set_type(typeId);
        msg.set_index(index);
        CConnectionMgr::GetSingleton().Call(robot->m_connectorID, "CNServer.Remove", &msg);
        return true;
    }
    return false;
}