//
//  CRobot.cpp
//  Robot
//
//  Created by PU on 13-4-16.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#include "CRobot.h"
#include "CConnectionMgr.h"
#include "ConfigManager.h"
#include "LogicFloat2.h"
#include "msg.pb.h"
#include "CRBuilding.h"
#include "LogicCharacter.h"
#include "LogicFinishNowCommand.h"
#include "LogicHeroSummonComponent.h"
#include <stdlib.h>

Use_NS_GameLogic
using namespace GameHub;
using namespace std;

#define ReadyToBattleTickTime   100
#define BattleEndTickTime   250

CRobot::CRobot(GameHub::uint32 connectorID, int nVillageConfig, int nIndex, string udid)
:LogicTicker::LogicTicker(),
m_connectorID(connectorID), m_UID(udid), m_GateKey(""), m_eConnectState(CS_None), m_TotalTickTime(0),
m_bLogin(false), m_GameState(GS_None), m_BattleTickCount(-1), m_ReadyToBattleTickCount(0), m_OriginYuanBao(0), m_OriginWuHun(0), m_uConfigIndex(nIndex), m_bNameOK(false), m_bLogOut(false)
{
    //m_LogOutTime = random() % (600 * 5);
    LogicTickManager::Instance()->AddTicker(this);
    m_LogicAvatar = LogicAvatarManager::Instance()->CreateAvatar(connectorID, false);
    m_bAllBuilding = m_UID.length() > 0;
    m_bThinkFighting = m_bAllBuilding;
    
    m_pBuildInfo = new CRBuilding();
    m_pBuildInfo->ReadRobotVillageXML(nVillageConfig);
    m_pEnemy = NULL;
    GH_INFO("分配到的配置文件id=%d",nVillageConfig);
}

CRobot::~CRobot()
{
    m_LogicAvatar->Clear();
    LogicTickManager::Instance()->RemoveTicker(this);
    LogicAvatarManager::Instance()->DestroyAvatar(m_connectorID);
    
    if (m_pBuildInfo) {
        delete m_pBuildInfo;
        m_pBuildInfo = NULL;
    }
}

void CRobot::Tick()
{
    if(m_eConnectState == CS_CNS && m_bLogin == true)
    {
        switch (m_GameState)
        {
            case GS_Edit:
            {
                if(!m_bAllBuilding && m_bNameOK)
                    ExtendVillageAll();
                
                //ask for fighting
                if (m_bThinkFighting && m_ReadyToBattleTickCount++ >= ReadyToBattleTickTime)
                {
                    m_ReadyToBattleTickCount = 0;
                    AskFighting();
                }
            }
                break;
            case GS_Battle:
            {
                //SendArmy();
                if(m_BattleTickCount >= 0 && (++m_BattleTickCount >= BattleEndTickTime))
                    ReturnHome();
            }
                break;
            default:
                break;
        }
        //if(m_TotalTickTime >= m_LogOutTime)LogOut();
        //if(m_bAllBuilding && m_TotalTickTime > 600)LogOut();
    }
    else if(m_eConnectState == CS_Reconnect && m_TotalTickTime >= 50)
        ReConnect();
    
    if (m_bLogOut)
    {
        LogOut();
        m_bLogOut = false;
    }
    ++m_TotalTickTime;
}

//add by wyc 2013-12-26 取名
void CRobot::AskForName()
{
    //取名
    rpc::UpdatePlayerInfo update;
    update.set_name(RUserData::Instance()->GetName());
    CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.SetPlayerName",&update);
}

void CRobot::Login()
{
    m_TotalTickTime = m_ReadyToBattleTickCount = 0;
    rpc::Login login;
    login.set_uid(m_UID.c_str());
    login.set_gatekey(m_GateKey.c_str());
    //login.set_appid(110);
    //login.set_openid("openid");
    //login.set_openkey("openkey");
    //login.set_platformtype(1);
    //login.set_imtype((rpc::IMType)1);
    
    //login.set_thirdpartyid("1234567890");
    //login.set_auth_code("0987654321");
    login.set_channelid((rpc::GameLocation)RUserData::Instance()->m_ChannelID);//支付宝账号1 服务器账号0
    CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.Login", &login);
    cout<<"LOGIN"<<endl;
}

