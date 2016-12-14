#include "FightCalculator.h"
#include <queue>
#include "LogicCharacter.h"
#include "LogicBuilding.h"
#include "ConfigManager.h"
#include "LogicVillage.h"
#include "TimeSystem.h"
#include "LogicBattle.h"
#include "LogicSkillManager.h"
#include "LogicAvatar.h"
#include "LogicClanComponent.h"
#include "CConnectionMgr.h"
#include "FileStream.h"
#include "XMLParser.h"
#include "LogicCombatComponent.h"
#include "LogSpellSkillComponent.h"
#include "LogicHeroSummonComponent.h"
#ifdef __linux__
#include <stdio.h>
#endif

using namespace std;
Use_NS_GameLogic
FightCalculator::FightCalculator()
{}

FightCalculator* FightCalculator::Instance()
{
    static FightCalculator inst;
    return &inst;
}

#define ADD_SINGLEBUILDING(className,funcName) \
{ \
pVillage->mutable_##funcName()->mutable_p()->set_x(x); \
pVillage->mutable_##funcName()->mutable_p()->set_y(y); \
pVillage->mutable_##funcName()->set_level(lvl); \
pVillage->mutable_##funcName()->set_hp(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,lvl,"Hitpoints").c_str())); \
pVillage->mutable_##funcName()->set_storage_gold(0); \
pVillage->mutable_##funcName()->set_storage_food(0); \
} \

#define ADD_MULTIBUILDING(className,funcName) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
pVillage->mutable_##funcName(index)->set_level(lvl); \
pVillage->mutable_##funcName(index)->set_hp(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,lvl,"Hitpoints").c_str())); \
pVillage->mutable_##funcName(index)->set_upgrade_time(0); \
} \

#define ADD_MULTIBUILDING_XBOW(className,funcName) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
pVillage->mutable_##funcName(index)->set_level(lvl); \
pVillage->mutable_##funcName(index)->set_hp(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,lvl,"Hitpoints").c_str())); \
pVillage->mutable_##funcName(index)->set_upgrade_time(0); \
pVillage->mutable_##funcName(index)->set_ammocount(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,lvl,"AmmoCount").c_str())); \
} \

#define ADD_MULTIGOLDSTORAGEBUILDING(className,funcName) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
pVillage->mutable_##funcName(index)->set_level(lvl); \
pVillage->mutable_##funcName(index)->set_hp(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,lvl,"Hitpoints").c_str())); \
pVillage->mutable_##funcName(index)->set_upgrade_time(0); \
pVillage->mutable_##funcName(index)->set_storage_gold(0); \
} \

#define ADD_MULTIFOODSTORAGEBUILDING(className,funcName) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
pVillage->mutable_##funcName(index)->set_level(lvl); \
pVillage->mutable_##funcName(index)->set_hp(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,lvl,"Hitpoints").c_str())); \
pVillage->mutable_##funcName(index)->set_upgrade_time(0); \
pVillage->mutable_##funcName(index)->set_storage_food(0); \
} \

#define ADD_MULTIRESPRODUCTBUILDING(className,funcName) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
pVillage->mutable_##funcName(index)->set_level(lvl); \
pVillage->mutable_##funcName(index)->set_hp(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,lvl,"Hitpoints").c_str())); \
pVillage->mutable_##funcName(index)->set_upgrade_time(0); \
pVillage->mutable_##funcName(index)->set_last_op_time(0); \
pVillage->mutable_##funcName(index)->set_res_count(0); \
} \

#define ADD_MULTIBUILDING_NOLEVEL(className,funcName) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
pVillage->mutable_##funcName(index)->set_hp(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,lvl,"Hitpoints").c_str())); \
} \

#define ADD_MULTIBUILDING_NOLEVEL_NOHP(className,funcName) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
} \

#define ADD_MULTIBARRIER(className,funcName) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
pVillage->mutable_##funcName(index)->set_remove_time(0); \
pVillage->mutable_##funcName(index)->set_gem_num(0); \
} \

#define ADD_GENERALHOUSE(className,funcName,heroname,herolvl) \
{ \
pVillage->add_##funcName(); \
int index = pVillage->funcName##_size() - 1; \
pVillage->mutable_##funcName(index)->mutable_p()->set_x(x); \
pVillage->mutable_##funcName(index)->mutable_p()->set_y(y); \
pVillage->mutable_##funcName(index)->add_hero(); \
rpc::Character* HeroData = pVillage->mutable_##funcName(index)->mutable_hero(0)->mutable_character(); \
HeroData->set_level(herolvl); \
HeroData->set_count(1); \
HeroData->set_type(::CharacterClassNameToIdEnum(heroname)); \
pVillage->mutable_##funcName(index)->set_selectedhero(GameLogic::CharacterClassNameToIdEnum(heroname)); \
pVillage->mutable_##funcName(index)->set_hp(atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,1,"Hitpoints").c_str())); \
} \


