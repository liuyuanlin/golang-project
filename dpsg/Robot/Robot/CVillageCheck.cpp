//
//  CVillageCheck.cpp
//  Robot
//
//  Created by 王云川 on 13-12-24.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#include "CVillageCheck.h"
#include "CConnectionMgr.h"
#include "ConfigManager.h"
#include "LogicFloat2.h"
#include "msg.pb.h"
#include "LogicCharacter.h"
#include "LogicFinishNowCommand.h"
#include "LogicHeroSummonComponent.h"
#include "LogicCharacter.h"
#include <stdlib.h>

Use_NS_GameLogic
using namespace GameHub;
using namespace std;

CVillageCheck* CVillageCheck::Instance()
{
    static CVillageCheck crBuilding;
    return &crBuilding;
}

//Constructor
CVillageCheck::CVillageCheck(void)
:LogicTicker::LogicTicker()
{
    LogicTickManager::Instance()->AddTicker(this);
    m_LogicAvatar = LogicAvatarManager::Instance()->CreateAvatar(123, false);

}

//Destructor
CVillageCheck::~CVillageCheck(void)
{
    m_LogicAvatar->Clear();
    LogicTickManager::Instance()->RemoveTicker(this);
    LogicAvatarManager::Instance()->DestroyAvatar(123);
}

//Set money
void CVillageCheck::SetMoney(void)
{
    m_LogicAvatar->GetVillage()->StoreGold(10000000);
    m_LogicAvatar->GetVillage()->StoreFood(10000000);
    m_LogicAvatar->GetVillage()->StoreYuanBao(10000000);
    m_LogicAvatar->GetVillage()->StoreWuHun(10000000);
}

//create base
void CVillageCheck::CreateBaseBuild(RobotBuildingInfo *buildInfo)
{
    rpc::BuildingId_IdType typeId = BuildingClassNameToIdEnum(buildInfo->templateName);
    int index = m_LogicAvatar->GetVillage()->GetBuildingCount(typeId);
    int level = 1;
    
    LogicBuildingData* pData = iNew LogicBuildingData();
    pData->_categoryName = buildInfo->templateName;
    pData->_idx = index;
    pData->_hp = atoi(ConfigManager::Instance()->GetAttribute(BUILDINGCONFIG, pData->_categoryName, level, "Hitpoints").c_str());
    LogicBuilding* pObj = (LogicBuilding*)m_LogicAvatar->GetGameObjectManager()->CreateObject(LOGICBUILDING,level,pData);
    //TODO:添加对pObj的处理
    pObj->SetGridPosition(buildInfo->x,buildInfo->y);
    m_LogicAvatar->GetVillage()->AddBuilding(typeId,pObj);
}

//Create base
void CVillageCheck::CreateBaseBuild(unsigned int uFileIndex)
{
    CRBuilding* pRBuild = CRBuilding::Instance();
    RobotBuildingInfo* pBuildInfo = NULL;
    unsigned int j = 0;
    while ((pBuildInfo = pRBuild->GetBuildingInfo(rpc::BuildingId_IdType_Worker, j))) {
        SetMoney();
        CreateBaseBuild(pBuildInfo);
        ++j;
    }
    
    j = 0;
    while ((pBuildInfo = pRBuild->GetBuildingInfo(rpc::BuildingId_IdType_Center, j))) {
        SetMoney();
        CreateBaseBuild(pBuildInfo);
        ++j;
    }
}

//Create Resource build
bool CVillageCheck::CreateResourceBuild(void)
{
    bool bOK = false;
    CRBuilding* crBuild = CRBuilding::Instance();
    RobotBuildingInfo* buildInfo = NULL;
    LogicBuilding* newBuilding = NULL;
    rpc::BuildingId_IdType idTypes[2] = {rpc::BuildingId_IdType_GoldStorage, rpc::BuildingId_IdType_FoodStorage};
    
    //升级所有资源库
    for(int i = 0; i != 2; ++i)
    {
        unsigned int jCount = m_LogicAvatar->GetVillage()->GetBuildingCount(idTypes[i]);
        
        for(int j = 0; j != jCount; ++j)
        {
            if((buildInfo = crBuild->GetBuildingInfo(idTypes[i], j)))
            {
                newBuilding = m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], j);
                if(!newBuilding)
                {
                    GH_INFO("没有该建筑=%d，无法升级", i);
                    continue;
                }
                int upgradeCount = buildInfo->level - newBuilding->GetLevel();
                //GH_INFO("当前等级=%d， 需要升级等级=%d", newBuilding->GetLevel(), upgradeCount);
                while (upgradeCount > 0)
                {
                    SetMoney();//取得GM资源
                    if(CRBuilding::Instance()->UpgradeBuilding(m_LogicAvatar, m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], newBuilding->GetLogicData()->_idx)))
                    {
                        bool tmpOK = CRBuilding::Instance()->FinishNowBuilding(m_LogicAvatar, m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], newBuilding->GetLogicData()->_idx));
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
        //GH_INFO("当前资源=%d，数量=%d", i, m_LogicAvatar->GetVillage()->GetBuildingCount(idTypes[i]));
        if((buildInfo = crBuild->GetBuildingInfo(idTypes[i], m_LogicAvatar->GetVillage()->GetBuildingCount(idTypes[i]))))
        {
            //建造
            //SetMoney();//取得GM资源
            newBuilding = crBuild->CreateNewBuilding(m_LogicAvatar, buildInfo);
            if(!newBuilding)
            {
                GH_INFO("创建资源建筑失败");
                continue;
            }
            bool tmpOK = crBuild->FinishNowBuilding(m_LogicAvatar, newBuilding);
            if(!bOK)bOK = tmpOK;
            
            //GH_INFO("创建资源建筑成功");
            
            //升级
            if(!bOK)continue;
            int upgradeCount = buildInfo->level - newBuilding->GetLevel();
            while (upgradeCount > 0)
            {
                SetMoney();//取得GM资源
                if(CRBuilding::Instance()->UpgradeBuilding(m_LogicAvatar, m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], newBuilding->GetLogicData()->_idx)))
                {
                    tmpOK = CRBuilding::Instance()->FinishNowBuilding(m_LogicAvatar, m_LogicAvatar->GetVillage()->GetBuilding(idTypes[i], newBuilding->GetLogicData()->_idx));
                    if(!bOK)bOK = tmpOK;
                    //GH_INFO("当前资源建筑=%d，升级成功=%d, 当前等级=%d", i, bOK, newBuilding->GetLevel());
                }
                --upgradeCount;
            }
            if(tmpOK)--i;
        }
    }
    return bOK;
}