void CRobot::LogOut()
{
    m_TotalTickTime = m_ReadyToBattleTickCount = 0;
    CConnectionMgr::GetSingleton().DisconnectFromGame(m_connectorID);
    m_bLogin = false;
    m_eConnectState = CS_Reconnect;
    m_GameState = GS_None;
    m_BattleTickCount = -1;
    LogicAvatarManager::Instance()->DestroyAvatar(m_connectorID);
    //cout<<"LOGOUT"<<endl;
    GH_INFO("机器人%d下线休息了", m_connectorID);
}

void CRobot::ReConnect()
{
    m_eConnectState = CS_None;
    m_GameState = GS_None;
    m_BattleTickCount = -1;
    
    m_connectorID = CConnectionMgr::GetSingleton().ConnectToGame(RUserData::Instance()->m_vecIp[m_uConfigIndex].c_str(), RUserData::Instance()->m_vecPort[m_uConfigIndex]);
    m_LogicAvatar = LogicAvatarManager::Instance()->CreateAvatar(m_connectorID, false);
    CConnectionMgr::GetSingleton().m_RobotMap[m_connectorID] = this;
}

void CRobot::GMFunc()
{
    rpc::Ping xx;
    CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.AddMoneyForGM", &xx);
    m_LogicAvatar->GetVillage()->StoreGold(10000000);
    m_LogicAvatar->GetVillage()->StoreFood(10000000);
    m_LogicAvatar->GetVillage()->StoreYuanBao(10000000);
    m_LogicAvatar->GetVillage()->StoreWuHun(10000000);
}

void CRobot::OnResourceChange(unsigned int food,unsigned int maxFood,
                              unsigned int gold,unsigned int maxGold,
                              unsigned int gem,unsigned int wuhun)
{}
void CRobot::OnStoreUnit(unsigned int barrackId,unsigned int firecampId,
                         rpc::CharacterType typeId,unsigned int level,
                         bool moveToImmediately, bool bReward)
{}
void CRobot::OnStoreUnitEnd()
{}

void CRobot::CreateVillage(const rpc::VillageInfo &info,bool battlefield)
{
    LogicAvatarManager::Instance()->DestroyAvatar(m_connectorID);
    m_LogicAvatar = LogicAvatarManager::Instance()->CreateAvatar(m_connectorID, false);
    m_LogicAvatar->GetVillage()->SetResourceChangeListener(this);
    m_LogicAvatar->GetVillage()->SetUnitStoredListener(this);
    m_LogicAvatar->GetBattle()->SetBattleInfoChangeListener(this);
    m_LogicAvatar->GetVillage()->SetYuanBao(m_OriginYuanBao);
    m_LogicAvatar->GetVillage()->SetWuHun(m_OriginWuHun);
    
    m_VillageInfo = info;
    //m_pWorkerMgr = iNew WorkerManager();
    //m_pWorkerMgr->SetVillage(m_pVillage);
    m_pBuildInfo->SyncBuildings(m_LogicAvatar, m_VillageInfo, battlefield);
}