void FightCalculator::ReadMissionConfigList(string missinConfig)
{
    const ConfigManager::AttributeTable& missionTable = ConfigManager::Instance()->GetAttributeTable(missinConfig);
    //printf("ReadMissionConfigList %ld, %s\n", missionTable.size(), missinConfig.c_str());
    ConfigManager::AttributeTable::const_iterator itr = missionTable.begin();
    for (; itr!= missionTable.end(); ++itr)
    {
        string mapName = itr->second.at(0).find("ResName")->second;
        //get file data
        GameHub::FileStream file;
        char s[512] = "";
        getcwd(s, 512);
        string binPath = s;
        string fullfilename = binPath + "/../designer/scene/" + mapName;
        //printf("ReadMissionConfigList111 %s\n", fullfilename.c_str());
        file.Open(fullfilename.c_str());
        unsigned int size = file.GetFileSize();
        assert(size != 0);
        char* buffer = GH_NEW char[size];
        file.GetBuffer(buffer);
        //parse pve file data
        rpc::VillageInfo* pVillage = new rpc::VillageInfo();
        XMLParser* parser = new XMLParser();
        if(parser->Parse(buffer, size))
        {
            if(parser->SetToFirstChild("building")) do
            {
                string name = parser->GetString("name");
                int x = atoi(parser->GetString("x").c_str());
                int y = atoi(parser->GetString("y").c_str());
                int lvl = atoi(parser->GetString("lvl").c_str());
                int herolvl = 1;
                string heroname = parser->GetString("arg1");
                if(parser->GetString("arg2").size())
                    herolvl = atoi(parser->GetString("arg2").c_str());
                if (name == "TownHall")
                    ADD_SINGLEBUILDING(TownHall, center)
                else if(name == "Worker")
                    ADD_MULTIBUILDING_NOLEVEL(Worker, worker)
                else if(name == "GoldStorage")
                    ADD_MULTIGOLDSTORAGEBUILDING(GoldStorage, goldstorage)
                else if(name == "FoodStorage")
                    ADD_MULTIFOODSTORAGEBUILDING(FoodStorage, foodstorage)
                else if(name == "Farm")
                    ADD_MULTIRESPRODUCTBUILDING(Farm, farm)
                else if(name == "GoldMine")
                    ADD_MULTIRESPRODUCTBUILDING(GoldMine, goldmine)
                else if(name == "AllianceCastle")
                    ADD_MULTIBUILDING(AllianceCastle, alliancecastle)
                else if(name == "Laboratory")
                    ADD_MULTIBUILDING(Laboratory, laboratory)
                else if(name == "Walls")
                    ADD_MULTIBUILDING(Walls, wall)
                else if(name == "Barrack")
                    ADD_MULTIBUILDING(Barrack, barrack)
                else if(name == "TroopHousing")
                    ADD_MULTIBUILDING(TroopHousing, troophosing)
                else if(name == "ArcherTower")
                    ADD_MULTIBUILDING(ArcherTower, archertower)
                else if(name == "AirDefense")
                    ADD_MULTIBUILDING(AirDefense, airdefense)
                else if(name == "Cannon")
                    ADD_MULTIBUILDING(Cannon, cannon)
                else if(name == "Mortar")
                    ADD_MULTIBUILDING(Mortar, mortar)
                else if(name == "WizardTower")
                    ADD_MULTIBUILDING(WizardTower, wizardtower)
                else if(name == "TeslaTower")
                    ADD_MULTIBUILDING(TeslaTower, teslatower)
                else if(name == "XBow")
                    ADD_MULTIBUILDING_XBOW(XBow, xbow)
                else if(name == "SpellForge")
                    ADD_MULTIBUILDING(SpellForge, spellforge)
                else if(name == "Bomb")
                    ADD_MULTIBUILDING_NOLEVEL_NOHP(Bomb, bomb)
                else if(name == "GiantBomb")
                    ADD_MULTIBUILDING_NOLEVEL_NOHP(GiantBomb, giantbomb)
                else if(name == "Eject")
                    ADD_MULTIBUILDING_NOLEVEL_NOHP(Eject, eject)
                else if(name == "GeneralHouse")
                    ADD_GENERALHOUSE(GeneralHouse,generalhouse,heroname,herolvl)
                else if(name == "Barrier1")
                    ADD_MULTIBARRIER(Barrier1,barrier1)
                else if(name == "Barrier2")
                    ADD_MULTIBARRIER(Barrier2,barrier2)
                else if(name == "Barrier3")
                    ADD_MULTIBARRIER(Barrier3,barrier3)
                else if(name == "Barrier4")
                    ADD_MULTIBARRIER(Barrier4,barrier4)
                else if(name == "Barrier5")
                    ADD_MULTIBARRIER(Barrier5,barrier5)
                else if(name == "Barrier6")
                    ADD_MULTIBARRIER(Barrier6,barrier6)

            }
            while(parser->SetToNextChild("building"));
        }
        GH_DELETE[] buffer;
        delete parser;
        m_PVEVillageMap[itr->first] = pVillage;
        //printf("ReadMissionConfigList... %s\n", itr->first.c_str());
    }
}


#define CREATE_SINGLEBUILDING(typeId,className,funcName) \
{ \
    const ::rpc::Position& pos = village->funcName().p(); \
    unsigned int level = village->funcName().level(); \
    if(level == 0 && village->funcName().upgrade_time()) \
        level = 1; \
    unsigned int food = village->funcName().storage_food(); \
    unsigned int gold = village->funcName().storage_gold(); \
    LogicBuildingData* pData = iNew LogicBuildingData(); \
    pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,level,"Hitpoints").c_str()); \
    pData->_idx = 0; \
    pData->_bEnemy = true; \
    pData->_foodCount = food; \
    pData->_goldCount = gold; \
    pData->_categoryName = BuildingIdEnumToString(typeId); \
    LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
    pObj->SetGridPosition(pos.x(),pos.y()); \
    pVillage->AddBuilding(typeId,pObj); \
} \

