//
//  ClientDispacher.cpp
//  dpsg
//
//  Created by chopdown on 13-1-14.
//
//

#include "CRobotDispacher.h"
#include "CConnectionMgr.h"
#include "Log.h"
#include "TimeSystem.h"
#include "NetWork.h"
#include "ConnectorEx.h"
#include <sstream>

#include "LogicVillage.h"

using namespace std;
Use_NS_GameLogic

#define REGISTER_CALLBACK_IMP(msg) this->registerMessageCallback<rpc::msg>(Net::delegate::from_method<CRobotDispacher, &CRobotDispacher::OnSync##msg>(this))

#define PROCESS_MESSAGE_START(msg) void CRobotDispacher::OnSync##msg(google::protobuf::Message* m, void* pContext){ rpc::msg* rst = static_cast<rpc::msg*>(m);
#define PROCESS_MESSAGE_END }

CRobotDispacher::CRobotDispacher(){
    REGISTER_CALLBACK_IMP(LoginResult);
    REGISTER_CALLBACK_IMP(PlayerInfo);
    REGISTER_CALLBACK_IMP(VillageInfo);
    REGISTER_CALLBACK_IMP(PingResult);
    REGISTER_CALLBACK_IMP(RpcErrorResponse);
    REGISTER_CALLBACK_IMP(MatchPlayer);
    REGISTER_CALLBACK_IMP(MatchPlayerResult);
    REGISTER_CALLBACK_IMP(LoginCnsInfo);
    REGISTER_CALLBACK_IMP(SyncError);
    REGISTER_CALLBACK_IMP(UpdatePlayerInfo);
    REGISTER_CALLBACK_IMP(Msg);
}

PROCESS_MESSAGE_START(LoginResult)
    GH_INFO("Login Return Code %d\n", rst->result());
    switch (rst->result())
    {
        case rpc::LoginResult::OK:
        {
            TimeSystem::Instance()->SynchronizeWithServer(rst->server_time());
            Net::CConnectorEx* pConnectorEx = static_cast<Net::CConnectorEx*>(pContext);
            CRobot* robot = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
            robot->m_bLogin = true;
        }
            break;
        case rpc::LoginResult::SERVERERROR:
        {
        }
        default:
            break;
    }
PROCESS_MESSAGE_END;


PROCESS_MESSAGE_START(RpcErrorResponse)
    std::stringstream title;
    title << "call rpc failed:" << rst->method();
PROCESS_MESSAGE_END;

PROCESS_MESSAGE_START(SyncError)
if (rst->has_text()) {
    GH_INFO("SyncError : %s", rst->text().c_str());
}
PROCESS_MESSAGE_END;

PROCESS_MESSAGE_START(PlayerInfo)
    GH_INFO("PlayerInfo Return Code %s\n", rst->base().uid().c_str());

    Net::CConnectorEx* pConnectorEx = static_cast<Net::CConnectorEx*>(pContext);
    CRobot* robot = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
    robot->SetUDID(rst->base().uid());
    robot->m_OriginYuanBao = rst->extra().diamonds();
    robot->m_OriginWuHun = rst->extra().wuhun();
PROCESS_MESSAGE_END;

PROCESS_MESSAGE_START(VillageInfo)
    GH_INFO("VillageInfo Return hp %d\n", rst->center().hp());
    Net::CConnectorEx* pConnectorEx = static_cast<Net::CConnectorEx*>(pContext);
    CRobot* robot = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());

    if (RUserData::Instance()->IsOpenBuild())
    {
        robot->CreateVillage(*rst, false);
        robot->m_BattleTickCount = -1;
        robot->m_GameState = CRobot::GS_Edit;
    
        //取名
        if (!robot->GetIsNewRobot())     robot->AskForName();
    }
    else
        robot->LogOut();
PROCESS_MESSAGE_END;


PROCESS_MESSAGE_START(MatchPlayer)
    //服务器返回的战斗匹配信息
    GH_INFO("MatchPlayer Return Code %d\n", rst->act());
    Net::CConnectorEx* pConnectorEx = static_cast<Net::CConnectorEx*>(pContext);
    CRobot* robot = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
    robot->m_MatchPlayerInfo = *rst;
    robot->CreateVillage(rst->v(), true);
    GH_INFO("Charcter count = %d", rst->own_char_size());
    robot->m_GameState = CRobot::GS_Battle;
    robot->SendArmy();
PROCESS_MESSAGE_END;


PROCESS_MESSAGE_START(MatchPlayerResult)
    GH_INFO("MatchPlayerResult Return Code %d\n", rst->result());
    Net::CConnectorEx* pConnectorEx = static_cast<Net::CConnectorEx*>(pContext);
    CRobot* robot = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
    switch (rst->result())
    {
        case rpc::MatchPlayerResult::OK:
        {
            GH_INFO("匹配成功，机器人出兵！！！！");
        }
            break;
        case rpc::MatchPlayerResult::NOTEXIST:
        {
            cout<<"NotExist"<<endl;
        }
            break;
        case rpc::MatchPlayerResult::SERVERERROR:
        {
            cout<<"ServerError"<<endl;
        }
            break;
        case rpc::MatchPlayerResult::ISONFIRE:
        {
            cout<<"IsOnFire"<<endl;
        }
            break;
        case rpc::MatchPlayerResult::MATCHNOTHING:
        {
            GH_INFO("都是渣渣，根本就匹配不到,下线休息");
            robot->LogOut();
        }
            break;
        default:
            break;
    }
}

PROCESS_MESSAGE_START(PingResult)
    TimeSystem::Instance()->SynchronizeWithServer(rst->server_time());
PROCESS_MESSAGE_END;

vector<string> splitstring(string input,const char* deli,bool ignoreEmpty = true)
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

PROCESS_MESSAGE_START(LoginCnsInfo)
    std::string ip = rst->cnsip();
    GH_INFO("得到cnsIP=%s", ip.c_str());
    if(ip.length())
    {
        vector<std::string> parts = splitstring(ip, ":");
        Net::CConnectorEx* pConnectorEx = static_cast<Net::CConnectorEx*>(pContext);
        CConnectionMgr::GetSingleton().TryConnectToCNS(pConnectorEx, rst->gsinfo(),parts[0].c_str(),atoi(parts[1].c_str()));
    }
PROCESS_MESSAGE_END;

//add by wyc 2013-12-26   取名成功
PROCESS_MESSAGE_START(UpdatePlayerInfo)
    Net::CConnectorEx* pConnectorEx = static_cast<Net::CConnectorEx*>(pContext);
    CRobot* robot = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
    if (robot->m_bNameOK && RUserData::Instance()->m_bOpenFight)
    {
        robot->m_GameState = CRobot::GS_Edit;
        robot->SetDoFight(true);
    }
    if (rst->name() != "")
    {
        string sName = rst->name();
        GH_INFO("取名成功,名字=%s", sName.c_str());
        robot->m_bNameOK = true;
    }
PROCESS_MESSAGE_END;

//消息
PROCESS_MESSAGE_START(Msg)
    Net::CConnectorEx* pConnectorEx = static_cast<Net::CConnectorEx*>(pContext);
    CRobot* robot = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
    //重名了，取名失败
    string sName = rst->code();
    std::cout << "收到服务器消息：" << rst->code() << std::endl;
    if (rst->code() == "TID_SAME_NAME")
    {
        robot->m_bNameOK = false;
        robot->AskForName();
    }
PROCESS_MESSAGE_END;








