//
//  ClientConnectionMgr.h
//  dpsg
//
//  Created by chopdown on 13-1-14.
//
//

#ifndef dpsg_ClientConnectionMgr_h
#define dpsg_ClientConnectionMgr_h

#include "Types.h"
#include "NetWork.h"
#include <google/protobuf/message.h>
#include "CRobot.h"

class CConnectionMgr: public Net::CNetWork {
private:
    CConnectionMgr();
    virtual ~CConnectionMgr();
    
public:
    GameHub::uint32 ConnectToGame(const char* sIp, GameHub::uint16 uPort);
    void DisconnectFromGame(GameHub::uint32 connectorID);
    
    void Test();
    
    //connect to chat
    GameHub::uint32 ConnectToChat(const char* pIp, GameHub::uint16 uPort);
    static void OnChatDisconnected(Net::CConnectorEx*		pConnectorEx);
    static void OnChatConnectFailed(Net::CConnectorEx*		pConnectorEx);
    static void OnChatConnectted(Net::CConnectorEx*			pConnectorEx);
    static void OnChatSomeDataSend(Net::CConnectorEx*	pConnectorEx);
    static void OnChatSomeDataRecv(Net::CConnectorEx*	pConnectorEx);
    static void OnChatPingServer(Net::CConnectorEx*	pConnectorEx);
    
    
    static CConnectionMgr& GetSingleton();
    static void ReleaseSingleton();

    static void ConnectorExFuncOnDisconnected(Net::CConnectorEx*		pConnectorEx);
    static void ConnectorExFuncOnConnectFailed(Net::CConnectorEx*		pConnectorEx);
    static void ConnectorExFuncOnConnectted(Net::CConnectorEx*			pConnectorEx);
    static void ConnectorExFuncOnSomeDataSend(Net::CConnectorEx*	pConnectorEx);
    static void ConnectorExFuncOnSomeDataRecv(Net::CConnectorEx*	pConnectorEx);

    static void ConnectorExFuncOnPingServer(Net::CConnectorEx*	pConnectorEx);
    
    
    void Call(GameHub::uint32 connectorID, const char* sCmd, google::protobuf::Message* msg);
    
    void TryConnectToCNS(Net::CConnectorEx* pConnectorEx, std::string gatekey, std::string ip, GameHub::uint32 port);
    
    CRobot* GetRobot(GameHub::uint32 connectorID);
    std::map<GameHub::uint32, CRobot*> m_RobotMap;
    
    //check the robot
    void CheckRobot(void);
    
private:
    static CConnectionMgr* s_pConectionMgr;
};

#endif