#define CREATE_MULTIBUILDING_RESOURCE_PRODUCE(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className >& buildingList = village->funcName(); \
    for (int i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        if(level == 0 && _building.upgrade_time()) \
            level = 1; \
        string cateName = BuildingIdEnumToString(typeId); \
        unsigned int optime = TimeSystem::Instance()->ServerTimeToLocalTime(_building.last_op_time()); \
        TimeValue tv; \
        LogicTimer::GetTime(&tv); \
        unsigned int timePass = tv._Seconds - optime; \
        int rate = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG, cateName, level, "ResourcePerHour").c_str()); \
        unsigned int resCount = rate / 3600.0f * timePass; \
        int resMax = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG, cateName, level, "ResourceMax").c_str()); \
        if (resCount > resMax) \
            resCount = resMax; \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,level,"Hitpoints").c_str()); \
        pData->_idx = i; \
        pData->_bEnemy = true; \
        if (bIsPVE) \
        { \
            pData->_goldCount = 0; \
            pData->_foodCount = 0; \
        } \
        else \
        { \
            pData->_goldCount = resCount; \
            pData->_foodCount = resCount; \
        } \
        pData->_categoryName = cateName; \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_MULTIBUILDING_GOLDSTORED(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village->funcName(); \
    for (int i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        if(level == 0 && _building.upgrade_time()) \
            level = 1; \
        unsigned int gold = _building.storage_gold(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,level,"Hitpoints").c_str()); \
        pData->_idx = i; \
        pData->_bEnemy = true; \
        pData->_goldCount = gold; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_MULTIBUILDING_FOODSTORED(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village->funcName(); \
    for (int i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        if(level == 0 && _building.upgrade_time()) \
            level = 1; \
        unsigned int food = _building.storage_food(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,level,"Hitpoints").c_str()); \
        pData->_idx = i; \
        pData->_bEnemy = true; \
        pData->_foodCount = food; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_MULTIBUILDING_NOLEVEL(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village->funcName(); \
    for (int i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,1,"Hitpoints").c_str()); \
        pData->_idx = i; \
        pData->_bEnemy = true; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,1,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATE_MULTIBUILDING(typeId,className,funcName,useConfigHP) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village->funcName(); \
    int size = buildingList.size(); \
    for (int i=0;i<size;++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        if(level == 0 && _building.upgrade_time()) \
            level = 1; \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        if(useConfigHP) \
            pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,level,"Hitpoints").c_str()); \
        else \
            pData->_hp = _building.hp(); \
        pData->_idx = i; \
        pData->_bEnemy = true; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

#define CREATEMULTIBUILDING_CHARACTERISTIC(typeId,className,funcName,useConfigHP) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village->funcName(); \
    int size = buildingList.size(); \
    for (int i=0;i<size;++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        unsigned int level = _building.level(); \
        if(level == 0 && _building.upgrade_time()) \
        level = 1; \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        if(useConfigHP) \
        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,#className,level,"Hitpoints").c_str()); \
        else \
        pData->_hp = _building.hp(); \
        pData->_idx = i; \
        pData->_bEnemy = true; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,level,pData); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
        LogicCombatComponent* combatComp = pObj->GetComponent<LogicCombatComponent>();\
        combatComp->SetAmmoCount(_building.ammocount()); \
        combatComp->SetAttackMode(_building.altattackrange()); \
        if (bIsPVE) \
            combatComp->SetAttackMode(2); \
    } \
} \

#define CREATE_DECOBUILDING(typeId,className,funcName) \
{ \
    const ::google::protobuf::RepeatedPtrField< ::rpc::className>& buildingList = village->funcName(); \
    for (int i=0;i<buildingList.size();++i) \
    { \
        const ::rpc::className& _building = buildingList.Get(i); \
        const ::rpc::Position& pos = _building.p(); \
        LogicBuildingData* pData = iNew LogicBuildingData(); \
        pData->_hp = 0; \
        pData->_idx = i; \
        pData->_bEnemy = true; \
        pData->_categoryName = BuildingIdEnumToString(typeId); \
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,1,pData); \
        pObj->SetBuildingListener(this); \
        pObj->SetGridPosition(pos.x(),pos.y()); \
        pVillage->AddBuilding(typeId,pObj); \
    } \
} \

////////////////////////////////////////////////////////
#define STORE_SINGLE_BUILDINGINFO(typeId,funcName) \
{ \
    LogicBuilding* pBuilding = LogicVillage::Instance()->GetBuilding(typeId,0); \
    LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
    village->mutable_##funcName()->set_hp(pData->_hp); \
} \

#define STORE_TOWNHALLINFO(typeId,funcName) \
{ \
    LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,0); \
    LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
    village->mutable_##funcName()->set_hp(pData->_hp); \
    village->mutable_##funcName()->set_storage_food(pData->_foodCount); \
    village->mutable_##funcName()->set_storage_gold(pData->_goldCount); \
} \

#define STORE_MULTI_BUILDINGINFO(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = village->mutable_##funcName(); \
    for (int i=0;i<buildingList->size();++i) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        village->mutable_##funcName(i)->set_hp(pData->_hp); \
    } \
} \

#define STORE_MULTI_BUILDINGINFO_CHARACTERISTIC(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = village->mutable_##funcName(); \
    for (int i=0;i<buildingList->size();++i) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        village->mutable_##funcName(i)->set_hp(pData->_hp); \
        LogicCombatComponent* combatComp = pBuilding->GetComponent<LogicCombatComponent>(); \
        village->mutable_##funcName(i)->set_ammocount(combatComp->GetAmmoCount()); \
    } \
} \

#define STORE_MULTI_GOLDSTORAGE_BUILDINGINFO(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = village->mutable_##funcName(); \
    for (int i=0;i<buildingList->size();++i) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        village->mutable_##funcName(i)->set_hp(pData->_hp); \
        village->mutable_##funcName(i)->set_storage_gold(pData->_goldCount); \
    } \
} \

