//
//  ClientConnectionMgr.cpp
//  dpsg
//
//  Created by chopdown on 13-1-14.
//
//

#include "CConnectionMgr.h"
#include "Log.h"
#include "ConnectorEx.h"
#include "CRobotDispacher.h"
#include "snappy.h"

#include "Timer.h"
using namespace GameHub;

CConnectionMgr* CConnectionMgr::s_pConectionMgr = NULL;


CConnectionMgr& CConnectionMgr::GetSingleton()
{
    if(NULL == s_pConectionMgr)
    {
        s_pConectionMgr = GH_NEW_T(CConnectionMgr, MEMCATEGORY_GENERAL, "")();
    }
    return * s_pConectionMgr;
}

void CConnectionMgr::ReleaseSingleton()
{
    SAFE_RELEASE(s_pConectionMgr);
}

CConnectionMgr::CConnectionMgr()
{}

CConnectionMgr::~CConnectionMgr()
{
    std::map<GameHub::uint32, CRobot*>::iterator it = m_RobotMap.begin();
    while (it != m_RobotMap.end())
    {
        delete it->second;
        ++it;
    }
    m_RobotMap.clear();
}


void CConnectionMgr::Call(GameHub::uint32 connectorID, const char* sCmd, google::protobuf::Message* msg){
    rpc::Request req;
    req.set_method(sCmd);
    std::string data;
    msg->SerializeToString(&data);
    req.set_serialized_request(data);
    data.clear();
    req.SerializeToString(&data);
    
    std::string output;
    snappy::Compress(data.data(), data.size(), &output);
    
    int32 size = output.size();
    
    this->ConnectorExSendData(connectorID, &size, sizeof(size));
    this->ConnectorExSendData(connectorID, output.c_str(), size);
}

GameHub::uint32 CConnectionMgr::ConnectToGame(const char* sIp, GameHub::uint16 uPort){
    return this->Connect(sIp, uPort, &CConnectionMgr::ConnectorExFuncOnDisconnected,
                  &CConnectionMgr::ConnectorExFuncOnConnectFailed,
                  &CConnectionMgr::ConnectorExFuncOnConnectted,
                  &CConnectionMgr::ConnectorExFuncOnSomeDataSend,
                  &CConnectionMgr::ConnectorExFuncOnSomeDataRecv,
                  &CConnectionMgr::ConnectorExFuncOnPingServer);
}

void CConnectionMgr::Test(){
    GH_INFO("Test\n");
}
void CConnectionMgr::DisconnectFromGame(GameHub::uint32 connectorID)
{
    if(CConnectionMgr::GetSingleton().ShutDownConnectorEx(connectorID))
    {
        GetRobot(connectorID)->m_eConnectState = CRobot::CS_None;
        GetRobot(connectorID)->m_GameState = CRobot::GS_None;
        GetRobot(connectorID)->m_BattleTickCount = -1;
        m_RobotMap.erase(connectorID);
        LogicAvatarManager::Instance()->DestroyAvatar(connectorID);
        GH_INFO("DisconnectFromGame --- RobotCount:%d", CConnectionMgr::GetSingleton().m_RobotMap.size());
    }
}
void CConnectionMgr::ConnectorExFuncOnDisconnected(Net::CConnectorEx*	pConnectorEx)
{
    GH_INFO("Server Disconnected --- RobotCount:%d", CConnectionMgr::GetSingleton().m_RobotMap.size());
    CRobot* r = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
    CConnectionMgr::GetSingleton().m_RobotMap.erase(pConnectorEx->GetId());
    LogicAvatarManager::Instance()->DestroyAvatar(pConnectorEx->GetId());
    r->m_bLogin = false;
    r->m_eConnectState = CRobot::CS_Reconnect;
    r->m_GameState = CRobot::GS_None;
    r->m_BattleTickCount = -1;
}

