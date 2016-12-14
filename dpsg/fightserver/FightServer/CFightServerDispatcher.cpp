//
//  CFightServerDispatcher.cpp
//  FightServer
//
//  Created by keenblue on 13-3-30.
//  Copyright (c) 2013年 keenblue. All rights reserved.
//

#include "CFightServerDispatcher.h"
#include "FightCalculator.h"
#include "NetWork.h"
#include "CAcceptorEx.h"
#include "ConfigManager.h"

using namespace Net;

#define REGISTER_CALLBACK_IMP(msg) this->registerMessageCallback<rpc::msg>(Net::delegate::from_method<CFsDispacher, &CFsDispacher::OnSync##msg>(this))

#define PROCESS_MESSAGE_START(msg) void CFsDispacher::OnSync##msg(google::protobuf::Message* m, void* pContext){ rpc::msg* rst = static_cast<rpc::msg*>(m);
#define PROCESS_MESSAGE_END }


CFsDispacher::CFsDispacher()
{
    REGISTER_CALLBACK_IMP(AttackBegin);
    REGISTER_CALLBACK_IMP(AttackEnd);
    REGISTER_CALLBACK_IMP(PVEAttackBegin);
}

CFsDispacher::~CFsDispacher()
{
    
}

PROCESS_MESSAGE_START(AttackBegin)
    printf("attack begin !!!!!!!!!!!!!!!!!!!!!!!!!!!\n");
    unsigned int t0 = GameLogic::LogicTimer::GetTimeMs();
    BattleResult ret = FightCalculator::Instance()->CalculateBattleResult(*rst);
    CAcceptorEx* pAcceptor = static_cast<CAcceptorEx*>(pContext);
    if(pAcceptor)
    {
        rpc::AttackEnd attackEnd;
        attackEnd.set_playerlid(ret._playerlid);
        attackEnd.mutable_v()->CopyFrom(ret.m_vi);
        attackEnd.set_goldstolen(ret._goldStolen);
        attackEnd.set_foodstolen(ret._foodStolen);
        attackEnd.set_damagepercent(ret._damagePercent);
        attackEnd.set_starts(ret._stars);
        attackEnd.set_trophy(ret._trophy);
        attackEnd.set_exp(ret._exp);
        attackEnd.set_wuhun(ret._wuhun);
        
        rpc::Request req;
        req.set_method("CNServer.AttackEnd");
        std::string data;
        attackEnd.SerializeToString(&data);
        req.set_serialized_request(data);
        data.clear();
        req.SerializeToString(&data);
        
        std::string output;
        snappy::Compress(data.data(), data.size(), &output);
        
        unsigned int size = output.size();
        
        pAcceptor->Send(&size, sizeof(size));
        pAcceptor->Send(output.c_str(), size);
#ifdef _SUPER_DEBUG_
        printf("!!!!!!!!!!!!!!!!!!!!playid--: %d\n",ret._playerlid);
#endif
    }
unsigned int t1 = GameLogic::LogicTimer::GetTimeMs();
unsigned int deltaTime = t1 - t0;
printf("---battle time cost: %d 毫秒\n",deltaTime);
PROCESS_MESSAGE_END;

PROCESS_MESSAGE_START(AttackEnd)
printf("AttackEnd !!!!!!!!!!!!!!!!!!!!!!!!!!!\n");
PROCESS_MESSAGE_END;

PROCESS_MESSAGE_START(PVEAttackBegin)
printf("PVEAttackBegin begin !!!!!!!!!!!!!!!!!!!!!!!!!!!\n");
unsigned int t0 = GameLogic::LogicTimer::GetTimeMs();
BattleResult ret = FightCalculator::Instance()->CalculateBattleResult(*rst);
CAcceptorEx* pAcceptor = static_cast<CAcceptorEx*>(pContext);
std::cout << "pContext !!!!!!!!!!!!!!!!!!!!" << endl;
if(pAcceptor)
{
    rpc::PVEAttackBegin ab = *rst;
    int goldCount,foodCount,diamondCount;
    if (ab.mutable_stage()->has_currentgold())
    {
        goldCount = ab.mutable_stage()->currentgold();
        foodCount = ab.mutable_stage()->currentfood();
        diamondCount = ab.mutable_stage()->currentdiamond();
#ifdef _SUPER_DEBUG_
        printf("!!!!!!!!!!!!!!!!!!!!pve get gold:%d\n",goldCount);
        printf("!!!!!!!!!!!!!!!!!!!!pve get food:%d\n",foodCount);
#endif
    }
    else
    {
        int id = ab.mutable_stage()->stageid();
        char idstr[32] = "";
        sprintf(idstr,"%d",id);
        goldCount = atoi(ConfigManager::Instance()->GetAttribute("misson.config",
                                                                 idstr, 1, "GoldStorage").c_str());
        foodCount = atoi(ConfigManager::Instance()->GetAttribute("misson.config",
                                                                 idstr, 1, "FoodStorage").c_str());
        diamondCount = 0;
    }
    int remainGold = goldCount - ret._goldStolen;
#ifdef _SUPER_DEBUG_
    printf("111remainGold: %d\n", remainGold);
#endif
    if(remainGold < 0)
        remainGold = 0;
    int remainFood = foodCount - ret._foodStolen;
#ifdef _SUPER_DEBUG_
    printf("111remainFood: %d\n", remainFood);
#endif
    if(remainFood < 0)
        remainFood = 0;
    rpc::PVEAttackEnd attackEnd;
    attackEnd.set_playerid(ret._playerlid);
    attackEnd.set_goldstolen(ret._goldStolen);
    attackEnd.set_foodstolen(ret._foodStolen);
    
    attackEnd.mutable_stage()->set_stageid(rst->mutable_stage()->stageid());
    attackEnd.mutable_stage()->set_stars(ret._stars);
    attackEnd.mutable_stage()->set_currentgold(remainGold);
    attackEnd.mutable_stage()->set_currentfood(remainFood);
    attackEnd.mutable_stage()->set_currentdiamond(0);
    attackEnd.set_exp(ret._exp);
    
    rpc::Request req;
    req.set_method("CNServer.PVEAttackEnd");
    std::string data;
    attackEnd.SerializeToString(&data);
    req.set_serialized_request(data);
    data.clear();
    req.SerializeToString(&data);
    
    std::string output;
    snappy::Compress(data.data(), data.size(), &output);
    
    unsigned int size = output.size();
    printf("pve send size of size, %ld, %d", uint32(sizeof(size)), size);
    PipeResult nRet = pAcceptor->Send(&size, sizeof(size));
    if(ePR_OK != nRet)
       printf("pve send head error !!!!!!!!!!, %d", nRet);
    
    nRet = pAcceptor->Send(output.c_str(), size);
    if(ePR_OK != nRet)
        printf("pve send content error !!!!!!!!!!, %d", nRet);
#ifdef  _SUPER_DEBUG_
    printf("222remainFood: %d\n", remainFood);
    printf("222remainGold: %d\n", remainGold);
#endif
    printf("*********************pve get gold:%d\n",ret._goldStolen);
    printf("*********************pve get food:%d\n",ret._foodStolen);
    printf("*********************pve get star:%d\n",ret._stars);
    printf("*********************pve get exp: %d\n",ret._exp);
}
unsigned int t1 = GameLogic::LogicTimer::GetTimeMs();
unsigned int deltaTime = t1 - t0;
printf("---pve battle time cost: %d 毫秒\n",deltaTime);
PROCESS_MESSAGE_END;