#define STORE_MULTI_GOLDPROD_BUILDINGINFO(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = village->mutable_##funcName(); \
    for (int i=0;i<buildingList->size();++i) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        village->mutable_##funcName(i)->set_hp(pData->_hp); \
        village->mutable_##funcName(i)->set_res_count(pData->_goldCount); \
    } \
} \

#define STORE_MULTI_FOODSTORAGE_BUILDINGINFO(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = village->mutable_##funcName(); \
    for (int i=0;i<buildingList->size();++i) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        village->mutable_##funcName(i)->set_hp(pData->_hp); \
        village->mutable_##funcName(i)->set_storage_food(pData->_foodCount); \
    } \
} \

#define STORE_MULTI_FOODPROD_BUILDINGINFO(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = village->mutable_##funcName(); \
    for (int i=0;i<buildingList->size();++i) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        village->mutable_##funcName(i)->set_hp(pData->_hp); \
        village->mutable_##funcName(i)->set_res_count(pData->_foodCount); \
    } \
} \

#define STORE_ALLIANCE_BUILDINGINFO(typeId,className,funcName) \
{ \
    ::google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = village->mutable_##funcName(); \
    if (buildingList->size() > 0) \
    { \
        LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,0); \
        LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
        village->mutable_##funcName(0)->set_hp(pData->_hp); \
        const vector<AllianceArmy>& allianceArmyIdList = m_AllianceCharacterMap[uidkey]; \
        for (size_t i=0;i<allianceArmyIdList.size();++i) \
        { \
            const AllianceArmy aa = allianceArmyIdList[i]; \
            LogicGameObject* obj = pAvatar->GetGameObjectManager()->GetObjectById(aa._id); \
            if (!obj || obj->GetState() == LogicGameObject::OS_Dead) \
            { \
                ::google::protobuf::RepeatedPtrField< ::rpc::Character >* armyList = village->mutable_alliancecastle(0)->mutable_characters(); \
                for (int c=0;c<armyList->size();++c) \
                { \
                    rpc::Character* pRole = armyList->Mutable(c); \
                    if(pRole->type() == aa._type) \
                    { \
                        int roleCount = pRole->count(); \
                        roleCount--; \
                        if(roleCount < 0) \
                            roleCount = 0; \
                        pRole->set_count(roleCount); \
                        break; \
                    } \
                } \
            } \
        } \
    } \
} \



#define PRINT_HP(typeId,className,funcName) \
{ \
google::protobuf::RepeatedPtrField< ::rpc::className>* buildingList = village->mutable_##funcName(); \
for (int i=0;i<buildingList->size();++i) \
{ \
    LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,i); \
    LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
    if(pData->_hp > 0) \
        printf("----%s---%d---\n",pData->_categoryName.c_str(),pData->_hp); \
} \
} \

#define PRINT_HP_S(typeId) \
{ \
LogicBuilding* pBuilding = pVillage->GetBuilding(typeId,0); \
LogicBuildingData* pData = (LogicBuildingData*)pBuilding->GetLogicData(); \
if(pData->_hp > 0) \
printf("----%s---%d---\n",pData->_categoryName.c_str(),pData->_hp); \
} \