void CRobot::ExtendVillageAll()
{
    //CRBuilding::Instance()->ReadRobotVillageXML();
    m_bAllBuilding = true;
    
    //CRBuilding* crBuild = CRBuilding::Instance();
    RobotBuildingInfo* buildInfo = NULL;
    LogicBuilding* newBuilding = NULL;
    
    //清除所有障碍物
    for(int32 i = rpc::BuildingId_IdType_Barrier1; i != rpc::BuildingId_IdType_End; ++i)
    {
        GMFunc();//取得GM资源
        unsigned int jCount = m_LogicAvatar->GetVillage()->GetBuildingCount((rpc::BuildingId_IdType)i);
        for(int j = 0; j != jCount; ++j)
        {
            newBuilding = m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, 0);
            if(newBuilding)
            {
                if(m_pBuildInfo->RemoveBuilding(this, newBuilding))
                {
                    if(m_pBuildInfo->FinishNowBuilding(this, newBuilding))
                    {
                        m_LogicAvatar->GetVillage()->RemoveBuilding((rpc::BuildingId_IdType)i, newBuilding->GetLogicData()->_idx);
                    }
                }
            }
        }
    }
    
    //重新放置位置
    int posX = 4;
    for(int32 i = rpc::BuildingId_IdType_Center; i != rpc::BuildingId_IdType_End; ++i)
    {
        unsigned int jCount = m_LogicAvatar->GetVillage()->GetBuildingCount((rpc::BuildingId_IdType)i);
        for(int j = 0; j != jCount; ++j)
        {
            newBuilding = m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, j);
            if(newBuilding)
            {
                newBuilding->SetGridPosition(posX, 4);
                
                rpc::MoveTo mt;
                mt.mutable_p()->set_x(posX);
                mt.mutable_p()->set_y(4);
                rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(newBuilding->GetLogicData()->_categoryName);
                mt.mutable_id()->set_type(typeId);
                mt.mutable_id()->set_index(newBuilding->GetLogicData()->_idx);
                CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.MoveTo", &mt);
                
                posX += newBuilding->GetBuildingSize();
            }
        }
    }
    
    for(int32 i = rpc::BuildingId_IdType_Center; i != rpc::BuildingId_IdType_End; ++i)
    {
        unsigned int jCount = m_LogicAvatar->GetVillage()->GetBuildingCount((rpc::BuildingId_IdType)i);
        for(int j = 0; j != jCount; ++j)
        {
            buildInfo = m_pBuildInfo->GetBuildingInfo((rpc::BuildingId_IdType)i, j);
            newBuilding = m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, j);
            if(buildInfo && newBuilding)
            {
                newBuilding->SetGridPosition(buildInfo->x, buildInfo->y);
                
                rpc::MoveTo mt;
                mt.mutable_p()->set_x(buildInfo->x);
                mt.mutable_p()->set_y(buildInfo->y);
                rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(newBuilding->GetLogicData()->_categoryName);
                mt.mutable_id()->set_type(typeId);
                mt.mutable_id()->set_index(newBuilding->GetLogicData()->_idx);
                CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.MoveTo", &mt);
            }
        }
    }
    
    //建造
    //在这里建造从配置里面读取的村庄信息
    int upgradeCount = 0;
    for(int32 i = rpc::BuildingId_IdType_Center; i != rpc::BuildingId_IdType_End; ++i)
    {
        unsigned int j = 0;
        while ((buildInfo = m_pBuildInfo->GetBuildingInfo((rpc::BuildingId_IdType)i, j)))
        {
            newBuilding = m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, j);
            if(!newBuilding)
            {
                //新建造建筑
                GMFunc();//取得GM资源
                newBuilding = m_pBuildInfo->CreateNewBuilding(this, buildInfo);
                if(!newBuilding)
                {
                    if(BuildResourceBuilding())continue;
                    else break;
                }
                if(!m_pBuildInfo->FinishNowBuilding(this, newBuilding))break;
            }
            //升级建筑
            if(newBuilding)
            {
                upgradeCount = buildInfo->level - newBuilding->GetLevel();
                while (upgradeCount > 0)
                {
                    GMFunc();//取得GM资源
                    bool bUpgrade = false;

                    if((bUpgrade = m_pBuildInfo->UpgradeBuilding(this, m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, newBuilding->GetLogicData()->_idx))))
                    {
                        if (rpc::BuildingId_IdType_Wall != i)
                        {
                            if(!m_pBuildInfo->FinishNowBuilding(this, m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, newBuilding->GetLogicData()->_idx)))break;
                        }
                    }
                    if(!bUpgrade)
                    {
                        if(BuildResourceBuilding())continue;
                        else break;
                    }
                    --upgradeCount;
                }
            }
            else break;
            ++j;
        }
    }
    
    //GH_INFO("建造完成");
    
    //设置英雄
    int heroBuildingCount = m_LogicAvatar->GetVillage()->GetBuildingCount(rpc::BuildingId_IdType_GeneralHouse);
    for(int i = 0; i != heroBuildingCount; ++i)
    {
        buildInfo = m_pBuildInfo->GetBuildingInfo(rpc::BuildingId_IdType_GeneralHouse, i);
        if(!buildInfo)continue;
        newBuilding = m_LogicAvatar->GetVillage()->GetBuilding(rpc::BuildingId_IdType_GeneralHouse, i);
        if(!newBuilding)continue;
        
        //解锁英雄
        stringstream commandData;
        commandData.write((char*)&i,sizeof(int));
        commandData.write((char*)&buildInfo->arg1,sizeof(int));
        if(m_LogicAvatar->GetCommandManager()->PushCommand(Command_UnlockHero, commandData))
        {
            //notify server
            rpc::HeroChoose msg;
            msg.set_idx(i);
            msg.set_type((rpc::CharacterType)buildInfo->arg1);
            CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.HeroCreate", &msg);
        }
        
        //选择英雄
        rpc::HeroChoose msg;
        msg.set_idx(i);
        msg.set_type((rpc::CharacterType)buildInfo->arg1);
        CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.HeroChoose", &msg);
        
        LogicHeroSummonComponent* heroComponent = newBuilding->GetComponent<LogicHeroSummonComponent>();
        HeroData* heroData = heroComponent->GetHeroData((rpc::CharacterType)buildInfo->arg1);
        if(!heroData)continue;
        int heroLevel = buildInfo->arg2 - heroData->_level;
        while (heroLevel > 0)
        {
            --heroLevel;
            //升级英雄
            TimeValue tv;
            LogicTimer::GetTime(&tv);
            unsigned int startTime = tv._Seconds;
            commandData.clear();
            commandData.write((char*)&i,sizeof(int));
            commandData.write((char*)&buildInfo->arg1,sizeof(int));
            commandData.write((char*)&startTime, sizeof(int));
            if(m_LogicAvatar->GetCommandManager()->PushCommand(Command_HeroUpgrade, commandData))
            {
                //notify server
                rpc::HeroChoose msg;
                msg.set_idx(i);
                msg.set_type((rpc::CharacterType)buildInfo->arg1);
                CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.HeroUpgrade", &msg);
            }
            
            //秒升级英雄
            commandData.clear();
            commandData.write((char*)&i,sizeof(int));
            commandData.write((char*)&buildInfo->arg1,sizeof(int));
            if(m_LogicAvatar->GetCommandManager()->PushCommand(Command_HeroUpgradeFinishNow, commandData))
            {
                //notify server
                rpc::HeroChoose msg;
                msg.set_idx(i);
                msg.set_type((rpc::CharacterType)buildInfo->arg1);
                CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.HeroFinishNow", &msg);
            }
        }
    }
    //GH_INFO("设置英雄");
    
    //随机令旗
    int centerLevel = m_LogicAvatar->GetVillage()->GetBuilding(rpc::BuildingId_IdType_Center, 0)->GetLevel();
    int getTrophy = 0;
    if(centerLevel > 1)
    {
        TrophyInfo* pTrophy = RUserData::Instance()->m_vecTrophyInfo[centerLevel-1];
        if (pTrophy->uMaxNumber > 0)
            getTrophy = pTrophy->uBaseNumber + random()%(pTrophy->uMaxNumber - pTrophy->uBaseNumber);
        else
            getTrophy = 0;
    }
    if (getTrophy > 0)
    {
        rpc::UpdatePlayerInfo msg;
        msg.set_trophy(getTrophy);
        CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.AddTrophy", &msg);
    }
    
    //Now we think about the fighting
    m_bThinkFighting = RUserData::Instance()->m_bOpenFight;
    m_TotalTickTime = 0;
    if (!m_bThinkFighting)
        LogOut();
    GH_INFO("机器人ID=%d已完成建筑，压入成功，主机等级=%d， 随机令旗=%d", m_connectorID, centerLevel, getTrophy);
}
bool CRobot::BuildResourceBuilding()
{
    bool bOK = false;
    //CRBuilding* crBuild = CRBuilding::Instance();
    RobotBuildingInfo* buildInfo = NULL;
    LogicBuilding* newBuilding = NULL;
    rpc::BuildingId_IdType idTypes[2] = {rpc::BuildingId_IdType_GoldStorage, rpc::BuildingId_IdType_FoodStorage};
    
    //升级所有资源库
    for(int i = 0; i != 2; ++i)
    {
        unsigned int jCount = m_LogicAvatar->GetVillage()->GetBuildingCount(idTypes[i]);
        for(int j = 0; j != jCount; ++j)
        {
            if((buildInfo = m_pBuildInfo->GetBuildingInfo(idTypes[i], j)))
            {
                newBuilding = m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], j);
                if(!newBuilding)continue;
                int upgradeCount = buildInfo->level - newBuilding->GetLevel();
                while (upgradeCount > 0)
                {
                    GMFunc();//取得GM资源
                    if(m_pBuildInfo->UpgradeBuilding(this, m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], newBuilding->GetLogicData()->_idx)))
                    {
                        bool tmpOK = m_pBuildInfo->FinishNowBuilding(this, m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], newBuilding->GetLogicData()->_idx));
                        if(!bOK)bOK = tmpOK;
                    }
                    --upgradeCount;
                }
            }
        }
    }
    
    //新建资源库
    for(int i = 0; i != 2; ++i)
    {
        if((buildInfo = m_pBuildInfo->GetBuildingInfo(idTypes[i], m_LogicAvatar->GetVillage()->GetBuildingCount(idTypes[i]))))
        {
            //建造
            GMFunc();//取得GM资源
            newBuilding = m_pBuildInfo->CreateNewBuilding(this, buildInfo);
            if(!newBuilding)continue;
            bool tmpOK = m_pBuildInfo->FinishNowBuilding(this, newBuilding);
            if(!bOK)bOK = tmpOK;
            
            //升级
            if(!bOK)continue;
            int upgradeCount = buildInfo->level - newBuilding->GetLevel();
            while (upgradeCount > 0)
            {
                GMFunc();//取得GM资源
                if(m_pBuildInfo->UpgradeBuilding(this, m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], newBuilding->GetLogicData()->_idx)))
                {
                    tmpOK = m_pBuildInfo->FinishNowBuilding(this, m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], newBuilding->GetLogicData()->_idx));
                    if(!bOK)bOK = tmpOK;
                }
                --upgradeCount;
            }
            if(tmpOK)--i;
        }
    }
    return bOK;
}

