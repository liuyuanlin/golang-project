//
//  CVillageCheck.h
//  Robot
//
//  Created by 王云川 on 13-12-24.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#ifndef Robot_CVillageCheck_h
#define Robot_CVillageCheck_h

#include <iostream>

#include "Types.h"
#include "LogicTickManager.h"
#include "LogicAvatar.h"
#include "RUserData.h"
#include "CRBuilding.h"

Use_NS_GameLogic

class CVillageCheck : public LogicTicker, public ResChangeListener, public UnitStoredListener, public BattleInfoChangeListener
{
public:
    static CVillageCheck* Instance();
    CVillageCheck(void);
    ~CVillageCheck(void);
    
    virtual void Tick(){};
    
    //村庄
    virtual void OnResourceChange(unsigned int food,unsigned int maxFood,
                                  unsigned int gold,unsigned int maxGold,
                                  unsigned int gem,unsigned int wuhun){};
    virtual void OnStoreUnit(unsigned int barrackId,unsigned int firecampId,
                             rpc::CharacterType typeId,unsigned int level,
                             bool moveToImmediately, bool bReward = false){};
    virtual void OnStoreUnitEnd(){};
    
    //战斗
    virtual void OnAddRobGoldCount(unsigned int addRobGoldCount){};
    virtual void OnAddRobFoodCount(unsigned int addRobFoodCount){};
    virtual void OnStarCountChange(unsigned int starCount){};
    virtual void On100PercentDestroy(unsigned char addStarCount){};
    virtual void OnArmyAllDead(){};
    virtual void OnCenterDead(unsigned int centerLevel){};
    
    //set money
    void SetMoney();
    //Create base builing(like:worker、center)
    void CreateBaseBuild(unsigned int uFileIndex);
    void CreateBaseBuild(RobotBuildingInfo* buildInfo);
    //Create Gold and Food
    bool CreateResourceBuild();
    //Create and upgrade build
    bool CreateBuild(int nIndex, unsigned int uFileIndex);
    //Create village
    void CreateVillage(void);
    
    //Check
    bool CheckVillage();
    
    //member
private:
    LogicAvatar*        m_LogicAvatar;              //玩家信息
    rpc::VillageInfo    m_VillageInfo;              //村庄信息
};
#endif