unsigned int uidkey = 0;
#define DeployInterval      2
BattleResult FightCalculator::CalculateBattleResult(rpc::VillageInfo* village, const ::google::protobuf::RepeatedPtrField< ::rpc::AttackerInfo> &attackerList, const ::google::protobuf::RepeatedPtrField< ::rpc::SpellInfo> &spellList,const ::rpc::ClanForceInfo* allianceArmy,int src_trophy,int tar_trophy,int totalCount,uint64_t playerId,bool bRestoreVillage,bool bIsPVE,int goldCount,int foodCount,int diamond)
{
    printf("---------battle total time is:%d\n",totalCount);
    uidkey++;
    if (!bIsPVE)
        m_pPVPVillageInfo = village;
    //rebuild village
    LogicAvatar* pAvatar = LogicAvatarManager::Instance()->CreateAvatar(uidkey,true);
    LogicBattle* br = pAvatar->GetBattle();
    br->Initialize();
    br->SetBattleInfoChangeListener(this);
    br->SetTrophy(src_trophy, tar_trophy);
    LogicVillage* pVillage = pAvatar->GetVillage();
    LogicGameObjectManager* pObjMgr = pAvatar->GetGameObjectManager();
    pVillage->InitVillage(44, 44,true);
    CREATE_MULTIBUILDING_NOLEVEL(rpc::BuildingId_IdType_Worker, Worker, worker)
    //CREATE_MULTIBUILDING_NOLEVEL(rpc::BuildingId_IdType_GeneralHouse, GeneralHouse, generalhouse)
    const ::google::protobuf::RepeatedPtrField< ::rpc::GeneralHouse>& buildingList = village->generalhouse();
    for (int i=0;i<buildingList.size();++i)
    {
        const ::rpc::GeneralHouse& _building = buildingList.Get(i);
        const ::rpc::Position& pos = _building.p();
        LogicBuildingData* pData = iNew LogicBuildingData();
        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG,"GeneralHouse",1,"Hitpoints").c_str());
        pData->_idx = i;
        pData->_bEnemy = true;
        pData->_categoryName = BuildingIdEnumToString(rpc::BuildingId_IdType_GeneralHouse);
        LogicBuilding* pObj = (LogicBuilding*)pObjMgr->CreateObject(LOGICBUILDING,1,pData);
        pObj->SetGridPosition(pos.x(),pos.y());
        pVillage->AddBuilding(rpc::BuildingId_IdType_GeneralHouse,pObj);
    }
    CREATE_SINGLEBUILDING(rpc::BuildingId_IdType_Center, TownHall, center)
    
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_AllianceCastle, AllianceCastle, alliancecastle,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Laboratory, Laboratory, laboratory,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Wall, Wall, wall,false)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Barrack, Barrack, barrack,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_TroopHousing, TroopHousing, troophosing,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_ArcherTower, ArcherTower, archertower,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_AirDefense, AirDefense, airdefense,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Cannon, Cannon, cannon,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_Mortar, Mortar, mortar,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_WizardTower, WizardTower, wizardtower,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_TeslaTower, TeslaTower, teslatower,true)
    CREATEMULTIBUILDING_CHARACTERISTIC(rpc::BuildingId_IdType_XBow, XBow, xbow,true)
    CREATE_MULTIBUILDING(rpc::BuildingId_IdType_SpellForge, SpellForge, spellforge,true)
    
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Bomb, Bomb, bomb)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_GiantBomb, GiantBomb, giantbomb)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Eject, Eject, eject)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Barrier1, Barrier1, barrier1)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Barrier2, Barrier2, barrier2)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Barrier3, Barrier3, barrier3)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Barrier4, Barrier4, barrier4)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Barrier5, Barrier5, barrier5)
    CREATE_DECOBUILDING(rpc::BuildingId_IdType_Barrier6, Barrier6, barrier6)
    
    CREATE_MULTIBUILDING_RESOURCE_PRODUCE(rpc::BuildingId_IdType_GoldMine, GoldMine, goldmine)
    CREATE_MULTIBUILDING_RESOURCE_PRODUCE(rpc::BuildingId_IdType_Farm, Farm, farm)
    CREATE_MULTIBUILDING_GOLDSTORED(rpc::BuildingId_IdType_GoldStorage, GoldStorage, goldstorage)
    CREATE_MULTIBUILDING_FOODSTORED(rpc::BuildingId_IdType_FoodStorage, FoodStorage, foodstorage)
    
    //store external resource
    if (bIsPVE)
    {
        pVillage->StoreGold(goldCount);
        pVillage->StoreFood(foodCount);
    }
    //fill alliance
    LogicBuilding* pAllianceCastle = pVillage->GetBuilding(rpc::BuildingId_IdType_AllianceCastle, 0);
    if (pAllianceCastle)
    {
        LogicClanComponent* pClanComp = pAllianceCastle->GetComponent<LogicClanComponent>();
        if (pClanComp)
        {
            const google::protobuf::RepeatedPtrField<rpc::Character>& charList = village->alliancecastle(0).characters();
            for (int i=0; i<charList.size(); ++i)
            {
                const rpc::Character& role = charList.Get(i);
                pClanComp->AddArmy(role.type(), role.level(),role.count());
            }
            pClanComp->SetAllianceArmyDeployListener(this);
        }
    }
    //create hero
    const google::protobuf::RepeatedPtrField<rpc::GeneralHouse>& generalHouses = village->generalhouse();
    for(int i=0;i<generalHouses.size();++i)
    {
        LogicBuilding* pBuilding = pAvatar->GetVillage()->GetBuilding(rpc::BuildingId_IdType_GeneralHouse, i);
        if (pBuilding)
        {
            LogicHeroSummonComponent* pHeroComp = pBuilding->GetComponent<LogicHeroSummonComponent>();
            rpc::GeneralHouse house = generalHouses.Get(i);
            if(house.has_selectedhero())
            {
                rpc::CharacterType heroType = house.selectedhero();
                for (int j=0; j<house.hero().size(); ++j)
                {
                    if(house.hero(j).character().type() == heroType)
                    {
                        int level = house.hero(j).character().level();
                        LogicCharacterData* pData = iNew LogicCharacterData();
                        pData->_categoryName = CharacterIdEnumToString(heroType);
                        pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(CHARACTERCONFIG, pData->_categoryName, level, "Hitpoints").c_str());
                        pData->_bEnemy = true;
                        LogicCharacter* pHero = (LogicCharacter*)pObjMgr->CreateObject(LOGICCHARACTER,level,pData);
                        const Float2& buildingPos = pBuilding->GetSubGridPosition();
                        pHero->SetSubGridPosition(buildingPos.x, buildingPos.y);
                        pHeroComp->SetHeroId(pHero->GetId());
                        break;
                    }
                }
            }
        }
    }
    //deploy charcter and spells
    queue<rpc::AttackerInfo> attackerQueue;
    queue<rpc::SpellInfo> spellsQueue;
    queue<rpc::AttackerInfo> allianceArmyQueue;
    for (int i=0; i<attackerList.size(); ++i)
    {
        const rpc::AttackerInfo& attacker = attackerList.Get(i);
        if(pVillage->IsBlock(attacker.p().x()*0.5, attacker.p().y()*0.5, DeployBlock))
        {
            while (!attackerQueue.empty()) {attackerQueue.pop();}
            BattleResult ret;
            ret._playerlid = playerId;
            ret.m_vi.CopyFrom(*village);
            ret._goldStolen = 0;
            ret._foodStolen = 0;
            ret._damagePercent = 0;
            ret._stars = 0;
            ret._trophy = br->GetTrophy(false);
            printf("---battle Cheater-- Player ID: %lld\n", playerId);
            
            map<unsigned int,vector<AllianceArmy> >::iterator itr = m_AllianceCharacterMap.find(uidkey);
            if(itr != m_AllianceCharacterMap.end())
                m_AllianceCharacterMap.erase(itr);
            //clear created data
            LogicAvatarManager::Instance()->DestroyAvatar(uidkey);
            return ret;
        }
        attackerQueue.push(attackerList.Get(i));
    }
    for (int i=0; i<spellList.size(); ++i)
        spellsQueue.push(spellList.Get(i));
    
    //calculation circle
    std::cout << " -- calculation circle -- " << std::endl;
    for (int i=0; i<totalCount; i++)
    {
        LogicTickManager::Instance()->Tick();
        //drop units
        if (attackerQueue.size())
        {
            rpc::AttackerInfo attacker = attackerQueue.front();
            while (attacker.droptime() <= i)
            {
                std::cout << "CNServer.AttackerInfo : " << attacker.droptime() << " , " << attacker.type() << " , " << attacker.level() << " , " << attacker.p().x() << " , " << attacker.p().y() << std::endl;
                
                LogicCharacterData* pData = iNew LogicCharacterData();
                pData->_categoryName = CharacterIdEnumToString(attacker.type());
                pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(CHARACTERCONFIG, pData->_categoryName, attacker.level(), "Hitpoints").c_str());
                pData->_bEnemy = false;
                LogicCharacter* pObj = (LogicCharacter*)pObjMgr->CreateObject(LOGICCHARACTER,attacker.level(),pData);
                pObj->SetSubGridPosition(attacker.p().x(), attacker.p().y());
                
                //check next character
                attackerQueue.pop();
                if (attackerQueue.empty())
                    break;
                else
                    attacker = attackerQueue.front();
            }
        }
        //drop alliance castle
        if (allianceArmy && allianceArmy->droptime() <= i)
        {
            rpc::Position alliancePos = allianceArmy->p();
            const ::rpc::ClanForce& af = allianceArmy->clan_force();
            for (unsigned int f=0; f<af.char__size(); ++f)
            {
                const rpc::Character alliance = af.char_(f);
                rpc::AttackerInfo attacker;
                attacker.set_droptime(i+DeployInterval*(f+1));
                attacker.mutable_p()->set_x(alliancePos.x());
                attacker.mutable_p()->set_y(alliancePos.y());
                attacker.set_type(alliance.type());
                attacker.set_level(alliance.level());
                allianceArmyQueue.push(attacker);
            }
            allianceArmy = NULL;
        }
        //deploy alliance army
        if (allianceArmyQueue.size())
        {
            rpc::AttackerInfo attacker = allianceArmyQueue.front();
            while (attacker.droptime() <= i)
            {
#ifdef _SUPER_DEBUG_
                std::cout << " allianceArmyAttackerinfo : " << attacker.droptime() << " , " << attacker.type() << " , " << attacker.level() << " , " << attacker.p().x() << " , " << attacker.p().y() << std::endl;
#endif
                LogicCharacterData* pData = iNew LogicCharacterData();
                pData->_categoryName = CharacterIdEnumToString(attacker.type());
                pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(CHARACTERCONFIG, pData->_categoryName, attacker.level(), "Hitpoints").c_str());
                pData->_bEnemy = false;
                LogicCharacter* pObj = (LogicCharacter*)pObjMgr->CreateObject(LOGICCHARACTER,attacker.level(),pData);
                pObj->SetSubGridPosition(attacker.p().x(), attacker.p().y());
                
                //check next character
                allianceArmyQueue.pop();
                if (allianceArmyQueue.empty())
                    break;
                else
                    attacker = allianceArmyQueue.front();
            }
        }
        //drop spell
        if (spellsQueue.size())
        {
            rpc::SpellInfo spell = spellsQueue.front();
            while (spell.droptime() <= i)
            {
#ifdef _SUPER_DEBUG_
                std::cout << " CNServer.SpellInfo : " << spell.droptime() << " , " << spell.type() << " , " << spell.level() << " , " << spell.p().x() << " , " << spell.p().y() << std::endl;
#endif
                LogicData* spellData = iNew LogicData();
                spellData->_categoryName = SpellIdEnumToString(spell.type());
                LogicGameObject* logicSpell = pObjMgr->CreateObject(spell.level(), spellData);
                logicSpell->AddComponent(iNew LogSpellSkillComponent(logicSpell));
                logicSpell->SetSubGridPosition(spell.p().x(),spell.p().y());
                UnitCamp oldcamp = logicSpell->GetCamp();
                oldcamp._bAttackable = false;
                oldcamp._bEnemy=false;
                oldcamp._bAirTarget = true;
                oldcamp._bGroudTarget = true;
                logicSpell->SetCamp(oldcamp);
                spellsQueue.pop();
                if(spellsQueue.empty())
                    break;
                else
                    spell = spellsQueue.front();
            }
        }