void CConnectionMgr::ConnectorExFuncOnConnectFailed(Net::CConnectorEx* pConnectorEx)
{
    CConnectionMgr::GetSingleton().DisconnectFromGame(pConnectorEx->GetId());
	GH_INFO("Connect Server failed, check ip and port --- RobotCount:%d", CConnectionMgr::GetSingleton().m_RobotMap.size());
}
void CConnectionMgr::TryConnectToCNS(Net::CConnectorEx* pConnectorEx, std::string gatekey, std::string ip, GameHub::uint32 port)
{
    CRobot* r = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
    if (r->m_eConnectState == CRobot::CS_Gate)
    {
        DisconnectFromGame(pConnectorEx->GetId());
        
        //connect to cns
        r->m_connectorID = CConnectionMgr::GetSingleton().ConnectToGame(ip.c_str(),port);
        r->m_eConnectState = CRobot::CS_Gate;
        r->m_GateKey = gatekey;
        CConnectionMgr::GetSingleton().m_RobotMap[r->m_connectorID] = r;
        
//        CRobot* robot = iNew CRobot(CConnectionMgr::GetSingleton().ConnectToGame(ip.c_str(),port), uid);
//        robot->m_eConnectState = CRobot::CS_Gate;
//        robot->m_GateKey = gatekey;
//        CConnectionMgr::GetSingleton().m_RobotMap[robot->m_connectorID] = robot;
    }
}
void CConnectionMgr::ConnectorExFuncOnConnectted(Net::CConnectorEx* pConnectorEx)
{
    GH_INFO("OnConnectted --- RobotCount:%d", CConnectionMgr::GetSingleton().m_RobotMap.size());
	const uint32 MAX_SIZE = 1024 * 1024 * 5;
	pConnectorEx->SetMaxRecvBufSize(MAX_SIZE);
	pConnectorEx->SetMaxSendBufSize(MAX_SIZE);
    
    CRobot* r = CConnectionMgr::GetSingleton().GetRobot(pConnectorEx->GetId());
    GH_ASSERT(r->m_eConnectState != CRobot::CS_CNS);
    if (r->m_eConnectState == CRobot::CS_None)
    {
        r->m_eConnectState = CRobot::CS_Gate;
    }
    else if(r->m_eConnectState == CRobot::CS_Gate)
    {
        r->m_eConnectState = CRobot::CS_CNS;
        r->Login();
    }
}

void CConnectionMgr::ConnectorExFuncOnSomeDataSend(Net::CConnectorEx* pConnectorEx)
{
    
}

void CConnectionMgr::ConnectorExFuncOnSomeDataRecv(Net::CConnectorEx* pConnectorEx)
{    
    static CRobotDispacher oDispacher;
  
	uint32 uProccessed=0;
	oDispacher.LoopDispatch(pConnectorEx->GetRecvData(), pConnectorEx->GetRecvDataSize(), uProccessed, pConnectorEx);
	pConnectorEx->PopRecvData(uProccessed);
}

void CConnectionMgr::ConnectorExFuncOnPingServer(Net::CConnectorEx* pConnectorEx)
{
    static rpc::Ping ping;
    CConnectionMgr::GetSingleton().Call(pConnectorEx->GetId(), "CNServer.Ping", &ping);
}

CRobot* CConnectionMgr::GetRobot(GameHub::uint32 connectorID)
{
    std::map<GameHub::uint32, CRobot*>::iterator it = m_RobotMap.find(connectorID);
    if(it != m_RobotMap.end())
        return it->second;
    return NULL;
}

//connect to chat server
GameHub::uint32 CConnectionMgr::ConnectToChat(const char* pIp, GameHub::uint16 uPort)
{
    return this->Connect(pIp, uPort, &CConnectionMgr::OnChatDisconnected,
                         &CConnectionMgr::OnChatConnectFailed,
                         &CConnectionMgr::OnChatConnectted,
                         &CConnectionMgr::OnChatSomeDataSend,
                         &CConnectionMgr::OnChatSomeDataRecv,
                         &CConnectionMgr::OnChatPingServer);
}

void CConnectionMgr::OnChatDisconnected(Net::CConnectorEx*	pConnectorEx)
{
    GH_INFO("断开与chat server的连接");
}
void CConnectionMgr::OnChatConnectFailed(Net::CConnectorEx*	pConnectorEx)
{
    GH_INFO("与chat server的连接失败");
}
void CConnectionMgr::OnChatConnectted(Net::CConnectorEx*	pConnectorEx)
{
    GH_INFO("成功连上chat server");
}
void CConnectionMgr::OnChatSomeDataSend(Net::CConnectorEx*	pConnectorEx)
{
}
void CConnectionMgr::OnChatSomeDataRecv(Net::CConnectorEx*	pConnectorEx)
{
}
void CConnectionMgr::OnChatPingServer(Net::CConnectorEx*	pConnectorEx)
{
    
}

//check the robot, if the count of the robot less than an number, will generate some robot
void CConnectionMgr::CheckRobot()
{
    
}
