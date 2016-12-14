//
//  CRobot.h
//  Robot
//
//  Created by PU on 13-4-16.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#ifndef __Robot__CRobot__
#define __Robot__CRobot__

#include <iostream>

#include "Types.h"
#include "LogicTickManager.h"
#include "LogicAvatar.h"
#include "RUserData.h"
//#include "CRBuilding.h"

Use_NS_GameLogic

class CRBuilding;

class CRobot : public LogicTicker, public ResChangeListener, public UnitStoredListener, public BattleInfoChangeListener
{
public:
    CRobot(GameHub::uint32 connectorID, int nVillageConfig, int nIndex, string udid = "");
    ~CRobot();
    
    virtual void Tick();
    
    void SetUDID(string udid){if(m_UID == udid)return; m_UID = udid; RUserData::Instance()->SetUDID(m_UID);}
    string GetUDID(){return m_UID;}
    
    void Login();
    void LogOut();
    void ReConnect();
    
    void GMFunc();
    
    //村庄
    virtual void OnResourceChange(unsigned int food,unsigned int maxFood,
                                  unsigned int gold,unsigned int maxGold,
                                  unsigned int gem,unsigned int wuhun);
    virtual void OnStoreUnit(unsigned int barrackId,unsigned int firecampId,
                             rpc::CharacterType typeId,unsigned int level,
                             bool moveToImmediately, bool bReward = false);
    virtual void OnStoreUnitEnd();
    void CreateVillage(const rpc::VillageInfo &info,bool battlefield);
    void ExtendVillageAll();
    bool BuildResourceBuilding();
    
    //add by wyc 2013-12-26 取名
    void        AskForName();
    bool        GetIsNewRobot() {return m_bAllBuilding;};
    
    
    //战斗
    virtual void OnAddRobGoldCount(unsigned int addRobGoldCount);
    virtual void OnAddRobFoodCount(unsigned int addRobFoodCount);
    virtual void OnStarCountChange(unsigned int starCount);
    virtual void On100PercentDestroy(unsigned char addStarCount);
    virtual void OnArmyAllDead();
    virtual void OnCenterDead(unsigned int centerLevel);
    
    void SearchEnemy();
    void SendArmy();
    void ReturnHome();
    void TrainingAllArmy();
    bool TrainingArmy(rpc::CharacterType typeId, unsigned int count);
    bool FinishNowTrainArmy();
    
    enum ConnectState
    {
        CS_None,
        CS_Gate,
        CS_CNS,
        CS_Reconnect,
    };
    ConnectState    m_eConnectState;
    
    GameHub::uint32 m_connectorID;
    std::string     m_GateKey;
    
    LogicAvatar*    m_LogicAvatar;
    bool            m_bLogin;
    
    enum GameState
    {
        GS_None,
        GS_Edit,
        GS_Battle,
    };
    GameState    m_GameState;
    
    rpc::MatchPlayer    m_MatchPlayerInfo;
    int m_BattleTickCount;
    
    //add by wyc 2013-12-26 玩家分配所得的村庄信息
    CRBuilding*     m_pBuildInfo;                   //each robot's building infomation
    unsigned int    m_uConfigIndex;                 //robot's building config
    bool            m_bNameOK;
    
    unsigned int    m_OriginYuanBao;
    unsigned int    m_OriginWuHun;
    
    //战斗相关
private:
    bool            m_bThinkFighting;               //考虑战斗
    LogicAvatar*    m_pEnemy;                       //被攻击的敌人
    bool            m_bLogOut;                      //下线
    
public:
    //检查战斗信息
    void    AskFighting();
    void    SetDoFight(bool bFight) {m_bThinkFighting = bFight; m_ReadyToBattleTickCount = 0;};
    
    //处理匹配敌人信息
    //void    SetEnemyInfo(const rpc::VillageInfo &stInfo);
    
    //产生一个可放兵的位置
    void    GeneratePos(GameLogic::Float2 &stPos);
    
    
private:
    string          m_UID;
    rpc::VillageInfo    m_VillageInfo;
    
    //unsigned int    m_LogOutTime;
    unsigned int    m_TotalTickTime;
    
    bool            m_bAllBuilding;                 //did all the building had built successfully
    int             m_ReadyToBattleTickCount;
};

#endif /* defined(__Robot__CRobot__) */