#ifdef _SUPER_DEBUG_
        std::cout << "$$$$$$$$$$$$$$$$$$$$$$current frame : " << i << "!!!!!!!!!!!!!!!!!!" << std::endl;
#endif
    }
    //test code begin (do not remove)
    /*PRINT_HP(rpc::BuildingId_IdType_Worker, Worker, worker)
    PRINT_HP(rpc::BuildingId_IdType_GeneralHouse, GeneralHouse, generalhouse)
    PRINT_HP_S(rpc::BuildingId_IdType_Center)
    
    PRINT_HP(rpc::BuildingId_IdType_AllianceCastle, AllianceCastle, alliancecastle)
    PRINT_HP(rpc::BuildingId_IdType_Laboratory, Laboratory, laboratory)
    PRINT_HP(rpc::BuildingId_IdType_Barrack, Barrack, barrack)
    PRINT_HP(rpc::BuildingId_IdType_TroopHousing, TroopHousing, troophosing)
    PRINT_HP(rpc::BuildingId_IdType_ArcherTower, ArcherTower, archertower)
    PRINT_HP(rpc::BuildingId_IdType_AirDefense, AirDefense, airdefense)
    PRINT_HP(rpc::BuildingId_IdType_Cannon, Cannon, cannon)
    PRINT_HP(rpc::BuildingId_IdType_Mortar, Mortar, mortar)
    PRINT_HP(rpc::BuildingId_IdType_WizardTower, WizardTower, wizardtower)
    PRINT_HP(rpc::BuildingId_IdType_TeslaTower, TeslaTower, teslatower)
    PRINT_HP(rpc::BuildingId_IdType_XBow, XBow, xbow)
    PRINT_HP(rpc::BuildingId_IdType_SpellForge, SpellForge, spellforge)
    
    PRINT_HP(rpc::BuildingId_IdType_Bomb, Bomb, bomb)
    PRINT_HP(rpc::BuildingId_IdType_GiantBomb, GiantBomb, giantbomb)
    PRINT_HP(rpc::BuildingId_IdType_Eject, Eject, eject)
    PRINT_HP(rpc::BuildingId_IdType_Barrier1, Barrier1, barrier1)
    PRINT_HP(rpc::BuildingId_IdType_Barrier2, Barrier2, barrier2)
    PRINT_HP(rpc::BuildingId_IdType_Barrier3, Barrier3, barrier3)
    PRINT_HP(rpc::BuildingId_IdType_Barrier4, Barrier4, barrier4)
    PRINT_HP(rpc::BuildingId_IdType_Barrier5, Barrier5, barrier5)
    PRINT_HP(rpc::BuildingId_IdType_Barrier6, Barrier6, barrier6)
    
    PRINT_HP(rpc::BuildingId_IdType_GoldMine, GoldMine, goldmine)
    PRINT_HP(rpc::BuildingId_IdType_Farm, Farm, farm)
    PRINT_HP(rpc::BuildingId_IdType_GoldStorage, GoldStorage, goldstorage)
    PRINT_HP(rpc::BuildingId_IdType_FoodStorage, FoodStorage, foodstorage)
    
    const CharacterMap& xx = pAvatar->GetCharacterMap();
    CharacterMap::const_iterator aa_itr = xx.begin();
    for (; aa_itr != xx.end(); ++aa_itr)
    {
        printf("-role--%s---camp:%d--laststate---%d--\n",aa_itr->second->GetLogicData()->_categoryName.c_str(),
               aa_itr->second->GetCamp()._bEnemy,(int)aa_itr->second->GetState());
    }*/
    //test code end
    if(bRestoreVillage)
    {
        //set back building properties
        STORE_TOWNHALLINFO(rpc::BuildingId_IdType_Center,center)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_Worker,Worker,worker)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_Laboratory, Laboratory, laboratory)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_Barrack, Barrack, barrack)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_TroopHousing, TroopHousing, troophosing)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_ArcherTower, ArcherTower, archertower)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_AirDefense, AirDefense, airdefense)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_Cannon, Cannon, cannon)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_Mortar, Mortar, mortar)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_TeslaTower, TeslaTower, teslatower)
        STORE_MULTI_BUILDINGINFO_CHARACTERISTIC(rpc::BuildingId_IdType_XBow, XBow, xbow)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_SpellForge, SpellForge, spellforge)
        STORE_MULTI_BUILDINGINFO(rpc::BuildingId_IdType_GeneralHouse,GeneralHouse,generalhouse)
        STORE_MULTI_FOODPROD_BUILDINGINFO(rpc::BuildingId_IdType_Farm, Farm, farm)
        STORE_MULTI_FOODSTORAGE_BUILDINGINFO(rpc::BuildingId_IdType_FoodStorage, FoodStorage, foodstorage)
        STORE_MULTI_GOLDPROD_BUILDINGINFO(rpc::BuildingId_IdType_GoldMine, GoldMine, goldmine)
        STORE_MULTI_GOLDSTORAGE_BUILDINGINFO(rpc::BuildingId_IdType_GoldStorage, GoldStorage, goldstorage)
        STORE_ALLIANCE_BUILDINGINFO(rpc::BuildingId_IdType_AllianceCastle, AllianceCastle, alliancecastle);
        
        unsigned int iCount = village->bomb_size();
        if(iCount)
        {
            ::google::protobuf::RepeatedPtrField< ::rpc::Bomb > newBombs;
            for(int i = 0; i != iCount; ++i)
            {
                rpc::Bomb* bomb = village->mutable_bomb(i);
                if(bomb->has_p())
                    newBombs.Add()->CopyFrom(*bomb);
            }
            
            village->mutable_bomb()->Clear();
            village->mutable_bomb()->CopyFrom(newBombs);
        }
        
        unsigned int giantBombCount = village->giantbomb_size();
        if(giantBombCount)
        {
            ::google::protobuf::RepeatedPtrField< ::rpc::GiantBomb > newGiantBombs;
            for(int i = 0; i != giantBombCount; ++i)
            {
                rpc::GiantBomb* giantBomb = village->mutable_giantbomb(i);
                if(giantBomb->has_p())
                    newGiantBombs.Add()->CopyFrom(*giantBomb);
            }
            village->mutable_giantbomb()->Clear();
            village->mutable_giantbomb()->CopyFrom(newGiantBombs);
        }
    }
    //STORE_MULTI_BUILDINGINFO();
    BattleResult ret;
    ret._playerlid = playerId;
    ret.m_vi.CopyFrom(*village);
    ret._goldStolen = br->GetRobGold();
    ret._foodStolen = br->GetRobFood();
    ret._damagePercent = br->GetDestoyPer();
    ret._stars = br->GetStarCount();
    ret._trophy = br->GetTrophy(br->isWin());
    ret._exp = m_CenterLevel;
    ret._wuhun = roundf(2 * ret._stars * (src_trophy / 60.0f));
    printf("---battle Is Win: %d\n",br->isWin());
    printf("---battle Destroy Per: %d%%\n",br->GetDestoyPer());
    printf("---battle Star Count: %d\n",br->GetStarCount());
    printf("---battle GetTrophy: %d\n",ret._trophy);
    printf("---battle Rob gold: %d\n",ret._goldStolen);
    printf("---battle Rob food: %d\n",ret._foodStolen);
    m_pPVPVillageInfo = NULL;
    
    map<unsigned int,vector<AllianceArmy> >::iterator itr = m_AllianceCharacterMap.find(uidkey);
    if(itr != m_AllianceCharacterMap.end())
        m_AllianceCharacterMap.erase(itr);
    //clear created data
    LogicAvatarManager::Instance()->DestroyAvatar(uidkey);
    return ret;
}