void CRobot::OnAddRobGoldCount(unsigned int addRobGoldCount)
{}
void CRobot::OnAddRobFoodCount(unsigned int addRobFoodCount)
{}
void CRobot::OnStarCountChange(unsigned int starCount)
{}
void CRobot::On100PercentDestroy(unsigned char addStarCount)
{
    GH_INFO("好牛逼，机器人%d完全摧毁了敌人", m_connectorID);
    ReturnHome();
}
void CRobot::OnArmyAllDead()
{
    GH_INFO("渣渣，机器人%d完败", m_connectorID);
    ReturnHome();
}
void CRobot::OnCenterDead(unsigned int centerLevel)
{}

//请战
void CRobot::AskFighting()
{
    //Check the army
    GH_INFO("机器人%d请战", m_connectorID);
    TrainingAllArmy();
    SearchEnemy();
    m_bThinkFighting = false;
}

////Set enemy info
//void CRobot::SetEnemyInfo(const rpc::VillageInfo &stInfo)
//{
//    //
//    if (m_pEnemy)
//        LogicAvatarManager::Instance()->DestroyAvatar(m_connectorID+10000);
//    m_pEnemy = LogicAvatarManager::Instance()->CreateAvatar(m_connectorID+10000, false);
//    CRBuilding::Instance()->SyncBuildings(m_pEnemy, stInfo, true);
//}