void CVillageCheck::CreateVillage(void)
{
    m_LogicAvatar->Clear();
    m_LogicAvatar = LogicAvatarManager::Instance()->CreateAvatar(123, false);
    m_LogicAvatar->GetVillage()->SetResourceChangeListener(this);
    m_LogicAvatar->GetVillage()->SetUnitStoredListener(this);
    m_LogicAvatar->GetBattle()->SetBattleInfoChangeListener(this);
}

//create build
bool CVillageCheck::CreateBuild(int nIndex, unsigned int uFileIndex)
{
    CRBuilding* pRBuild = CRBuilding::Instance();
    RobotBuildingInfo* pBuildInfo = NULL;
    LogicBuilding* pBuilding = NULL;
    unsigned int j = 0;
    int i = nIndex;
    int upgradeCount = 0;
    while ((pBuildInfo = pRBuild->GetBuildingInfo((rpc::BuildingId_IdType)i, j)))
    {
        pBuilding = m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, j);
        if(!pBuilding)
        {
            //加钱
            SetMoney();
            pBuilding = pRBuild->CreateNewBuilding(m_LogicAvatar, pBuildInfo);
            
            if(!pBuilding)
            {
                //修建金库何木头
                if(CreateResourceBuild())
                    continue;
                else
                {
                    GH_INFO("创建建筑失败，建筑类型=%d,配置表id=%d,该类别第%d条", i, uFileIndex, j);
                    return false;
                    break;
                }
            }
            //快速完成失败
            if(!pRBuild->FinishNowBuilding(m_LogicAvatar, pBuilding))
            {
                GH_INFO("快速完成建筑失败，建筑类型=%d,配置表id=%d,该类别第%d条", i, uFileIndex,j);
                return false;
                break;
            }
            //GH_INFO("成功创建建筑=%d", i);
        }
        
        //升级建筑到配置等级
        if(pBuilding)
        {
            upgradeCount = pBuildInfo->level - pBuilding->GetLevel();
            while (upgradeCount > 0)
            {
                SetMoney();//取得GM资源
                bool bUpgrade = false;
                if((bUpgrade = CRBuilding::Instance()->UpgradeBuilding(m_LogicAvatar, m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, pBuilding->GetLogicData()->_idx))))
                {
                    if (rpc::BuildingId_IdType_Wall != i)
                    {
                        if(!CRBuilding::Instance()->FinishNowBuilding(m_LogicAvatar, m_LogicAvatar->GetVillage()->GetBuilding((rpc::BuildingId_IdType)i, pBuilding->GetLogicData()->_idx)))
                        {
                            GH_INFO("快速完成建筑失败，建筑类型=%d,配置表id=%d,该类别第%d条", i, uFileIndex,j);
                            return false;
                            break;
                        }
                    }
                    //GH_INFO("建筑=%d升级成功,当前建筑等级=%d,需要提升到=%d", i, pBuilding->GetLevel(), pBuildInfo->level);
                }
                if(!bUpgrade)
                {
                    //修建金库何木头
                    if(CreateResourceBuild())
                    {
                        //GH_INFO("先升级资源库来升级该建筑=%d", i);
                        continue;
                    }
                    else
                    {
                        GH_INFO("升级失败，建筑类型=%d,配置表id=%d,该类别第%d条, 需要升级的等级%d", i, uFileIndex,j, upgradeCount);
                        return false;
                        break;
                    }
                }
                --upgradeCount;
            }
        }
        else break;
        ++j;
    }
    return true;
}

//Check
bool CVillageCheck::CheckVillage(void)
{
    
    unsigned int nVillageCount = RUserData::Instance()->GetVillageCount();
    if (0 == nVillageCount)
    {
        GH_INFO("当前没有村庄配置，请检测配置路径\n");
    }

    //遍历所有村庄配置
    while (nVillageCount)
    {
        CreateVillage();
        m_LogicAvatar->GetVillage()->InitVillage(44, 44, false);
        
        //测试该配置是否能够完成
        CRBuilding::Instance()->ReadRobotVillageXML(nVillageCount);
        GH_INFO("开始检测村庄配置，当前=%d张配置", nVillageCount);
        CreateBaseBuild(nVillageCount);
        
        for(int32 i = rpc::BuildingId_IdType_Center; i != rpc::BuildingId_IdType_End; ++i)
        {
            if (!CreateBuild(i, nVillageCount)) {
                return false;
            }
        }
        
        --nVillageCount;
    }
    
    return true;

}