BattleResult FightCalculator::CalculateBattleResult(rpc::AttackBegin &ab)
{ 
    m_CenterLevel = 0;
    
    const rpc::ClanForceInfo* allianceArmyInfo = NULL;
    if (ab.has_clan_force_info())
        allianceArmyInfo = &(ab.clan_force_info());
    return CalculateBattleResult(ab.mutable_v(),
                                 ab.attackunits(),
                                 ab.spells(),
                                 allianceArmyInfo,
                                 ab.src_trophy(),
                                 ab.tar_trophy(),
                                 ab.totaltime(),
                                 ab.playerlid(),
                                 true);
}

BattleResult FightCalculator::CalculateBattleResult(rpc::PVEAttackBegin &ab)
{
    int id = ab.mutable_stage()->stageid();
    if(id <= 0 || id > m_PVEVillageMap.size())
    {
        BattleResult ret;
        ret._playerlid = ab.playerlid();
        ret._goldStolen = 0;
        ret._foodStolen = 0;
        ret._damagePercent = 0;
        ret._stars = 0;
        ret._trophy = 0;
        return ret;
    }
    char idstr[32] = "";
    sprintf(idstr,"%d",id);
    int goldCount,foodCount,diamondCount;
    if (ab.mutable_stage()->has_currentgold())
    {
        goldCount = ab.mutable_stage()->currentgold();
        foodCount = ab.mutable_stage()->currentfood();
        diamondCount = ab.mutable_stage()->currentdiamond();
    }
    else
    {
        goldCount = atoi(ConfigManager::Instance()->GetAttribute("misson.config",
                                                            idstr, 1, "GoldStorage").c_str());
        foodCount = atoi(ConfigManager::Instance()->GetAttribute("misson.config",
                                                                 idstr, 1, "FoodStorage").c_str());
        diamondCount = 0;
    }
    const rpc::ClanForceInfo* allianceArmyInfo = NULL;
    if (ab.has_clan_force_info())
        allianceArmyInfo = &(ab.clan_force_info());
    return CalculateBattleResult(m_PVEVillageMap[idstr],
                                 ab.attackunits(),
                                 ab.spells(),
                                 allianceArmyInfo,
                                 0,
                                 0,
                                 ab.totaltime(),
                                 ab.playerlid(),
                                 false,
                                 true,
                                 goldCount,
                                 foodCount,
                                 diamondCount);
}