void CRobot::SearchEnemy()
{
    m_ReadyToBattleTickCount = 0;
    //rpc
    rpc::Ping xx;
    CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.RandomMatch",&xx);
    cout<<"Search Enemy"<<endl;
}

//Generate a pos
void CRobot::GeneratePos(GameLogic::Float2 &stPos)
{
    int x = random() % 44;
    int y = random() % 44;
    
    if (m_LogicAvatar && m_LogicAvatar->GetVillage()->IsBlock(x, y, DeployBlock))
    {
        stPos.x = (float)x;
        stPos.y = (float)y;
    }
    else
        GeneratePos(stPos);
}

//Set the army when attack others
void CRobot::SendArmy()
{
    if(m_MatchPlayerInfo.own_char_size() <= 0)
        return;
    if(m_BattleTickCount < 0)
        m_BattleTickCount = 0;
    
    Float2 isoPos = Float2(0, 0);
    GeneratePos(isoPos);
    
    //GH_INFO("随机到的攻击位置X=%f, Y=%f", isoPos.x, isoPos.y);
    int iCount = m_MatchPlayerInfo.own_char_size();
    for(int i = 0; i != iCount; ++i)
    {
        rpc::Character* charater = m_MatchPlayerInfo.mutable_own_char(i);
        if(charater->count())
        {
            string cateName = CharacterIdEnumToString(charater->type());
            int level = charater->level();
            
            LogicCharacterData* pDatas = iNew LogicCharacterData();
            pDatas->_categoryName = cateName;
            pDatas->_hp = atoi(ConfigManager::Instance()->GetAttribute(CHARACTERCONFIG, cateName, level, "Hitpoints").c_str());
            pDatas->_bEnemy = false;
            LogicCharacter* pLogicCharacter = (LogicCharacter*)m_LogicAvatar->GetGameObjectManager()->CreateObject(LOGICCHARACTER, level, pDatas);
            pLogicCharacter->SetSubGridPosition(isoPos.x,isoPos.y);
            charater->set_count(charater->count() - 1);
            
            rpc::AttackerInfo attackerInfo;
            attackerInfo.set_droptime(m_BattleTickCount);
            attackerInfo.set_type(charater->type());
            attackerInfo.set_level(charater->level());
            attackerInfo.mutable_p()->set_x(isoPos.x);
            attackerInfo.mutable_p()->set_y(isoPos.y);
            CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.AttackerInfo", &attackerInfo);
            GH_INFO("Attack info");
            //m_LogicAvatar->GetBattle()->AddSendArmy(charater);
            return;
        }
    }
}

