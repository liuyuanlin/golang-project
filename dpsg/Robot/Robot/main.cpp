//
//  main.cpp
//  Robot
//
//  Created by PU on 13-4-15.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#include <iostream>

#include "BaseTickMgr.h"
#include "CConnectionMgr.h"
#include "NetWork.h"
#include "ConfigManager.h"

#include "FileStream.h"
#include "GameLogicDefinitions.h"
#include "LogicInitializer.h"
#include "RUserData.h"
#include "CRBuilding.h"
//add by wyc 2013-12-25 add village check 
#include "CVillageCheck.h"


using namespace GameLogic;
using namespace Net;

class a : public GameHub::CTick
{
public:
    a() {};
};

vector<string> _splitstring(string input,const char* deli,bool ignoreEmpty = true)
{
	int nend=0;
	int nbegin=0;
	vector<string> outlist;
	while(nend != -1)
	{
		nend = input.find(deli, nbegin);
		if(nend == -1)
		{
			string str = input.substr(nbegin, input.length()-nbegin);
			if(ignoreEmpty)
			{
				if(str != "")
					outlist.push_back(str);
			}
			else
				outlist.push_back(str);
		}
		else
		{
			string str = input.substr(nbegin, nend-nbegin);
			if(ignoreEmpty)
			{
				if(str != "")
					outlist.push_back(str);
			}
			else
				outlist.push_back(str);
		}
		nbegin = nend + strlen(deli);
	}
	return outlist;
}
void ReadConfigFile(string config,string binPath)
{
    GameHub::FileStream file;
    string filename = binPath + "/../designer/" + config;
    file.Open(filename.c_str());
    unsigned int size = file.GetFileSize();
    if (size)
    {
        char* buffer = GH_NEW char[size];
        file.GetBuffer(buffer);
        ConfigManager::Instance()->LoadConfigFromBuffer(config, buffer, size);
        GH_DELETE[] buffer;
    }
}

int main(int argc, const char * argv[])
{
    a b;
    GameHub::GetTickMgr()->RegisterTick(333, &b);
    
    char s[512];
    getcwd(s, 512);
    
    string binPath = s;
    ReadConfigFile(BUILDINGCONFIG,binPath);
    ReadConfigFile(CHARACTERCONFIG,binPath);
    ReadConfigFile(PROJECTILE,binPath);
    ReadConfigFile(TOWNHALLCONFIG, binPath);
    
    RUserData::Instance()->m_Path = binPath + "/RobotCfg";
    RUserData::Instance()->Init();
    
    //Is just for chat server test
    if (RUserData::Instance()->IsJustForChat())
    {
        int nConfigCount = RUserData::Instance()->m_uChatCount;
        int nRobotCount = 0;
        while (nConfigCount > 0)
        {
            GH_INFO("Chat server 配置数量=%d", nConfigCount);
            nRobotCount = RUserData::Instance()->m_vecChatRobotCount[nConfigCount-1];
            GH_INFO("该条配置机器人数量=%d, IP=%s, Port=%d", nRobotCount, RUserData::Instance()->m_vecChatIp[nConfigCount-1].c_str(), RUserData::Instance()->m_vecChatPort[nConfigCount-1]);
            while (nRobotCount > 0)
            {
                CConnectionMgr::GetSingleton().ConnectToChat(RUserData::Instance()->m_vecChatIp[nConfigCount-1].c_str(),
                                                             RUserData::Instance()->m_vecChatPort[nConfigCount-1]);
                usleep(30 * 1000);
                
                --nRobotCount;
            }
            --nConfigCount;
        }
        
        GH_INFO("连接到Chat server的操作执行完毕");
    }
    else
    {
        CRBuilding::Instance()->Init();
        
        //检测所有村庄配置的合法性
        if (!CVillageCheck::Instance()->CheckVillage())
        {
            GH_INFO("请策划检查以上错误配置\n");
            return 0;
        }
        
        //TODO 村庄分配
        int nVillageCount = RUserData::Instance()->m_uConfigCount;
        GH_INFO("当前压入配置数量=%d", nVillageCount);
        int nIndex = 0;
        int nVillageLevel = 0;
        int nExistCount = 0;
        int nRobotCount = 0;
        int nRobotConfig = 0;
        
        while (nIndex < nVillageCount) {
            //当前批次机器人本数
            nVillageLevel = RUserData::Instance()->m_vecLevel[nIndex];
            GH_INFO("当前批次机器人本数=%d", nVillageLevel);
            //当前本数配置数量
            nExistCount = CRBuilding::Instance()->GetCountOnLevel(nVillageLevel);
            if (0 == nExistCount) {
                //GH_INFO("策划配置有问题了，没有本数=%d的配置，请找策划配置", nVillageLevel);
                //return 0;
                ++nIndex;
                continue;
            }
            //当前批次压入机器人数量
            nRobotCount = RUserData::Instance()->m_vecRobotCount[nIndex];
            
            //压入当前配置机器人
            while (nRobotCount) {
                --nExistCount;
                if (nExistCount < 0)
                {
                    nExistCount = CRBuilding::Instance()->GetCountOnLevel(nVillageLevel) - 1;
                }
                //获取具体配置表
                nRobotConfig = CRBuilding::Instance()->GetIndexOnLevel(nVillageLevel, nExistCount);
                
                CRobot* robot = iNew CRobot(CConnectionMgr::GetSingleton().ConnectToGame(RUserData::Instance()->m_vecIp[nIndex].c_str(), RUserData::Instance()->m_vecPort[nIndex]), nRobotConfig, nIndex, RUserData::Instance()->GetUDID(nRobotCount - RUserData::Instance()->m_MaxRobotCount));
                CConnectionMgr::GetSingleton().m_RobotMap[robot->m_connectorID] = robot;
                GH_INFO("机器人=%d完成了初始化", nRobotCount);
                --nRobotCount;
            }
            
            ++nIndex;
        }
    }
    
    TimeValue tv;
    LogicTimer::GetTime(&tv);
    unsigned long long int lasttime = tv._Seconds*1000000.0 + tv._MicroSeconds;
    while (1)
    {
        CConnectionMgr::GetSingleton().DispatchEvents();
        
//        //机器人压入成功
//        if (CConnectionMgr::GetSingleton().m_RobotMap.size() == 0 && !RUserData::Instance()->IsJustForChat())
//        {
//            GH_INFO("机器人已经全部压入成功！！！");
//            return 0;
//        }
        
        LogicTimer::GetTime(&tv);
        unsigned long long int time = tv._Seconds*1000000.0 + tv._MicroSeconds;
        unsigned long long int deltaTime = time - lasttime;
        double de = deltaTime / 1000000.0;
        LogicTickManager::Instance()->Tick(de);
        lasttime = time;
        
        usleep(30 * 1000);
    }

    std::cout << "Hello, World!\n";
    return 0;
}