void FightCalculator::OnAddRobGoldCount(unsigned int addRobGoldCount)
{}
void FightCalculator::OnAddRobFoodCount(unsigned int addRobFoodCount)
{}
void FightCalculator::OnStarCountChange(unsigned int starCount)
{}
void FightCalculator::On100PercentDestroy(unsigned char addStarCount)
{}
void FightCalculator::OnArmyAllDead()
{}
void FightCalculator::OnCenterDead(unsigned int centerLevel)
{
    m_CenterLevel = centerLevel;
}

void FightCalculator::BombDead(string cateName,unsigned int index)
{
    if(!m_pPVPVillageInfo)
        return;
    
    if(cateName == "Bomb" && m_pPVPVillageInfo->bomb_size() > index)
        m_pPVPVillageInfo->mutable_bomb(index)->Clear();
    else if(cateName == "GiantBomb" && m_pPVPVillageInfo->giantbomb_size() > index)
        m_pPVPVillageInfo->mutable_giantbomb(index)->Clear();
    //else if(cateName == "Eject" && m_pPVPVillageInfo->eject_size() > index)
        //m_pPVPVillageInfo->mutable_eject(index)->Clear();
}
void FightCalculator::OnAllianceArmyDeploy(GameLogic::LogicAvatar* avatar,rpc::CharacterType type, int level, GameLogic::Float2 isoPos)
{
    LogicCharacterData* pData = iNew LogicCharacterData();
    pData->_categoryName = CharacterIdEnumToString(type);
    pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(CHARACTERCONFIG, pData->_categoryName, level, "Hitpoints").c_str());
    pData->_bEnemy = true;
    LogicCharacter* pRole = (LogicCharacter*)avatar->GetGameObjectManager()->CreateObject(LOGICCHARACTER,level,pData);
    pRole->SetSubGridPosition(isoPos.x, isoPos.y);
    AllianceArmy aa;
    aa._id = pRole->GetId();
    aa._level = level;
    aa._type = type;
    m_AllianceCharacterMap[avatar->GetAvatarId()].push_back(aa);
    
}