void CRobot::ReturnHome()
{
    rpc::NotifyBattleEnd nbe;
    nbe.set_totaltime(m_BattleTickCount);
    CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.NotifyBattleEnd", &nbe);
    
    m_BattleTickCount = 0;
    rpc::Ping xx;
    CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.ReturnHome", &xx);
    m_bLogOut = true;
    cout<<"Return Home"<<endl;
}

void CRobot::TrainingAllArmy()
{
    int i = 10;
    while(i--)
    {
        GMFunc();
        bool bTraining = false;
        for(int i = rpc::Barbarian; i != rpc::PEKKA + 1; ++i)
        {
            bTraining = TrainingArmy((rpc::CharacterType)i, 11 - i);
            if(bTraining)
                FinishNowTrainArmy();
        }
    }
}

bool CRobot::TrainingArmy(rpc::CharacterType typeId, unsigned int count)
{
    int idx = 0;
    int enumId = (int)typeId;
    int level = 1;
    stringstream commandData;
    commandData.write((char*)&idx, sizeof(int));
    commandData.write((char*)&enumId, sizeof(int));
    commandData.write((char*)&level, sizeof(int));
    if(m_LogicAvatar->GetCommandManager()->PushCommand(Command_TrainingUnit, commandData))
    {
        m_LogicAvatar->GetCommandManager()->Tick();
        
        rpc::Training msg;
        msg.mutable_character()->set_count(count);
        msg.mutable_character()->set_type(typeId);
        msg.set_index(idx);
        CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.Training", &msg);
        //GH_INFO("Training success");
        return true;
    }
    return false;
}

bool CRobot::FinishNowTrainArmy()
{
    //GH_INFO("Finish train army");
    rpc::BuildingId_IdType typeId = rpc::BuildingId_IdType_Barrack;
    unsigned int buildingIdx = 0;
    int finishType = LogicFinishNowCommand::FT_Training;
    stringstream commandData;
    commandData.write((char*)&typeId,sizeof(int));
    commandData.write((char*)&buildingIdx,sizeof(int));
    commandData.write((char*)&finishType, sizeof(int));
    if(m_LogicAvatar->GetCommandManager()->PushCommand(Command_FinishNow, commandData))
    {
        m_LogicAvatar->GetCommandManager()->Tick();
        
        //send rpc
        rpc::BuildingId msg;
        msg.set_type(typeId);
        msg.set_index(buildingIdx);
        CConnectionMgr::GetSingleton().Call(m_connectorID, "CNServer.Barrack_FinishNow", &msg);
        return true;
    }
    return false;
}